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
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCollectdCassandraReceiverProvidesAllMetrics(t *testing.T) {
	metricNames := []string{
		"counter.cassandra.ClientRequest.CASRead.Latency.Count",
		"counter.cassandra.ClientRequest.CASRead.TotalLatency.Count",
		"counter.cassandra.ClientRequest.CASWrite.Latency.Count",
		"counter.cassandra.ClientRequest.CASWrite.TotalLatency.Count",
		"counter.cassandra.ClientRequest.RangeSlice.Latency.Count",
		"counter.cassandra.ClientRequest.RangeSlice.Timeouts.Count",
		"counter.cassandra.ClientRequest.RangeSlice.TotalLatency.Count",
		"counter.cassandra.ClientRequest.RangeSlice.Unavailables.Count",
		"counter.cassandra.ClientRequest.Read.Latency.Count",
		"counter.cassandra.ClientRequest.Read.Timeouts.Count",
		"counter.cassandra.ClientRequest.Read.TotalLatency.Count",
		"counter.cassandra.ClientRequest.Read.Unavailables.Count",
		"counter.cassandra.ClientRequest.Write.Latency.Count",
		"counter.cassandra.ClientRequest.Write.Timeouts.Count",
		"counter.cassandra.ClientRequest.Write.TotalLatency.Count",
		"counter.cassandra.ClientRequest.Write.Unavailables.Count",
		"counter.cassandra.Compaction.TotalCompactionsCompleted.Count",
		"counter.cassandra.Storage.Exceptions.Count",
		"counter.cassandra.Storage.Load.Count",
		"counter.cassandra.Storage.TotalHints.Count",
		"counter.cassandra.Storage.TotalHintsInProgress.Count",
		"gauge.cassandra.ClientRequest.CASRead.Latency.50thPercentile",
		"gauge.cassandra.ClientRequest.CASRead.Latency.99thPercentile",
		"gauge.cassandra.ClientRequest.CASRead.Latency.Max",
		"gauge.cassandra.ClientRequest.CASWrite.Latency.50thPercentile",
		"gauge.cassandra.ClientRequest.CASWrite.Latency.99thPercentile",
		"gauge.cassandra.ClientRequest.CASWrite.Latency.Max",
		"gauge.cassandra.ClientRequest.RangeSlice.Latency.50thPercentile",
		"gauge.cassandra.ClientRequest.RangeSlice.Latency.99thPercentile",
		"gauge.cassandra.ClientRequest.RangeSlice.Latency.Max",
		"gauge.cassandra.ClientRequest.Read.Latency.50thPercentile",
		"gauge.cassandra.ClientRequest.Read.Latency.99thPercentile",
		"gauge.cassandra.ClientRequest.Read.Latency.Max",
		"gauge.cassandra.ClientRequest.Write.Latency.50thPercentile",
		"gauge.cassandra.ClientRequest.Write.Latency.99thPercentile",
		"gauge.cassandra.ClientRequest.Write.Latency.Max",
		"gauge.cassandra.Compaction.PendingTasks.Value",
		"gauge.jvm.threads.count",
		"gauge.loaded_classes",
		"invocations",
		"jmx_memory.committed",
		"jmx_memory.init",
		"jmx_memory.max",
		"jmx_memory.used",
		"total_time_in_ms.collection_time",
	}
	checkMetricsPresence(t, metricNames, "all_metrics_config.yaml")
}

func TestCollectdCassandraReceiverProvidesDefaultMetrics(t *testing.T) {
	metricNames := []string{
		"counter.cassandra.ClientRequest.RangeSlice.Latency.Count",
		"counter.cassandra.ClientRequest.RangeSlice.Timeouts.Count",
		"counter.cassandra.ClientRequest.RangeSlice.Unavailables.Count",
		"counter.cassandra.ClientRequest.Read.Latency.Count",
		"counter.cassandra.ClientRequest.Read.Timeouts.Count",
		"counter.cassandra.ClientRequest.Read.Unavailables.Count",
		"counter.cassandra.ClientRequest.Write.Latency.Count",
		"counter.cassandra.ClientRequest.Write.Timeouts.Count",
		"counter.cassandra.ClientRequest.Write.Unavailables.Count",
		"counter.cassandra.Storage.Load.Count",
		"counter.cassandra.Storage.TotalHintsInProgress.Count",
		"gauge.cassandra.ClientRequest.RangeSlice.Latency.99thPercentile",
		"gauge.cassandra.ClientRequest.Read.Latency.50thPercentile",
		"gauge.cassandra.ClientRequest.Read.Latency.99thPercentile",
		"gauge.cassandra.ClientRequest.Read.Latency.Max",
		"gauge.cassandra.ClientRequest.Write.Latency.50thPercentile",
		"gauge.cassandra.ClientRequest.Write.Latency.99thPercentile",
		"gauge.cassandra.ClientRequest.Write.Latency.Max",
		"gauge.cassandra.Compaction.PendingTasks.Value",
		"gauge.jvm.threads.count",
		"gauge.loaded_classes",
		"invocations",
		"jmx_memory.committed",
		"jmx_memory.init",
		"jmx_memory.max",
		"jmx_memory.used",
		"total_time_in_ms.collection_time",
	}
	checkMetricsPresence(t, metricNames, "default_metrics_config.yaml")
}

func checkMetricsPresence(t *testing.T, metricNames []string, configFile string) {
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

	missingMetrics := make(map[string]any, len(metricNames))
	for _, m := range metricNames {
		missingMetrics[m] = struct{}{}
	}

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		for i := 0; i < len(sink.AllMetrics()); i++ {
			m := sink.AllMetrics()[i]
			for j := 0; j < m.ResourceMetrics().Len(); j++ {
				rm := m.ResourceMetrics().At(j)
				for k := 0; k < rm.ScopeMetrics().Len(); k++ {
					sm := rm.ScopeMetrics().At(k)
					for l := 0; l < sm.Metrics().Len(); l++ {
						delete(missingMetrics, sm.Metrics().At(l).Name())
					}
				}
			}
		}
		msg := "Missing metrics:\n"
		for k := range missingMetrics {
			msg += fmt.Sprintf("- %q\n", k)
		}
		assert.Len(tt, missingMetrics, 0, msg)
	}, 1*time.Minute, 1*time.Second)
}
