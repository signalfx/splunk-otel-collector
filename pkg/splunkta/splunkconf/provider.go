// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkconf is a confmap.Provider that reads a Splunk .conf tree and
// emits an in-memory pipeline (default name logs/uf) built from the input and
// output stanzas. It is Retrieve-only: it does not watch files. Reload rides the
// collector's existing trigger (SIGHUP or an OpAMP-pushed config), consistent
// with UF's pull/triggered model.
//
// URI form: splunkconf://<config-root>?pipeline=<name>
//
// The config root is the directory holding system/ and apps/*/, read verbatim.
// The provider knows nothing about $SPLUNK_HOME, $SPLUNK_ETC, or any install
// layout: the caller resolves those and passes a plain path, so the same URI
// always resolves the same way regardless of ambient env, and a relocated tree
// needs no special handling (splunkconf:///etc/splunk-config).
//
// The pipeline (default "uf") names the emitted pipeline (logs/<pipeline>).
// Enable/disable of the whole pipeline is done at the launch level by including
// or omitting the --config=splunkconf://... argument.
package splunkconf

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"

	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

const schemeName = "splunkconf"

// pipelineRegexp restricts the pipeline name to a safe component-name segment.
var pipelineRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

type provider struct {
	reg *registry
}

// NewFactory returns a confmap.ProviderFactory for the splunkconf scheme.
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
	confRoot, pipeline, err := parseURI(uri)
	if err != nil {
		return nil, err
	}

	frag, err := p.build(confRoot, pipeline)
	if err != nil {
		return nil, err
	}
	return confmap.NewRetrieved(frag)
}

func parseURI(uri string) (confRoot, pipeline string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", "", fmt.Errorf("splunkconf: invalid uri %q: %w", uri, err)
	}
	if u.Scheme != schemeName {
		return "", "", fmt.Errorf("splunkconf: unexpected scheme %q", u.Scheme)
	}
	// The config root is the opaque path: host + path (splunkconf:///opt/splunk/etc
	// -> host="", path="/opt/splunk/etc"; splunkconf://etc -> host="etc").
	confRoot = u.Host + u.Path
	if confRoot == "" {
		return "", "", fmt.Errorf("splunkconf: empty config root in uri %q", uri)
	}
	pipeline = u.Query().Get("pipeline")
	if pipeline == "" {
		pipeline = "uf"
	}
	if !pipelineRegexp.MatchString(pipeline) {
		return "", "", fmt.Errorf("splunkconf: invalid pipeline %q (must match %s)", pipeline, pipelineRegexp.String())
	}
	return confRoot, pipeline, nil
}

// build reads inputs.conf, outputs.conf, props.conf, and transforms.conf across the Splunk
// conf search path and assembles the confmap fragment with all receiver configs wired with
// props and transforms.
func (p *provider) build(confRoot, pipeline string) (map[string]any, error) {
	dirs := tabuilder.ConfRootDirs(confRoot)

	inputs, err := tabuilder.ReadInputs(dirs)
	if err != nil {
		return nil, fmt.Errorf("splunkconf: read inputs.conf: %w", err)
	}

	// The full merged props/transforms set is attached to every emitted
	// receiver, matching how splunk_inputs feeds the same set to each
	// sub-receiver. Per-source matching stays receiver-internal.
	props, err := tabuilder.ReadProps(dirs)
	if err != nil {
		return nil, fmt.Errorf("splunkconf: read props.conf: %w", err)
	}
	transforms, err := tabuilder.ReadTransforms(dirs)
	if err != nil {
		return nil, fmt.Errorf("splunkconf: read transforms.conf: %w", err)
	}

	recvs, skipped, err := mapInputs(inputs, props, transforms)
	if err != nil {
		return nil, err
	}
	if len(recvs) == 0 {
		return nil, fmt.Errorf("splunkconf: no supported input stanzas found under %s (skipped: %v)", confRoot, skipped)
	}

	merged, err := tabuilder.ReadOutputsFromDirs(dirs)
	if err != nil {
		return nil, fmt.Errorf("splunkconf: read outputs.conf: %w", err)
	}
	// No exporter is hardcoded: mapOutputs emits one exporter per output stanza
	// it recognizes and fails with a specific error for an unsupported stanza
	// kind or an outputs.conf with no output stanzas at all.
	exps, err := mapOutputs(p.reg.outputs, merged)
	if err != nil {
		return nil, fmt.Errorf("splunkconf: outputs.conf under %s: %w", confRoot, err)
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
