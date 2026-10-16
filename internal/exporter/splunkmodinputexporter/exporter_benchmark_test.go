// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package splunkmodinputexporter

import (
	"io"
	"testing"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func BenchmarkModinputExporter(b *testing.B) {
	for _, size := range []struct {
		name  string
		count int
	}{{name: "single", count: 1}, {name: "batch", count: 1000}} {
		for _, signal := range []string{"logs", "metrics", "traces"} {
			b.Run(signal+"/"+size.name, func(b *testing.B) {
				e := newModinputExporter(&Config{}, componenttest.NewNopTelemetrySettings(), io.Discard)
				logs, metrics, traces := plog.NewLogs(), pmetric.NewMetrics(), ptrace.NewTraces()
				lr := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords()
				dp := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
				dp.SetName("benchmark.metric")
				gauge := dp.SetEmptyGauge()
				spans := traces.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans()
				for range size.count {
					lr.AppendEmpty().Body().SetStr("benchmark log")
					gauge.DataPoints().AppendEmpty().SetIntValue(42)
					spans.AppendEmpty().SetName("benchmark span")
				}
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					var err error
					switch signal {
					case "logs":
						err = e.pushLogs(b.Context(), logs)
					case "metrics":
						err = e.pushMetrics(b.Context(), metrics)
					case "traces":
						err = e.pushTraces(b.Context(), traces)
					}
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
