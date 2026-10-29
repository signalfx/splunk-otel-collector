// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package splunkauthextension

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/extension/extensiontest"

	"github.com/signalfx/splunk-otel-collector/internal/modularinput"
)

func TestFactoryDefaultCredentials(t *testing.T) {
	for _, tc := range []struct {
		name, endpoint, sessionKey string
	}{
		{name: "empty environment"},
		{name: "modular input environment", endpoint: "https://127.0.0.1:8089", sessionKey: "env-session"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(modularinput.EnvManagementURI, tc.endpoint)
			t.Setenv(modularinput.EnvSessionKey, tc.sessionKey)
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig().(*Config)
			require.Equal(t, tc.endpoint, cfg.Endpoint)
			require.Equal(t, configopaque.String(tc.sessionKey), cfg.SessionKey)
			require.NoError(t, confmap.NewFromStringMap(map[string]any{
				"endpoint": "https://explicit.example.com:8089", "session_key": "explicit-session",
			}).Unmarshal(cfg))
			require.Equal(t, "https://explicit.example.com:8089", cfg.Endpoint)
			require.Equal(t, configopaque.String("explicit-session"), cfg.SessionKey)
			require.NoError(t, cfg.Validate())
			injected := NewFactoryWithTokenSource(sourceFunc(func(context.Context) ([]HecTokenConfig, error) { return nil, nil }))
			injectedCfg := injected.CreateDefaultConfig().(*Config)
			require.Empty(t, injectedCfg.Endpoint)
			require.Empty(t, injectedCfg.SessionKey)
			ext, err := injected.Create(t.Context(), extensiontest.NewNopSettings(injected.Type()), injectedCfg)
			require.NoError(t, err)
			require.NoError(t, ext.Shutdown(t.Context()))
		})
	}
}

func TestConfiguredTLS(t *testing.T) {
	api := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(tokenPageXML(0, 0, 0)))
	}))
	defer api.Close()
	caFile := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(caFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: api.Certificate().Raw}), 0o600))
	for _, tc := range []struct {
		name, ca          string
		skipVerify, valid bool
	}{
		{name: "untrusted certificate"},
		{name: "explicit skip verification", skipVerify: true, valid: true},
		{name: "configured CA", ca: caFile, valid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewDefaultConfig()
			cfg.Endpoint, cfg.SessionKey = api.URL, "session"
			cfg.TLS.InsecureSkipVerify = tc.skipVerify
			cfg.TLS.CAFile = tc.ca
			factory := NewFactory()
			ext, err := factory.Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
			require.NoError(t, err)
			err = ext.Start(t.Context(), componenttest.NewNopHost())
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "certificate")
			}
			require.NoError(t, ext.Shutdown(t.Context()))
		})
	}
	cfg := NewDefaultConfig()
	cfg.Endpoint, cfg.SessionKey = api.URL, "session"
	cfg.TLS.CAFile = "testdata/nonexistent-ca.pem"
	factory := NewFactory()
	_, err := factory.Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
	require.Error(t, err)
	cfg.TLS.CAFile = ""
	_, err = NewFactoryWithTokenSource(sourceFunc(func(context.Context) ([]HecTokenConfig, error) { return nil, nil })).Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
	require.ErrorContains(t, err, "cannot be combined")
}

func TestFactoryCreate(t *testing.T) {
	factory := NewFactoryWithTokenSource(sourceFunc(func(_ context.Context) ([]HecTokenConfig, error) { return nil, nil }))
	cfg := factory.CreateDefaultConfig().(*Config)
	require.Equal(t, 3*time.Minute, cfg.RefreshInterval)
	require.Equal(t, 30*time.Second, cfg.RefreshTimeout)
	require.Equal(t, 10*time.Minute, cfg.MaxStaleness)
	require.NoError(t, cfg.Validate())
	ext, err := factory.Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)
	_, err = NewFactoryWithTokenSource(nil).Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), cfg)
	require.EqualError(t, err, "endpoint and session_key are required without an injected token source")
	for _, field := range []string{"interval", "timeout", "staleness"} {
		for _, value := range []time.Duration{0, -time.Second} {
			invalid := *cfg
			switch field {
			case "interval":
				invalid.RefreshInterval = value
			case "timeout":
				invalid.RefreshTimeout = value
			case "staleness":
				invalid.MaxStaleness = value
			}
			require.Error(t, invalid.Validate())
			_, err = factory.Create(t.Context(), extensiontest.NewNopSettings(factory.Type()), &invalid)
			require.Error(t, err)
		}
	}
}
