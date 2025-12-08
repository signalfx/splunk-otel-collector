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

	metadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/signalfx/golib/v3/event"     //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/signalfx/golib/v3/trace"     //nolint:staticcheck // SA1019: deprecated package still in use
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pipeline"
	otelcolreceiver "go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.uber.org/zap"

	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"

	"github.com/signalfx/splunk-otel-collector/pkg/receiver/smartagentreceiver/converter"
)

const internalTransport = "internal"

// output is an implementation of a Smart Agent FilteringOutput that receives datapoints, events, and dimension updates
// from a configured monitor.  It will forward all datapoints to the nextMetricsConsumer, all dimension updates to the
// nextDimensionClients as determined by the associated items in Config.MetadataClients, and all events to the
// nextLogsConsumer.
type output struct {
	nextMetricsConsumer  consumer.Metrics
	nextLogsConsumer     consumer.Logs
	nextTracesConsumer   consumer.Traces
	extraDimensions      map[string]string
	logger               *zap.Logger
	reporter             *receiverhelper.ObsReport
	translator           converter.Translator
	monitorFiltering     *monitorFiltering
	receiverID           component.ID
	nextDimensionClients []metadata.MetadataExporter
}

var (
	_ types.Output          = (*output)(nil)
	_ types.FilteringOutput = (*output)(nil)
)

// Deprecated: This is a temporary workaround for the following issue:
// https://github.com/open-telemetry/opentelemetry-collector/issues/7370
type getExporters interface {
	GetExporters() map[pipeline.Signal]map[component.ID]component.Component
}

func newOutput(
	config Config, filtering *monitorFiltering, nextMetricsConsumer consumer.Metrics,
	nextLogsConsumer consumer.Logs, nextTracesConsumer consumer.Traces, host component.Host,
	params otelcolreceiver.Settings,
) (*output, error) {
	obsReceiver, err := receiverhelper.NewObsReport(receiverhelper.ObsReportSettings{
		ReceiverID:             params.ID,
		Transport:              internalTransport,
		ReceiverCreateSettings: params,
	})
	if err != nil {
		return nil, err
	}
	return &output{
		receiverID:           params.ID,
		nextMetricsConsumer:  nextMetricsConsumer,
		nextLogsConsumer:     nextLogsConsumer,
		nextTracesConsumer:   nextTracesConsumer,
		nextDimensionClients: getMetadataExporters(config, host, nextMetricsConsumer, params.Logger),
		logger:               params.Logger,
		translator:           converter.NewTranslator(params.Logger),
		extraDimensions:      map[string]string{},
		monitorFiltering:     filtering,
		reporter:             obsReceiver,
	}, nil
}

// getMetadataExporters walks through obtained Config.MetadataClients and returns all matching registered MetadataExporters,
// if any.  At this time the SignalFx exporter is the only supported use case and adopter of this type.
func getMetadataExporters(
	cfg Config, host component.Host, nextMetricsConsumer consumer.Metrics, logger *zap.Logger,
) []metadata.MetadataExporter {
	var exporters []metadata.MetadataExporter

	exporters, noClientsSpecified := getDimensionClientsFromMetricsExporters(cfg.DimensionClients, host, nextMetricsConsumer, logger)

	if len(exporters) == 0 && noClientsSpecified {
		sfxExporter := getLoneSFxExporter(host, pipeline.SignalMetrics)
		if sfxExporter != nil {
			if sfx, ok := sfxExporter.(metadata.MetadataExporter); ok {
				exporters = append(exporters, sfx)
			}
		}
	}

	if len(exporters) == 0 {
		logger.Debug("no dimension updates are possible as no valid dimensionClients have been provided and next pipeline component isn't a MetadataExporter")
	}

	return exporters
}

// getDimensionClientsFromMetricsExporters will walk through all provided config.DimensionClients and retrieve matching registered
// MetricsExporters, the only truly supported component type.
// If config.MetadataClients is nil, it will return a slice with nextMetricsConsumer if it's a MetricsExporter.
func getDimensionClientsFromMetricsExporters(
	specifiedClients []string, host component.Host, nextMetricsConsumer consumer.Metrics, logger *zap.Logger,
) (clients []metadata.MetadataExporter, wasNil bool) {
	if specifiedClients == nil {
		wasNil = true
		// default to nextMetricsConsumer if no clients have been provided
		if asMetadataExporter, ok := nextMetricsConsumer.(metadata.MetadataExporter); ok {
			clients = append(clients, asMetadataExporter)
		}
		return clients, wasNil
	}

	ge, ok := host.(getExporters)
	if !ok {
		return clients, wasNil
	}
	exporters := ge.GetExporters()

	if builtExporters, ok := exporters[pipeline.SignalMetrics]; ok {
		for _, client := range specifiedClients {
			var found bool
			for exporterConfig, exporter := range builtExporters {
				if exporterConfig.String() == client {
					if asMetadataExporter, ok := exporter.(metadata.MetadataExporter); ok {
						clients = append(clients, asMetadataExporter)
					}
					found = true
				}
			}
			if !found {
				logger.Info(
					"specified dimension client is not an available exporter",
					zap.String("client", client),
				)
			}
		}
	}
	return clients, wasNil
}

func getLoneSFxExporter(host component.Host, exporterType pipeline.Signal) component.Component {
	var sfxExporter component.Component
	ge, ok := host.(getExporters)
	if !ok {
		return sfxExporter
	}
	exporters := ge.GetExporters()

	if builtExporters, ok := exporters[exporterType]; ok {
		for exporterConfig, exporter := range builtExporters {
			if exporterConfig.Type().String() == "signalfx" {
				if sfxExporter == nil {
					sfxExporter = exporter
				} else { // we've already found one so no lone instance to use as default
					return nil
				}
			}
		}
	}
	return sfxExporter
}

func (out *output) AddDatapointExclusionFilter(filter dpfilters.DatapointFilter) {
	out.logger.Debug("AddDatapointExclusionFilter has been called", zap.Any("filter", filter))
	out.monitorFiltering.AddDatapointExclusionFilter(filter)
}

func (out *output) EnabledMetrics() []string {
	out.logger.Debug("EnabledMetrics has been called.")
	return out.monitorFiltering.EnabledMetrics()
}

func (out *output) HasEnabledMetricInGroup(group string) bool {
	out.logger.Debug("HasEnabledMetricInGroup has been called", zap.String("group", group))
	return out.monitorFiltering.HasEnabledMetricInGroup(group)
}

func (out *output) HasAnyExtraMetrics() bool {
	out.logger.Debug("HasAnyExtraMetrics has been called.")
	return out.monitorFiltering.HasAnyExtraMetrics()
}

// Copy clones the output to provide to child monitors with their own extraDimensions.
func (out *output) Copy() types.Output {
	out.logger.Debug("Copying out", zap.Any("out", out))
	cp := *out
	cp.extraDimensions = utils.CloneStringMap(out.extraDimensions)
	return &cp
}

func (out *output) SendMetrics(metrics ...pmetric.Metric) {
	if out.nextMetricsConsumer == nil {
		return
	}

	ctx := out.reporter.StartMetricsOp(context.Background())

	metrics = out.filterMetrics(metrics)
	pm := pmetric.NewMetrics()
	rm := pm.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	for _, dp := range metrics {
		for k, v := range out.extraDimensions {
			switch dp.Type() {
			case pmetric.MetricTypeGauge:
				for i := 0; i < dp.Gauge().DataPoints().Len(); i++ {
					dp.Gauge().DataPoints().At(i).Attributes().PutStr(k, v)
				}
			case pmetric.MetricTypeSum:
				for i := 0; i < dp.Sum().DataPoints().Len(); i++ {
					dp.Sum().DataPoints().At(i).Attributes().PutStr(k, v)
				}
			default:
				out.logger.Error("Unsupported metric type", zap.Any("type", dp.Type()), zap.String("name", dp.Name()))
			}
		}
		dp.MoveTo(sm.Metrics().AppendEmpty())
	}

	numPoints := pm.MetricCount()
	err := out.nextMetricsConsumer.ConsumeMetrics(context.Background(), pm)
	out.reporter.EndMetricsOp(ctx, typeStr, numPoints, err)
}

func (out *output) SendDatapoints(datapoints ...*datapoint.Datapoint) {
	if out.nextMetricsConsumer == nil {
		return
	}

	ctx := out.reporter.StartMetricsOp(context.Background())

	datapoints = out.filterDatapoints(datapoints)
	for _, dp := range datapoints {
		// out's extraDimensions take priority over datapoint's
		dp.Dimensions = utils.MergeStringMaps(dp.Dimensions, out.extraDimensions)
	}

	metrics, err := out.translator.ToMetrics(datapoints)
	if err != nil {
		out.logger.Error("error converting SFx datapoints to pmetric.Metrics", zap.Error(err))
	}

	numPoints := metrics.DataPointCount()
	err = out.nextMetricsConsumer.ConsumeMetrics(context.Background(), metrics)
	out.reporter.EndMetricsOp(ctx, typeStr, numPoints, err)
}

func (out *output) SendEvent(event *event.Event) {
	if out.nextLogsConsumer == nil {
		return
	}

	logs, err := out.translator.ToLogs(event)
	if err != nil {
		out.logger.Error("error converting SFx events to ptrace.Traces", zap.Error(err))
	}

	err = out.nextLogsConsumer.ConsumeLogs(context.Background(), logs)
	if err != nil {
		out.logger.Debug("SendEvent has failed", zap.Error(err))
	}
}

func (out *output) SendSpans(spans ...*trace.Span) {
	if out.nextTracesConsumer == nil {
		return
	}

	traces, err := out.translator.ToTraces(spans)
	if err != nil {
		out.logger.Error("error converting SFx spans to ptrace.Traces", zap.Error(err))
	}

	err = out.nextTracesConsumer.ConsumeTraces(context.Background(), traces)
	if err != nil {
		out.logger.Debug("SendSpans has failed", zap.Error(err))
	}
}

func (out *output) SendDimensionUpdate(dimension *types.Dimension) {
	if len(out.nextDimensionClients) == 0 {
		return
	}

	metadataUpdate := dimensionToMetadataUpdate(*dimension)
	for _, consumerInst := range out.nextDimensionClients {
		err := consumerInst.ConsumeMetadata([]*metadata.MetadataUpdate{&metadataUpdate})
		if err != nil {
			out.logger.Debug("SendDimensionUpdate has failed", zap.Error(err))
		}
	}
}

func (out *output) AddExtraDimension(key, value string) {
	out.logger.Debug("Adding extra dimension", zap.String("key", key), zap.String("value", value))
	out.extraDimensions[key] = value
}

func (out *output) filterMetrics(metrics []pmetric.Metric) []pmetric.Metric {
	if out.monitorFiltering.filterSet == nil {
		return metrics
	}
	filteredMetrics := make([]pmetric.Metric, 0, len(metrics))
	for _, m := range metrics {
		atLeastOneDataPoint := false
		switch m.Type() {
		case pmetric.MetricTypeGauge:
			m.Gauge().DataPoints().RemoveIf(func(point pmetric.NumberDataPoint) bool {
				return out.monitorFiltering.filterSet.MatchesMetricDataPoint(m.Name(), point.Attributes())
			})
			atLeastOneDataPoint = m.Gauge().DataPoints().Len() > 0
		case pmetric.MetricTypeSum:
			m.Sum().DataPoints().RemoveIf(func(point pmetric.NumberDataPoint) bool {
				return out.monitorFiltering.filterSet.MatchesMetricDataPoint(m.Name(), point.Attributes())
			})
			atLeastOneDataPoint = m.Sum().DataPoints().Len() > 0
		default:
			panic("unsupported metric type")
		}
		if atLeastOneDataPoint {
			filteredMetrics = append(filteredMetrics, m)
		}
	}
	return filteredMetrics
}

func (out *output) filterDatapoints(datapoints []*datapoint.Datapoint) []*datapoint.Datapoint {
	if out.monitorFiltering.filterSet == nil {
		return datapoints
	}
	filteredDatapoints := make([]*datapoint.Datapoint, 0, len(datapoints))
	for _, dp := range datapoints {
		if !out.monitorFiltering.filterSet.Matches(dp) {
			filteredDatapoints = append(filteredDatapoints, dp)
		}
	}
	return filteredDatapoints
}
