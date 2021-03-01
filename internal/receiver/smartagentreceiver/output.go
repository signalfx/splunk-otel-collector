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
	"time"

	metadata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/obsreport"
	"go.uber.org/zap"
)

const internalTransport = "internal"

// Output is an implementation of a Smart Agent FilteringOutput that receives datapoints, events, and dimension updates
// from a configured monitor.  It will forward all datapoints to the nextConsumer, all dimension updates to the
// nextMetricMetadataClients, and all events to the nextLogMetadataClients as determined by the associated
// items in Config.MetadataClients.
type Output struct {
	receiverName              string
	nextConsumer              consumer.MetricsConsumer
	nextMetricMetadataClients []*metadata.MetadataExporter
	nextLogMetadataClients    []*consumer.LogsConsumer
	logger                    *zap.Logger
	converter                 Converter
	extraDimensions           map[string]string
}

var _ types.Output = (*Output)(nil)
var _ types.FilteringOutput = (*Output)(nil)

func NewOutput(config Config, nextConsumer consumer.MetricsConsumer, host component.Host, logger *zap.Logger) *Output {
	metadataExporters := getMetadataExporters(config, host, &nextConsumer, logger)
	logConsumers := getLogsConsumers(config, host, &nextConsumer, logger)
	return &Output{
		receiverName:              config.Name(),
		nextConsumer:              nextConsumer,
		nextMetricMetadataClients: metadataExporters,
		nextLogMetadataClients:    logConsumers,
		logger:                    logger,
		converter:                 Converter{logger: logger},
		extraDimensions:           map[string]string{},
	}
}

// getMetadataExporters walks through obtained Config.MetadataClients and returns all matching registered MetadataExporters,
// if any.  At this time the SignalFx exporter is the only supported use case and adopter of this type.
func getMetadataExporters(
	config Config, host component.Host, nextConsumer *consumer.MetricsConsumer, logger *zap.Logger,
) []*metadata.MetadataExporter {
	var exporters []*metadata.MetadataExporter

	metadataClients := getMetadataClients(config, host, nextConsumer, logger)
	for _, client := range metadataClients {
		if metadataExporter, ok := (*client).(metadata.MetadataExporter); ok {
			exporters = append(exporters, &metadataExporter)
		} else {
			logger.Info("cannot send dimension updates to metadataClient", zap.Any("client", *client))
		}
	}

	if len(exporters) == 0 {
		logger.Debug("no dimension updates are possible as no valid metadataClients have been provided and next pipeline component isn't a MetadataExporter")
	}

	return exporters
}

// getLogsConsumers walks through obtained Config.MetadataClients and returns all matching registered LogsConsumers,
// if any.  At this time the SignalFx exporter is the only real target use case, but it's unexported and
// as implemented all specified combination MetricsExporters and LogsConsumers will be returned.
func getLogsConsumers(
	config Config, host component.Host, nextConsumer *consumer.MetricsConsumer, logger *zap.Logger,
) []*consumer.LogsConsumer {
	var consumers []*consumer.LogsConsumer

	metadataClients := getMetadataClients(config, host, nextConsumer, logger)
	for _, client := range metadataClients {
		if logsExporter, ok := (*client).(consumer.LogsConsumer); ok {
			consumers = append(consumers, &logsExporter)
		} else {
			logger.Info("cannot send events to metadataClient", zap.Any("client", *client))
		}
	}

	if len(consumers) == 0 {
		logger.Debug("no SFx events are possible as no valid metadataClients have been provided and next pipeline component isn't a LogsConsumer")
	}

	return consumers
}

// getMetadataClients will walk through all provided config.MetadataClients and retrieve matching registered
// MetricsExporters, the only truly supported component type.
// If config.MetadataClients is nil, it will return a slice with nextConsumer if it's a MetricsExporter.
func getMetadataClients(
	config Config, host component.Host, nextConsumer *consumer.MetricsConsumer, logger *zap.Logger,
) []*component.MetricsExporter {
	var clients []*component.MetricsExporter
	if config.MetadataClients == nil {
		// default to nextConsumer if no metadata clients have been provided
		if metricsExporter, ok := (*nextConsumer).(component.MetricsExporter); ok {
			clients = append(clients, &metricsExporter)
		}
		return clients
	}

	builtExporters := host.GetExporters()[configmodels.MetricsDataType]
	for _, client := range *config.MetadataClients {
		var found bool
		for exporterConfig, exporter := range builtExporters {
			if exporterConfig.Name() == client {
				if metricsExporter, ok := exporter.(component.MetricsExporter); ok {
					clients = append(clients, &metricsExporter)
					found = true
					break
				} else {
					logger.Info(
						"specified metadataClient is not a valid MetricsConsumer",
						zap.String("client", client),
					)
				}
			}
		}
		if !found {
			logger.Info("specified metadataClient is not an available exporter", zap.String("client", client))
		}
	}
	return clients
}

func (output *Output) AddDatapointExclusionFilter(filter dpfilters.DatapointFilter) {
	output.logger.Debug("AddDatapointExclusionFilter has been called", zap.Any("filter", filter))
}

func (output *Output) EnabledMetrics() []string {
	output.logger.Debug("EnabledMetrics has been called.")
	return []string{}
}

func (output *Output) HasEnabledMetricInGroup(group string) bool {
	output.logger.Debug("HasEnabledMetricInGroup has been called", zap.String("group", group))
	return true
}

func (output *Output) HasAnyExtraMetrics() bool {
	output.logger.Debug("HasAnyExtraMetrics has been called.")
	return true
}

// Some monitors will clone their Output to provide to child monitors with their own extraDimensions
func (output *Output) Copy() types.Output {
	output.logger.Debug("Copying Output", zap.Any("output", output))
	cp := *output
	cp.extraDimensions = utils.CloneStringMap(output.extraDimensions)
	return &cp
}

func (output *Output) SendDatapoints(datapoints ...*datapoint.Datapoint) {
	ctx := obsreport.ReceiverContext(context.Background(), output.receiverName, internalTransport)
	ctx = obsreport.StartMetricsReceiveOp(ctx, typeStr, internalTransport)

	for _, dp := range datapoints {
		// Output's extraDimensions take priority over datapoint's
		dp.Dimensions = utils.MergeStringMaps(dp.Dimensions, output.extraDimensions)
	}

	metrics, numDropped := output.converter.toMetrics(datapoints, time.Now())
	if numDropped > 0 {
		output.logger.Debug("SendDatapoints has dropped points", zap.Int("numDropped", numDropped))
	}

	_, numPoints := metrics.MetricAndDataPointCount()
	err := output.nextConsumer.ConsumeMetrics(context.Background(), metrics)
	obsreport.EndMetricsReceiveOp(ctx, typeStr, numPoints, err)
}

func (output *Output) SendEvent(event *event.Event) {
	if len(output.nextLogMetadataClients) == 0 {
		return
	}

	logRecord := eventToLog(event, output.logger)
	for _, logsConsumer := range output.nextLogMetadataClients {
		err := (*logsConsumer).ConsumeLogs(context.Background(), logRecord)
		if err != nil {
			output.logger.Debug("SendEvent has failed", zap.Error(err))
		}
	}
}

func (output *Output) SendSpans(spans ...*trace.Span) {
	output.logger.Debug("SendSpans has been called.", zap.Any("Span", spans))
}

func (output *Output) SendDimensionUpdate(dimension *types.Dimension) {
	if len(output.nextMetricMetadataClients) == 0 {
		return
	}

	metadataUpdate := dimensionToMetadataUpdate(*dimension)
	for _, consumer := range output.nextMetricMetadataClients {
		exporter := *consumer
		err := exporter.ConsumeMetadata([]*metadata.MetadataUpdate{&metadataUpdate})
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
	output.logger.Debug("AddExtraSpanTag has been called.", zap.String("key", key), zap.String("value", value))
}

func (output *Output) RemoveExtraSpanTag(key string) {
	output.logger.Debug("RemoveExtraSpanTag has been called.", zap.String("key", key))
}

func (output *Output) AddDefaultSpanTag(key, value string) {
	output.logger.Debug("AddDefaultSpanTag has been called.", zap.String("key", key), zap.String("value", value))
}

func (output *Output) RemoveDefaultSpanTag(key string) {
	output.logger.Debug("RemoveDefaultSpanTag has been called.", zap.String("key", key))
}
