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
	"go.opentelemetry.io/collector/consumer/pdata"
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
	me := mockMetadataClient{}

	output := NewOutput(Config{}, &me, componenttest.NewNopHost(), zap.NewNop())

	dim := types.Dimension{
		Name:  "my_dimension",
		Value: "my_dimension_value",
		Properties: map[string]string{
			"property": "property_value",
		},
	}
	output.SendDimensionUpdate(&dim)
	received := me.receivedMetadataUpdates
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
	me := mockMetadataClient{name: "mockmetadataexporter"}
	output := NewOutput(
		Config{
			DimensionClients: []string{"mockmetadataexporter", "exampleexporter", "metricsreceiver", "notareceiver", "notreal"},
		}, &componenttest.ExampleExporterConsumer{},
		&hostWithExporters{exporter: &me},
		zap.NewNop(),
	)

	dim := types.Dimension{
		Name: "error",
	}
	output.SendDimensionUpdate(&dim)
	received := me.receivedMetadataUpdates
	require.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "has_errored", update.ResourceIDKey)
}

func TestSendDimensionUpdateFromNextConsumerMetadataExporters(t *testing.T) {
	me := mockMetadataClient{}
	output := NewOutput(Config{}, &me, componenttest.NewNopHost(), zap.NewNop())

	dim := types.Dimension{
		Name: "error",
	}
	output.SendDimensionUpdate(&dim)
	received := me.receivedMetadataUpdates
	require.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "has_errored", update.ResourceIDKey)
}

func TestSendEvent(t *testing.T) {
	me := mockMetadataClient{}

	output := NewOutput(Config{}, &me, componenttest.NewNopHost(), zap.NewNop())

	event := event.Event{
		EventType: "my_event",
		Properties: map[string]interface{}{
			"property": "property_value",
		},
	}
	output.SendEvent(&event)
	received := me.receivedLogs
	require.Equal(t, 1, len(received))
	log := received[0]
	logRecord := log.ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0)
	assert.Equal(t, "my_event", logRecord.Name())
	attributes := logRecord.Attributes()
	eventProperties, ok := attributes.Get("com.splunk.signalfx.event_properties")
	require.True(t, ok)
	val, ok := eventProperties.MapVal().Get("property")
	require.True(t, ok)
	assert.Equal(t, "property_value", val.StringVal())
}

func TestSendEventWithInvalidExporter(t *testing.T) {
	output := NewOutput(Config{}, &metricsReceiver{}, componenttest.NewNopHost(), zap.NewNop())
	event := event.Event{EventType: "error"}

	// doesn't panic
	output.SendEvent(&event)
}

func TestSendEventWithoutMetadataClients(t *testing.T) {
	output := NewOutput(Config{
		EventClients: []string{},
	}, consumertest.NewMetricsNop(),
		componenttest.NewNopHost(),
		zap.NewNop(),
	)
	// doesn't panic
	output.SendEvent(&event.Event{})
}

func TestSendEventFromConfigMetadataExporters(t *testing.T) {
	me := mockMetadataClient{name: "mockmetadataexporter"}
	output := NewOutput(
		Config{
			EventClients: []string{"mockmetadataexporter", "exampleexporter", "notareceiver", "notreal"},
		}, &componenttest.ExampleExporterConsumer{},
		&hostWithExporters{exporter: &me},
		zap.NewNop(),
	)

	event := event.Event{
		EventType: "error",
	}
	output.SendEvent(&event)
	received := me.receivedLogs
	require.Equal(t, 1, len(received))
	log := received[0]
	logRecord := log.ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0)
	assert.Equal(t, "has_errored", logRecord.Name())
}

func TestDimensionClientDefaultsToSFxExporter(t *testing.T) {
	me := mockMetadataClient{name: "signalfx"}
	output := NewOutput(
		Config{DimensionClients: nil}, &componenttest.ExampleExporterConsumer{},
		&hostWithExporters{exporter: &me},
		zap.NewNop(),
	)

	dim := types.Dimension{
		Name: "some_dimension",
	}
	output.SendDimensionUpdate(&dim)
	received := me.receivedMetadataUpdates
	require.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "some_dimension", update.ResourceIDKey)
}

func TestDimensionClientDefaultsRequiresLoneSFxExporter(t *testing.T) {
	me := mockMetadataClient{name: "signalfx"}
	output := NewOutput(
		Config{DimensionClients: nil}, &componenttest.ExampleExporterConsumer{},
		&hostWithTwoSFxExporters{sfxExporter: &me},
		zap.NewNop(),
	)

	dim := types.Dimension{
		Name: "some_dimension",
	}
	output.SendDimensionUpdate(&dim)
	received := me.receivedMetadataUpdates
	require.Zero(t, len(received))
}

func TestEventClientDefaultsToSFxExporter(t *testing.T) {
	me := mockMetadataClient{name: "signalfx"}
	output := NewOutput(
		Config{EventClients: nil}, &metricsReceiver{},
		&hostWithExporters{exporter: &me},
		zap.NewNop(),
	)

	event := event.Event{
		EventType: "my_event",
	}
	output.SendEvent(&event)
	received := me.receivedLogs
	require.Equal(t, 1, len(received))
	log := received[0]
	logRecord := log.ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0)
	assert.Equal(t, "my_event", logRecord.Name())
}

func TestEventClientDefaultsRequiresLoneSFxExporter(t *testing.T) {
	me := mockMetadataClient{name: "signalfx"}
	output := NewOutput(
		Config{EventClients: nil}, &metricsReceiver{},
		&hostWithTwoSFxExporters{sfxExporter: &me},
		zap.NewNop(),
	)

	event := event.Event{
		EventType: "my_event",
	}
	output.SendEvent(&event)
	received := me.receivedLogs
	require.Zero(t, len(received))
}

type mockMetadataClient struct {
	name                    string
	receivedMetadataUpdates []*metadata.MetadataUpdate
	receivedLogs            []pdata.Logs
	componenttest.ExampleExporterConsumer
}

func (me *mockMetadataClient) ConsumeMetadata(updates []*metadata.MetadataUpdate) error {
	me.receivedMetadataUpdates = append(me.receivedMetadataUpdates, updates...)

	if updates[0].ResourceIDKey == "error" {
		updates[0].ResourceIDKey = "has_errored"
		return fmt.Errorf("some error")
	}
	return nil
}

func (me *mockMetadataClient) ConsumeLogs(ctx context.Context, logs pdata.Logs) error {
	me.receivedLogs = append(me.receivedLogs, logs)

	logRecord := logs.ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0)
	if logRecord.Name() == "error" {
		logRecord.SetName("has_errored")
		return fmt.Errorf("some error")
	}
	return nil
}

type notAReceiver struct{ component.Component }

type metricsReceiver struct{ component.Component }

func (mr *metricsReceiver) ConsumeMetrics(context.Context, pdata.Metrics) error { return nil }

type hostWithExporters struct {
	exporter *mockMetadataClient
	componenttest.NopHost
}

func getExporters() map[configmodels.DataType]map[configmodels.Exporter]component.Exporter {
	exporters := map[configmodels.DataType]map[configmodels.Exporter]component.Exporter{}
	logExporterMap := map[configmodels.Exporter]component.Exporter{}
	exporters[configmodels.LogsDataType] = logExporterMap

	metricExporterMap := map[configmodels.Exporter]component.Exporter{}
	exporters[configmodels.MetricsDataType] = metricExporterMap

	exampleExporterFactory := componenttest.ExampleExporterFactory{}
	exampleExporter, _ := exampleExporterFactory.CreateMetricsExporter(
		context.Background(), component.ExporterCreateParams{}, nil,
	)
	metricExporterMap[exampleExporterFactory.CreateDefaultConfig()] = exampleExporter

	receiver := namedEntity{name: "metricsreceiver"}
	metricExporterMap[&receiver] = &metricsReceiver{}

	notReceiver := namedEntity{name: "notareceiver"}
	metricExporterMap[&notReceiver] = &notAReceiver{}

	return exporters
}

func (h *hostWithExporters) GetExporters() map[configmodels.DataType]map[configmodels.Exporter]component.Exporter {
	exporters := getExporters()
	exporterMap := exporters[configmodels.MetricsDataType]

	me := namedEntity{name: h.exporter.name, _type: h.exporter.name}
	exporterMap[&me] = component.MetricsExporter(h.exporter)

	exporterMap = exporters[configmodels.LogsDataType]
	exporterMap[&me] = component.LogsExporter(h.exporter)
	return exporters
}

type hostWithTwoSFxExporters struct {
	sfxExporter *mockMetadataClient
	componenttest.NopHost
}

func (h *hostWithTwoSFxExporters) GetExporters() map[configmodels.DataType]map[configmodels.Exporter]component.Exporter {
	exporters := getExporters()
	exporterMap := exporters[configmodels.MetricsDataType]

	meOne := namedEntity{name: "sfx1", _type: "signalfx"}
	exporterMap[&meOne] = component.MetricsExporter(h.sfxExporter)

	meTwo := namedEntity{name: "sfx2", _type: "signalfx"}
	exporterMap[&meTwo] = component.MetricsExporter(h.sfxExporter)

	exporterMap = exporters[configmodels.LogsDataType]
	exporterMap[&meOne] = component.LogsExporter(h.sfxExporter)
	exporterMap[&meTwo] = component.LogsExporter(h.sfxExporter)

	return exporters
}

type namedEntity struct {
	name  string
	_type string
}

func (ne *namedEntity) Type() configmodels.Type { return configmodels.Type(ne._type) }
func (ne *namedEntity) Name() string            { return ne.name }
func (ne *namedEntity) SetName(_ string)        {}
