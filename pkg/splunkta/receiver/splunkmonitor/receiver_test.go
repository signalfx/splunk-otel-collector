// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkmonitor

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
				"path":          "/var/log/syslog",
				"index":         "main",
				"host":          "h1",
				"charset":       "UTF-8",
				"truncate":      512,
				"event_breaker": "BRK",
				"followTail":    "1",
				"blacklist":     "\\.gz$",
			},
			extra: map[string]string{
				"followTail": "1",
				"blacklist":  "\\.gz$",
			},
			stanza: "monitor:///var/log/syslog",
			params: conf.Params{
				{Name: "index", Value: "main"},
				{Name: "host", Value: "h1"},
				{Name: "blacklist", Value: "\\.gz$"},
				{Name: "followTail", Value: "1"},
				{Name: "CHARSET", Value: "UTF-8"},
				{Name: "EVENT_BREAKER", Value: "BRK"},
				{Name: "TRUNCATE", Value: "512"},
			},
		},
		{
			name:   "target only, so nothing overrides a props.conf default",
			yaml:   map[string]any{"path": "/var/log/a.log"},
			stanza: "monitor:///var/log/a.log",
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
			cfg:  Config{Path: "/var/log/syslog"},
		},
		{
			name:   "no path",
			cfg:    Config{Index: "main"},
			errMsg: "path is required",
		},
		{
			name:   "whitespace-only path",
			cfg:    Config{Path: "  "},
			errMsg: "path is required",
		},
		{
			name:   "negative truncate",
			cfg:    Config{Path: "/var/log/syslog", Truncate: -1},
			errMsg: "truncate must not be negative",
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
	cfg := &Config{Path: "/var/log/syslog", Index: "main"}
	r, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, r)
}
