// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkwineventlog

import (
	"context"
	"runtime"
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
				"event_log_name": "Application",
				"index":          "wineventlog",
				"source":         "WinEventLog:Application",
				"sourcetype":     "WinEventLog",
				"host":           "h1",
				"current_only":   "1",
			},
			extra:  map[string]string{"current_only": "1"},
			stanza: "wineventlog://Application",
			params: conf.Params{
				{Name: "index", Value: "wineventlog"},
				{Name: "source", Value: "WinEventLog:Application"},
				{Name: "sourcetype", Value: "WinEventLog"},
				{Name: "host", Value: "h1"},
				{Name: "current_only", Value: "1"},
			},
		},
		{
			name:   "target only, so nothing overrides a props.conf default",
			yaml:   map[string]any{"event_log_name": "Security"},
			stanza: "wineventlog://Security",
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
			cfg:  Config{EventLogName: "Application"},
		},
		{
			name:   "empty event log name",
			cfg:    Config{Index: "wineventlog"},
			errMsg: "event_log_name is required",
		},
		{
			name:   "whitespace only event log name",
			cfg:    Config{EventLogName: "   "},
			errMsg: "event_log_name is required",
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

// The underlying wineventlogreceiver is build-tagged: on Windows it builds,
// everywhere else its factory returns an explicit error. The wrapper delegates
// either way, so the expectation is platform-dependent rather than skipped.
func TestFactory(t *testing.T) {
	f := NewFactory()
	cfg := &Config{EventLogName: "Application", Index: "wineventlog"}
	r, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	if runtime.GOOS != "windows" {
		require.ErrorContains(t, err, "wineventlog is not supported outside Windows environments")
		require.Nil(t, r)
		return
	}
	require.NoError(t, err)
	require.NotNil(t, r)
}
