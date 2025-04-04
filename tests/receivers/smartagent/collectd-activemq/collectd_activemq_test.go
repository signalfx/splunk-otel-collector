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
	"fmt"
	"testing"
	"time"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
	"github.com/stretchr/testify/assert"
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
	checkMetricsPresence(t, metricNames, "all_metrics_config.yaml")
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
	checkMetricsPresence(t, metricNames, "default_metrics_config.yaml")
}

func checkMetricsPresence(t *testing.T, metricNames []string, configFile string) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer(configFile)
	t.Cleanup(shutdown)

	missingMetrics := make(map[string]any, len(metricNames))
	for _, m := range metricNames {
		missingMetrics[m] = struct{}{}
	}

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		for i := 0; i < len(tc.OTLPReceiverSink.AllMetrics()); i++ {
			m := tc.OTLPReceiverSink.AllMetrics()[i]
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
