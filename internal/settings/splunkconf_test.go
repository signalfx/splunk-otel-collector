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
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/configsource"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkconf"
)

// TestSplunkConfPipelineBuilds runs the collector's own build path over a
// splunkconf:// source: otelcol.DryRun resolves and validates it, then runs
// service.Validate and graph.Build, which wires the pipeline's receivers
// through to its exporters. The provider is wrapped the way main.go wraps every
// provider factory, so a .conf value holding a literal $ goes through config
// source resolution as it does in a real collector.
func TestSplunkConfPipelineBuilds(t *testing.T) {
	require.NoError(t, featuregate.GlobalRegistry().Set("enableTARunner", true))
	t.Cleanup(func() {
		require.NoError(t, featuregate.GlobalRegistry().Set("enableTARunner", false))
	})

	// A fixture local to this test, holding only stanzas that build on every
	// platform. The provider package's fixture covers more kinds, including
	// WinEventLog, whose receiver cannot be built off Windows.
	confRoot := filepath.Join("testdata", "splunkconf", "etc")

	col, err := otelcol.NewCollector(otelcol.CollectorSettings{
		BuildInfo: component.NewDefaultBuildInfo(),
		Factories: components.Get,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{"splunkconf://" + confRoot + "?pipeline=uf"},
				ProviderFactories: []confmap.ProviderFactory{
					configsource.New(zap.NewNop(), nil).Wrap(splunkconf.NewFactory()),
				},
			},
		},
	})
	require.NoError(t, err)

	require.NoError(t, col.DryRun(context.Background()),
		"the .conf-derived receivers must build a real graph through to splunk_hec")
}
