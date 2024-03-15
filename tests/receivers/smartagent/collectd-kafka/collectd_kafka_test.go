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

func TestCollectdKafkaReceiversAllBrokerMetrics(t *testing.T) {
	metricNames := []string{
		"counter.kafka-bytes-in",
		"counter.kafka-bytes-out",
		"counter.kafka-isr-expands",
		"counter.kafka-isr-shrinks",
		"counter.kafka-leader-election-rate",
		"counter.kafka-messages-in",
		"counter.kafka-unclean-elections-rate",
		"counter.kafka.fetch-consumer.total-time.count",
		"counter.kafka.fetch-follower.total-time.count",
		"counter.kafka.produce.total-time.count",
		"gauge.jvm.threads.count",
		"gauge.kafka-active-controllers",
		"gauge.kafka-max-lag",
		"gauge.kafka-offline-partitions-count",
		"gauge.kafka-request-queue",
		"gauge.kafka-underreplicated-partitions",
		"gauge.kafka.fetch-consumer.total-time.99th",
		"gauge.kafka.fetch-consumer.total-time.median",
		"gauge.kafka.fetch-follower.total-time.99th",
		"gauge.kafka.fetch-follower.total-time.median",
		"gauge.kafka.produce.total-time.99th",
		"counter.kafka.logs.flush-time.count",
		"gauge.kafka.logs.flush-time.median",
		"gauge.kafka.logs.flush-time.99th",
		"gauge.kafka.produce.total-time.median",
		"gauge.loaded_classes",
		"invocations",
		"jmx_memory.committed",
		"jmx_memory.init",
		"jmx_memory.max",
		"jmx_memory.used",
		"total_time_in_ms.collection_time",
	}
	checkMetricsPresence(t, metricNames, "all_broker_metrics_config.yaml")
}

func TestCollectdKafkaReceiversAllConsumerMetrics(t *testing.T) {
	metricNames := []string{
		"gauge.jvm.threads.count",
		"gauge.kafka.consumer.bytes-consumed-rate",
		"gauge.kafka.consumer.fetch-rate",
		"gauge.kafka.consumer.fetch-size-avg",
		"gauge.kafka.consumer.records-consumed-rate",
		"gauge.kafka.consumer.records-lag-max",
		"gauge.loaded_classes",
		"invocations",
		"jmx_memory.committed",
		"jmx_memory.init",
		"jmx_memory.max",
		"jmx_memory.used",
		"total_time_in_ms.collection_time",
	}
	checkMetricsPresence(t, metricNames, "all_producer_metrics_config.yaml")
}

func TestCollectdKafkaReceiversAllProducerMetrics(t *testing.T) {
	metricNames := []string{
		"gauge.jvm.threads.count",
		"gauge.kafka.producer.byte-rate",
		"gauge.kafka.producer.compression-rate",
		"gauge.kafka.producer.io-wait-time-ns-avg",
		"gauge.kafka.producer.outgoing-byte-rate",
		"gauge.kafka.producer.record-error-rate",
		"gauge.kafka.producer.record-retry-rate",
		"gauge.kafka.producer.record-send-rate",
		"gauge.kafka.producer.request-latency-avg",
		"gauge.kafka.producer.request-rate",
		"gauge.kafka.producer.response-rate",
		"gauge.loaded_classes",
		"invocations",
		"jmx_memory.committed",
		"jmx_memory.init",
		"jmx_memory.max",
		"jmx_memory.used",
		"total_time_in_ms.collection_time",
	}
	checkMetricsPresence(t, metricNames, "all_producer_metrics_config.yaml")
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
