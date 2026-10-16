// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

// stubMapper claims whatever stanza names it is given, so the dispatch
// invariants can be exercised without a real receiver.
type stubMapper struct {
	receiver.Factory
	claims map[string]component.Config
	err    error
}

func (s stubMapper) MapInputs(Conf) (map[string]component.Config, error) {
	return s.claims, s.err
}

func stub(typeStr string, claims map[string]component.Config, err error) stubMapper {
	return stubMapper{
		Factory: receiver.NewFactory(component.MustNewType(typeStr), func() component.Config { return &struct{}{} }),
		claims:  claims,
		err:     err,
	}
}

// The provider, not a mapper, decides the component ID, so the same stanza
// always lands on the same ID regardless of which receiver claimed it.
func TestMapInputsAssignsIDs(t *testing.T) {
	p := &provider{mappers: []MappingReceiverFactory{
		stub("splunk_a", map[string]component.Config{"monitor:///var/log/foo": &struct{}{}}, nil),
	}}
	out, skipped, err := p.mapInputs(Conf{})
	require.NoError(t, err)
	require.Empty(t, skipped)
	require.Len(t, out, 1)
	require.Equal(t, "splunk_a/monitor-var-log-foo-31463e8f", out[0].id)
}

// Two mappers claiming one stanza is a registration bug, not a precedence
// question, so it fails rather than letting map iteration order pick a winner.
func TestMapInputsRejectsDoubleClaim(t *testing.T) {
	cfg := map[string]component.Config{"monitor:///x": &struct{}{}}
	p := &provider{mappers: []MappingReceiverFactory{
		stub("splunk_a", cfg, nil),
		stub("splunk_b", cfg, nil),
	}}
	_, _, err := p.mapInputs(Conf{})
	require.ErrorContains(t, err, "claimed by both")
}

// A stanza no mapper claims is reported, which is what makes an unsupported
// kind visible instead of silently collecting nothing.
func TestMapInputsReportsUnclaimed(t *testing.T) {
	tree, err := newConf(t.TempDir(), []conf.Input{
		{Configuration: conf.Configuration{Stanza: conf.Stanza{Name: "tcp-ssl://9998"}}},
	})
	require.NoError(t, err)

	p := &provider{mappers: []MappingReceiverFactory{stub("splunk_a", nil, nil)}}
	out, skipped, err := p.mapInputs(tree)
	require.NoError(t, err)
	require.Empty(t, out)
	require.Equal(t, []string{"tcp-ssl://9998"}, skipped)
}

// A mapper's error names the type, so a bad out-of-repo mapper is attributable.
func TestMapInputsWrapsMapperError(t *testing.T) {
	p := &provider{mappers: []MappingReceiverFactory{
		stub("splunk_a", nil, errors.New("boom")),
	}}
	_, _, err := p.mapInputs(Conf{})
	require.ErrorContains(t, err, "splunk_a: boom")
}

// A disabled stanza never reaches a mapper, so it is neither emitted nor
// reported as unclaimed.
func TestNewConfDropsDisabled(t *testing.T) {
	tree, err := newConf(t.TempDir(), []conf.Input{
		{Configuration: conf.Configuration{Stanza: conf.Stanza{
			Name:   "monitor:///off",
			Params: conf.Params{{Name: "disabled", Value: "1"}},
		}}},
	})
	require.NoError(t, err)
	require.Empty(t, tree.Inputs)
}

func TestParseURI(t *testing.T) {
	for _, tt := range []struct {
		name         string
		uri          string
		wantRoot     string
		wantPipeline string
		errMsg       string
	}{
		{
			name:         "pipeline defaults to uf",
			uri:          "splunkconf://testdata/etc",
			wantRoot:     "testdata/etc",
			wantPipeline: "uf",
		},
		{
			name:         "pipeline query param overrides the default",
			uri:          "splunkconf://testdata/etc?pipeline=custom",
			wantRoot:     "testdata/etc",
			wantPipeline: "custom",
		},
		{
			name:         "absolute config root",
			uri:          "splunkconf:///opt/splunk/etc",
			wantRoot:     "/opt/splunk/etc",
			wantPipeline: "uf",
		},
		{
			// A native Windows root is not a legal URI authority, so the root is
			// taken verbatim instead of through url.Parse. The space is what a
			// default install carries.
			name:         "windows config root with drive letter and spaces",
			uri:          `splunkconf://C:\Program Files\SplunkUniversalForwarder\etc?pipeline=uf`,
			wantRoot:     `C:\Program Files\SplunkUniversalForwarder\etc`,
			wantPipeline: "uf",
		},
		{
			name:   "wrong scheme",
			uri:    "file://testdata/etc",
			errMsg: `unexpected scheme in uri "file://testdata/etc"`,
		},
		{
			name:   "empty config root",
			uri:    "splunkconf://",
			errMsg: "empty config root",
		},
		{
			name:   "pipeline name outside the component-name charset",
			uri:    "splunkconf://testdata/etc?pipeline=Not/Valid",
			errMsg: `invalid pipeline "Not/Valid"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root, pipeline, err := parseURI(tt.uri)
			if tt.errMsg != "" {
				require.ErrorContains(t, err, tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantRoot, root)
			require.Equal(t, tt.wantPipeline, pipeline)
		})
	}
}

func TestMapOutputs(t *testing.T) {
	for _, tt := range []struct {
		name   string
		merged conf.Map
		wantID string
		errMsg string
	}{
		{
			name: "hecout maps to the contrib exporter",
			merged: conf.Map{
				"hecout": {"uri": "https://hec:8088", "httpEventCollectorToken": "tok"},
			},
			wantID: "splunk_hec/hecout-d6a24a0d",
		},
		{
			// S2S needs a real exporter that does not exist upstream, so it arrives
			// with the output-mapper registration API rather than being guessed at.
			name: "tcpout has no mapping yet and must not be dropped silently",
			merged: conf.Map{
				"tcpout:primary": {"server": "idx1:9997"},
			},
			errMsg: `unsupported output stanza(s) [tcpout:primary] (kind "tcpout")`,
		},
		{
			name:   "no output stanzas at all",
			merged: conf.Map{},
			errMsg: conf.ErrNoOutputStanzas.Error(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := mapOutputs(tt.merged)
			if tt.errMsg != "" {
				require.ErrorContains(t, err, tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.Len(t, out, 1)
			require.Equal(t, tt.wantID, out[0].id)
		})
	}
}

func TestStableNameOfUnslugabbleStanza(t *testing.T) {
	require.Regexp(t, "^[0-9a-f]{8}$", stableName(":::"))
}
