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

func TestCollectdSolrReceiverProvidesAllMetrics(t *testing.T) {
	checkGoldenFile(t, "all_metrics_config.yaml", "all_expected.yaml",
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreMetricValues(
			"gauge.solr.core_usablespace",
			"counter.solr.node_collections_requests",
			"counter.solr.node_metrics_requests",
			"counter.solr.http_2xx_responses",
			"counter.solr.http_requests",
			"gauge.solr.jetty_request_latency",
			"gauge.solr.jvm_heap_usage",
			"gauge.solr.jvm_total_memory_used",
			"gauge.solr.jvm_memory_pools_Metaspace_usage",
			"gauge.solr.searcher_warmup",
			"gauge.solr.core_totalspace",
			"gauge.solr.core_index_size",
			"gauge.solr.search_query_response",
			"gauge.solr.update_request_handler_response",
		),
		pmetrictest.IgnoreResourceAttributeValue("node"),
	)
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
	//require.NoError(t, err)

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(sink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		err := pmetrictest.CompareMetrics(expected, sink.AllMetrics()[len(sink.AllMetrics())-1], options...)
		assert.NoError(tt, err)
	}, 10*time.Second, 1*time.Second)
}
