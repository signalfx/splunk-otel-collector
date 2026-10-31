// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/extension/extensiontest"

	"github.com/signalfx/splunk-otel-collector/internal/modularinput"
)

func TestDefaultConfigEnablesCertificateVerification(t *testing.T) {
	require.False(t, createDefaultConfig().(*Config).TLS.InsecureSkipVerify)
}

func TestMissingManagementCredentials(t *testing.T) {
	t.Setenv(modularinput.EnvManagementURI, "")
	t.Setenv(modularinput.EnvSessionKey, "")
	for _, tc := range []struct {
		name, endpoint, sessionKey, want string
	}{
		{name: "both missing", want: "endpoint is required"},
		{name: "endpoint missing", sessionKey: "session", want: "endpoint is required"},
		{name: "session key missing", endpoint: "https://splunk:8089", want: "session_key is required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig().(*Config)
			require.NoError(t, confmap.NewFromStringMap(map[string]any{
				"endpoint": tc.endpoint, "session_key": tc.sessionKey,
			}).Unmarshal(cfg))
			require.EqualError(t, confmap.Validate(cfg), tc.want)
			_, err := factory.Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestConfiguredSourceValidation(t *testing.T) {
	for _, endpoint := range []string{"splunk:8089", "ftp://splunk:8089", "https://", "https://user:secret@splunk:8089", "https://splunk:8089?token=secret", "https://splunk:8089#fragment"} {
		t.Run(endpoint, func(t *testing.T) {
			cfg := createDefaultConfig().(*Config)
			cfg.Endpoint, cfg.SessionKey = endpoint, "secret"
			require.Error(t, cfg.Validate())
		})
	}
	for _, values := range []map[string]any{
		{"endpoint": "https://splunk:8089"},
		{"session_key": "secret"},
		{"unknown_field": true},
		{"token_source": "not configurable through YAML"},
		{"refresh_interval": "not a duration"},
	} {
		cfg := createDefaultConfig().(*Config)
		err := confmap.NewFromStringMap(values).Unmarshal(cfg)
		if err == nil {
			require.Error(t, cfg.Validate())
		}
	}
	secret := confmap.NewFromStringMap(map[string]any{"endpoint": "https://splunk:8089", "session_key": "secret"})
	cfg := createDefaultConfig().(*Config)
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
			cfg := createDefaultConfig().(*Config)
			cfg.Endpoint, cfg.SessionKey = "http://localhost:8089", "session"
			cfg.RefreshInterval, cfg.RefreshTimeout, cfg.MaxStaleness = tc.interval, tc.timeout, tc.staleness
			factory := NewFactory()
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
