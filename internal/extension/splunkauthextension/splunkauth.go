// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strings"

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

type splunkAuth struct {
	logger *zap.Logger
	header string
	scheme string
	tokens []HecTokenConfig
}

const (
	defaultHeader = "Authorization"
	defaultScheme = "Splunk"
)

func newSplunkAuth(cfg *Config, logger *zap.Logger) *splunkAuth {
	a := &splunkAuth{
		header: defaultHeader,
		scheme: defaultScheme,
		logger: logger,
	}
	a.setAuthorizationValues(cfg.Tokens) // Store tokens
	return a
}

func (b *splunkAuth) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (b *splunkAuth) setAuthorizationValues(tokens []HecTokenConfig) {
	values := make([]HecTokenConfig, len(tokens))
	for i, token := range tokens {
		if b.scheme != "" {
			values[i].Token = configopaque.String(b.scheme + " " + string(token.Token))
		} else {
			values[i] = token
		}
		values[i].AllowedIndexes = token.AllowedIndexes
		values[i].DefaultIndex = token.DefaultIndex
	}
	b.tokens = values
}

func (b *splunkAuth) Shutdown(_ context.Context) error {
	return nil
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
	expectedTokens := b.tokens
	for _, expectedToken := range expectedTokens {
		if subtle.ConstantTimeCompare([]byte(expectedToken.Token), []byte(token)) == 1 {
			return context.WithValue(ctx, ContextKey, expectedToken), nil // Authentication successful, token is valid
		}
	}
	return ctx, fmt.Errorf("scheme or token does not match: %s", token) // Token is invalid
}
