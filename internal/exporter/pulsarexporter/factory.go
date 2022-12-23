// Copyright  The OpenTelemetry Authors
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

package pulsarexporter

import (
	ctx "context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	typeStr                 = "pulsar"
	defaultMetricsTopic     = "otlp_metrics"
	defaultEncoding         = "otlp_proto"
	defaultBroker           = "pulsar://localhost:6651"
	defaultCompressionType  = "none"
	defaultCompressionLevel = "default"
	defaultHashingScheme    = "java_string_hash"
)

// FactoryOption applies changes to pulsarExporterFactory.
type FactoryOption func(factory *pulsarExporterFactory)

// NewFactory creates pulsar exporter factory.
func NewFactory(options ...FactoryOption) exporter.Factory {
	f := &pulsarExporterFactory{
		metricsMarshalers: metricsMarshalers(),
	}
	for _, option := range options {
		option(f)
	}
	return exporter.NewFactory(
		typeStr,
		createDefaultConfig,
		exporter.WithMetrics(f.createMetricsExporter, component.StabilityLevelAlpha),
	)
}

type pulsarExporterFactory struct {
	metricsMarshalers map[string]MetricsMarshaler
}

func createDefaultConfig() component.Config {

	return &Config{
		TimeoutSettings: exporterhelper.NewDefaultTimeoutSettings(),
		RetrySettings:   exporterhelper.NewDefaultRetrySettings(),
		QueueSettings:   exporterhelper.NewDefaultQueueSettings(),
		Broker:          defaultBroker,
		Topic:           defaultMetricsTopic,
		Encoding:        defaultEncoding,
		Producer: Producer{
			CompressionType:  defaultCompressionType,
			CompressionLevel: defaultCompressionLevel,
			HashingScheme:    defaultHashingScheme,
		},
		Authentication: Authentication{TLS: &configtls.TLSClientSetting{
			InsecureSkipVerify: true,
		}},
	}
}

func (f *pulsarExporterFactory) createMetricsExporter(
	ctx ctx.Context,
	settings exporter.CreateSettings,
	cfg component.Config,
) (exporter.Metrics, error) {
	oCfg := cfg.(*Config)
	if oCfg.Encoding == "otlp_json" {
		settings.Logger.Info("otlp_json is considered experimental and should not be used in a production environment")
	}
	exp, err := newMetricsExporter(*oCfg, settings, f.metricsMarshalers)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewMetricsExporter(
		ctx,
		settings,
		cfg,
		exp.metricsDataPusher,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(exporterhelper.TimeoutSettings{Timeout: 0}),
		exporterhelper.WithRetry(oCfg.RetrySettings),
		exporterhelper.WithQueue(oCfg.QueueSettings),
		exporterhelper.WithShutdown(exp.Close))
}
