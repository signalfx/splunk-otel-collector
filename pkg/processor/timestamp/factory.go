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

package timestampprocessor

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// typeStr is the value of "type" key in configuration.
	typeStr = "timestamp"
	// The stability level of the processor.
	stability = component.StabilityLevelDevelopment
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
func createDefaultConfig() component.Config {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(component.NewID(typeStr)),
		Offset:            "0h",
	}
}

func createTracesProcessor(
	ctx context.Context,
	set component.ProcessorCreateSettings,
	cfg component.Config,
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
	cfg component.Config,
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
	cfg component.Config,
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

func offsetFn(offset time.Duration) func(pcommon.Timestamp) pcommon.Timestamp {
	return func(ts pcommon.Timestamp) pcommon.Timestamp {
		if ts == zeroTs {
			return ts
		}
		t := ts.AsTime()
		t = t.Add(offset)
		return pcommon.NewTimestampFromTime(t)
	}
}
