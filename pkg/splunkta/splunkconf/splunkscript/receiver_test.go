// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkscript

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
		"script_filename": "/usr/local/bin/test.sh",
		"interval":        "30",
		"index":           "scripts",
		"source":          "test.sh",
		"sourcetype":      "script_out",
		"host":            "h1",
		"passAuth":        "splunk-system-user",
		"start_by_shell":  "false",
	})

	cfg := &Config{}
	require.NoError(t, cm.Unmarshal(cfg))

	require.Equal(t, "/usr/local/bin/test.sh", cfg.ScriptFilename)
	require.Equal(t, "30", cfg.Interval)
	require.Equal(t, "scripts", cfg.Index)
	require.Equal(t, "test.sh", cfg.Source)
	require.Equal(t, "script_out", cfg.Sourcetype)
	require.Equal(t, "h1", cfg.Host)
	require.Equal(t, map[string]string{"passAuth": "splunk-system-user", "start_by_shell": "false"}, cfg.Extra,
		"unmodeled params must survive in Extra")
}

// TestInputTranslation pins the typed config back to the stanza tabuilder
// dispatches on. Params are asserted as an exact ordered slice: the order is
// arbitrary to consumers but fixed, so a reordering here is a real change.
// interval is appended after the extras, being a type-specific param.
func TestInputTranslation(t *testing.T) {
	cfg := &Config{
		ScriptFilename: "/usr/local/bin/test.sh",
		Interval:       "30",
		Index:          "scripts",
		Source:         "test.sh",
		Sourcetype:     "script_out",
		Host:           "h1",
		Extra:          map[string]string{"passAuth": "splunk-system-user", "and_first": "x"},
	}

	input := cfg.input()
	require.Equal(t, "script:///usr/local/bin/test.sh", input.Configuration.Stanza.Name)
	require.Equal(t, conf.Params{
		{Name: "index", Value: "scripts"},
		{Name: "source", Value: "test.sh"},
		{Name: "sourcetype", Value: "script_out"},
		{Name: "host", Value: "h1"},
		{Name: "and_first", Value: "x"},
		{Name: "passAuth", Value: "splunk-system-user"},
		{Name: "interval", Value: "30"},
	}, input.Configuration.Stanza.Params)
}

// TestInputTranslationMinimal proves an unset field emits no param at all,
// rather than an empty one that would override a props.conf default.
func TestInputTranslationMinimal(t *testing.T) {
	cfg := &Config{ScriptFilename: "/usr/local/bin/test.sh"}
	input := cfg.input()
	require.Equal(t, "script:///usr/local/bin/test.sh", input.Configuration.Stanza.Name)
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
			cfg:  Config{ScriptFilename: "/usr/local/bin/test.sh"},
		},
		{
			name:   "empty script filename",
			cfg:    Config{Index: "scripts"},
			errMsg: "script_filename is required",
		},
		{
			name:   "whitespace only script filename",
			cfg:    Config{ScriptFilename: "   "},
			errMsg: "script_filename is required",
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

	cfg := &Config{ScriptFilename: "/usr/local/bin/test.sh", Index: "scripts"}
	r, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, r, "script is a kind tabuilder handles, so a receiver must be built")
}
