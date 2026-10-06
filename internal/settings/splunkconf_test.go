// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkconf"
)

// TestSplunkConfPipelineBuilds runs the collector's own build path over a merged
// config: otelcol.DryRun resolves base.yaml plus a splunkconf:// source into one
// otelcol.Config, validates the whole thing, then runs service.Validate and
// graph.Build, which actually wires each pipeline's receivers through to its
// exporters.
//
// This lives in the root module on purpose. otelcol and service pull in the otel
// resource detectors and with them k8s, prometheus and gopsutil; the root module
// already depends on all of it, while pkg/splunkta is a leaf library that should
// not. Running it here also exercises the real registered component set and the
// enableTARunner gate rather than a hand-built factory map.
func TestSplunkConfPipelineBuilds(t *testing.T) {
	require.NoError(t, featuregate.GlobalRegistry().Set("enableTARunner", true))
	t.Cleanup(func() {
		require.NoError(t, featuregate.GlobalRegistry().Set("enableTARunner", false))
	})

	// A fixture local to this test, holding only stanzas that build on every
	// platform. The provider package's fixture covers more kinds, including
	// WinEventLog, whose receiver cannot be built off Windows.
	confRoot := filepath.Join("testdata", "splunkconf", "etc")
	base := "file:" + filepath.Join("testdata", "splunkconf", "base.yaml")

	col, err := otelcol.NewCollector(otelcol.CollectorSettings{
		BuildInfo: component.NewDefaultBuildInfo(),
		Factories: components.Get,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{base, "splunkconf://" + confRoot + "?pipeline=uf"},
				ProviderFactories: []confmap.ProviderFactory{
					fileprovider.NewFactory(),
					splunkconf.NewFactory(),
				},
			},
		},
	})
	require.NoError(t, err)

	require.NoError(t, col.DryRun(context.Background()),
		"the merged config must build a real graph, wiring the .conf-derived receivers through to splunk_hec")
}
