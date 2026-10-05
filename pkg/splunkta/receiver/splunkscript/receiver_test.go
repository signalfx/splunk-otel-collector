// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkscript

import (
	"context"
	"path/filepath"
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
// stanza is what tabuilder dispatches on. interval lands after the extras, being
// a type-specific param rather than a resource attribute.
func TestConfigToStanza(t *testing.T) {
	for _, tt := range []struct {
		name   string
		yaml   map[string]any
		extra  map[string]string
		stanza string
		appDir string
		params conf.Params
	}{
		{
			name: "every modeled field plus an unmodeled one",
			yaml: map[string]any{
				"script_filename": "/usr/local/bin/test.sh",
				"interval":        "30",
				"index":           "scripts",
				"source":          "test.sh",
				"sourcetype":      "script_out",
				"host":            "h1",
				"passAuth":        "splunk-system-user",
			},
			extra:  map[string]string{"passAuth": "splunk-system-user"},
			stanza: "script:///usr/local/bin/test.sh",
			params: conf.Params{
				{Name: "index", Value: "scripts"},
				{Name: "source", Value: "test.sh"},
				{Name: "sourcetype", Value: "script_out"},
				{Name: "host", Value: "h1"},
				{Name: "passAuth", Value: "splunk-system-user"},
				{Name: "interval", Value: "30"},
			},
		},
		{
			name:   "target only, so nothing overrides a props.conf default",
			yaml:   map[string]any{"script_filename": "/usr/local/bin/x.sh"},
			stanza: "script:///usr/local/bin/x.sh",
			params: conf.Params{},
		},
		{
			name: "app-relative target carries the app dir through to Input",
			yaml: map[string]any{
				"script_filename": "./bin/foo.sh",
				"app_dir":         "/opt/splunk/etc/apps/splunk_ta_nix",
			},
			stanza: "script://./bin/foo.sh",
			appDir: "/opt/splunk/etc/apps/splunk_ta_nix",
			params: conf.Params{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewFactory().CreateDefaultConfig().(*Config)
			require.NoError(t, confmap.NewFromStringMap(tt.yaml).Unmarshal(cfg))
			require.Equal(t, tt.extra, cfg.Extra)

			in := cfg.input()
			require.Equal(t, tt.stanza, in.Configuration.Stanza.Name)
			require.Equal(t, tt.appDir, in.AppDir,
				"AppDir must survive: script.DetermineCommandName resolves and sandboxes a relative target against it")
			require.Equal(t, tt.params, in.Configuration.Stanza.Params)
		})
	}
}

func TestValidate(t *testing.T) {
	// Absoluteness is platform-dependent and Validate mirrors
	// script.DetermineCommandName's filepath.IsAbs check, so a POSIX path like
	// /usr/local/bin/test.sh is relative on Windows. Derive one that is absolute
	// wherever the test runs.
	absScript := filepath.Join(t.TempDir(), "test.sh")

	for _, tt := range []struct {
		name   string
		errMsg string
		cfg    Config
	}{
		{
			name: "absolute target needs no app dir",
			cfg:  Config{ScriptFilename: absScript},
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
		{
			name: "relative target with an app dir",
			cfg:  Config{ScriptFilename: "./bin/foo.sh", AppDir: "/opt/splunk/etc/apps/splunk_ta_nix"},
		},
		{
			name:   "relative target without an app dir would resolve against the collector cwd",
			cfg:    Config{ScriptFilename: "./bin/foo.sh"},
			errMsg: "app_dir is required to resolve the relative script_filename",
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
	cfg := &Config{ScriptFilename: "/usr/local/bin/test.sh", Index: "scripts"}
	r, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, r)
}
