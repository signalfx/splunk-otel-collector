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

// TestConfigUnmarshal covers the shape a .conf config source emits: modeled
// stanza params bind to typed fields, and anything unmodeled lands in Extra
// instead of failing strict unmarshal.
func TestConfigUnmarshal(t *testing.T) {
	cm := confmap.NewFromStringMap(map[string]any{
		"include":       []any{"/var/log/syslog"},
		"exclude":       []any{"*.gz"},
		"index":         "main",
		"sourcetype":    "syslog",
		"charset":       "UTF-8",
		"truncate":      10000,
		"event_breaker": `([\r\n]+)`,
		"followTail":    "1",
		"whitelist":     `\.log$`,
	})

	cfg := &Config{}
	require.NoError(t, cm.Unmarshal(cfg))

	require.Equal(t, []string{"/var/log/syslog"}, cfg.Include)
	require.Equal(t, []string{"*.gz"}, cfg.Exclude)
	require.Equal(t, "main", cfg.Index)
	require.Equal(t, "syslog", cfg.Sourcetype)
	require.Equal(t, "UTF-8", cfg.Charset)
	require.Equal(t, 10000, cfg.Truncate)
	require.Equal(t, `([\r\n]+)`, cfg.EventBreaker)
	require.Equal(t, map[string]string{"followTail": "1", "whitelist": `\.log$`}, cfg.Extra,
		"unmodeled params must survive in Extra")
}

// TestInputTranslation pins the typed config back to the stanza tabuilder
// dispatches on. Params are asserted as an exact ordered slice: the order is
// arbitrary to consumers but fixed, so a reordering here is a real change.
func TestInputTranslation(t *testing.T) {
	cfg := &Config{
		Include:      []string{"/var/log/syslog"},
		Exclude:      []string{"*.gz", "*.bz2"},
		Index:        "main",
		Host:         "h1",
		Charset:      "UTF-8",
		Truncate:     512,
		EventBreaker: "BRK",
		Extra:        map[string]string{"zeta": "z", "alpha": "a"},
	}

	input := cfg.input()
	require.Equal(t, "monitor:///var/log/syslog", input.Configuration.Stanza.Name)
	require.Equal(t, conf.Params{
		{Name: "index", Value: "main"},
		{Name: "host", Value: "h1"},
		{Name: "alpha", Value: "a"},
		{Name: "zeta", Value: "z"},
		{Name: "blacklist", Value: "*.gz|*.bz2"},
		{Name: "CHARSET", Value: "UTF-8"},
		{Name: "EVENT_BREAKER", Value: "BRK"},
		{Name: "TRUNCATE", Value: "512"},
	}, input.Configuration.Stanza.Params)
}

// TestInputTranslationMinimal proves an unset field emits no param at all,
// rather than an empty one that would override a props.conf default.
func TestInputTranslationMinimal(t *testing.T) {
	cfg := &Config{Include: []string{"/var/log/a.log"}}
	input := cfg.input()
	require.Equal(t, "monitor:///var/log/a.log", input.Configuration.Stanza.Name)
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
			cfg:  Config{Include: []string{"/var/log/syslog"}},
		},
		{
			name:   "no include",
			cfg:    Config{Index: "main"},
			errMsg: "include is required",
		},
		{
			name:   "empty include path",
			cfg:    Config{Include: []string{"  "}},
			errMsg: "must not contain empty paths",
		},
		{
			name:   "negative truncate",
			cfg:    Config{Include: []string{"/var/log/syslog"}, Truncate: -1},
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
	require.Equal(t, TypeStr, f.Type().String())
	require.Equal(t, &Config{}, f.CreateDefaultConfig())

	cfg := &Config{Include: []string{"/var/log/syslog"}, Index: "main"}
	r, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, r, "monitor is a kind tabuilder handles, so a receiver must be built")
}
