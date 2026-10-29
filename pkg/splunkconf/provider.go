// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkconf is a confmap.Provider that reads a Splunk .conf tree and
// emits an in-memory pipeline (default name logs/uf) built from the input and
// output stanzas, with the registered processors between them. It is
// Retrieve-only: it does not watch files. Reload rides the collector's existing
// trigger (SIGHUP or an OpAMP-pushed config), consistent with UF's
// pull/triggered model.
//
// URI form: splunkconf://<config-root>?pipeline=<name>
//
// The config root is the directory holding system/ and apps/*/, read verbatim,
// in any OS-supported path format:
//
//	splunkconf://etc                              relative path
//	splunkconf:///opt/splunk/etc                  absolute path
//	splunkconf://C:\Program Files\Splunk\etc      absolute path with drive letter
//
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
	"strings"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
)

const schemeName = "splunkconf"

// nopReceiverID carries no name segment because there is no stanza to derive one
// from, and its config is empty.
const nopReceiverID = "nop"

// pipelineRegexp restricts the pipeline name to a safe component-name segment.
var pipelineRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

type provider struct {
	logger           *zap.Logger
	inputMappers     []MappingReceiverFactory
	outputMappers    []MappingExporterFactory
	processorMappers []MappingProcessorFactory
}

// NewFactory returns a confmap.ProviderFactory for the splunkconf scheme. Pass
// WithInputMappers to register the receivers that map input stanzas,
// WithOutputMappers the exporters that map output stanzas, and
// WithProcessorMappers the processors that go between them.
func NewFactory(opts ...Option) confmap.ProviderFactory {
	return confmap.NewProviderFactory(func(set confmap.ProviderSettings) confmap.Provider {
		// Only confmap.NewResolver substitutes a nop logger for a nil one, and
		// Create is reachable without it.
		logger := set.Logger
		if logger == nil {
			logger = zap.NewNop()
		}
		p := &provider{logger: logger}
		for _, opt := range opts {
			opt(p)
		}
		return p
	})
}

func (*provider) Scheme() string { return schemeName }

func (*provider) Shutdown(context.Context) error { return nil }

// Retrieve reads the .conf tree and returns the emitted pipeline fragment. The
// watcher is intentionally ignored, matching the built-in file provider.
func (p *provider) Retrieve(_ context.Context, uri string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
	confRoot, pipeline, err := parseURI(uri)
	if err != nil {
		return nil, err
	}
	frag, err := p.build(confRoot, pipeline)
	if err != nil {
		return nil, err
	}
	return confmap.NewRetrieved(escapeDollars(frag))
}

// escapeDollars doubles every $ in the emitted string values: a .conf value may
// hold a literal $, which the collector's config resolver would otherwise read
// as an env var or config source reference, and which it collapses back to a
// single $ once resolution is done.
func escapeDollars(v any) any {
	switch t := v.(type) {
	case string:
		return strings.ReplaceAll(t, "$", "$$")
	case map[string]any:
		for k, e := range t {
			t[k] = escapeDollars(e)
		}
	case []any:
		for i, e := range t {
			t[i] = escapeDollars(e)
		}
	}
	return v
}

// parseURI splits the URI into its config root and pipeline name. The root is
// taken verbatim rather than through url.Parse, matching the built-in file
// provider: a native Windows path is a legal config root, and its backslashes
// and drive letter are not a legal URI authority.
func parseURI(uri string) (confRoot, pipeline string, err error) {
	rest, ok := strings.CutPrefix(uri, schemeName+":")
	if !ok {
		return "", "", fmt.Errorf("splunkconf: unexpected scheme in uri %q", uri)
	}
	rest, rawQuery, _ := strings.Cut(rest, "?")
	confRoot = strings.TrimPrefix(rest, "//")
	if confRoot == "" {
		return "", "", fmt.Errorf("splunkconf: empty config root in uri %q", uri)
	}
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", "", fmt.Errorf("splunkconf: invalid query in uri %q: %w", uri, err)
	}
	pipeline = query.Get("pipeline")
	if pipeline == "" {
		pipeline = "uf"
	}
	if !pipelineRegexp.MatchString(pipeline) {
		return "", "", fmt.Errorf("splunkconf: invalid pipeline %q (must match %s)", pipeline, pipelineRegexp.String())
	}
	return confRoot, pipeline, nil
}

// build reads inputs.conf, outputs.conf, props.conf and transforms.conf across
// the Splunk conf search path and assembles the confmap fragment.
func (p *provider) build(confRoot, pipeline string) (map[string]any, error) {
	dirs := ConfRootDirs(confRoot)

	inputs, schemeDefaults, err := ReadInputs(dirs)
	if err != nil {
		return nil, fmt.Errorf("splunkconf: read inputs.conf: %w", err)
	}

	tree, err := newConf(confRoot, inputs, schemeDefaults)
	if err != nil {
		return nil, err
	}
	recvs, skipped, err := p.mapInputs(tree)
	if err != nil {
		return nil, err
	}
	// An idle agent is a normal state, so a tree with nothing to read still
	// resolves rather than refusing to start.
	if len(recvs) == 0 {
		p.logger.Info("splunkconf: no supported input stanzas, reading from nop",
			zap.String("config_root", confRoot),
			zap.Strings("skipped", skipped))
		recvs = []emitted{{id: nopReceiverID, cfg: map[string]any{}}}
	}

	merged, err := ReadOutputs(dirs)
	if err != nil {
		return nil, fmt.Errorf("splunkconf: read outputs.conf: %w", err)
	}
	exps, skippedOut, err := p.mapOutputs(merged)
	if err != nil {
		return nil, fmt.Errorf("splunkconf: outputs.conf under %s: %w", confRoot, err)
	}
	if len(exps) == 0 {
		return nil, fmt.Errorf("splunkconf: no supported output stanzas found under %s (skipped: %v)", confRoot, skippedOut)
	}

	procs, err := p.mapProcessors(tree)
	if err != nil {
		return nil, err
	}

	receivers := map[string]any{}
	recvIDs := make([]string, 0, len(recvs))
	for _, e := range recvs {
		receivers[e.id] = e.cfg
		recvIDs = append(recvIDs, e.id)
	}
	sort.Strings(recvIDs)

	exporters := map[string]any{}
	expIDs := make([]string, 0, len(exps))
	for _, e := range exps {
		exporters[e.id] = e.cfg
		expIDs = append(expIDs, e.id)
	}
	sort.Strings(expIDs)

	// Processors keep the order mapProcessors returned them in, which is
	// registration order: a pipeline's processors run in sequence, so sorting
	// them the way receivers and exporters are sorted would change behavior.
	processors := map[string]any{}
	procIDs := make([]string, 0, len(procs))
	for _, e := range procs {
		processors[e.id] = e.cfg
		procIDs = append(procIDs, e.id)
	}

	logs := map[string]any{
		"receivers": toAnySlice(recvIDs),
		"exporters": toAnySlice(expIDs),
	}
	frag := map[string]any{
		"receivers": receivers,
		"exporters": exporters,
		"service": map[string]any{
			"pipelines": map[string]any{
				"logs/" + pipeline: logs,
			},
		},
	}
	// Both keys are left out entirely when nothing mapped, rather than emitted
	// empty: a provider with no processor mappers has to produce the fragment it
	// did before there were any.
	if len(procIDs) > 0 {
		frag["processors"] = processors
		logs["processors"] = toAnySlice(procIDs)
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
