// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf

import (
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/processor"
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

// stubOutputMapper is the output-side counterpart of stubMapper.
type stubOutputMapper struct {
	exporter.Factory
	claims map[string]component.Config
	err    error
}

func (s stubOutputMapper) MapOutputs([]Output) (map[string]component.Config, error) {
	return s.claims, s.err
}

func outStub(typeStr string, claims map[string]component.Config, err error) stubOutputMapper {
	return stubOutputMapper{
		Factory: exporter.NewFactory(component.MustNewType(typeStr), func() component.Config { return &struct{}{} }),
		claims:  claims,
		err:     err,
	}
}

func procStub(typeStr string, instances map[string]component.Config, err error) MappingProcessorFactory {
	return NewMappingProcessor(
		processor.NewFactory(component.MustNewType(typeStr), func() component.Config { return &struct{}{} }),
		func(Conf) (map[string]component.Config, error) { return instances, err },
	)
}

func TestMapProcessorsAssignsIDs(t *testing.T) {
	p := &provider{processorMappers: []MappingProcessorFactory{
		procStub("splunk_linebreak", map[string]component.Config{"": &struct{}{}}, nil),
		procStub("splunk_meta", map[string]component.Config{"named": &struct{}{}}, nil),
	}}

	got, err := p.mapProcessors(Conf{})
	require.NoError(t, err)
	require.Equal(t, []string{"splunk_linebreak", "splunk_meta/named"}, ids(got))
}

func TestMapProcessorsKeepsRegistrationOrder(t *testing.T) {
	p := &provider{processorMappers: []MappingProcessorFactory{
		procStub("zzz_first", map[string]component.Config{"": &struct{}{}}, nil),
		procStub("aaa_second", map[string]component.Config{"": &struct{}{}}, nil),
	}}

	got, err := p.mapProcessors(Conf{})
	require.NoError(t, err)
	require.Equal(t, []string{"zzz_first", "aaa_second"}, ids(got))
}

func TestMapProcessorsOrdersOneMappersInstancesByName(t *testing.T) {
	p := &provider{processorMappers: []MappingProcessorFactory{
		procStub("splunk_p", map[string]component.Config{
			"c": &struct{}{}, "a": &struct{}{}, "b": &struct{}{},
		}, nil),
	}}

	got, err := p.mapProcessors(Conf{})
	require.NoError(t, err)
	require.Equal(t, []string{"splunk_p/a", "splunk_p/b", "splunk_p/c"}, ids(got))
}

func TestMapProcessorsNoInstances(t *testing.T) {
	p := &provider{processorMappers: []MappingProcessorFactory{
		procStub("splunk_p", nil, nil),
	}}

	got, err := p.mapProcessors(Conf{})
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestMapProcessorsError(t *testing.T) {
	p := &provider{processorMappers: []MappingProcessorFactory{
		procStub("splunk_p", nil, errors.New("boom")),
	}}

	_, err := p.mapProcessors(Conf{})
	require.ErrorContains(t, err, "splunk_p")
	require.ErrorContains(t, err, "boom")
}

func ids(es []emitted) []string {
	out := make([]string, 0, len(es))
	for _, e := range es {
		out = append(out, e.id)
	}
	return out
}

// The provider, not a mapper, decides the component ID, so the same stanza
// always lands on the same ID regardless of which receiver claimed it.
func TestMapInputsAssignsIDs(t *testing.T) {
	p := &provider{inputMappers: []MappingReceiverFactory{
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
	p := &provider{inputMappers: []MappingReceiverFactory{
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

	p := &provider{inputMappers: []MappingReceiverFactory{stub("splunk_a", nil, nil)}}
	out, skipped, err := p.mapInputs(tree)
	require.NoError(t, err)
	require.Empty(t, out)
	require.Equal(t, []string{"tcp-ssl://9998"}, skipped)
}

// A mapper's error names the type, so a bad out-of-repo mapper is attributable.
func TestMapInputsWrapsMapperError(t *testing.T) {
	p := &provider{inputMappers: []MappingReceiverFactory{
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
	const (
		hecout = "hecout"
		tcpout = "tcpout:primary"
	)
	hecStanza := map[string]string{"uri": "https://hec:8088", "httpEventCollectorToken": "tok"}
	tcpStanza := map[string]string{"server": "idx1:9997"}

	for _, tt := range []struct {
		merged      conf.Map
		name        string
		errMsg      string
		mappers     []MappingExporterFactory
		wantIDs     []string
		wantSkipped []string
	}{
		{
			// The provider models no output kind itself, so every stanza reaches a
			// registered mapper or none at all.
			name:    "a registered mapper's instances get provider-owned IDs",
			merged:  conf.Map{hecout: hecStanza, tcpout: tcpStanza},
			mappers: []MappingExporterFactory{outStub("splunk_s2s", map[string]component.Config{hecout: &struct{}{}, tcpout: &struct{}{}}, nil)},
			wantIDs: []string{"splunk_s2s/hecout-d6a24a0d", "splunk_s2s/tcpout-primary-bc98bd6f"},
		},
		{
			name:        "a stanza nobody claims is skipped alongside one that maps",
			merged:      conf.Map{hecout: hecStanza, tcpout: tcpStanza},
			mappers:     []MappingExporterFactory{outStub("splunk_hecout", map[string]component.Config{hecout: &struct{}{}}, nil)},
			wantIDs:     []string{"splunk_hecout/hecout-d6a24a0d"},
			wantSkipped: []string{tcpout},
		},
		{
			name:   "two registered mappers claiming one stanza",
			merged: conf.Map{tcpout: tcpStanza},
			mappers: []MappingExporterFactory{
				outStub("splunk_a", map[string]component.Config{tcpout: &struct{}{}}, nil),
				outStub("splunk_b", map[string]component.Config{tcpout: &struct{}{}}, nil),
			},
			errMsg: "claimed by both splunk_a and splunk_b",
		},
		{
			name:    "a mapper's error names the type",
			merged:  conf.Map{tcpout: tcpStanza},
			mappers: []MappingExporterFactory{outStub("splunk_s2s", nil, errors.New("boom"))},
			errMsg:  "splunk_s2s: boom",
		},
		{
			name:        "a kind nothing claims is reported as skipped",
			merged:      conf.Map{tcpout: tcpStanza},
			wantIDs:     []string{},
			wantSkipped: []string{tcpout},
		},
		{
			name:        "no mappers registered at all",
			merged:      conf.Map{hecout: hecStanza},
			wantIDs:     []string{},
			wantSkipped: []string{hecout},
		},
		{
			name:   "no output stanzas at all",
			merged: conf.Map{},
			errMsg: conf.ErrNoOutputStanzas.Error(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{outputMappers: tt.mappers}
			out, skipped, err := p.mapOutputs(tt.merged)
			if tt.errMsg != "" {
				require.ErrorContains(t, err, tt.errMsg)
				return
			}
			require.NoError(t, err)
			ids := make([]string, 0, len(out))
			for _, e := range out {
				ids = append(ids, e.id)
			}
			sort.Strings(ids)
			require.Equal(t, tt.wantIDs, ids)
			require.Equal(t, tt.wantSkipped, skipped)
		})
	}
}

func TestStableNameOfUnslugabbleStanza(t *testing.T) {
	require.Regexp(t, "^[0-9a-f]{8}$", stableName(":::"))
}

func TestNewMappingReceiver(t *testing.T) {
	want := map[string]component.Config{"monitor:///x": &struct{}{}}
	f := NewMappingReceiver(
		receiver.NewFactory(component.MustNewType("splunk_a"), func() component.Config { return &struct{}{} }),
		func(Conf) (map[string]component.Config, error) { return want, nil },
	)

	require.Equal(t, "splunk_a", f.Type().String())
	got, err := f.MapInputs(Conf{})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestNewMappingReceiverPropagatesError(t *testing.T) {
	f := NewMappingReceiver(
		receiver.NewFactory(component.MustNewType("splunk_a"), func() component.Config { return &struct{}{} }),
		func(Conf) (map[string]component.Config, error) { return nil, errors.New("boom") },
	)
	_, err := f.MapInputs(Conf{})
	require.EqualError(t, err, "boom")
}

func TestNewMappingExporter(t *testing.T) {
	want := map[string]component.Config{"hecout": &struct{}{}}
	f := NewMappingExporter(
		exporter.NewFactory(component.MustNewType("splunk_e"), func() component.Config { return &struct{}{} }),
		func([]Output) (map[string]component.Config, error) { return want, nil },
	)

	require.Equal(t, "splunk_e", f.Type().String())
	got, err := f.MapOutputs(nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
