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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/smartagentreceiver/converter"
)

const internalTransport = "internal"

// Output is an implementation of a Smart Agent FilteringOutput that receives datapoints, events, and dimension updates
// from a configured monitor.  It will forward all datapoints to the nextMetricsConsumer, all dimension updates to the
// nextDimensionClients as determined by the associated items in Config.MetadataClients, and all events to the
// nextLogsConsumer.
type Output struct {
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

var _ types.Output = (*Output)(nil)
var _ types.FilteringOutput = (*Output)(nil)

func NewOutput(
	config Config, filtering *monitorFiltering, nextMetricsConsumer consumer.Metrics,
	nextLogsConsumer consumer.Logs, nextTracesConsumer consumer.Traces, host component.Host,
	params component.ReceiverCreateSettings,
) *Output {
	return &Output{
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

func (output *Output) AddDatapointExclusionFilter(filter dpfilters.DatapointFilter) {
	output.logger.Debug("AddDatapointExclusionFilter has been called", zap.Any("filter", filter))
	output.monitorFiltering.AddDatapointExclusionFilter(filter)
}

func (output *Output) EnabledMetrics() []string {
	output.logger.Debug("EnabledMetrics has been called.")
	return output.monitorFiltering.EnabledMetrics()
}

func (output *Output) HasEnabledMetricInGroup(group string) bool {
	output.logger.Debug("HasEnabledMetricInGroup has been called", zap.String("group", group))
	return output.monitorFiltering.HasEnabledMetricInGroup(group)
}

func (output *Output) HasAnyExtraMetrics() bool {
	output.logger.Debug("HasAnyExtraMetrics has been called.")
	return output.monitorFiltering.HasAnyExtraMetrics()
}

// Copy clones the Output to provide to child monitors with their own extraDimensions.
func (output *Output) Copy() types.Output {
	output.logger.Debug("Copying Output", zap.Any("output", output))
	cp := *output
	cp.extraDimensions = utils.CloneStringMap(output.extraDimensions)
	cp.extraSpanTags = utils.CloneStringMap(output.extraSpanTags)
	cp.defaultSpanTags = utils.CloneStringMap(output.defaultSpanTags)
	return &cp
}

func (output *Output) SendDatapoints(datapoints ...*datapoint.Datapoint) {
	if output.nextMetricsConsumer == nil {
		return
	}

	ctx := output.reporter.StartMetricsOp(context.Background())

	datapoints = output.filterDatapoints(datapoints)
	for _, dp := range datapoints {
		// Output's extraDimensions take priority over datapoint's
		dp.Dimensions = utils.MergeStringMaps(dp.Dimensions, output.extraDimensions)
	}

	metrics, err := output.translator.ToMetrics(datapoints)
	if err != nil {
		output.logger.Error("error converting SFx datapoints to pdata.Traces", zap.Error(err))
	}

	numPoints := metrics.DataPointCount()
	err = output.nextMetricsConsumer.ConsumeMetrics(context.Background(), metrics)
	output.reporter.EndMetricsOp(ctx, typeStr, numPoints, err)
}

func (output *Output) SendEvent(event *event.Event) {
	if output.nextLogsConsumer == nil {
		return
	}

	logs, err := output.translator.ToLogs(event)
	if err != nil {
		output.logger.Error("error converting SFx events to pdata.Traces", zap.Error(err))
	}

	err = output.nextLogsConsumer.ConsumeLogs(context.Background(), logs)
	if err != nil {
		output.logger.Debug("SendEvent has failed", zap.Error(err))
	}
}

func (output *Output) SendSpans(spans ...*trace.Span) {
	if output.nextTracesConsumer == nil {
		return
	}

	for _, span := range spans {
		if span.Tags == nil {
			span.Tags = map[string]string{}
		}

		for name, value := range output.defaultSpanTags {
			// If the tags are already set, don't override
			if _, ok := span.Tags[name]; !ok {
				span.Tags[name] = value
			}
		}

		span.Tags = utils.MergeStringMaps(span.Tags, output.extraSpanTags)
	}

	traces, err := output.translator.ToTraces(spans)
	if err != nil {
		output.logger.Error("error converting SFx spans to pdata.Traces", zap.Error(err))
	}

	err = output.nextTracesConsumer.ConsumeTraces(context.Background(), traces)
	if err != nil {
		output.logger.Debug("SendSpans has failed", zap.Error(err))
	}
}

func (output *Output) SendDimensionUpdate(dimension *types.Dimension) {
	if len(output.nextDimensionClients) == 0 {
		return
	}

	metadataUpdate := dimensionToMetadataUpdate(*dimension)
	for _, consumer := range output.nextDimensionClients {
		err := consumer.ConsumeMetadata([]*metadata.MetadataUpdate{&metadataUpdate})
		if err != nil {
			output.logger.Debug("SendDimensionUpdate has failed", zap.Error(err))
		}
	}
}

func (output *Output) AddExtraDimension(key, value string) {
	output.logger.Debug("Adding extra dimension", zap.String("key", key), zap.String("value", value))
	output.extraDimensions[key] = value
}

func (output *Output) RemoveExtraDimension(key string) {
	output.logger.Debug("Removing extra dimension", zap.String("key", key))
	delete(output.extraDimensions, key)
}

func (output *Output) AddExtraSpanTag(key, value string) {
	output.extraSpanTags[key] = value
}

func (output *Output) RemoveExtraSpanTag(key string) {
	delete(output.extraSpanTags, key)
}

func (output *Output) AddDefaultSpanTag(key, value string) {
	output.defaultSpanTags[key] = value
}

func (output *Output) RemoveDefaultSpanTag(key string) {
	delete(output.defaultSpanTags, key)
}

func (output *Output) filterDatapoints(datapoints []*datapoint.Datapoint) []*datapoint.Datapoint {
	filteredDatapoints := make([]*datapoint.Datapoint, 0, len(datapoints))
	for _, dp := range datapoints {
		if output.monitorFiltering.filterSet == nil || !output.monitorFiltering.filterSet.Matches(dp) {
			filteredDatapoints = append(filteredDatapoints, dp)
		}
	}
	return filteredDatapoints
}
