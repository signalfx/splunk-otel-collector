// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package stdoutexporter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

var testDataDir = "../../../cmd/splunk-connect-for-otlp/testdata"

// TelemetryType represents supported telemetry signals.
type TelemetryType string

const (
	TelemetryTypeMetrics TelemetryType = "metrics"
	TelemetryTypeTraces  TelemetryType = "traces"
	TelemetryTypeLogs    TelemetryType = "logs"
)

func BenchmarkStdoutExporter(b *testing.B) {
	stdoutWriter = func([]byte) error { return nil }

	settings := exportertest.NewNopSettings(exportertest.NopType)
	cfg := createDefaultConfig().(*Config)
	ctx := context.Background()

	tests := []struct {
		name          string
		telType       TelemetryType
		inputFilePath string
	}{
		{
			name:          "metrics",
			telType:       TelemetryTypeMetrics,
			inputFilePath: filepath.Join(testDataDir, "otlp_metrics.json"),
		},
		{
			name:          "large metric data set",
			telType:       TelemetryTypeMetrics,
			inputFilePath: filepath.Join(testDataDir, "otlp_metrics_big.json"),
		},
		{
			name:          "traces",
			telType:       TelemetryTypeTraces,
			inputFilePath: filepath.Join(testDataDir, "otlp_traces.json"),
		},
		{
			name:          "large trace data set",
			telType:       TelemetryTypeTraces,
			inputFilePath: filepath.Join(testDataDir, "otlp_traces_big.json"),
		},
		{
			name:          "logs",
			telType:       TelemetryTypeLogs,
			inputFilePath: filepath.Join(testDataDir, "otlp_logs.json"),
		},
		{
			name:          "large log data set",
			telType:       TelemetryTypeLogs,
			inputFilePath: filepath.Join(testDataDir, "otlp_logs_big.json"),
		},
	}

	for _, tt := range tests {
		tt := tt
		b.Run(tt.name, func(b *testing.B) {
			var consume func(context.Context) error
			switch tt.telType {
			case TelemetryTypeMetrics:
				metrics := LoadMetricsFromFile(b, tt.inputFilePath)
				consume = setupMetricsExporter(b, ctx, settings, cfg, metrics)
			case TelemetryTypeTraces:
				traces := LoadTracesFromFile(b, tt.inputFilePath)
				consume = setupTracesExporter(b, ctx, settings, cfg, traces)
			case TelemetryTypeLogs:
				logs := LoadLogsFromFile(b, tt.inputFilePath)
				consume = setupLogsExporter(b, ctx, settings, cfg, logs)
			default:
				b.Fatalf("unknown telemetry type: %v", tt.telType)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := consume(ctx); err != nil {
					b.Fatalf("%s failed: %v", tt.name, err)
				}
			}
		})
	}
}

func setupMetricsExporter(b *testing.B, ctx context.Context, settings exporter.Settings, cfg *Config, metrics pmetric.Metrics) func(context.Context) error {
	b.Helper()
	exp, err := newMetricsExporter(ctx, settings, cfg)
	if err != nil {
		b.Fatalf("failed to create metrics exporter: %v", err)
	}
	if err = exp.Start(ctx, componenttest.NewNopHost()); err != nil {
		b.Fatalf("failed to start metrics exporter: %v", err)
	}
	return func(ctx context.Context) error {
		return exp.ConsumeMetrics(ctx, metrics)
	}
}

func setupTracesExporter(b *testing.B, ctx context.Context, settings exporter.Settings, cfg *Config, traces ptrace.Traces) func(context.Context) error {
	b.Helper()
	exp, err := newTracesExporter(ctx, settings, cfg)
	if err != nil {
		b.Fatalf("failed to create traces exporter: %v", err)
	}
	if err = exp.Start(ctx, componenttest.NewNopHost()); err != nil {
		b.Fatalf("failed to start traces exporter: %v", err)
	}
	return func(ctx context.Context) error {
		return exp.ConsumeTraces(ctx, traces)
	}
}

func setupLogsExporter(b *testing.B, ctx context.Context, settings exporter.Settings, cfg *Config, logs plog.Logs) func(context.Context) error {
	b.Helper()
	exp, err := newLogsExporter(ctx, settings, cfg)
	if err != nil {
		b.Fatalf("failed to create logs exporter: %v", err)
	}
	if err = exp.Start(ctx, componenttest.NewNopHost()); err != nil {
		b.Fatalf("failed to start logs exporter: %v", err)
	}
	return func(ctx context.Context) error {
		return exp.ConsumeLogs(ctx, logs)
	}
}

func LoadLogsFromFile(tb testing.TB, path string) plog.Logs {
	tb.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		tb.Fatalf("failed to read logs fixture: %v", err)
	}
	unmarshaler := plog.JSONUnmarshaler{}
	logs, err := unmarshaler.UnmarshalLogs(data)
	if err != nil {
		tb.Fatalf("failed to unmarshal logs fixture: %v", err)
	}
	return logs
}

func LoadMetricsFromFile(tb testing.TB, path string) pmetric.Metrics {
	tb.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		tb.Fatalf("failed to read metrics fixture: %v", err)
	}
	unmarshaler := pmetric.JSONUnmarshaler{}
	metrics, err := unmarshaler.UnmarshalMetrics(data)
	if err != nil {
		tb.Fatalf("failed to unmarshal metrics fixture: %v", err)
	}
	return metrics
}

func LoadTracesFromFile(tb testing.TB, path string) ptrace.Traces {
	tb.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		tb.Fatalf("failed to read traces fixture: %v", err)
	}
	unmarshaler := ptrace.JSONUnmarshaler{}
	traces, err := unmarshaler.UnmarshalTraces(data)
	if err != nil {
		tb.Fatalf("failed to unmarshal traces fixture: %v", err)
	}
	return traces
}
