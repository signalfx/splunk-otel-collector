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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestSplunkPlatformLogsEffectiveConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	collector, shutdown := tc.SplunkOtelCollectorContainer(
		"",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithArgs(
				"--config", "/etc/otel/collector/splunk_logs_config_linux.yaml",
			).WithEnv(map[string]string{
				"SPLUNK_PLATFORM_TOKEN":              "test-token",
				"SPLUNK_PLATFORM_URL":                "https://http-inputs-test.splunkcloud.com/services/collector",
				"SPLUNK_PLATFORM_LOGS_INDEX":         "main",
				"SPLUNK_LISTEN_INTERFACE":            "127.0.0.1",
				"SPLUNK_FILE_STORAGE_EXTENSION_PATH": "/tmp/filelogs",
			})
		},
	)
	defer shutdown()

	config := collector.EffectiveConfig(t)

	// Verify the config contains only the platform logs components (no o11y components).
	service, ok := config["service"].(map[string]any)
	require.True(t, ok)
	pipelines, ok := service["pipelines"].(map[string]any)
	require.True(t, ok)

	// Only the platform logs pipeline is present.
	require.Contains(t, pipelines, "logs/hec")
	require.NotContains(t, pipelines, "metrics")
	require.NotContains(t, pipelines, "traces")

	logsHec, ok := pipelines["logs/hec"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, []any{"file_log/varlog"}, logsHec["receivers"])
	require.Equal(t, []any{"memory_limiter", "resourcedetection"}, logsHec["processors"])
	require.Equal(t, []any{"splunk_hec/logs"}, logsHec["exporters"])

	// The platform HEC exporter points at the configured URL and index.
	exporters, ok := config["exporters"].(map[string]any)
	require.True(t, ok)
	hecLogs, ok := exporters["splunk_hec/logs"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "https://http-inputs-test.splunkcloud.com/services/collector", hecLogs["endpoint"])
	require.Equal(t, "main", hecLogs["index"])
	require.Equal(t, "<redacted>", hecLogs["token"])

	// Only the platform logs extensions are present (no o11y extensions).
	extensions, ok := config["extensions"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, extensions, "health_check")
	require.Contains(t, extensions, "zpages")
	require.Contains(t, extensions, "file_storage/filelogs")
	require.NotContains(t, extensions, "smartagent")
	require.NotContains(t, extensions, "headers_setter")

	// Extensions in the service block match.
	serviceExtensions, ok := service["extensions"].([]any)
	require.True(t, ok)
	require.ElementsMatch(t, []any{"health_check", "zpages", "file_storage/filelogs", "config_source_telemetry"}, serviceExtensions)

	// Only the platform logs processors are present.
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

func TestSplunkPlatformLogsWithO11yEffectiveConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	collector, shutdown := tc.SplunkOtelCollectorContainer(
		"",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithArgs(
				"--config", "/etc/otel/collector/agent_config.yaml",
				"--config", "/etc/otel/collector/splunk_logs_config_linux.yaml",
				"--feature-gates=confmap.enableMergeAppendOption",
			).WithEnv(map[string]string{
				"SPLUNK_ACCESS_TOKEN":                "not.real",
				"SPLUNK_REALM":                       "not.real",
				"SPLUNK_HEC_TOKEN":                   "not.real",
				"SPLUNK_INGEST_URL":                  "http://127.0.0.1:0",
				"SPLUNK_API_URL":                     "http://127.0.0.1:0",
				"SPLUNK_HEC_URL":                     "http://127.0.0.1:0",
				"SPLUNK_PLATFORM_TOKEN":              "test-token",
				"SPLUNK_PLATFORM_URL":                "https://http-inputs-test.splunkcloud.com/services/collector",
				"SPLUNK_PLATFORM_LOGS_INDEX":         "main",
				"SPLUNK_LISTEN_INTERFACE":            "127.0.0.1",
				"SPLUNK_FILE_STORAGE_EXTENSION_PATH": "/tmp/filelogs",
			})
		},
	)
	defer shutdown()

	config := collector.EffectiveConfig(t)

	// Verify the merged config contains both the o11y pipelines from agent_config.yaml
	// and the platform logs pipeline from splunk_logs_config_linux.yaml.
	service, ok := config["service"].(map[string]any)
	require.True(t, ok)
	pipelines, ok := service["pipelines"].(map[string]any)
	require.True(t, ok)

	// agent_config.yaml pipelines are present
	require.Contains(t, pipelines, "metrics")
	require.Contains(t, pipelines, "traces")
	require.Contains(t, pipelines, "logs")

	// splunk_logs_config_linux.yaml pipeline is present
	require.Contains(t, pipelines, "logs/hec")

	// logs/hec uses the platform exporter
	logsHec, ok := pipelines["logs/hec"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, []any{"splunk_hec/logs"}, logsHec["exporters"])
	require.Equal(t, []any{"file_log/varlog"}, logsHec["receivers"])

	// splunk_hec/logs exporter is present in the merged config
	exporters, ok := config["exporters"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, exporters, "splunk_hec/logs")

	// Both configs' extensions are present in the merged service extensions list.
	// agent_config.yaml declares: headers_setter, health_check, http_forwarder,
	// http_forwarder/opamp_splunk_o11y, zpages, smartagent (opamp/splunk_o11y is
	// removed at startup unless --feature-gates=+splunk.opamp.enabled; config_source_telemetry
	// is injected by the collector itself).
	// splunk_logs_config_linux.yaml adds: file_storage/filelogs.
	serviceExtensions, ok := service["extensions"].([]any)
	require.True(t, ok)
	require.ElementsMatch(t, []any{
		"headers_setter",
		"health_check",
		"http_forwarder",
		"http_forwarder/opamp_splunk_o11y",
		"zpages",
		"smartagent",
		"config_source_telemetry",
		"file_storage/filelogs",
	}, serviceExtensions)

	// Both extension definitions are present in the extensions config block.
	extensions, ok := config["extensions"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, extensions, "health_check")
	require.Contains(t, extensions, "smartagent")
	require.Contains(t, extensions, "headers_setter")
	require.Contains(t, extensions, "http_forwarder")
	require.Contains(t, extensions, "http_forwarder/opamp_splunk_o11y")
	require.Contains(t, extensions, "zpages")
	require.Contains(t, extensions, "file_storage/filelogs")

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
	}, processors)
}

func TestSplunkPlatformLogsConfig(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("syslog-based log collection only supported on Linux")
	}
	tc := testutils.NewHECTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownHECReceiverSink()

	path, err := filepath.Abs("../../cmd/otelcol/config/collector/splunk_logs_config_linux.yaml")
	require.NoError(t, err)

	testMessage := fmt.Sprintf("splunk-platform-logs-test-message-%s", tc.ID)

	// Write the message before starting the collector so it is present in
	// /var/log/auth.log (matched by *auth* glob) when the collector reads from beginning.
	require.NoError(t, exec.Command("logger", "-p", "auth.info", "-t", "otelcol", testMessage).Run())

	_, shutdown := tc.SplunkOtelCollectorProcess(path,
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(map[string]string{
				"SPLUNK_PLATFORM_TOKEN":              "test-token",
				"SPLUNK_PLATFORM_URL":                tc.HECEndpointForCollector,
				"SPLUNK_PLATFORM_LOGS_INDEX":         "test-index",
				"SPLUNK_MEMORY_LIMIT_MIB":            "256",
				"SPLUNK_FILE_STORAGE_EXTENSION_PATH": t.TempDir(),
				"SPLUNK_LISTEN_INTERFACE":            "127.0.0.1",
			})
		},
	)
	defer shutdown()

	assertTestMessageReceived(t, tc.HECReceiverSink, testMessage)
}

// TestSplunkPlatformLogsWithO11yDataFlow verifies that when both the agent and platform logs configs
// are loaded together, the platform logs pipeline delivers file logs to the HEC sink and the o11y
// metrics pipeline (host_metrics → otlp_grpc/gateway) delivers metrics to the OTLP sink.
func TestSplunkPlatformLogsWithO11yDataFlow(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("syslog-based log collection only supported on Linux")
	}

	// Use NewTestcase for the OTLP sink (metrics) and create a separate HEC sink for platform logs.
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
	logsPath, err := filepath.Abs("../../cmd/otelcol/config/collector/splunk_logs_config_linux.yaml")
	require.NoError(t, err)

	testMessage := fmt.Sprintf("splunk-platform-logs-test-message-%s", tc.ID)

	// Write the message before starting the collector so it is present in
	// /var/log/auth.log (matched by *auth* glob) when the collector reads from beginning.
	require.NoError(t, exec.Command("logger", "-p", "auth.info", "-t", "otelcol", testMessage).Run())

	// Write an override config to a temp file that points otlp_grpc/gateway at the test OTLP sink
	// and redirects the metrics pipeline to use it. SPLUNK_GATEWAY_URL appends :4317 which
	// conflicts with the randomly-assigned test port, so we override the endpoint here instead.
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
					"--config", logsPath,
					"--config", overrideFile.Name(),
					"--feature-gates=confmap.enableMergeAppendOption",
					"--set=service.telemetry.metrics.level=none",
				).
				WithEnv(map[string]string{
					"SPLUNK_ACCESS_TOKEN":                "not.real",
					"SPLUNK_REALM":                       "not.real",
					"SPLUNK_HEC_TOKEN":                   "not.real",
					"SPLUNK_INGEST_URL":                  "http://127.0.0.1:0",
					"SPLUNK_API_URL":                     "http://127.0.0.1:0",
					"SPLUNK_HEC_URL":                     "http://127.0.0.1:0",
					"SPLUNK_GATEWAY_URL":                 "127.0.0.1",
					"SPLUNK_PLATFORM_TOKEN":              "test-token",
					"SPLUNK_PLATFORM_URL":                hecEndpointForCollector,
					"SPLUNK_PLATFORM_LOGS_INDEX":         "test-index",
					"SPLUNK_MEMORY_LIMIT_MIB":            "256",
					"SPLUNK_FILE_STORAGE_EXTENSION_PATH": t.TempDir(),
					"SPLUNK_LISTEN_INTERFACE":            "127.0.0.1",
				})
		},
	)
	defer shutdown()

	// Verify platform logs flow: auth.log message arrives at the HEC sink.
	assertTestMessageReceived(t, hecSink, testMessage)

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

func assertTestMessageReceived(t *testing.T, sink *testutils.HECReceiverSink, testMessage string) {
	t.Helper()
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		require.Positive(c, sink.LogRecordCount())

		found := false
		for _, logs := range sink.AllLogs() {
			for i := range logs.ResourceLogs().Len() {
				rl := logs.ResourceLogs().At(i)
				for j := range rl.ScopeLogs().Len() {
					for k := range rl.ScopeLogs().At(j).LogRecords().Len() {
						record := rl.ScopeLogs().At(j).LogRecords().At(k)
						if strings.Contains(record.Body().Str(), testMessage) {
							found = true

							sourcetype, ok := rl.Resource().Attributes().Get("com.splunk.sourcetype")
							assert.True(c, ok, "com.splunk.sourcetype attribute missing")
							assert.Equal(c, "linux:varlog", sourcetype.Str())

							source, ok := rl.Resource().Attributes().Get("com.splunk.source")
							assert.True(c, ok, "com.splunk.source attribute missing")
							assert.Equal(c, "/var/log/auth.log", source.Str())
						}
					}
				}
			}
		}
		require.True(c, found, "test log message not found in HEC sink")
	}, 60*time.Second, 500*time.Millisecond)
}
