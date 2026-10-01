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

// TestConfigUnmarshal covers the shape a .conf config source emits: modeled
// stanza params bind to typed fields, and anything unmodeled lands in Extra
// instead of failing strict unmarshal.
func TestConfigUnmarshal(t *testing.T) {
	cm := confmap.NewFromStringMap(map[string]any{
		"file_path":   "/var/spool/splunk/batch.log",
		"index":       "batch",
		"source":      "batch.log",
		"sourcetype":  "batch_out",
		"host":        "h1",
		"move_policy": "sinkhole",
		"crcSalt":     "<SOURCE>",
	})

	cfg := &Config{}
	require.NoError(t, cm.Unmarshal(cfg))

	require.Equal(t, "/var/spool/splunk/batch.log", cfg.FilePath)
	require.Equal(t, "batch", cfg.Index)
	require.Equal(t, "batch.log", cfg.Source)
	require.Equal(t, "batch_out", cfg.Sourcetype)
	require.Equal(t, "h1", cfg.Host)
	require.Equal(t, map[string]string{"move_policy": "sinkhole", "crcSalt": "<SOURCE>"}, cfg.Extra,
		"unmodeled params must survive in Extra")
}

// TestInputTranslation pins the typed config back to the stanza tabuilder
// dispatches on. Params are asserted as an exact ordered slice: the order is
// arbitrary to consumers but fixed, so a reordering here is a real change.
func TestInputTranslation(t *testing.T) {
	cfg := &Config{
		FilePath:   "/var/spool/splunk/batch.log",
		Index:      "batch",
		Source:     "batch.log",
		Sourcetype: "batch_out",
		Host:       "h1",
		Extra:      map[string]string{"move_policy": "sinkhole", "crcSalt": "<SOURCE>"},
	}

	input := cfg.input()
	require.Equal(t, "batch:///var/spool/splunk/batch.log", input.Configuration.Stanza.Name)
	require.Equal(t, conf.Params{
		{Name: "index", Value: "batch"},
		{Name: "source", Value: "batch.log"},
		{Name: "sourcetype", Value: "batch_out"},
		{Name: "host", Value: "h1"},
		{Name: "crcSalt", Value: "<SOURCE>"},
		{Name: "move_policy", Value: "sinkhole"},
	}, input.Configuration.Stanza.Params)
}

// TestInputTranslationMinimal proves an unset field emits no param at all,
// rather than an empty one that would override a props.conf default.
func TestInputTranslationMinimal(t *testing.T) {
	cfg := &Config{FilePath: "/var/spool/splunk/batch.log"}
	input := cfg.input()
	require.Equal(t, "batch:///var/spool/splunk/batch.log", input.Configuration.Stanza.Name)
	require.Empty(t, input.Configuration.Stanza.Params)
}

func TestValidate(t *testing.T) {
	for _, tt := range []struct {
		name   string
		cfg    Config
		errMsg string
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
	require.Equal(t, TypeStr, f.Type().String())
	require.Equal(t, &Config{}, f.CreateDefaultConfig())

	cfg := &Config{FilePath: "/var/spool/splunk/batch.log", Index: "batch"}
	r, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, r, "batch is a kind tabuilder handles, so a receiver must be built")
}
