// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkudp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

// TestConfigUnmarshal covers the shape a .conf config source emits: modeled
// stanza params bind to typed fields, and anything unmodeled lands in Extra
// instead of failing strict unmarshal.
func TestConfigUnmarshal(t *testing.T) {
	cm := confmap.NewFromStringMap(map[string]any{
		"listen_address":  "10.0.0.5",
		"port":            45515,
		"index":           "net",
		"source":          "udp:45515",
		"sourcetype":      "syslog",
		"host":            "h1",
		"queueSize":       "500KB",
		"connection_host": "ip",
	})

	cfg := &Config{}
	require.NoError(t, cm.Unmarshal(cfg))

	require.Equal(t, "10.0.0.5", cfg.ListenAddress)
	require.Equal(t, 45515, cfg.Port)
	require.Equal(t, "net", cfg.Index)
	require.Equal(t, "udp:45515", cfg.Source)
	require.Equal(t, "syslog", cfg.Sourcetype)
	require.Equal(t, "h1", cfg.Host)
	require.Equal(t, map[string]string{"queueSize": "500KB", "connection_host": "ip"}, cfg.Extra,
		"unmodeled params must survive in Extra")
}

// TestInputTranslation pins the typed config back to the stanza tabuilder
// dispatches on. Params are asserted as an exact ordered slice: the order is
// arbitrary to consumers but fixed, so a reordering here is a real change.
func TestInputTranslation(t *testing.T) {
	cfg := &Config{
		ListenAddress: "10.0.0.5",
		Port:          45515,
		Index:         "net",
		Source:        "udp:45515",
		Sourcetype:    "syslog",
		Host:          "h1",
		Extra:         map[string]string{"queueSize": "500KB", "connection_host": "ip"},
	}

	input := cfg.input()
	require.Equal(t, "udp://10.0.0.5:45515", input.Configuration.Stanza.Name)
	require.Equal(t, conf.Params{
		{Name: "index", Value: "net"},
		{Name: "source", Value: "udp:45515"},
		{Name: "sourcetype", Value: "syslog"},
		{Name: "host", Value: "h1"},
		{Name: "connection_host", Value: "ip"},
		{Name: "queueSize", Value: "500KB"},
	}, input.Configuration.Stanza.Params)
}

// TestInputTranslationPortOnlyForm covers UF's port-only stanza form
// ([udp://5514]), which leaves the address empty and means every interface.
func TestInputTranslationPortOnlyForm(t *testing.T) {
	cfg := &Config{Port: 45515}
	input := cfg.input()
	require.Equal(t, "udp://0.0.0.0:45515", input.Configuration.Stanza.Name)
	require.Empty(t, input.Configuration.Stanza.Params)
}

func TestValidate(t *testing.T) {
	for _, tt := range []struct {
		name   string
		errMsg string
		cfg    Config
	}{
		{
			name: "valid",
			cfg:  Config{Port: 45515},
		},
		{
			name:   "zero port",
			cfg:    Config{},
			errMsg: "port must be between 1 and 65535, got 0",
		},
		{
			name:   "negative port",
			cfg:    Config{Port: -1},
			errMsg: "port must be between 1 and 65535, got -1",
		},
		{
			name:   "port above range",
			cfg:    Config{Port: 65536},
			errMsg: "port must be between 1 and 65535, got 65536",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.errMsg == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.errMsg)
		})
	}
}

func TestFactory(t *testing.T) {
	f := NewFactory()
	require.Equal(t, TypeStr, f.Type().String())
	require.Equal(t, &Config{}, f.CreateDefaultConfig())

	cfg := &Config{Port: 45515, Index: "net"}
	r, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, r, "udp is a kind tabuilder handles, so a receiver must be built")
}
