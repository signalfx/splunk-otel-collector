// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/file"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/split"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

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
		name   string
		params conf.Params
		want   map[string]helper.ExprStringConfig
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
