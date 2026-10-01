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

// TestConfigUnmarshal covers the shape a .conf config source emits: modeled
// stanza params bind to typed fields, and anything unmodeled lands in Extra
// instead of failing strict unmarshal.
func TestConfigUnmarshal(t *testing.T) {
	cm := confmap.NewFromStringMap(map[string]any{
		"event_log_name":     "Application",
		"index":              "wineventlog",
		"source":             "WinEventLog:Application",
		"sourcetype":         "WinEventLog",
		"host":               "h1",
		"current_only":       "1",
		"evt_resolve_ad_obj": "0",
	})

	cfg := &Config{}
	require.NoError(t, cm.Unmarshal(cfg))

	require.Equal(t, "Application", cfg.EventLogName)
	require.Equal(t, "wineventlog", cfg.Index)
	require.Equal(t, "WinEventLog:Application", cfg.Source)
	require.Equal(t, "WinEventLog", cfg.Sourcetype)
	require.Equal(t, "h1", cfg.Host)
	require.Equal(t, map[string]string{"current_only": "1", "evt_resolve_ad_obj": "0"}, cfg.Extra,
		"unmodeled params must survive in Extra")
}

// TestInputTranslation pins the typed config back to the stanza tabuilder
// dispatches on. Params are asserted as an exact ordered slice: the order is
// arbitrary to consumers but fixed, so a reordering here is a real change.
func TestInputTranslation(t *testing.T) {
	cfg := &Config{
		EventLogName: "Application",
		Index:        "wineventlog",
		Source:       "WinEventLog:Application",
		Sourcetype:   "WinEventLog",
		Host:         "h1",
		Extra:        map[string]string{"current_only": "1", "evt_resolve_ad_obj": "0"},
	}

	input := cfg.input()
	require.Equal(t, "wineventlog://Application", input.Configuration.Stanza.Name)
	require.Equal(t, conf.Params{
		{Name: "index", Value: "wineventlog"},
		{Name: "source", Value: "WinEventLog:Application"},
		{Name: "sourcetype", Value: "WinEventLog"},
		{Name: "host", Value: "h1"},
		{Name: "current_only", Value: "1"},
		{Name: "evt_resolve_ad_obj", Value: "0"},
	}, input.Configuration.Stanza.Params)
}

// TestInputTranslationMinimal proves an unset field emits no param at all,
// rather than an empty one that would override a props.conf default.
func TestInputTranslationMinimal(t *testing.T) {
	cfg := &Config{EventLogName: "Security"}
	input := cfg.input()
	require.Equal(t, "wineventlog://Security", input.Configuration.Stanza.Name)
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

// TestFactory asserts both platform branches. The underlying
// wineventlogreceiver is build-tagged: on Windows it builds, everywhere else its
// factory returns an explicit error. The wrapper delegates either way, so the
// expectation is platform-dependent rather than skipped.
func TestFactory(t *testing.T) {
	f := NewFactory()
	require.Equal(t, TypeStr, f.Type().String())
	require.Equal(t, &Config{}, f.CreateDefaultConfig())

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
