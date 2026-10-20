// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package splunkmodinputexporter

import (
	"context"
	"os"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/signalfx/splunk-otel-collector/internal/exporter/splunkmodinputexporter/internal/metadata"
)

// NewFactory creates a factory for the splunk_modinput exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exporter.WithLogs(newLogsExporter, metadata.LogsStability),
		exporter.WithMetrics(newMetricsExporter, metadata.MetricsStability),
		exporter.WithTraces(newTracesExporter, metadata.TracesStability),
	)
}

func newLogsExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Logs, error) {
	e := newModinputExporter(cfg.(*Config), set.TelemetrySettings, os.Stdout)
	return exporterhelper.NewLogs(ctx, set, cfg, e.pushLogs,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}))
}

func newMetricsExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Metrics, error) {
	e := newModinputExporter(cfg.(*Config), set.TelemetrySettings, os.Stdout)
	return exporterhelper.NewMetrics(ctx, set, cfg, e.pushMetrics,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}))
}

func newTracesExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Traces, error) {
	e := newModinputExporter(cfg.(*Config), set.TelemetrySettings, os.Stdout)
	return exporterhelper.NewTraces(ctx, set, cfg, e.pushTraces,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}))
}
