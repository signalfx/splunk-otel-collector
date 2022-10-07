// Copyright OpenTelemetry Authors
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
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestOutput(t *testing.T) {
	rcs := component.ReceiverCreateSettings{}
	rcs.Logger = zap.NewNop()
	output := newOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(),
		consumertest.NewNop(), consumertest.NewNop(),
		componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	output.AddDatapointExclusionFilter(dpfilters.DatapointFilter(nil))
	assert.False(t, output.HasAnyExtraMetrics())
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

func TestHasEnabledMetric(t *testing.T) {
	monitorFiltering, err := newMonitorFiltering(&saconfig.MonitorConfig{}, &monitors.Metadata{
		DefaultMetrics: utils.StringSet("mem.used"),
		Metrics: map[string]monitors.MetricInfo{
			"mem.used": {Type: datapoint.Counter, Group: "mem"},
			"mem.free": {Type: datapoint.Counter, Group: "mem"},
		},
		Groups: utils.StringSet("mem"),
	}, zap.NewNop())
	require.NoError(t, err)
	output := newOutput(
		Config{}, monitorFiltering, consumertest.NewNop(), consumertest.NewNop(),
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	assert.Equal(t, []string{"mem.used"}, output.EnabledMetrics())

	// Empty metadata
	monitorFiltering, err = newMonitorFiltering(&saconfig.MonitorConfig{}, nil, zap.NewNop())
	require.NoError(t, err)
	output = newOutput(
		Config{}, monitorFiltering, consumertest.NewNop(), consumertest.NewNop(),
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	assert.Empty(t, output.EnabledMetrics())
}

func TestHasEnabledMetricInGroup(t *testing.T) {
	monitorFiltering, err := newMonitorFiltering(&saconfig.MonitorConfig{}, &monitors.Metadata{
		DefaultMetrics: utils.StringSet("mem.used"),
		Metrics: map[string]monitors.MetricInfo{
			"cpu.min":  {Type: datapoint.Gauge, Group: "cpu"},
			"cpu.max":  {Type: datapoint.Gauge, Group: "cpu"},
			"mem.used": {Type: datapoint.Counter, Group: "mem"},
			"mem.free": {Type: datapoint.Counter, Group: "mem"},
		},
		Groups: utils.StringSet("mem"),
		GroupMetricsMap: map[string][]string{
			"cpu": {"cpu.min", "cpu.max"},
			"mem": {"mem.free", "mem.used"},
		},
	}, zap.NewNop())
	require.NoError(t, err)
	output := newOutput(
		Config{}, monitorFiltering, consumertest.NewNop(), consumertest.NewNop(),
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	assert.True(t, output.HasEnabledMetricInGroup("mem"))
	assert.False(t, output.HasEnabledMetricInGroup("cpu"))

	// Empty metadata
	monitorFiltering, err = newMonitorFiltering(&saconfig.MonitorConfig{}, nil, zap.NewNop())
	require.NoError(t, err)
	output = newOutput(
		Config{}, monitorFiltering, consumertest.NewNop(), consumertest.NewNop(),
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	assert.False(t, output.HasEnabledMetricInGroup("any"))
}

func TestExtraDimensions(t *testing.T) {
	o := newOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), consumertest.NewNop(),
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	assert.Empty(t, o.extraDimensions)

	o.RemoveExtraDimension("not_a_known_dimension_name")

	o.AddExtraDimension("a_dimension_name", "a_value")
	assert.Equal(t, "a_value", o.extraDimensions["a_dimension_name"])

	cp, ok := o.Copy().(*output)
	require.True(t, ok)
	assert.Equal(t, "a_value", cp.extraDimensions["a_dimension_name"])

	cp.RemoveExtraDimension("a_dimension_name")
	assert.Empty(t, cp.extraDimensions["a_dimension_name"])
	assert.Equal(t, "a_value", o.extraDimensions["a_dimension_name"])

	cp.AddExtraDimension("another_dimension_name", "another_dimension_value")
	assert.Equal(t, "another_dimension_value", cp.extraDimensions["another_dimension_name"])
	assert.Empty(t, o.extraDimensions["another_dimension_name"])
}

func TestSendDimensionUpdate(t *testing.T) {
	mmc := mockMetadataClient{id: config.NewComponentID("signalfx")}
	output := newOutput(
		Config{}, fakeMonitorFiltering(), &mmc, consumertest.NewNop(), consumertest.NewNop(),
		componenttest.NewNopHost(), newReceiverCreateSettings(),
	)

	dim := types.Dimension{
		Name:  "my_dimension",
		Value: "my_dimension_value",
		Properties: map[string]string{
			"property": "property_value",
		},
	}
	output.SendDimensionUpdate(&dim)
	received := mmc.receivedMetadataUpdates
	assert.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "my_dimension", update.ResourceIDKey)
	assert.Equal(t, metadata.ResourceID("my_dimension_value"), update.ResourceID)
	assert.Equal(t, map[string]string{"property": "property_value"}, update.MetadataToUpdate)
}

func TestSendDimensionUpdateWithInvalidExporter(t *testing.T) {
	output := newOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), consumertest.NewNop(),
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)
	dim := types.Dimension{Name: "error"}

	// doesn't panic
	output.SendDimensionUpdate(&dim)
}

func TestSendDimensionUpdateFromConfigMetadataExporters(t *testing.T) {
	mmc := mockMetadataClient{id: config.NewComponentID("mockmetadataexporter")}
	output := newOutput(
		Config{
			DimensionClients: []string{"mockmetadataexporter", "exampleexporter", "metricsreceiver", "notareceiver", "notreal"},
		},
		fakeMonitorFiltering(),
		consumertest.NewNop(),
		consumertest.NewNop(),
		consumertest.NewNop(),
		&hostWithExporters{exporter: &mmc},
		newReceiverCreateSettings(),
	)

	dim := types.Dimension{
		Name: "error",
	}
	output.SendDimensionUpdate(&dim)
	received := mmc.receivedMetadataUpdates
	require.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "has_errored", update.ResourceIDKey)
}

func TestSendDimensionUpdateFromNextConsumerMetadataExporters(t *testing.T) {
	mmc := mockMetadataClient{id: config.NewComponentID("signalfx")}
	output := newOutput(
		Config{}, fakeMonitorFiltering(), &mmc, consumertest.NewNop(),
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)

	dim := types.Dimension{
		Name: "error",
	}
	output.SendDimensionUpdate(&dim)
	received := mmc.receivedMetadataUpdates
	require.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "has_errored", update.ResourceIDKey)
}

func TestSendEvent(t *testing.T) {
	mmc := mockMetadataClient{id: config.NewComponentID("signalfx")}
	output := newOutput(
		Config{}, fakeMonitorFiltering(), consumertest.NewNop(), &mmc,
		consumertest.NewNop(), componenttest.NewNopHost(), newReceiverCreateSettings(),
	)

	event := event.Event{
		EventType: "my_event",
		Properties: map[string]any{
			"property": "property_value",
		},
	}
	output.SendEvent(&event)
	received := mmc.receivedLogs
	require.Equal(t, 1, len(received))
	log := received[0]
	logRecord := log.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	attributes := logRecord.Attributes()
	eventType, ok := attributes.Get("com.splunk.signalfx.event_type")
	require.True(t, ok)
	assert.Equal(t, "my_event", eventType.Str())
	eventProperties, ok := attributes.Get("com.splunk.signalfx.event_properties")
	require.True(t, ok)
	val, ok := eventProperties.Map().Get("property")
	require.True(t, ok)
	assert.Equal(t, "property_value", val.Str())
}

func TestDimensionClientDefaultsToSFxExporter(t *testing.T) {
	mmc := mockMetadataClient{id: config.NewComponentID("signalfx")}
	output := newOutput(
		Config{DimensionClients: nil},
		fakeMonitorFiltering(),
		consumertest.NewNop(),
		consumertest.NewNop(),
		consumertest.NewNop(),
		&hostWithExporters{exporter: &mmc},
		newReceiverCreateSettings(),
	)

	dim := types.Dimension{
		Name: "some_dimension",
	}
	output.SendDimensionUpdate(&dim)
	received := mmc.receivedMetadataUpdates
	require.Equal(t, 1, len(received))
	update := *(received[0])
	assert.Equal(t, "some_dimension", update.ResourceIDKey)
}

func TestDimensionClientDefaultsRequiresLoneSFxExporter(t *testing.T) {
	mmc := mockMetadataClient{id: config.NewComponentID("signalfx")}
	output := newOutput(
		Config{DimensionClients: nil},
		fakeMonitorFiltering(),
		consumertest.NewNop(),
		consumertest.NewNop(),
		consumertest.NewNop(),
		&hostWithTwoSFxExporters{sfxExporter: &mmc},
		newReceiverCreateSettings(),
	)

	dim := types.Dimension{
		Name: "some_dimension",
	}
	output.SendDimensionUpdate(&dim)
	received := mmc.receivedMetadataUpdates
	require.Zero(t, len(received))
}

func fakeMonitorFiltering() *monitorFiltering {
	return &monitorFiltering{
		filterSet:       &dpfilters.FilterSet{},
		metadata:        &monitors.Metadata{},
		hasExtraMetrics: false,
	}
}

type mockMetadataClient struct {
	id                      config.ComponentID
	receivedMetadataUpdates []*metadata.MetadataUpdate
	receivedLogs            []plog.Logs
}

func (mmc *mockMetadataClient) Capabilities() consumer.Capabilities {
	panic("implement me")
}

func (mmc *mockMetadataClient) Start(_ context.Context, _ component.Host) error {
	panic("implement me")
}

func (mmc *mockMetadataClient) Shutdown(_ context.Context) error {
	panic("implement me")
}

func (mmc *mockMetadataClient) ConsumeMetrics(_ context.Context, _ pmetric.Metrics) error {
	panic("implement me")
}

func (mmc *mockMetadataClient) ConsumeMetadata(updates []*metadata.MetadataUpdate) error {
	mmc.receivedMetadataUpdates = append(mmc.receivedMetadataUpdates, updates...)

	if updates[0].ResourceIDKey == "error" {
		updates[0].ResourceIDKey = "has_errored"
		return fmt.Errorf("some error")
	}
	return nil
}

func (mmc *mockMetadataClient) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	mmc.receivedLogs = append(mmc.receivedLogs, logs)
	return nil
}

type notAReceiver struct{ component.Component }

type mockMetricsReceiver struct{ component.Component }

func (mr *mockMetricsReceiver) ConsumeMetrics(context.Context, pmetric.Metrics) error { return nil }

type nopHost struct{}

func (h *nopHost) ReportFatalError(_ error) {}
func (h *nopHost) GetFactory(_ component.Kind, _ config.Type) component.Factory {
	return nil
}
func (h *nopHost) GetExtensions() map[config.ComponentID]component.Extension {
	return nil
}

type hostWithExporters struct {
	*nopHost
	exporter *mockMetadataClient
}

func getExporters() map[config.DataType]map[config.ComponentID]component.Exporter {
	exporters := map[config.DataType]map[config.ComponentID]component.Exporter{}
	metricExporterMap := map[config.ComponentID]component.Exporter{}
	exporters[config.MetricsDataType] = metricExporterMap

	exampleExporterFactory := componenttest.NewNopExporterFactory()
	exampleExporter, _ := exampleExporterFactory.CreateMetricsExporter(
		context.Background(), component.ExporterCreateSettings{}, nil,
	)

	metricExporterMap[exampleExporterFactory.CreateDefaultConfig().ID()] = exampleExporter
	metricExporterMap[config.NewComponentID("metricsreceiver")] = &mockMetricsReceiver{}
	metricExporterMap[config.NewComponentID("notareceiver")] = &notAReceiver{}

	return exporters
}

func (h *hostWithExporters) GetExporters() map[config.DataType]map[config.ComponentID]component.Exporter {
	exporters := getExporters()
	exporterMap := exporters[config.MetricsDataType]

	// Add internal exporter to the list.
	exporterMap[h.exporter.id] = component.MetricsExporter(h.exporter)
	return exporters
}

type hostWithTwoSFxExporters struct {
	*nopHost
	sfxExporter *mockMetadataClient
}

func (h *hostWithTwoSFxExporters) GetExporters() map[config.DataType]map[config.ComponentID]component.Exporter {
	exporters := getExporters()
	exporterMap := exporters[config.MetricsDataType]

	meOne := config.NewComponentIDWithName("signalfx", "sfx1")
	exporterMap[meOne] = component.MetricsExporter(h.sfxExporter)

	meTwo := config.NewComponentIDWithName("signalfx", "sfx2")
	exporterMap[meTwo] = component.MetricsExporter(h.sfxExporter)
	return exporters
}
