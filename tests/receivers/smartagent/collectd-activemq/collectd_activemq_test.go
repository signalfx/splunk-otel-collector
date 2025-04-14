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

func TestCollectdActiveMQReceiverProvidesAllMetrics(t *testing.T) {
	metricNames := []string{
		"counter.amq.TotalConnectionsCount",
		"gauge.amq.TotalConsumerCount",
		"gauge.amq.TotalDequeueCount",
		"gauge.amq.TotalEnqueueCount",
		"gauge.amq.TotalMessageCount",
		"gauge.amq.TotalProducerCount",
		"gauge.amq.queue.AverageBlockedTime",
		"gauge.amq.queue.AverageEnqueueTime",
		"gauge.amq.queue.AverageMessageSize",
		"gauge.amq.queue.BlockedSends",
		"gauge.amq.queue.ConsumerCount",
		"gauge.amq.queue.DequeueCount",
		"gauge.amq.queue.EnqueueCount",
		"gauge.amq.queue.ExpiredCount",
		"gauge.amq.queue.ForwardCount",
		"gauge.amq.queue.InFlightCount",
		"gauge.amq.queue.ProducerCount",
		"gauge.amq.queue.QueueSize",
		"gauge.amq.queue.TotalBlockedTime",
		"gauge.amq.topic.AverageBlockedTime",
		"gauge.amq.topic.AverageEnqueueTime",
		"gauge.amq.topic.AverageMessageSize",
		"gauge.amq.topic.BlockedSends",
		"gauge.amq.topic.ConsumerCount",
		"gauge.amq.topic.DequeueCount",
		"gauge.amq.topic.EnqueueCount",
		"gauge.amq.topic.ExpiredCount",
		"gauge.amq.topic.ForwardCount",
		"gauge.amq.topic.InFlightCount",
		"gauge.amq.topic.ProducerCount",
		"gauge.amq.topic.QueueSize",
		"gauge.amq.topic.TotalBlockedTime",
		"gauge.jvm.threads.count",
		"gauge.loaded_classes",
		"invocations",
		"jmx_memory.committed",
		"jmx_memory.init",
		"jmx_memory.max",
		"jmx_memory.used",
		"total_time_in_ms.collection_time",
	}
	testutils.CheckMetricsPresence(t, metricNames, "all_metrics_config.yaml")
}
func TestCollectdActiveMQReceiverProvidesDefaultMetrics(t *testing.T) {
	metricNames := []string{
		"counter.amq.TotalConnectionsCount",
		"gauge.amq.TotalConsumerCount",
		"gauge.amq.TotalEnqueueCount",
		"gauge.amq.TotalMessageCount",
		"gauge.amq.TotalProducerCount",
		"gauge.amq.queue.AverageEnqueueTime",
		"gauge.amq.queue.ConsumerCount",
		"gauge.amq.queue.DequeueCount",
		"gauge.amq.queue.EnqueueCount",
		"gauge.amq.queue.ExpiredCount",
		"gauge.amq.queue.InFlightCount",
		"gauge.amq.queue.ProducerCount",
		"gauge.amq.queue.QueueSize",
		"gauge.amq.topic.AverageEnqueueTime",
		"gauge.amq.topic.ConsumerCount",
		"gauge.amq.topic.EnqueueCount",
		"gauge.amq.topic.ExpiredCount",
		"gauge.amq.topic.InFlightCount",
		"gauge.amq.topic.ProducerCount",
		"gauge.amq.topic.QueueSize",
		"gauge.jvm.threads.count",
		"gauge.loaded_classes",
		"invocations",
		"jmx_memory.committed",
		"jmx_memory.init",
		"jmx_memory.max",
		"jmx_memory.used",
		"total_time_in_ms.collection_time",
	}
	testutils.CheckMetricsPresence(t, metricNames, "default_metrics_config.yaml")
}
