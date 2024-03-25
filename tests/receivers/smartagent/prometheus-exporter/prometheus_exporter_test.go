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
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestPrometheusExporterProvidesOTelInternalMetrics(t *testing.T) {
	checkGoldenFile(t, "internal_metrics_config.yaml", "expected_internal.yaml",
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreMetricAttributeValue("service_instance_id"),
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
	checkGoldenFile(t, "httpd_metrics_config.yaml", "expected_httpd.yaml",
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreStartTimestamp())
}

func TestPrometheusExporterScrapesTargetsWithFilter(t *testing.T) {
	checkGoldenFile(t, "httpd_metrics_config_with_filter.yaml", "expected_httpd_filtered.yaml",
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreStartTimestamp())
}

func checkGoldenFile(t *testing.T, configFile string, expectedFilePath string, options ...pmetrictest.CompareMetricsOption) {
	f := otlpreceiver.NewFactory()
	port := testutils.GetAvailablePort(t)
	c := f.CreateDefaultConfig().(*otlpreceiver.Config)
	c.GRPC.NetAddr.Endpoint = fmt.Sprintf("localhost:%d", port)
	sink := &consumertest.MetricsSink{}
	receiver, err := f.CreateMetricsReceiver(context.Background(), receivertest.NewNopCreateSettings(), c, sink)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, receiver.Shutdown(context.Background()))
	})
	logger, _ := zap.NewDevelopment()

	dockerHost := "0.0.0.0"
	if runtime.GOOS == "darwin" {
		dockerHost = "host.docker.internal"
	}
	p, err := testutils.NewCollectorContainer().
		WithConfigPath(filepath.Join("testdata", configFile)).
		WithLogger(logger).
		WithEnv(map[string]string{"OTLP_ENDPOINT": fmt.Sprintf("%s:%d", dockerHost, port)}).
		Build()
	require.NoError(t, err)
	require.NoError(t, p.Start())
	t.Cleanup(func() {
		require.NoError(t, p.Shutdown())
	})

	expected, err := golden.ReadMetrics(filepath.Join("testdata", expectedFilePath))
	require.NoError(t, err)

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(sink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		err := pmetrictest.CompareMetrics(expected, sink.AllMetrics()[len(sink.AllMetrics())-1], options...)
		assert.NoError(tt, err)
	}, 30*time.Second, 1*time.Second)
}
