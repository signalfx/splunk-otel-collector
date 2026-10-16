// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package splunkmodinputexporter

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/goccy/go-json"
	translator "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/splunk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	DefaultSourceTypeLabel = "com.splunk.sourcetype"
	DefaultSourceLabel     = "com.splunk.source"
	DefaultIndexLabel      = "com.splunk.index"
	DefaultNameLabel       = "otel.log.name"
)

func prepareLogs() plog.Logs {
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().SetName("test")
	ts := pcommon.Timestamp(0)
	logRecord := sl.LogRecords().AppendEmpty()
	logRecord.Body().SetStr("test log")
	logRecord.Attributes().PutStr(DefaultNameLabel, "test- label")
	logRecord.Attributes().PutStr("host.name", "myhost")
	logRecord.Attributes().PutStr("custom", "custom")
	logRecord.SetTimestamp(ts)
	return logs
}

func prepareLogsNonDefaultParams(index, source, sourcetype, event string) plog.Logs {
	logs := plog.NewLogs()
	rl := logs.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().SetName("test")
	ts := pcommon.Timestamp(0)

	logRecord := sl.LogRecords().AppendEmpty()
	logRecord.Body().SetStr(event)
	logRecord.Attributes().PutStr(DefaultNameLabel, "label")
	logRecord.Attributes().PutStr(DefaultSourceLabel, source)
	logRecord.Attributes().PutStr(DefaultSourceTypeLabel, sourcetype)
	logRecord.Attributes().PutStr(DefaultIndexLabel, index)
	logRecord.Attributes().PutStr("host.name", "myhost")
	logRecord.Attributes().PutStr("custom", "custom")
	logRecord.SetTimestamp(ts)
	return logs
}

func prepareMetricsData(metricName string) pmetric.Metrics {
	metricData := pmetric.NewMetrics()
	metric := metricData.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	g := metric.SetEmptyGauge()
	g.DataPoints().AppendEmpty().SetDoubleValue(132.929)
	metric.SetName(metricName)
	return metricData
}

func prepareTracesData(index, source, sourcetype string) ptrace.Traces {
	ts := pcommon.Timestamp(0)

	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr(DefaultSourceLabel, source)
	rs.Resource().Attributes().PutStr("host.name", "myhost")
	rs.Resource().Attributes().PutStr(DefaultSourceTypeLabel, sourcetype)
	rs.Resource().Attributes().PutStr(DefaultIndexLabel, index)
	ils := rs.ScopeSpans().AppendEmpty()
	initSpan("myspan", ts, ils.Spans().AppendEmpty())
	return traces
}

func initSpan(name string, ts pcommon.Timestamp, span ptrace.Span) {
	span.Attributes().PutStr("foo", "bar")
	span.SetName(name)
	span.SetStartTimestamp(ts)
	spanLink := span.Links().AppendEmpty()
	spanLink.TraceState().FromRaw("OK")
	bytes, _ := hex.DecodeString("12345678")
	var traceID [16]byte
	copy(traceID[:], bytes)
	spanLink.SetTraceID(traceID)
	bytes, _ = hex.DecodeString("1234")
	var spanID [8]byte
	copy(spanID[:], bytes)
	spanLink.SetSpanID(spanID)
	spanLink.Attributes().PutInt("foo", 1)
	spanLink.Attributes().PutBool("bar", false)
	foobarContents := spanLink.Attributes().PutEmptySlice("foobar")
	foobarContents.AppendEmpty().SetStr("a")
	foobarContents.AppendEmpty().SetStr("b")

	spanEvent := span.Events().AppendEmpty()
	spanEvent.Attributes().PutStr("foo", "bar")
	spanEvent.SetName("myEvent")
	spanEvent.SetTimestamp(ts + 3)
}

// TestHECJSONEncoding verifies the HEC event format written to stdout; no HTTP
// request or HEC endpoint is involved.
func TestHECJSONEncoding(t *testing.T) {
	for _, signal := range []string{"logs", "logs_attributes", "metrics", "traces"} {
		t.Run(signal, func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			e := newModinputExporter(&Config{}, componenttest.NewNopTelemetrySettings(), &out)
			var err error
			var fixture string
			switch signal {
			case "logs":
				ld := prepareLogs()
				ld.MarkReadOnly()
				err = e.pushLogs(t.Context(), ld)
				fixture = "expected_hec_log.json"
			case "logs_attributes":
				ld := prepareLogsNonDefaultParams("sck-otel", "otel-source", "sck-otel-st", "This is my new event! And some number 101")
				ld.MarkReadOnly()
				err = e.pushLogs(t.Context(), ld)
				fixture = "expected_hec_log_non_default_index.json"
			case "metrics":
				md := prepareMetricsData("test.metric")
				md.MarkReadOnly()
				err = e.pushMetrics(t.Context(), md)
				fixture = "expected_hec_metric.json"
			case "traces":
				td := prepareTracesData("sck-traces", "trace-source", "trace-sourcetype")
				td.MarkReadOnly()
				err = e.pushTraces(t.Context(), td)
				fixture = "expected_hec_trace.json"
			}
			require.NoError(t, err)
			expected, err := os.ReadFile("testdata/" + fixture)
			require.NoError(t, err)
			require.JSONEq(t, string(expected), out.String())
			require.True(t, strings.HasSuffix(out.String(), "\n"))
		})
	}
}

func TestConfigDefaultsAndAttributePrecedence(t *testing.T) {
	t.Parallel()
	for _, signal := range []string{"logs", "metrics", "traces"} {
		for _, attributes := range []bool{false, true} {
			t.Run(signal+"/attributes="+map[bool]string{false: "false", true: "true"}[attributes], func(t *testing.T) {
				t.Parallel()
				var out bytes.Buffer
				e := newModinputExporter(&Config{Source: "configured-source", Sourcetype: "configured-type", Index: "configured-index"}, componenttest.NewNopTelemetrySettings(), &out)
				var attrs pcommon.Map
				var push func() error
				switch signal {
				case "logs":
					ld := prepareLogs()
					attrs = ld.ResourceLogs().At(0).Resource().Attributes()
					push = func() error { return e.pushLogs(t.Context(), ld) }
				case "metrics":
					md := prepareMetricsData("test.metric")
					attrs = md.ResourceMetrics().At(0).Resource().Attributes()
					push = func() error { return e.pushMetrics(t.Context(), md) }
				case "traces":
					td := prepareTracesData("", "", "")
					attrs = td.ResourceSpans().At(0).Resource().Attributes()
					attrs.Clear()
					push = func() error { return e.pushTraces(t.Context(), td) }
				}
				prefix := "configured-"
				if attributes {
					prefix = "attribute-"
					attrs.PutStr(DefaultSourceLabel, prefix+"source")
					attrs.PutStr(DefaultSourceTypeLabel, prefix+"type")
					attrs.PutStr(DefaultIndexLabel, prefix+"index")
				}
				require.NoError(t, push())
				var event translator.Event
				require.NoError(t, json.Unmarshal(out.Bytes(), &event))
				assert.Equal(t, prefix+"source", event.Source)
				assert.Equal(t, prefix+"type", event.SourceType)
				assert.Equal(t, prefix+"index", event.Index)
			})
		}
	}
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(b []byte) (int, error) { return f(b) }

func TestWriteErrors(t *testing.T) {
	t.Parallel()
	writeErr := errors.New("broken pipe")
	for _, tt := range []struct {
		writer io.Writer
		want   error
		name   string
	}{
		{name: "write failure", writer: writerFunc(func([]byte) (int, error) { return 0, writeErr }), want: writeErr},
		{name: "short write", writer: writerFunc(func(b []byte) (int, error) { return len(b) - 1, nil }), want: io.ErrShortWrite},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, signal := range []string{"logs", "metrics", "traces"} {
				e := newModinputExporter(&Config{}, componenttest.NewNopTelemetrySettings(), tt.writer)
				var err error
				switch signal {
				case "logs":
					err = e.pushLogs(t.Context(), prepareLogs())
				case "metrics":
					err = e.pushMetrics(t.Context(), prepareMetricsData("metric"))
				case "traces":
					err = e.pushTraces(t.Context(), prepareTracesData("", "", ""))
				}
				require.ErrorIs(t, err, tt.want)
			}
		})
	}
}

func TestMarshalErrorIsPermanent(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	e := newModinputExporter(&Config{}, componenttest.NewNopTelemetrySettings(), &out)
	err := e.writeEvent(&translator.Event{Event: make(chan int)})
	require.Error(t, err)
	assert.True(t, consumererror.IsPermanent(err))
	assert.Empty(t, out.String())
}

func TestEmptyAndCanceledData(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	e := newModinputExporter(&Config{}, componenttest.NewNopTelemetrySettings(), &out)
	require.NoError(t, e.pushLogs(t.Context(), plog.NewLogs()))
	require.NoError(t, e.pushMetrics(t.Context(), pmetric.NewMetrics()))
	require.NoError(t, e.pushTraces(t.Context(), ptrace.NewTraces()))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, e.pushLogs(ctx, prepareLogs()), context.Canceled)
	require.ErrorIs(t, e.pushMetrics(ctx, prepareMetricsData("metric")), context.Canceled)
	require.ErrorIs(t, e.pushTraces(ctx, prepareTracesData("", "", "")), context.Canceled)
	assert.Empty(t, out.String())
}

func TestConcurrentInstances(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			e := newModinputExporter(&Config{}, componenttest.NewNopTelemetrySettings(), &out)
			assert.NoError(t, e.pushLogs(t.Context(), prepareLogs()))
			assert.NoError(t, e.pushMetrics(t.Context(), prepareMetricsData("metric")))
			assert.NoError(t, e.pushTraces(t.Context(), prepareTracesData("", "", "")))
		})
	}
	wg.Wait()
	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	require.Len(t, lines, 60)
	for _, line := range lines {
		assert.True(t, json.Valid([]byte(line)))
	}
}

// Exercise the public factory with a real stdout pipe as the mock backend.
func TestFactoryOutput(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	original := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = original; _ = w.Close(); _ = r.Close() })
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Source = "factory-source"
	set := exportertest.NewNopSettings(factory.Type())
	e, err := factory.CreateLogs(t.Context(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, e.Start(t.Context(), componenttest.NewNopHost()))
	require.NoError(t, e.ConsumeLogs(t.Context(), prepareLogs()))
	require.NoError(t, e.Shutdown(t.Context()))
	require.NoError(t, w.Close())
	os.Stdout = original
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"source":"factory-source"`)
}

func TestMetricsCancellationBetweenEvents(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	writes := 0
	writer := writerFunc(func(b []byte) (int, error) {
		writes++
		cancel()
		return len(b), nil
	})
	md := prepareMetricsData("test.metric")
	md.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().AppendEmpty().SetIntValue(42)
	e := newModinputExporter(&Config{}, componenttest.NewNopTelemetrySettings(), writer)
	require.ErrorIs(t, e.pushMetrics(ctx, md), context.Canceled)
	assert.Equal(t, 1, writes)
}
