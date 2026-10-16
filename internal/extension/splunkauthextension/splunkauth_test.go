// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.uber.org/zap"
)

func TestAuthenticateDisabledToken(t *testing.T) {
	disabled := HecTokenConfig{Token: configopaque.String("disabled-token"), Disabled: true}
	enabled := HecTokenConfig{
		Token:          configopaque.String("enabled-token"),
		DefaultIndex:   "main",
		AllowedIndexes: []string{"main", "other"},
	}
	a := newSplunkAuth(&Config{Tokens: []HecTokenConfig{disabled, enabled}}, zap.NewNop())
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
