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
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"go.opentelemetry.io/collector/component"
	collectorConfig "go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/obsreport"
	"go.uber.org/zap"

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
	extraSpanTags        map[string]string
	defaultSpanTags      map[string]string
	logger               *zap.Logger
	reporter             *obsreport.Receiver
	translator           converter.Translator
	monitorFiltering     *monitorFiltering
	receiverID           collectorConfig.ComponentID
	nextDimensionClients []metadata.MetadataExporter
}

var _ types.Output = (*output)(nil)
var _ types.FilteringOutput = (*output)(nil)

func newOutput(
	config Config, filtering *monitorFiltering, nextMetricsConsumer consumer.Metrics,
	nextLogsConsumer consumer.Logs, nextTracesConsumer consumer.Traces, host component.Host,
	params component.ReceiverCreateSettings,
) *output {
	return &output{
		receiverID:           config.ID(),
		nextMetricsConsumer:  nextMetricsConsumer,
		nextLogsConsumer:     nextLogsConsumer,
		nextTracesConsumer:   nextTracesConsumer,
		nextDimensionClients: getMetadataExporters(config, host, nextMetricsConsumer, params.Logger),
		logger:               params.Logger,
		translator:           converter.NewTranslator(params.Logger),
		extraDimensions:      map[string]string{},
		extraSpanTags:        map[string]string{},
		defaultSpanTags:      map[string]string{},
		monitorFiltering:     filtering,
		reporter: obsreport.NewReceiver(obsreport.ReceiverSettings{
			ReceiverID:             config.ID(),
			Transport:              internalTransport,
			ReceiverCreateSettings: params,
		}),
	}

}

// getMetadataExporters walks through obtained Config.MetadataClients and returns all matching registered MetadataExporters,
// if any.  At this time the SignalFx exporter is the only supported use case and adopter of this type.
func getMetadataExporters(
	cfg Config, host component.Host, nextMetricsConsumer consumer.Metrics, logger *zap.Logger,
) []metadata.MetadataExporter {
	var exporters []metadata.MetadataExporter

	exporters, noClientsSpecified := getDimensionClientsFromMetricsExporters(cfg.DimensionClients, host, nextMetricsConsumer, logger)

	if len(exporters) == 0 && noClientsSpecified {
		sfxExporter := getLoneSFxExporter(host, collectorConfig.MetricsDataType)
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
		return
	}

	if builtExporters, ok := host.GetExporters()[collectorConfig.MetricsDataType]; ok {
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
	return
}

func getLoneSFxExporter(host component.Host, exporterType collectorConfig.DataType) component.Exporter {
	var sfxExporter component.Exporter
	if builtExporters, ok := host.GetExporters()[exporterType]; ok {
		for exporterConfig, exporter := range builtExporters {
			if exporterConfig.Type() == "signalfx" {
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
	cp.extraSpanTags = utils.CloneStringMap(out.extraSpanTags)
	cp.defaultSpanTags = utils.CloneStringMap(out.defaultSpanTags)
	return &cp
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
		out.logger.Error("error converting SFx datapoints to ptrace.Traces", zap.Error(err))
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

	for _, span := range spans {
		if span.Tags == nil {
			span.Tags = map[string]string{}
		}

		for name, value := range out.defaultSpanTags {
			// If the tags are already set, don't override
			if _, ok := span.Tags[name]; !ok {
				span.Tags[name] = value
			}
		}

		span.Tags = utils.MergeStringMaps(span.Tags, out.extraSpanTags)
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
	for _, consumer := range out.nextDimensionClients {
		err := consumer.ConsumeMetadata([]*metadata.MetadataUpdate{&metadataUpdate})
		if err != nil {
			out.logger.Debug("SendDimensionUpdate has failed", zap.Error(err))
		}
	}
}

func (out *output) AddExtraDimension(key, value string) {
	out.logger.Debug("Adding extra dimension", zap.String("key", key), zap.String("value", value))
	out.extraDimensions[key] = value
}

func (out *output) RemoveExtraDimension(key string) {
	out.logger.Debug("Removing extra dimension", zap.String("key", key))
	delete(out.extraDimensions, key)
}

func (out *output) AddExtraSpanTag(key, value string) {
	out.extraSpanTags[key] = value
}

func (out *output) RemoveExtraSpanTag(key string) {
	delete(out.extraSpanTags, key)
}

func (out *output) AddDefaultSpanTag(key, value string) {
	out.defaultSpanTags[key] = value
}

func (out *output) RemoveDefaultSpanTag(key string) {
	delete(out.defaultSpanTags, key)
}

func (out *output) filterDatapoints(datapoints []*datapoint.Datapoint) []*datapoint.Datapoint {
	filteredDatapoints := make([]*datapoint.Datapoint, 0, len(datapoints))
	for _, dp := range datapoints {
		if out.monitorFiltering.filterSet == nil || !out.monitorFiltering.filterSet.Matches(dp) {
			filteredDatapoints = append(filteredDatapoints, dp)
		}
	}
	return filteredDatapoints
}
