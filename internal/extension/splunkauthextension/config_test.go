// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/extension/extensiontest"
)

func TestConfiguredSourceValidation(t *testing.T) {
	for _, endpoint := range []string{"splunk:8089", "ftp://splunk:8089", "https://", "https://user:secret@splunk:8089", "https://splunk:8089?token=secret", "https://splunk:8089#fragment"} {
		t.Run(endpoint, func(t *testing.T) {
			cfg := NewDefaultConfig()
			cfg.Endpoint, cfg.SessionKey = endpoint, "secret"
			require.Error(t, cfg.Validate())
		})
	}
	for _, values := range []map[string]any{
		{"endpoint": "https://splunk:8089"},
		{"session_key": "secret"},
		{"unknown_field": true},
		{"refresh_interval": "not a duration"},
	} {
		cfg := NewDefaultConfig()
		err := confmap.NewFromStringMap(values).Unmarshal(cfg)
		if err == nil {
			require.Error(t, cfg.Validate())
		}
	}
	secret := confmap.NewFromStringMap(map[string]any{"endpoint": "https://splunk:8089", "session_key": "secret"})
	cfg := NewDefaultConfig()
	require.NoError(t, secret.Unmarshal(cfg))
	require.NoError(t, cfg.Validate())
	redacted := confmap.New()
	require.NoError(t, redacted.Marshal(cfg))
	require.NotContains(t, redacted.ToStringMap()["session_key"], "secret")
}

func TestValidateRefreshWindow(t *testing.T) {
	const maxDuration = time.Duration(1<<63 - 1)
	for _, tc := range []struct {
		name      string
		interval  time.Duration
		timeout   time.Duration
		staleness time.Duration
		valid     bool
	}{
		{name: "below interval", interval: time.Minute, timeout: time.Second, staleness: 30 * time.Second},
		{name: "equal to interval", interval: time.Minute, timeout: time.Second, staleness: time.Minute},
		{name: "below sum", interval: time.Minute, timeout: time.Minute, staleness: 90 * time.Second},
		{name: "equal to sum", interval: time.Minute, timeout: time.Minute, staleness: 2 * time.Minute},
		{name: "above sum", interval: time.Minute, timeout: time.Minute, staleness: 2*time.Minute + time.Nanosecond, valid: true},
		{name: "sum would overflow", interval: maxDuration - time.Second, timeout: 2 * time.Second, staleness: maxDuration},
		{name: "large valid durations", interval: maxDuration - 2*time.Second, timeout: time.Second, staleness: maxDuration, valid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{RefreshInterval: tc.interval, RefreshTimeout: tc.timeout, MaxStaleness: tc.staleness}
			factory := NewFactoryWithTokenSource(sourceFunc(func(context.Context) ([]HecTokenConfig, error) { return nil, nil }))
			_, err := factory.Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
			if tc.valid {
				require.NoError(t, cfg.Validate())
				require.NoError(t, err)
			} else {
				require.EqualError(t, cfg.Validate(), "max_staleness must be greater than refresh_interval + refresh_timeout")
				require.EqualError(t, err, "max_staleness must be greater than refresh_interval + refresh_timeout")
			}
		})
	}
}
