// Copyright  Splunk, Inc.
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

package timestamp

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// typeStr is the value of "type" key in configuration.
	typeStr = "timestamp"
	// The stability level of the processor.
	stability = component.StabilityLevelInDevelopment
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Attributes processor.
func NewFactory() component.ProcessorFactory {
	return component.NewProcessorFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesProcessor(createTracesProcessor, stability),
		component.WithLogsProcessor(createLogsProcessor, stability),
		component.WithMetricsProcessor(createMetricsProcessor, stability))
}

// Note: This isn't a valid configuration because the processor would do no work.
func createDefaultConfig() config.Processor {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(typeStr)),
		Offset:            "0h",
	}
}

func createTracesProcessor(
	ctx context.Context,
	set component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Traces,
) (component.TracesProcessor, error) {
	oCfg := cfg.(*Config)
	offset, _ := time.ParseDuration(oCfg.Offset)

	return processorhelper.NewTracesProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		newSpanAttributesProcessor(set.Logger, offsetFn(offset)),
		processorhelper.WithCapabilities(processorCapabilities))
}

func createLogsProcessor(
	ctx context.Context,
	set component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Logs,
) (component.LogsProcessor, error) {
	oCfg := cfg.(*Config)
	offset, _ := time.ParseDuration(oCfg.Offset)

	return processorhelper.NewLogsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		newLogAttributesProcessor(set.Logger, offsetFn(offset)),
		processorhelper.WithCapabilities(processorCapabilities))
}

func createMetricsProcessor(
	ctx context.Context,
	set component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Metrics,
) (component.MetricsProcessor, error) {
	oCfg := cfg.(*Config)
	offset, _ := time.ParseDuration(oCfg.Offset)
	
	return processorhelper.NewMetricsProcessor(
		ctx,
		set,
		cfg,
		nextConsumer,
		newMetricAttributesProcessor(set.Logger, offsetFn(offset)),
		processorhelper.WithCapabilities(processorCapabilities))
}
