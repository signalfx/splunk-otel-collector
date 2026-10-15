// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/file"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/split"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

func TestSimpleRegexGlobs(t *testing.T) {
	if filepath.Separator != '/' {
		t.Skip("Splunk_TA_nix path discovery is supported on Unix systems")
	}

	path := filepath.Join(string(filepath.Separator), "var", "log")
	tests := []struct {
		name       string
		expression string
		want       []string
	}{
		{
			name:       "Splunk TA nix whitelist",
			expression: `(\.log|log$|messages|secure|auth|mesg$|cron$|acpid$|\.out)`,
			want: []string{
				filepath.Join(path, "*.log*"),
				filepath.Join(path, "*log"),
				filepath.Join(path, "*messages*"),
				filepath.Join(path, "*secure*"),
				filepath.Join(path, "*auth*"),
				filepath.Join(path, "*mesg"),
				filepath.Join(path, "*cron"),
				filepath.Join(path, "*acpid"),
				filepath.Join(path, "*.out*"),
			},
		},
		{
			name:       "different whitelist",
			expression: `(access\.log$|error|audit)`,
			want: []string{
				filepath.Join(path, "*access.log"),
				filepath.Join(path, "*error*"),
				filepath.Join(path, "*audit*"),
			},
		},
		{
			name:       "glob metacharacters are literals",
			expression: `\*\?\[\\\{\}`,
			want:       []string{filepath.Join(path, `*\*\?\[\\\{\}*`)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := simpleRegexGlobs(path, tt.expression)
			require.True(t, ok)
			require.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestSimpleRegexGlobsFallsBack(t *testing.T) {
	if filepath.Separator != '/' {
		t.Skip("Splunk_TA_nix path discovery is supported on Unix systems")
	}

	tests := []struct {
		name       string
		path       string
		expression string
	}{
		{name: "empty", path: "/var/log", expression: ""},
		{name: "malformed", path: "/var/log", expression: `(`},
		{name: "wildcard", path: "/var/log", expression: `auth.*`},
		{name: "character class", path: "/var/log", expression: `[ab]`},
		{name: "repetition", path: "/var/log", expression: `auth+`},
		{name: "case insensitive", path: "/var/log", expression: `(?i)auth`},
		{name: "start anchored", path: "/var/log", expression: `^auth`},
		{name: "path separator", path: "/var/log", expression: `var/log`},
		{name: "literal matches monitored path", path: "/var/auth", expression: `auth`},
		{name: "end anchor before branch end", path: "/var/log", expression: `auth$log`},
		{name: "multiple end anchors", path: "/var/log", expression: `auth$$`},
		{name: "duplicate alternatives", path: "/var/log", expression: `(auth|auth)`},
		{name: "factored alternation", path: "/var/log", expression: `(foo|foobar)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := simpleRegexGlobs(tt.path, tt.expression)
			require.False(t, ok)
			require.Nil(t, got)
		})
	}
}

func TestApplyIncludeExclude(t *testing.T) {
	if filepath.Separator != '/' {
		t.Skip("Splunk_TA_nix path discovery is supported on Unix systems")
	}

	t.Run("converts the Splunk TA nix expressions", func(t *testing.T) {
		path := "/var/log"
		oc := file.NewConfig()
		stanza := conf.Stanza{
			Name: "monitor:///var/log",
			Params: conf.Params{
				{Name: "whitelist", Value: `(\.log|log$|messages|secure|auth|mesg$|cron$|acpid$|\.out)`},
				{Name: "blacklist", Value: `(lastlog|anaconda\.syslog)`},
			},
		}

		ApplyIncludeExclude(oc, path, stanza, "monitor", zap.NewNop())

		require.ElementsMatch(t, []string{
			filepath.Join(path, "*.log*"),
			filepath.Join(path, "*log"),
			filepath.Join(path, "*messages*"),
			filepath.Join(path, "*secure*"),
			filepath.Join(path, "*auth*"),
			filepath.Join(path, "*mesg"),
			filepath.Join(path, "*cron"),
			filepath.Join(path, "*acpid"),
			filepath.Join(path, "*.out*"),
		}, oc.Include)
		require.ElementsMatch(t, []string{
			filepath.Join(path, "*lastlog*"),
			filepath.Join(path, "*anaconda.syslog*"),
		}, oc.Exclude)
	})

	t.Run("falls back for unsupported expressions", func(t *testing.T) {
		path := "/var/log"
		oc := file.NewConfig()
		stanza := conf.Stanza{
			Name: "monitor:///var/log",
			Params: conf.Params{
				{Name: "whitelist", Value: `.*\.log$`},
				{Name: "blacklist", Value: `.*debug.*`},
			},
		}

		ApplyIncludeExclude(oc, path, stanza, "monitor", zap.NewNop())

		require.Equal(t, []string{filepath.Join(path, "*")}, oc.Include)
		require.Empty(t, oc.Exclude)
	})

	t.Run("preserves a globbed stanza path", func(t *testing.T) {
		path := "/var/log/*/app?.log"
		oc := file.NewConfig()
		stanza := conf.Stanza{
			Name: "monitor://" + path,
			Params: conf.Params{
				{Name: "whitelist", Value: `auth`},
				{Name: "blacklist", Value: `debug`},
			},
		}

		ApplyIncludeExclude(oc, path, stanza, "monitor", zap.NewNop())

		require.Equal(t, []string{path}, oc.Include)
		require.Empty(t, oc.Exclude)
	})

	t.Run("empty whitelist expands a directory", func(t *testing.T) {
		path := t.TempDir()
		oc := file.NewConfig()
		stanza := conf.Stanza{
			Name:   "monitor://" + path,
			Params: conf.Params{{Name: "whitelist", Value: ""}, {Name: "blacklist", Value: ""}},
		}

		ApplyIncludeExclude(oc, path, stanza, "monitor", zap.NewNop())

		require.Equal(t, []string{filepath.Join(path, "*")}, oc.Include)
		require.Empty(t, oc.Exclude)
	})

	t.Run("missing whitelist expands a directory", func(t *testing.T) {
		path := t.TempDir()
		oc := file.NewConfig()

		ApplyIncludeExclude(oc, path, conf.Stanza{Name: "monitor://" + path}, "monitor", zap.NewNop())

		require.Equal(t, []string{filepath.Join(path, "*")}, oc.Include)
	})

	t.Run("missing whitelist preserves a file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "app.log")
		require.NoError(t, os.WriteFile(path, []byte("log"), 0o600))
		oc := file.NewConfig()

		ApplyIncludeExclude(oc, path, conf.Stanza{Name: "monitor://" + path}, "monitor", zap.NewNop())

		require.Equal(t, []string{path}, oc.Include)
	})
}

// TestApplyStanzaConfigSplitsOnNewlines pins the event-boundary contract: split
// on newlines and drop the terminator, the way Splunk indexes a line. A
// LineStartPattern here would end each token at the next line start and so keep
// the "\n" in the body.
func TestApplyStanzaConfigSplitsOnNewlines(t *testing.T) {
	oc := file.NewConfig()
	ApplyStanzaConfig(oc, conf.Stanza{Name: "monitor:///var/log/foo.log"})

	require.Equal(t, split.Config{}, oc.SplitConfig)
	// Whitespace inside the line is content, so only the delimiter is dropped.
	require.True(t, oc.TrimConfig.PreserveLeading)
	require.True(t, oc.TrimConfig.PreserveTrailing)
	require.Equal(t, "beginning", oc.StartAt)
	require.Equal(t, "utf-8", oc.Encoding)
	require.True(t, oc.IncludeFilePath)
}

// TestApplyStanzaConfigAttributes covers which stanza params reach the file
// input as attributes. A param the stanza omits must not appear at all, so a
// later operator can tell "unset" from "set to empty".
func TestApplyStanzaConfigAttributes(t *testing.T) {
	for _, tt := range []struct {
		want   map[string]helper.ExprStringConfig
		name   string
		params conf.Params
	}{
		{
			name:   "no params sets no attributes",
			params: conf.Params{},
			want:   map[string]helper.ExprStringConfig{},
		},
		{
			name: "every honored param",
			params: conf.Params{
				{Name: "host", Value: "myhost"},
				{Name: "index", Value: "myindex"},
				{Name: "sourcetype", Value: "myst"},
				{Name: "source", Value: "mysrc"},
			},
			want: map[string]helper.ExprStringConfig{
				"host":       "myhost",
				"index":      "myindex",
				"sourcetype": "myst",
				"source":     "mysrc",
			},
		},
		{
			name:   "unhonored params are ignored",
			params: conf.Params{{Name: "crcSalt", Value: "<SOURCE>"}, {Name: "followTail", Value: "1"}},
			want:   map[string]helper.ExprStringConfig{},
		},
		{
			name:   "an empty value is still set",
			params: conf.Params{{Name: "host", Value: ""}},
			want:   map[string]helper.ExprStringConfig{"host": ""},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oc := file.NewConfig()
			ApplyStanzaConfig(oc, conf.Stanza{Name: "monitor:///var/log/foo.log", Params: tt.params})
			require.Equal(t, tt.want, oc.Attributes)
		})
	}
}
