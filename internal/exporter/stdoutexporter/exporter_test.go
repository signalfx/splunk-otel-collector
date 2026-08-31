// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package stdoutexporter

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/signalfx/splunk-otel-collector/internal/auth"
)

const (
	DefaultSourceTypeLabel = "com.splunk.sourcetype"
	DefaultSourceLabel     = "com.splunk.source"
	DefaultIndexLabel      = "com.splunk.index"
	DefaultNameLabel       = "otel.log.name"
)

var configPath = "./testdata/test_config.yaml"

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

type cfg struct {
	event      string
	index      string
	source     string
	sourcetype string
}

type telemetryType string

var (
	metricsType = telemetryType("metrics")
	logsType    = telemetryType("logs")
	tracesType  = telemetryType("traces")
)

type testCfg struct {
	name                   string
	config                 *cfg
	startTime              string
	telType                telemetryType
	expectedResultFilePath string
}

func logsTest(t *testing.T, test testCfg) {
	settings := exportertest.NewNopSettings(exportertest.NopType)
	var logs plog.Logs
	if test.config.index != "main" {
		logs = prepareLogsNonDefaultParams(test.config.index, test.config.source, test.config.sourcetype, test.config.event)
	} else {
		logs = prepareLogs()
	}

	exporter, err := newLogsExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)

	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)
	out := CaptureStdout(t, func() {
		err = exporter.ConsumeLogs(context.WithValue(t.Context(), auth.ContextKey, auth.HecTokenConfig{
			AllowedIndexes: []string{
				"main",
				test.config.index,
			},
		}), logs)
	})

	require.NotEmpty(t, out)
	require.NoError(t, err, "Must not error while sending log data")
	expectedJSON, err := os.ReadFile(test.expectedResultFilePath)
	require.NoError(t, err)
	require.JSONEq(t, out, string(expectedJSON))
}

func metricsTest(t *testing.T, test testCfg) {
	settings := exportertest.NewNopSettings(exportertest.NopType)
	metricData := prepareMetricsData(test.config.event)

	exporter, err := newMetricsExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)
	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)

	out := CaptureStdout(t, func() {
		err = exporter.ConsumeMetrics(context.WithValue(t.Context(), auth.ContextKey, auth.HecTokenConfig{
			AllowedIndexes: []string{
				"main",
				test.config.index,
			},
		}), metricData)
	})
	require.NotEmpty(t, out)
	require.NoError(t, err, "Must not error while sending metric data")
	expectedJSON, err := os.ReadFile(test.expectedResultFilePath)
	require.NoError(t, err)
	require.JSONEq(t, out, string(expectedJSON))
}

func tracesTest(t *testing.T, test testCfg) {
	settings := exportertest.NewNopSettings(exportertest.NopType)
	tracesData := prepareTracesData(test.config.index, test.config.source, test.config.sourcetype)

	exporter, err := newTracesExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)
	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)

	out := CaptureStdout(t, func() {
		err = exporter.ConsumeTraces(context.WithValue(t.Context(), auth.ContextKey, auth.HecTokenConfig{
			AllowedIndexes: []string{
				"main",
				test.config.index,
			},
		}), tracesData)
	})
	require.NotEmpty(t, out)
	require.NoError(t, err, "Must not error while sending trace data")
	expectedJSON, err := os.ReadFile(test.expectedResultFilePath)
	require.NoError(t, err)
	require.JSONEq(t, out, string(expectedJSON))
}

func TestSplunkHecExporter(t *testing.T) {
	eventIndex, err := getConfigVariable(configPath, "EVENT_INDEX")
	require.NoError(t, err)
	metricIndex, err := getConfigVariable(configPath, "METRIC_INDEX")
	require.NoError(t, err)
	traceIndex, err := getConfigVariable(configPath, "TRACE_INDEX")
	require.NoError(t, err)
	tests := []testCfg{
		{
			name: "Events to Splunk - logs",
			config: &cfg{
				event:      "test log",
				index:      "main",
				source:     "otel",
				sourcetype: "st-otel",
			},
			startTime:              "-3h@h",
			telType:                logsType,
			expectedResultFilePath: "./testdata/expected_hec_log.json",
		},
		{
			name: "Events to Splunk - Non default index",
			config: &cfg{
				event:      "This is my new event! And some number 101",
				index:      eventIndex,
				source:     "otel-source",
				sourcetype: "sck-otel-st",
			},
			startTime:              "-1m@m",
			telType:                logsType,
			expectedResultFilePath: "./testdata/expected_hec_log_non_default_index.json",
		},
		{
			name: "Events to Splunk - metrics",
			config: &cfg{
				event:      "test.metric",
				index:      metricIndex,
				source:     "otel",
				sourcetype: "st-otel",
			},
			startTime:              "",
			telType:                metricsType,
			expectedResultFilePath: "./testdata/expected_hec_metric.json",
		},
		{
			name: "Events to Splunk - traces",
			config: &cfg{
				event:      "",
				index:      traceIndex,
				source:     "trace-source",
				sourcetype: "trace-sourcetype",
			},
			startTime:              "-1m@m",
			telType:                tracesType,
			expectedResultFilePath: "./testdata/expected_hec_trace.json",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.telType {
			case logsType:
				logsTest(t, test)
			case metricsType:
				metricsTest(t, test)
			case tracesType:
				tracesTest(t, test)
			default:
				assert.Fail(t, "Telemetry type must be set to one of the following values: metrics, traces, or logs.")
			}
		})
	}
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

func TestBadIndexLogs(t *testing.T) {
	logs := prepareLogsNonDefaultParams("foo", "", "", "event")
	settings := exportertest.NewNopSettings(exportertest.NopType)

	exporter, err := newLogsExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)

	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)
	err = exporter.ConsumeLogs(context.WithValue(t.Context(), auth.ContextKey, auth.HecTokenConfig{
		AllowedIndexes: []string{
			"main",
		},
	}), logs)

	require.EqualError(t, err, `index "foo" is not allowed`)
}

func TestBadIndexMetrics(t *testing.T) {
	m := prepareMetricsData("foo")
	m.ResourceMetrics().At(0).Resource().Attributes().PutStr("com.splunk.index", "foo")
	settings := exportertest.NewNopSettings(exportertest.NopType)

	exporter, err := newMetricsExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)

	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)
	err = exporter.ConsumeMetrics(context.WithValue(t.Context(), auth.ContextKey, auth.HecTokenConfig{
		AllowedIndexes: []string{
			"main",
		},
	}), m)

	require.EqualError(t, err, `index "foo" is not allowed`)
}

func TestBadIndexTraces(t *testing.T) {
	td := prepareTracesData("bar", "", "")
	settings := exportertest.NewNopSettings(exportertest.NopType)

	exporter, err := newTracesExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)

	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)
	err = exporter.ConsumeTraces(context.WithValue(t.Context(), auth.ContextKey, auth.HecTokenConfig{
		AllowedIndexes: []string{
			"main",
		},
	}), td)

	require.EqualError(t, err, `index "bar" is not allowed`)
}

func TestBadIndexTracesResource(t *testing.T) {
	td := prepareTracesData("foo", "", "")
	settings := exportertest.NewNopSettings(exportertest.NopType)

	exporter, err := newTracesExporter(t.Context(), settings, createDefaultConfig())
	require.NoError(t, err)

	err = exporter.Start(t.Context(), componenttest.NewNopHost())
	require.NoError(t, err)
	err = exporter.ConsumeTraces(context.WithValue(t.Context(), auth.ContextKey, auth.HecTokenConfig{
		AllowedIndexes: []string{
			"main",
		},
	}), td)

	require.EqualError(t, err, `index "foo" is not allowed`)
}

type IntegrationTestsConfig struct {
	Host           string `yaml:"HOST"`
	User           string `yaml:"USER"`
	Password       string `yaml:"PASSWORD"`
	UIPort         string `yaml:"UI_PORT"`
	HecPort        string `yaml:"HEC_PORT"`
	ManagementPort string `yaml:"MANAGEMENT_PORT"`
	EventIndex     string `yaml:"EVENT_INDEX"`
	MetricIndex    string `yaml:"METRIC_INDEX"`
	TraceIndex     string `yaml:"TRACE_INDEX"`
	HecToken       string `yaml:"HEC_TOKEN"`
	SplunkImage    string `yaml:"SPLUNK_IMAGE"`
}

func getConfigVariable(configPath, key string) (string, error) {
	// Read YAML file
	fileData, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	var config IntegrationTestsConfig
	err = yaml.Unmarshal(fileData, &config)
	if err != nil {
		return "", fmt.Errorf("error decoding YAML: %w", err)
	}

	switch key {
	case "EVENT_INDEX":
		return config.EventIndex, nil
	case "METRIC_INDEX":
		return config.MetricIndex, nil
	case "TRACE_INDEX":
		return config.TraceIndex, nil
	default:
		fmt.Println("invalid field")
		return "None", fmt.Errorf("invalid field %v", key)
	}
}

// CaptureStdout captures all stdout output generated by fn and returns it as a string.
func CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	os.Stdout = w

	outputCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		outputCh <- buf.String()
	}()

	fn()

	// TODO: Find a way to synchronize without sleep.
	time.Sleep(1 * time.Second)

	_ = w.Close()
	os.Stdout = original

	return <-outputCh
}
