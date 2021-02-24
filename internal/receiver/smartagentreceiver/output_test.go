// Copyright 2021, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentreceiver

import (
	"context"
	"fmt"
	"testing"

	metadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
)

func TestOutput(t *testing.T) {
	output := NewOutput(Config{}, consumertest.NewMetricsNop(), componenttest.NewNopHost(), zap.NewNop())
	output.AddDatapointExclusionFilter(dpfilters.DatapointFilter(nil))
	assert.Empty(t, output.EnabledMetrics())
	assert.True(t, output.HasEnabledMetricInGroup(""))
	assert.True(t, output.HasAnyExtraMetrics())
	assert.NotSame(t, &output, output.Copy())
	output.SendDatapoints()
	output.SendEvent(new(event.Event))
	output.SendSpans()
	output.SendDimensionUpdate(new(types.Dimension))
	output.AddExtraDimension("", "")
	output.RemoveExtraDimension("")
	output.AddExtraSpanTag("", "")
	output.RemoveExtraSpanTag("")
	output.AddDefaultSpanTag("", "")
	output.RemoveDefaultSpanTag("")
}

func TestExtraDimensions(t *testing.T) {
	output := NewOutput(Config{}, consumertest.NewMetricsNop(), componenttest.NewNopHost(), zap.NewNop())
	assert.Empty(t, output.extraDimensions)

	output.RemoveExtraDimension("not_a_known_dimension_name")

	output.AddExtraDimension("a_dimension_name", "a_value")
	assert.Equal(t, "a_value", output.extraDimensions["a_dimension_name"])

	cp, ok := output.Copy().(*Output)
	require.True(t, ok)
	assert.Equal(t, "a_value", cp.extraDimensions["a_dimension_name"])

	cp.RemoveExtraDimension("a_dimension_name")
	assert.Empty(t, cp.extraDimensions["a_dimension_name"])
	assert.Equal(t, "a_value", output.extraDimensions["a_dimension_name"])

	cp.AddExtraDimension("another_dimension_name", "another_dimension_value")
	assert.Equal(t, "another_dimension_value", cp.extraDimensions["another_dimension_name"])
	assert.Empty(t, output.extraDimensions["another_dimension_name"])
}

func TestSendDimensionUpdate(t *testing.T) {
	me := mockMetadataExporter{}

	output := NewOutput(Config{}, &me, componenttest.NewNopHost(), zap.NewNop())

	dim := types.Dimension{
		Name:  "my_dimension",
		Value: "my_dimension_value",
		Properties: map[string]string{
			"property": "property_value",
		},
	}
	output.SendDimensionUpdate(&dim)
	received := me.received
	assert.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "my_dimension", update.ResourceIDKey)
	assert.Equal(t, metadata.ResourceID("my_dimension_value"), update.ResourceID)
	assert.Equal(t, map[string]string{"property": "property_value"}, update.MetadataToUpdate)
}

func TestSendDimensionUpdateWithInvalidExporter(t *testing.T) {
	output := NewOutput(Config{}, consumertest.NewMetricsNop(), componenttest.NewNopHost(), zap.NewNop())
	dim := types.Dimension{Name: "error"}

	// doesn't panic
	output.SendDimensionUpdate(&dim)
}

func TestSendDimensionUpdateFromConfigMetadataExporters(t *testing.T) {
	me := mockMetadataExporter{name: "mockmetadataexporter"}
	output := NewOutput(
		Config{
			MetadataClients: &[]string{"mockmetadataexporter", "exampleexporter"},
		}, &componenttest.ExampleExporterConsumer{},
		&hostWithExporters{exporter: &me},
		zap.NewNop(),
	)

	dim := types.Dimension{
		Name: "error",
	}
	output.SendDimensionUpdate(&dim)
	received := me.received
	require.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "has_errored", update.ResourceIDKey)
}

func TestSendDimensionUpdateFromNextConsumerMetadataExporters(t *testing.T) {
	me := mockMetadataExporter{}
	output := NewOutput(Config{}, &me, componenttest.NewNopHost(), zap.NewNop())

	dim := types.Dimension{
		Name: "error",
	}
	output.SendDimensionUpdate(&dim)
	received := me.received
	require.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "has_errored", update.ResourceIDKey)
}

type mockMetadataExporter struct {
	name     string
	received []*metadata.MetadataUpdate
	componenttest.ExampleExporterConsumer
}

func (me *mockMetadataExporter) ConsumeMetadata(updates []*metadata.MetadataUpdate) error {
	me.received = append(me.received, updates...)

	if updates[0].ResourceIDKey == "error" {
		updates[0].ResourceIDKey = "has_errored"
		return fmt.Errorf("some error")
	}
	return nil
}

type hostWithExporters struct {
	exporter *mockMetadataExporter
	componenttest.NopHost
}

func (h *hostWithExporters) GetExporters() map[configmodels.DataType]map[configmodels.Exporter]component.Exporter {
	exporters := map[configmodels.DataType]map[configmodels.Exporter]component.Exporter{}
	exporterMap := map[configmodels.Exporter]component.Exporter{}
	exporters[configmodels.MetricsDataType] = exporterMap

	me := namedEntity{name: h.exporter.name}
	exporterMap[&me] = component.MetricsExporter(h.exporter)

	exampleExporterFactory := componenttest.ExampleExporterFactory{}
	exampleExporter, _ := exampleExporterFactory.CreateMetricsExporter(
		context.Background(), component.ExporterCreateParams{}, nil,
	)
	exporterMap[exampleExporterFactory.CreateDefaultConfig()] = exampleExporter

	return exporters
}

type namedEntity struct {
	name string
}

func (ne *namedEntity) Type() configmodels.Type { return configmodels.Type(ne.name) }
func (ne *namedEntity) Name() string            { return ne.name }
func (ne *namedEntity) SetName(_ string)        {}
