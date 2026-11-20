// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/collector/config/configopaque"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/extensionauth"
	"go.uber.org/zap"
)

var (
	_ extension.Extension  = (*splunkAuth)(nil)
	_ extensionauth.Server = (*splunkAuth)(nil)
)

type Key string

const ContextKey Key = "hec"

// tokenSource supplies complete snapshots of tokens and their authorization metadata.
// Load must honor context cancellation and return an error for incomplete results.
// Sources can be shared by multiple extension instances and must support concurrent loads.
type tokenSource interface {
	Load(context.Context) ([]HecTokenConfig, error)
}

type tokenSnapshot struct {
	loadedAt time.Time
	tokens   []HecTokenConfig
}

type splunkAuth struct {
	logger   *zap.Logger
	source   tokenSource
	snapshot atomic.Pointer[tokenSnapshot]
	cancel   context.CancelFunc
	done     chan struct{}
	header   string
	scheme   string
	config   Config

	staleLogged atomic.Bool
}

const (
	defaultHeader = "Authorization"
	defaultScheme = "Splunk"
)

func newSplunkAuth(cfg *Config, source tokenSource, logger *zap.Logger) *splunkAuth {
	return &splunkAuth{
		header: defaultHeader,
		scheme: defaultScheme,
		logger: logger,
		config: *cfg,
		source: source,
	}
}

func (b *splunkAuth) Start(ctx context.Context, _ component.Host) error {
	if err := b.refresh(ctx); err != nil {
		return fmt.Errorf("initial HEC token load: %w", err)
	}
	// The Start context only covers startup. Shutdown owns the worker lifetime.
	workerCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	b.cancel = cancel
	b.done = make(chan struct{})
	go func() {
		defer close(b.done)
		ticker := time.NewTicker(b.config.RefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				if err := b.refresh(workerCtx); err != nil && workerCtx.Err() == nil {
					age := time.Since(b.snapshot.Load().loadedAt)
					if age >= b.config.MaxStaleness {
						b.logger.Error("HEC token refresh failed; snapshot exceeded max_staleness and authentication is being rejected", zap.Error(err), zap.Duration("snapshot_age", age))
					} else {
						b.logger.Warn("HEC token refresh failed; retaining previous snapshot until max_staleness", zap.Error(err), zap.Duration("snapshot_age", age))
					}
				}
			}
		}
	}()
	return nil
}

func (b *splunkAuth) refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, b.config.RefreshTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	// Measure age from the start of retrieval so slow pagination does not extend
	// the lifetime of authorization data already read from the server.
	loadedAt := time.Now()
	tokens, err := b.source.Load(ctx)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	values := make([]HecTokenConfig, len(tokens))
	for i, token := range tokens {
		values[i] = token
		values[i].Token = configopaque.String(b.scheme + " " + string(token.Token))
		values[i].AllowedIndexes = slices.Clone(token.AllowedIndexes)
	}
	b.snapshot.Store(&tokenSnapshot{tokens: values, loadedAt: loadedAt})
	if b.staleLogged.Swap(false) {
		b.logger.Info("HEC token refresh recovered; authentication restored")
	}
	return nil
}

func (b *splunkAuth) Shutdown(ctx context.Context) error {
	if b.cancel == nil {
		return nil
	}
	b.cancel()
	select {
	case <-b.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *splunkAuth) Authenticate(ctx context.Context, headers map[string][]string) (context.Context, error) {
	auth, ok := headers[strings.ToLower(b.header)]
	if !ok {
		auth, ok = headers[b.header]
	}
	if !ok || len(auth) == 0 {
		return ctx, fmt.Errorf("missing or empty authorization header: %s", b.header)
	}
	token := auth[0] // Extract token from authorization header
	snapshot := b.snapshot.Load()
	if snapshot == nil || time.Since(snapshot.loadedAt) >= b.config.MaxStaleness {
		// Log once per stale episode so an outage under load does not flood the log.
		if b.staleLogged.CompareAndSwap(false, true) {
			b.logger.Error("HEC token snapshot is stale; rejecting all authentication until a refresh succeeds", zap.Duration("max_staleness", b.config.MaxStaleness))
		}
		return ctx, errors.New("HEC token snapshot is unavailable or stale")
	}
	for _, expectedToken := range snapshot.tokens {
		if subtle.ConstantTimeCompare([]byte(expectedToken.Token), []byte(token)) == 1 {
			if expectedToken.Disabled {
				return ctx, errors.New("HEC token is disabled")
			}
			// Do not expose the cached slice to downstream consumers.
			expectedToken.AllowedIndexes = slices.Clone(expectedToken.AllowedIndexes)
			return context.WithValue(ctx, ContextKey, expectedToken), nil
		}
	}
	return ctx, errors.New("scheme or token does not match")
}
