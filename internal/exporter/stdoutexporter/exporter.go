// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package stdoutexporter

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/signalfx/splunk-otel-collector/internal/auth"

	"github.com/goccy/go-json"
	translator "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/splunk"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	defaultIndexLabel = "com.splunk.index"
)

var stdoutWriter = defaultStdoutWriter

func newLogsExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Logs, error) {
	e := &stdoutExporter{}

	return exporterhelper.NewLogs(ctx, set, cfg, e.ConsumeLogs,
		exporterhelper.WithCapabilities(consumer.Capabilities{
			MutatesData: false,
		}))
}

func newTracesExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Traces, error) {
	e := &stdoutExporter{}

	return exporterhelper.NewTraces(ctx, set, cfg, e.ConsumeTraces,
		exporterhelper.WithCapabilities(consumer.Capabilities{
			MutatesData: false,
		}))
}

func newMetricsExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Metrics, error) {
	e := &stdoutExporter{}

	return exporterhelper.NewMetrics(ctx, set, cfg, e.ConsumeMetrics,
		exporterhelper.WithCapabilities(consumer.Capabilities{
			MutatesData: false,
		}))
}

type stdoutExporter struct {
	TelemetrySettings component.TelemetrySettings
	index             string
	source            string
	sourcetype        string
}

func (se *stdoutExporter) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	toOtelAttrs := translator.DefaultHecToOtelAttrs()
	toHecAttrs := translator.DefaultOtelToHecFields()

	tokenConfig := ctx.Value(auth.ContextKey).(auth.HecTokenConfig)
	mapIndexes := make(map[string]struct{}, len(tokenConfig.AllowedIndexes))
	for _, index := range tokenConfig.AllowedIndexes {
		mapIndexes[index] = struct{}{}
	}
	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rl := ld.ResourceLogs().At(i)
		r := rl.Resource()
		for j := 0; j < rl.ScopeLogs().Len(); j++ {
			sl := rl.ScopeLogs().At(j)
			for k := 0; k < sl.LogRecords().Len(); k++ {
				logRecord := sl.LogRecords().At(k)
				if logIndex, ok := logRecord.Attributes().Get(defaultIndexLabel); ok {
					logIndexStr := logIndex.AsString()
					if logIndexStr != "" {
						if _, ok := mapIndexes[logIndexStr]; !ok {
							return fmt.Errorf("index %q is not allowed", logIndexStr)
						}
						continue
					}
				}
				if resourceIndex, ok := r.Attributes().Get(defaultIndexLabel); ok {
					resourceIndexStr := resourceIndex.AsString()
					if resourceIndexStr != "" {
						if _, ok := mapIndexes[resourceIndexStr]; !ok {
							return fmt.Errorf("index %q is not allowed", resourceIndexStr)
						}
						continue
					}
				}
			}
		}
	}

	var errs []error
	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rl := ld.ResourceLogs().At(i)
		r := rl.Resource()
		for j := 0; j < rl.ScopeLogs().Len(); j++ {
			sl := rl.ScopeLogs().At(j)
			for k := 0; k < sl.LogRecords().Len(); k++ {
				logRecord := sl.LogRecords().At(k)
				event := translator.LogToSplunkEvent(r, logRecord, toOtelAttrs, toHecAttrs, se.source, se.sourcetype, se.index)
				if event == nil {
					continue
				}
				b, err := json.Marshal(&event)
				if err != nil {
					errs = append(errs, err)
				} else {
					if err = se.writeToStdout(b); err != nil {
						errs = append(errs, err)
					}
				}
			}
		}
	}
	return errors.Join(errs...)
}

func (se *stdoutExporter) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	toOtelAttrs := translator.DefaultHecToOtelAttrs()

	tokenConfig := ctx.Value(auth.ContextKey).(auth.HecTokenConfig)
	allowedIndices := make(map[string]struct{}, len(tokenConfig.AllowedIndexes))
	for _, index := range tokenConfig.AllowedIndexes {
		allowedIndices[index] = struct{}{}
	}
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		r := rs.Resource()
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			for k := 0; k < ss.Spans().Len(); k++ {
				span := ss.Spans().At(k)
				if spanIndex, ok := span.Attributes().Get(defaultIndexLabel); ok {
					spanIndexStr := spanIndex.AsString()
					if spanIndexStr != "" {
						if _, ok := allowedIndices[spanIndexStr]; !ok {
							return fmt.Errorf("index %q is not allowed", spanIndexStr)
						}
						continue
					}
				}
				if resourceIndex, ok := r.Attributes().Get(defaultIndexLabel); ok {
					resourceIndexStr := resourceIndex.AsString()
					if resourceIndexStr != "" {
						if _, ok := allowedIndices[resourceIndexStr]; !ok {
							return fmt.Errorf("index %q is not allowed", resourceIndexStr)
						}
						continue
					}
				}
			}
		}
	}

	var errs []error
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		rs := td.ResourceSpans().At(i)
		r := rs.Resource()
		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			ss := rs.ScopeSpans().At(j)
			for k := 0; k < ss.Spans().Len(); k++ {
				span := ss.Spans().At(k)
				b, err := json.Marshal(translator.SpanToSplunkEvent(r, span, toOtelAttrs, se.source, se.sourcetype, se.index))
				if err != nil {
					errs = append(errs, err)
				} else {
					if err = se.writeToStdout(b); err != nil {
						errs = append(errs, err)
					}
				}
			}
		}
	}
	return errors.Join(errs...)
}

func (se *stdoutExporter) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	toOtelAttrs := translator.DefaultHecToOtelAttrs()

	tokenConfig := ctx.Value(auth.ContextKey).(auth.HecTokenConfig)
	allowedIndices := make(map[string]struct{}, len(tokenConfig.AllowedIndexes))
	for _, index := range tokenConfig.AllowedIndexes {
		allowedIndices[index] = struct{}{}
	}
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		r := rm.Resource()
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)
				switch m.Type() {
				case pmetric.MetricTypeEmpty:
				case pmetric.MetricTypeGauge:
					g := m.Gauge()
					for k := 0; k < g.DataPoints().Len(); k++ {
						dp := g.DataPoints().At(k)
						if metricIndex, ok := dp.Attributes().Get(defaultIndexLabel); ok {
							metricIndexStr := metricIndex.AsString()
							if metricIndexStr != "" {
								if _, ok := allowedIndices[metricIndexStr]; !ok {
									return fmt.Errorf("index %q is not allowed", metricIndexStr)
								}
								continue
							}
						}
					}
				case pmetric.MetricTypeSum:
					sum := m.Sum()
					for k := 0; k < sum.DataPoints().Len(); k++ {
						dp := sum.DataPoints().At(k)
						if metricIndex, ok := dp.Attributes().Get(defaultIndexLabel); ok {
							metricIndexStr := metricIndex.AsString()
							if metricIndexStr != "" {
								if _, ok := allowedIndices[metricIndexStr]; !ok {
									return fmt.Errorf("index %q is not allowed", metricIndexStr)
								}
							}
						}
					}
				case pmetric.MetricTypeHistogram:
					h := m.Histogram()
					for k := 0; k < h.DataPoints().Len(); k++ {
						dp := h.DataPoints().At(k)
						if metricIndex, ok := dp.Attributes().Get(defaultIndexLabel); ok {
							metricIndexStr := metricIndex.AsString()
							if metricIndexStr != "" {
								if _, ok := allowedIndices[metricIndexStr]; !ok {
									return fmt.Errorf("index %q is not allowed", metricIndexStr)
								}
							}
						}
					}
				case pmetric.MetricTypeSummary:
					ms := m.Summary()
					for k := 0; k < ms.DataPoints().Len(); k++ {
						dp := ms.DataPoints().At(k)
						if metricIndex, ok := dp.Attributes().Get(defaultIndexLabel); ok {
							metricIndexStr := metricIndex.AsString()
							if metricIndexStr != "" {
								if _, ok := allowedIndices[metricIndexStr]; !ok {
									return fmt.Errorf("index %q is not allowed", metricIndexStr)
								}
							}
						}
					}
				case pmetric.MetricTypeExponentialHistogram:
					h := m.ExponentialHistogram()
					for k := 0; k < h.DataPoints().Len(); k++ {
						dp := h.DataPoints().At(k)
						if metricIndex, ok := dp.Attributes().Get(defaultIndexLabel); ok {
							metricIndexStr := metricIndex.AsString()
							if metricIndexStr != "" {
								if _, ok := allowedIndices[metricIndexStr]; !ok {
									return fmt.Errorf("index %q is not allowed", metricIndexStr)
								}
							}
						}
					}
				}
				if resourceIndex, ok := r.Attributes().Get(defaultIndexLabel); ok {
					resourceIndexStr := resourceIndex.AsString()
					if resourceIndexStr != "" {
						if _, ok := allowedIndices[resourceIndexStr]; !ok {
							return fmt.Errorf("index %q is not allowed", resourceIndexStr)
						}
						continue
					}
				}
			}
		}
	}

	var errs []error
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		r := rm.Resource()
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)
				for _, result := range translator.MetricToSplunkEvent(r, m, se.TelemetrySettings.Logger, toOtelAttrs, se.source, se.sourcetype, se.index) {
					b, err := json.Marshal(result)
					if err != nil {
						errs = append(errs, err)
					} else {
						if err = se.writeToStdout(b); err != nil {
							errs = append(errs, err)
						}
					}
				}
			}
		}
	}
	return errors.Join(errs...)
}

func (se *stdoutExporter) writeToStdout(b []byte) error {
	return stdoutWriter(b)
}

func defaultStdoutWriter(b []byte) error {
	_, err := os.Stdout.Write(append(b, '\n'))
	return err
}
