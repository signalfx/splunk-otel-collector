// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestAuthenticateDisabledToken(t *testing.T) {
	disabled := HecTokenConfig{Token: configopaque.String("disabled-token"), Disabled: true}
	enabled := HecTokenConfig{
		Token:          configopaque.String("enabled-token"),
		DefaultIndex:   "main",
		AllowedIndexes: []string{"main", "other"},
	}
	a := newTestAuthenticator(t, sourceFunc(func(context.Context) ([]HecTokenConfig, error) {
		return []HecTokenConfig{disabled, enabled}, nil
	}), nil)
	for _, header := range []string{"Authorization", "authorization"} {
		t.Run(header, func(t *testing.T) {
			ctx := t.Context()
			got, err := a.Authenticate(ctx, map[string][]string{header: {"Splunk disabled-token"}})
			require.EqualError(t, err, "HEC token is disabled")
			require.Equal(t, ctx, got)
			require.Nil(t, got.Value(ContextKey))

			got, err = a.Authenticate(ctx, map[string][]string{header: {"Splunk enabled-token"}})
			require.NoError(t, err)
			expected := enabled
			expected.Token = configopaque.String("Splunk enabled-token")
			require.Equal(t, expected, got.Value(ContextKey))
		})
	}
}

type sourceFunc func(context.Context) ([]HecTokenConfig, error)

func (f sourceFunc) Load(ctx context.Context) ([]HecTokenConfig, error) { return f(ctx) }

func newTestAuthenticator(t *testing.T, source tokenSource, configure func(*Config)) *splunkAuth {
	t.Helper()
	cfg := createDefaultConfig().(*Config)
	if configure != nil {
		configure(cfg)
	}
	a := newSplunkAuth(cfg, source, zap.NewNop())
	require.NoError(t, a.Start(t.Context(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, a.Shutdown(context.Background())) })
	return a
}

func authenticate(a *splunkAuth, token string) (context.Context, error) {
	return a.Authenticate(context.Background(), map[string][]string{"Authorization": {"Splunk " + token}})
}

func TestPeriodicTokenReplacement(t *testing.T) {
	var mu sync.Mutex
	tokens := []HecTokenConfig{{Token: "old", DefaultIndex: "main", AllowedIndexes: []string{"main"}}}
	var loadErr error
	a := newTestAuthenticator(t, sourceFunc(func(context.Context) ([]HecTokenConfig, error) {
		mu.Lock()
		defer mu.Unlock()
		return tokens, loadErr
	}), func(cfg *Config) { cfg.RefreshInterval = time.Millisecond })
	_, err := authenticate(a, "old")
	require.NoError(t, err)
	mu.Lock()
	tokens = []HecTokenConfig{{Token: "new", DefaultIndex: "other", AllowedIndexes: []string{"other"}}}
	mu.Unlock()
	require.Eventually(t, func() bool {
		ctx, authErr := authenticate(a, "new")
		return authErr == nil && ctx.Value(ContextKey).(HecTokenConfig).DefaultIndex == "other"
	}, time.Second, time.Millisecond)
	_, err = authenticate(a, "old")
	require.Error(t, err)

	mu.Lock()
	tokens = []HecTokenConfig{{Token: "new", Disabled: true}}
	mu.Unlock()
	require.Eventually(t, func() bool {
		_, authErr := authenticate(a, "new")
		return authErr != nil && authErr.Error() == "HEC token is disabled"
	}, time.Second, time.Millisecond)

	mu.Lock()
	tokens = []HecTokenConfig{{Token: "new", DefaultIndex: "changed", AllowedIndexes: []string{"changed"}}}
	mu.Unlock()
	require.Eventually(t, func() bool {
		ctx, authErr := authenticate(a, "new")
		return authErr == nil && ctx.Value(ContextKey).(HecTokenConfig).DefaultIndex == "changed" &&
			slices.Equal(ctx.Value(ContextKey).(HecTokenConfig).AllowedIndexes, []string{"changed"})
	}, time.Second, time.Millisecond)

	// A failed load must not publish the partial result it returned.
	mu.Lock()
	tokens = []HecTokenConfig{{Token: "partial"}}
	loadErr = errors.New("page two failed")
	mu.Unlock()
	require.Error(t, a.refresh(t.Context()))
	_, err = authenticate(a, "new")
	require.NoError(t, err)
	_, err = authenticate(a, "partial")
	require.Error(t, err)

	mu.Lock()
	tokens, loadErr = nil, nil
	mu.Unlock()
	require.Eventually(t, func() bool {
		snapshot := a.snapshot.Load()
		return snapshot != nil && len(snapshot.tokens) == 0
	}, time.Second, time.Millisecond)
	_, err = authenticate(a, "new")
	require.Error(t, err)
}

func TestSnapshotOwnershipAndStaleness(t *testing.T) {
	tokens := []HecTokenConfig{{Token: "valid", AllowedIndexes: []string{"main"}}}
	a := newTestAuthenticator(t, sourceFunc(func(context.Context) ([]HecTokenConfig, error) {
		return tokens, nil
	}), nil)
	tokens[0].AllowedIndexes[0] = "source mutation"
	ctx, err := authenticate(a, "valid")
	require.NoError(t, err)
	metadata := ctx.Value(ContextKey).(HecTokenConfig)
	require.Equal(t, []string{"main"}, metadata.AllowedIndexes)
	metadata.AllowedIndexes[0] = "consumer mutation"
	ctx, err = authenticate(a, "valid")
	require.NoError(t, err)
	require.Equal(t, []string{"main"}, ctx.Value(ContextKey).(HecTokenConfig).AllowedIndexes)
	snapshot := a.snapshot.Load()
	a.snapshot.Store(&tokenSnapshot{tokens: snapshot.tokens, loadedAt: time.Now().Add(-a.config.MaxStaleness)})
	_, err = authenticate(a, "valid")
	require.EqualError(t, err, "HEC token snapshot is unavailable or stale")
	require.NoError(t, a.refresh(t.Context()))
	_, err = authenticate(a, "valid")
	require.NoError(t, err)
}

func TestInitialLoadFailure(t *testing.T) {
	for _, timeout := range []bool{false, true} {
		t.Run(strconv.FormatBool(timeout), func(t *testing.T) {
			cfg := createDefaultConfig().(*Config)
			cfg.RefreshTimeout = time.Millisecond
			a := newSplunkAuth(cfg, sourceFunc(func(ctx context.Context) ([]HecTokenConfig, error) {
				if timeout {
					<-ctx.Done()
					return nil, ctx.Err()
				}
				return nil, errors.New("unavailable")
			}), zap.NewNop())
			require.Error(t, a.Start(t.Context(), componenttest.NewNopHost()))
			require.Nil(t, a.snapshot.Load())
			require.NoError(t, a.Shutdown(t.Context()))
			_, err := authenticate(a, "anything")
			require.Error(t, err)
		})
	}
}

func TestShutdownCancelsRefresh(t *testing.T) {
	var calls atomic.Int32
	refreshing := make(chan struct{})
	a := newTestAuthenticator(t, sourceFunc(func(ctx context.Context) ([]HecTokenConfig, error) {
		if calls.Add(1) == 1 {
			return nil, nil
		}
		close(refreshing)
		<-ctx.Done()
		return nil, ctx.Err()
	}), func(cfg *Config) { cfg.RefreshInterval = time.Millisecond })
	select {
	case <-refreshing:
	case <-time.After(time.Second):
		t.Fatal("refresh did not start")
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	require.NoError(t, a.Shutdown(ctx))
	require.EqualValues(t, 2, calls.Load())
}

func TestConcurrentAuthenticationAndRefresh(t *testing.T) {
	var generation atomic.Int32
	a := newTestAuthenticator(t, sourceFunc(func(context.Context) ([]HecTokenConfig, error) {
		index := strconv.Itoa(int(generation.Add(1)))
		return []HecTokenConfig{{Token: "valid", DefaultIndex: index, AllowedIndexes: []string{index}}}, nil
	}), func(cfg *Config) { cfg.RefreshInterval = time.Millisecond })
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 1000 {
				ctx, err := authenticate(a, "valid")
				if err != nil {
					t.Errorf("authentication failed: %v", err)
					return
				}
				token := ctx.Value(ContextKey).(HecTokenConfig)
				assert.Equal(t, []string{token.DefaultIndex}, token.AllowedIndexes)
				token.AllowedIndexes[0] = "mutation"
			}
		})
	}
	wg.Wait()
}

func TestRefreshTimeoutAndRecovery(t *testing.T) {
	var calls atomic.Int32
	timedOut := make(chan struct{})
	a := newTestAuthenticator(t, sourceFunc(func(ctx context.Context) ([]HecTokenConfig, error) {
		switch calls.Add(1) {
		case 1:
			return []HecTokenConfig{{Token: "old"}}, nil
		case 2:
			<-ctx.Done()
			close(timedOut)
			// Even a source that returns success after its deadline cannot publish.
			return []HecTokenConfig{{Token: "late"}}, nil
		default:
			return []HecTokenConfig{{Token: "new"}}, nil
		}
	}), func(cfg *Config) {
		cfg.RefreshInterval = 10 * time.Millisecond
		cfg.RefreshTimeout = time.Millisecond
	})
	select {
	case <-timedOut:
	case <-time.After(time.Second):
		t.Fatal("refresh did not time out")
	}
	require.Eventually(t, func() bool {
		_, err := authenticate(a, "new")
		return err == nil
	}, time.Second, time.Millisecond)
	_, err := authenticate(a, "late")
	require.Error(t, err)
	_, err = authenticate(a, "old")
	require.Error(t, err)
}

func TestWorkerOutlivesStartContext(t *testing.T) {
	var calls atomic.Int32
	source := sourceFunc(func(context.Context) ([]HecTokenConfig, error) {
		calls.Add(1)
		return nil, nil
	})
	cfg := createDefaultConfig().(*Config)
	cfg.RefreshInterval = time.Millisecond
	a := newSplunkAuth(cfg, source, zap.NewNop())
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	require.NoError(t, a.Start(ctx, componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, a.Shutdown(context.Background())) })
	cancel()
	require.Eventually(t, func() bool { return calls.Load() > 1 }, time.Second, time.Millisecond)
}

func TestStaleLoggedOncePerEpisode(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	cfg := createDefaultConfig().(*Config)
	a := newSplunkAuth(cfg, sourceFunc(func(context.Context) ([]HecTokenConfig, error) {
		return []HecTokenConfig{{Token: "valid"}}, nil
	}), zap.New(core))
	require.NoError(t, a.Start(t.Context(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, a.Shutdown(context.Background())) })

	expire := func() {
		a.snapshot.Store(&tokenSnapshot{tokens: a.snapshot.Load().tokens, loadedAt: time.Now().Add(-cfg.MaxStaleness)})
	}
	for range 3 {
		expire()
		_, err := authenticate(a, "valid")
		require.Error(t, err)
	}
	require.Equal(t, 1, logs.Len())

	require.NoError(t, a.refresh(t.Context()))
	expire()
	_, err := authenticate(a, "valid")
	require.Error(t, err)
	require.Equal(t, 2, logs.Len())
}
