// Copyright Splunk, Inc.
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

//go:build integration

package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestSplunkPlatformMetricsEffectiveConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	collector, shutdown := tc.SplunkOtelCollectorContainer(
		"",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithArgs(
				"--config", "/etc/otel/collector/splunk_metrics_config_linux.yaml",
			).WithEnv(map[string]string{
				"SPLUNK_PLATFORM_TOKEN":         "test-token",
				"SPLUNK_PLATFORM_URL":           "https://http-inputs-test.splunkcloud.com/services/collector",
				"SPLUNK_PLATFORM_METRICS_INDEX": "main",
				"SPLUNK_LISTEN_INTERFACE":       "127.0.0.1",
			})
		},
	)
	defer shutdown()

	config := collector.EffectiveConfig(t)

	// Only the platform metrics pipeline is present (no logs, no traces, no o11y components).
	service, ok := config["service"].(map[string]any)
	require.True(t, ok)
	pipelines, ok := service["pipelines"].(map[string]any)
	require.True(t, ok)

	require.Contains(t, pipelines, "metrics/platform")
	require.NotContains(t, pipelines, "logs")
	require.NotContains(t, pipelines, "traces")
	require.NotContains(t, pipelines, "metrics") // no o11y metrics pipeline

	metricsPlatform, ok := pipelines["metrics/platform"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, []any{"host_metrics/platform"}, metricsPlatform["receivers"])
	require.Equal(t, []any{"memory_limiter", "resourcedetection"}, metricsPlatform["processors"])
	require.Equal(t, []any{"splunk_hec/metrics"}, metricsPlatform["exporters"])

	// The platform HEC exporter points at the configured URL and index.
	exporters, ok := config["exporters"].(map[string]any)
	require.True(t, ok)
	hecMetrics, ok := exporters["splunk_hec/metrics"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "https://http-inputs-test.splunkcloud.com/services/collector", hecMetrics["endpoint"])
	require.Equal(t, "main", hecMetrics["index"])
	require.Equal(t, "<redacted>", hecMetrics["token"])

	// Only health_check and zpages extensions are present (no file_storage, no smartagent, etc.).
	extensions, ok := config["extensions"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, extensions, "health_check")
	require.Contains(t, extensions, "zpages")
	require.NotContains(t, extensions, "file_storage/filelogs")
	require.NotContains(t, extensions, "smartagent")

	serviceExtensions, ok := service["extensions"].([]any)
	require.True(t, ok)
	require.ElementsMatch(t, []any{"health_check", "zpages", "config_source_telemetry"}, serviceExtensions)

	// Only the platform metrics processors are present.
	processors, ok := config["processors"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, map[string]any{
		"memory_limiter": map[string]any{
			"check_interval": "2s",
			"limit_mib":      460,
		},
		"resourcedetection": map[string]any{
			"detectors": []any{"gcp", "ecs", "ec2", "azure", "system"},
			"override":  true,
		},
	}, processors)
}

func TestSplunkPlatformMetricsWithO11yEffectiveConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	collector, shutdown := tc.SplunkOtelCollectorContainer(
		"",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithArgs(
				"--config", "/etc/otel/collector/agent_config.yaml",
				"--config", "/etc/otel/collector/splunk_metrics_config_linux.yaml",
				"--feature-gates=confmap.enableMergeAppendOption",
			).WithEnv(map[string]string{
				"SPLUNK_ACCESS_TOKEN":           "not.real",
				"SPLUNK_REALM":                  "not.real",
				"SPLUNK_HEC_TOKEN":              "not.real",
				"SPLUNK_INGEST_URL":             "http://127.0.0.1:0",
				"SPLUNK_API_URL":                "http://127.0.0.1:0",
				"SPLUNK_HEC_URL":                "http://127.0.0.1:0",
				"SPLUNK_PLATFORM_TOKEN":         "test-token",
				"SPLUNK_PLATFORM_URL":           "https://http-inputs-test.splunkcloud.com/services/collector",
				"SPLUNK_PLATFORM_METRICS_INDEX": "main",
				"SPLUNK_LISTEN_INTERFACE":       "127.0.0.1",
			})
		},
	)
	defer shutdown()

	config := collector.EffectiveConfig(t)

	// Verify both o11y pipelines from agent_config.yaml and the platform metrics pipeline are present.
	service, ok := config["service"].(map[string]any)
	require.True(t, ok)
	pipelines, ok := service["pipelines"].(map[string]any)
	require.True(t, ok)

	// agent_config.yaml pipelines are present.
	require.Contains(t, pipelines, "metrics")
	require.Contains(t, pipelines, "traces")
	require.Contains(t, pipelines, "logs")

	// splunk_metrics_config_linux.yaml pipeline is present.
	require.Contains(t, pipelines, "metrics/platform")

	metricsPlatform, ok := pipelines["metrics/platform"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, []any{"host_metrics/platform"}, metricsPlatform["receivers"])
	require.Equal(t, []any{"splunk_hec/metrics"}, metricsPlatform["exporters"])

	// splunk_hec/metrics exporter is present in the merged config.
	exporters, ok := config["exporters"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, exporters, "splunk_hec/metrics")

	// All extensions from agent_config.yaml plus health_check from metrics config are present.
	serviceExtensions, ok := service["extensions"].([]any)
	require.True(t, ok)
	require.Contains(t, serviceExtensions, "health_check")
	require.Contains(t, serviceExtensions, "headers_setter")
	require.Contains(t, serviceExtensions, "http_forwarder")
	require.Contains(t, serviceExtensions, "http_forwarder/opamp_splunk_o11y")
	require.Contains(t, serviceExtensions, "zpages")

	extensions, ok := config["extensions"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, extensions, "health_check")
	require.Contains(t, extensions, "headers_setter")
	require.Contains(t, extensions, "http_forwarder")
	require.Contains(t, extensions, "http_forwarder/opamp_splunk_o11y")
	require.Contains(t, extensions, "zpages")

	// Processors from both configs are present in the merged config.
	processors, ok := config["processors"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, map[string]any{
		"batch": map[string]any{
			"metadata_keys": []any{"X-SF-Token"},
		},
		"memory_limiter": map[string]any{
			"check_interval": "2s",
			"limit_mib":      460,
		},
		"resourcedetection": map[string]any{
			"detectors": []any{"gcp", "ecs", "ec2", "azure", "system"},
			"override":  true,
		},
		"resource/add_mode": map[string]any{
			"attributes": []any{
				map[string]any{
					"action": "insert",
					"value":  "agent",
					"key":    "otelcol.service.mode",
				},
			},
		},
	}, processors)
}

func TestSplunkPlatformMetricsConfig(t *testing.T) {
	tc := testutils.NewHECTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownHECReceiverSink()

	path, err := filepath.Abs("../../cmd/otelcol/config/collector/splunk_metrics_config_linux.yaml")
	require.NoError(t, err)

	_, shutdown := tc.SplunkOtelCollectorProcess(path,
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(map[string]string{
				"SPLUNK_PLATFORM_TOKEN":         "test-token",
				"SPLUNK_PLATFORM_URL":           tc.HECEndpointForCollector,
				"SPLUNK_PLATFORM_METRICS_INDEX": "test-index",
				"SPLUNK_MEMORY_LIMIT_MIB":       "256",
				"SPLUNK_LISTEN_INTERFACE":       "127.0.0.1",
			})
		},
	)
	defer shutdown()

	assertHostMetricsReceived(t, tc.HECReceiverSink)
}

// TestSplunkPlatformMetricsWithO11yDataFlow verifies that when both the agent and platform metrics configs
// are loaded together, the platform metrics pipeline delivers host metrics to the HEC sink and the o11y
// metrics pipeline (host_metrics → otlp_grpc/gateway) delivers metrics to the OTLP sink.
func TestSplunkPlatformMetricsWithO11yDataFlow(t *testing.T) {
	// Use NewTestcase for the OTLP sink (o11y metrics) and create a separate HEC sink for platform metrics.
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	hecSink, err := testutils.NewHECReceiverSink().WithEndpoint(fmt.Sprintf("0.0.0.0:%d", testutils.GetAvailablePort(t))).Build()
	require.NoError(t, err)
	require.NoError(t, hecSink.Start())
	defer func() { require.NoError(t, hecSink.Shutdown()) }()
	hecEndpointForCollector := "http://127.0.0.1:" + strings.Split(hecSink.Endpoint, ":")[1]

	agentPath, err := filepath.Abs("../../cmd/otelcol/config/collector/agent_config.yaml")
	require.NoError(t, err)
	metricsPath, err := filepath.Abs("../../cmd/otelcol/config/collector/splunk_metrics_config_linux.yaml")
	require.NoError(t, err)

	// Write an override config that points otlp_grpc/gateway at the test OTLP sink.
	overrideContent := "exporters:\n  otlp_grpc/gateway:\n    endpoint: " + tc.OTLPEndpointForCollector + "\n    tls:\n      insecure: true\nservice:\n  pipelines:\n    metrics:\n      exporters: [otlp_grpc/gateway]\n"
	overrideFile, err := os.CreateTemp(t.TempDir(), "otlp-override-*.yaml")
	require.NoError(t, err)
	_, err = overrideFile.WriteString(overrideContent)
	require.NoError(t, err)
	require.NoError(t, overrideFile.Close())

	_, shutdown := tc.SplunkOtelCollectorProcess("",
		func(collector testutils.Collector) testutils.Collector {
			return collector.
				WithArgs(
					"--set=service.telemetry.logs.level=debug",
					"--config", agentPath,
					"--config", metricsPath,
					"--config", overrideFile.Name(),
					"--feature-gates=confmap.enableMergeAppendOption",
					"--set=service.telemetry.metrics.level=none",
				).
				WithEnv(map[string]string{
					"SPLUNK_ACCESS_TOKEN":           "not.real",
					"SPLUNK_REALM":                  "not.real",
					"SPLUNK_HEC_TOKEN":              "not.real",
					"SPLUNK_INGEST_URL":             "http://127.0.0.1:0",
					"SPLUNK_API_URL":                "http://127.0.0.1:0",
					"SPLUNK_HEC_URL":                "http://127.0.0.1:0",
					"SPLUNK_GATEWAY_URL":            "127.0.0.1",
					"SPLUNK_PLATFORM_TOKEN":         "test-token",
					"SPLUNK_PLATFORM_URL":           hecEndpointForCollector,
					"SPLUNK_PLATFORM_METRICS_INDEX": "test-index",
					"SPLUNK_MEMORY_LIMIT_MIB":       "256",
					"SPLUNK_LISTEN_INTERFACE":       "127.0.0.1",
				})
		},
	)
	defer shutdown()

	// Verify platform metrics flow: host_metrics/platform data points arrive at the HEC sink.
	assertHostMetricsReceived(t, hecSink)

	// Verify o11y metrics flow: host_metrics (system.cpu.time) arrives at the OTLP sink.
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		found := false
		for _, metrics := range tc.OTLPReceiverSink.AllMetrics() {
			for i := range metrics.ResourceMetrics().Len() {
				rm := metrics.ResourceMetrics().At(i)
				for j := range rm.ScopeMetrics().Len() {
					for k := range rm.ScopeMetrics().At(j).Metrics().Len() {
						if rm.ScopeMetrics().At(j).Metrics().At(k).Name() == "system.cpu.time" {
							found = true
						}
					}
				}
			}
		}
		require.True(c, found, "host metric system.cpu.time not found in OTLP sink")
	}, 60*time.Second, 500*time.Millisecond)
}

func assertHostMetricsReceived(t *testing.T, sink *testutils.HECReceiverSink) {
	t.Helper()
	// splunk_hec/metrics sends metrics as HEC metric events; verify data points arrive.
	require.Eventually(t, func() bool {
		return sink.DataPointCount() > 0
	}, 60*time.Second, 500*time.Millisecond)
}
