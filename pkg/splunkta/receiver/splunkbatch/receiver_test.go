// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkbatch

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
				"file_path":   "/var/spool/splunk/batch.log",
				"index":       "batch",
				"source":      "batch.log",
				"sourcetype":  "batch_out",
				"host":        "h1",
				"move_policy": "sinkhole",
			},
			extra:  map[string]string{"move_policy": "sinkhole"},
			stanza: "batch:///var/spool/splunk/batch.log",
			params: conf.Params{
				{Name: "index", Value: "batch"},
				{Name: "source", Value: "batch.log"},
				{Name: "sourcetype", Value: "batch_out"},
				{Name: "host", Value: "h1"},
				{Name: "move_policy", Value: "sinkhole"},
			},
		},
		{
			name:   "target only, so nothing overrides a props.conf default",
			yaml:   map[string]any{"file_path": "/var/spool/splunk/x.log"},
			stanza: "batch:///var/spool/splunk/x.log",
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
			cfg:  Config{FilePath: "/var/spool/splunk/batch.log"},
		},
		{
			name:   "empty file path",
			cfg:    Config{Index: "batch"},
			errMsg: "file_path is required",
		},
		{
			name:   "whitespace only file path",
			cfg:    Config{FilePath: "   "},
			errMsg: "file_path is required",
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
	cfg := &Config{FilePath: "/var/spool/splunk/batch.log", Index: "batch"}
	r, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, r)
}
