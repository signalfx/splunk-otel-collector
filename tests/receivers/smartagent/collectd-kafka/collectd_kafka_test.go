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
	"testing"

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
	testutils.CheckMetricsPresence(t, metricNames, "all_broker_metrics_config.yaml")
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
	testutils.CheckMetricsPresence(t, metricNames, "all_consumer_metrics_config.yaml")
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
	testutils.CheckMetricsPresence(t, metricNames, "all_producer_metrics_config.yaml")
}
