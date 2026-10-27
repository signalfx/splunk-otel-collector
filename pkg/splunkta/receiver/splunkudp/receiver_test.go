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

// Unmarshalling a provider-shaped fragment into the factory's default config and
// translating it back to a stanza is one chain, and the one the collector drives:
// the mapstructure keys are the contract with the .conf config source, and the
// stanza is what tabuilder dispatches on.
func TestConfigToStanza(t *testing.T) {
	for _, tt := range []struct {
		name   string
		yaml   map[string]any
		extra  map[string]string
		stanza string
		params conf.Params
	}{
		{
			name: "every modeled field plus an unmodeled one",
			yaml: map[string]any{
				"listen_address":  "10.0.0.5",
				"port":            45515,
				"index":           "net",
				"source":          "udp:45515",
				"sourcetype":      "syslog",
				"host":            "h1",
				"connection_host": "ip",
			},
			extra:  map[string]string{"connection_host": "ip"},
			stanza: "udp://10.0.0.5:45515",
			params: conf.Params{
				{Name: "index", Value: "net"},
				{Name: "source", Value: "udp:45515"},
				{Name: "sourcetype", Value: "syslog"},
				{Name: "host", Value: "h1"},
				{Name: "connection_host", Value: "ip"},
			},
		},
		{
			name:   "port-only form, so every interface and nothing overrides a props.conf default",
			yaml:   map[string]any{"port": 45515},
			stanza: "udp://0.0.0.0:45515",
			params: conf.Params{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewFactory().CreateDefaultConfig().(*Config)
			require.NoError(t, confmap.NewFromStringMap(tt.yaml).Unmarshal(cfg))
			require.Equal(t, tt.extra, cfg.Extra)

			in := cfg.input()
			require.Equal(t, tt.stanza, in.Configuration.Stanza.Name)
			require.Equal(t, tt.params, in.Configuration.Stanza.Params)
		})
	}
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
	cfg := &Config{Port: 45515, Index: "net"}
	r, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, r)
}
