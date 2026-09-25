// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkhome is a confmap.Provider that reads a Splunk $SPLUNK_HOME
// .conf tree and emits an in-memory pipeline (default name logs/uf) built from
// the input and output stanzas. It is Retrieve-only: it does not watch files.
// Reload rides the collector's existing trigger (SIGHUP or an OpAMP-pushed
// config), consistent with UF's pull/triggered model.
//
// URI form: splunkhome://<SPLUNK_HOME>?pipeline=<name>
// The pipeline (default "uf") names the emitted pipeline (logs/<pipeline>).
// Enable/disable of the whole pipeline is done at the launch level by including
// or omitting the --config=splunkhome://... argument.
package splunkhome

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"

	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

const schemeName = "splunkhome"

// pipelineRegexp restricts the pipeline name to a safe component-name segment.
var pipelineRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

type provider struct {
	reg *registry
}

// NewFactory returns a confmap.ProviderFactory for the splunkhome scheme.
// Options register additional UF-ported components (e.g. an S2S exporter via
// WithOutputMapper) on top of the built-in mappers. With no options it emits
// only the built-ins (wrapper receivers + splunk_hecout), so existing callers
// are unaffected.
func NewFactory(opts ...Option) confmap.ProviderFactory {
	reg := newRegistry()
	for _, opt := range opts {
		opt(reg)
	}
	return confmap.NewProviderFactory(func(confmap.ProviderSettings) confmap.Provider {
		return &provider{reg: reg}
	})
}

func (*provider) Scheme() string { return schemeName }

func (*provider) Shutdown(context.Context) error { return nil }

// Retrieve reads the .conf tree and returns the emitted pipeline fragment.
// The watcher is intentionally ignored (Retrieve-only), matching the built-in
// file provider.
func (p *provider) Retrieve(_ context.Context, uri string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
	splunkHome, pipeline, err := parseURI(uri)
	if err != nil {
		return nil, err
	}

	frag, err := p.build(splunkHome, pipeline)
	if err != nil {
		return nil, err
	}
	return confmap.NewRetrieved(frag)
}

func parseURI(uri string) (splunkHome, pipeline string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", "", fmt.Errorf("splunkhome: invalid uri %q: %w", uri, err)
	}
	if u.Scheme != schemeName {
		return "", "", fmt.Errorf("splunkhome: unexpected scheme %q", u.Scheme)
	}
	// SPLUNK_HOME is the opaque path: host + path (splunkhome:///opt/splunk ->
	// host="", path="/opt/splunk"; splunkhome://opt/splunk -> host="opt").
	splunkHome = u.Host + u.Path
	if splunkHome == "" {
		return "", "", fmt.Errorf("splunkhome: empty SPLUNK_HOME in uri %q", uri)
	}
	pipeline = u.Query().Get("pipeline")
	if pipeline == "" {
		pipeline = "uf"
	}
	if !pipelineRegexp.MatchString(pipeline) {
		return "", "", fmt.Errorf("splunkhome: invalid pipeline %q (must match %s)", pipeline, pipelineRegexp.String())
	}
	return splunkHome, pipeline, nil
}

// build reads inputs.conf, outputs.conf, props.conf, and transforms.conf across the Splunk
// conf search path and assembles the confmap fragment with all receiver configs wired with
// props and transforms.
func (p *provider) build(splunkHome, pipeline string) (map[string]any, error) {
	dirs := tabuilder.ConfDirs(splunkHome)

	inputs, err := tabuilder.ReadInputs(dirs)
	if err != nil {
		return nil, fmt.Errorf("splunkhome: read inputs.conf: %w", err)
	}

	// The full merged props/transforms set is attached to every emitted
	// receiver, matching how splunk_inputs feeds the same set to each
	// sub-receiver. Per-source matching stays receiver-internal.
	props, err := tabuilder.ReadProps(dirs)
	if err != nil {
		return nil, fmt.Errorf("splunkhome: read props.conf: %w", err)
	}
	transforms, err := tabuilder.ReadTransforms(dirs)
	if err != nil {
		return nil, fmt.Errorf("splunkhome: read transforms.conf: %w", err)
	}

	recvs, skipped, err := mapInputs(inputs, props, transforms)
	if err != nil {
		return nil, err
	}
	if len(recvs) == 0 {
		return nil, fmt.Errorf("splunkhome: no supported input stanzas found under %s (skipped: %v)", splunkHome, skipped)
	}

	merged, err := tabuilder.ReadOutputs(splunkHome)
	if err != nil {
		return nil, fmt.Errorf("splunkhome: read outputs.conf: %w", err)
	}
	// No exporter is hardcoded: mapOutputs emits one exporter per output stanza
	// it recognizes and fails with a specific error for an unsupported stanza
	// kind or an outputs.conf with no output stanzas at all.
	exps, err := mapOutputs(p.reg.outputs, merged)
	if err != nil {
		return nil, fmt.Errorf("splunkhome: outputs.conf under %s: %w", splunkHome, err)
	}

	receivers := map[string]any{}
	var recvIDs []string
	for _, e := range recvs {
		receivers[e.ID] = e.Cfg
		recvIDs = append(recvIDs, e.ID)
	}
	sort.Strings(recvIDs)

	exporters := map[string]any{}
	var expIDs []string
	for _, e := range exps {
		exporters[e.ID] = e.Cfg
		expIDs = append(expIDs, e.ID)
	}
	sort.Strings(expIDs)

	pipelineID := "logs/" + pipeline
	frag := map[string]any{
		"receivers": receivers,
		"exporters": exporters,
		"service": map[string]any{
			"pipelines": map[string]any{
				pipelineID: map[string]any{
					"receivers": toAnySlice(recvIDs),
					"exporters": toAnySlice(expIDs),
				},
			},
		},
	}
	return frag, nil
}

func toAnySlice(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}
