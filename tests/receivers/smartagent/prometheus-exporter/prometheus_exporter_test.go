// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build smartagent_integration

package tests

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestPrometheusExporterProvidesOTelInternalMetrics(t *testing.T) {
	testutils.CheckGoldenFile(t, "internal_metrics_config.yaml", "expected_internal.yaml",
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreMetricAttributeValue("service_instance_id"),
		pmetrictest.IgnoreMetricAttributeValue("service_version"),
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreStartTimestamp(),
		pmetrictest.IgnoreMetricValues(
			"otelcol_exporter_sent_metric_points",
			"otelcol_process_cpu_seconds",
			"otelcol_process_memory_rss",
			"otelcol_process_runtime_heap_alloc_bytes",
			"otelcol_process_runtime_total_alloc_bytes",
			"otelcol_process_runtime_total_sys_memory_bytes",
			"otelcol_process_uptime",
			"otelcol_receiver_accepted_metric_points",
			"otelcol_rpc_client_duration",
			"otelcol_rpc_client_duration_bucket",
			"otelcol_rpc_client_duration_count",
			"otelcol_rpc_client_request_size",
			"otelcol_rpc_client_request_size_bucket",
			"otelcol_rpc_client_request_size_count",
			"otelcol_rpc_client_requests_per_rpc",
			"otelcol_rpc_client_requests_per_rpc_bucket",
			"otelcol_rpc_client_requests_per_rpc_count",
			"otelcol_rpc_client_response_size",
			"otelcol_rpc_client_response_size_bucket",
			"otelcol_rpc_client_response_size_count",
			"otelcol_rpc_client_responses_per_rpc",
			"otelcol_rpc_client_responses_per_rpc_bucket",
			"otelcol_rpc_client_responses_per_rpc_count",
		))
}

func TestPrometheusExporterScrapesTargets(t *testing.T) {
	testutils.CheckGoldenFile(t, "httpd_metrics_config.yaml", "expected_httpd.yaml",
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreStartTimestamp())
}

func TestPrometheusExporterScrapesTargetsWithFilter(t *testing.T) {
	testutils.CheckGoldenFile(t, "httpd_metrics_config_with_filter.yaml", "expected_httpd_filtered.yaml",
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreStartTimestamp())
}
