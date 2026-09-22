// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkhome is a confmap.Provider that reads a Splunk $SPLUNK_HOME
// .conf tree and emits an in-memory pipeline (default name logs/uf) built from
// the input and output stanzas. It is Retrieve-only: it does not watch files.
// Reload rides the collector's existing trigger (SIGHUP or an OpAMP-pushed
// config), consistent with UF's pull/triggered model.
//
// URI form: splunkhome://<SPLUNK_HOME>?prefix=<name>
// The prefix (default "uf") names the emitted pipeline (logs/<prefix>) and the
// component name suffix. Enable/disable of the whole pipeline is done at the
// launch level by including or omitting the --config=splunkhome://... argument.
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

// prefixRegexp restricts the pipeline/component name suffix to a safe segment.
// It must be a legal component-name part and yield readable IDs.
var prefixRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

type provider struct{}

// NewFactory returns a confmap.ProviderFactory for the splunkhome scheme.
func NewFactory() confmap.ProviderFactory {
	return confmap.NewProviderFactory(func(confmap.ProviderSettings) confmap.Provider {
		return &provider{}
	})
}

func (*provider) Scheme() string { return schemeName }

func (*provider) Shutdown(context.Context) error { return nil }

// Retrieve reads the .conf tree and returns the emitted pipeline fragment.
// The watcher is intentionally ignored (Retrieve-only), matching the built-in
// file provider.
func (p *provider) Retrieve(_ context.Context, uri string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
	splunkHome, prefix, err := parseURI(uri)
	if err != nil {
		return nil, err
	}

	frag, err := p.build(splunkHome, prefix)
	if err != nil {
		return nil, err
	}
	return confmap.NewRetrieved(frag)
}

func parseURI(uri string) (splunkHome, prefix string, err error) {
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
	prefix = u.Query().Get("prefix")
	if prefix == "" {
		prefix = "uf"
	}
	if !prefixRegexp.MatchString(prefix) {
		return "", "", fmt.Errorf("splunkhome: invalid prefix %q (must match %s)", prefix, prefixRegexp.String())
	}
	return splunkHome, prefix, nil
}

// build reads inputs.conf, outputs.conf, props.conf, and transforms.conf across the Splunk
// conf search path and assembles the confmap fragment with all receiver configs wired with
// props and transforms.
func (p *provider) build(splunkHome, prefix string) (map[string]any, error) {
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

	recvs, skipped, err := mapInputs(prefix, inputs, props, transforms)
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
	exps, err := mapOutputs(prefix, merged)
	if err != nil {
		return nil, fmt.Errorf("splunkhome: outputs.conf under %s: %w", splunkHome, err)
	}

	receivers := map[string]any{}
	var recvIDs []string
	for _, e := range recvs {
		receivers[e.id] = e.cfg
		recvIDs = append(recvIDs, e.id)
	}
	sort.Strings(recvIDs)

	exporters := map[string]any{}
	var expIDs []string
	for _, e := range exps {
		exporters[e.id] = e.cfg
		expIDs = append(expIDs, e.id)
	}
	sort.Strings(expIDs)

	pipelineID := "logs/" + prefix
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
