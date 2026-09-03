// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package baseline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
)

// TestNewBaselineBuilds verifies the generated baseline assembles cleanly and is
// non-empty across all component kinds.
func TestNewBaselineBuilds(t *testing.T) {
	factories, err := NewBaseline().Build()
	require.NoError(t, err)

	assert.NotEmpty(t, factories.Receivers)
	assert.NotEmpty(t, factories.Processors)
	assert.NotEmpty(t, factories.Exporters)
	assert.NotEmpty(t, factories.Extensions)
	assert.NotEmpty(t, factories.Connectors)
	assert.NotNil(t, factories.Telemetry)
}

// TestBaselineIsUpstreamOnly asserts the baseline carries no Splunk-specific
// components. Those are each flavor's delta, layered on top by the consumer.
func TestBaselineIsUpstreamOnly(t *testing.T) {
	factories, err := NewBaseline().Build()
	require.NoError(t, err)

	for _, splunk := range []string{
		"smartagent",
		"splunk_inputs",
		"discovery",
		"gnmi",
		"lightprometheus",
		"signalfxgatewayprometheusremotewrite",
	} {
		typ := component.MustNewType(splunk)
		assert.NotContains(t, factories.Receivers, typ, splunk)
		assert.NotContains(t, factories.Extensions, typ, splunk)
	}
}

// TestFlavorContribution demonstrates the composition seam: a flavor appends
// its factory to a slice and Build includes it without touching the baseline.
func TestFlavorContribution(t *testing.T) {
	exampleType := component.MustNewType("example_contrib_exporter")
	newExample := func() exporter.Factory {
		return exporter.NewFactory(exampleType, func() component.Config { return &struct{}{} })
	}

	baselineFactories, err := NewBaseline().Build()
	require.NoError(t, err)

	b := NewBaseline()
	b.AddExporters(newExample())
	withContrib, err := b.Build()
	require.NoError(t, err)

	require.NotContains(t, baselineFactories.Exporters, exampleType)
	assert.Contains(t, withContrib.Exporters, exampleType)
	assert.Len(t, withContrib.Exporters, len(baselineFactories.Exporters)+1)
}

// TestBuildReportsDuplicates verifies a colliding component type surfaces an
// error from MakeFactoryMap rather than silently overwriting.
func TestBuildReportsDuplicates(t *testing.T) {
	exampleType := component.MustNewType("example_contrib_exporter")
	mk := func() exporter.Factory {
		return exporter.NewFactory(exampleType, func() component.Config { return &struct{}{} })
	}
	b := NewBaseline()
	b.AddExporters(mk(), mk())
	_, err := b.Build()
	assert.Error(t, err)
}
