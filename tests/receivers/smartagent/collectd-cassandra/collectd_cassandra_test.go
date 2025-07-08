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
	testutils.CheckMetricsPresence(t, metricNames, "all_metrics_config.yaml")
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
	testutils.CheckMetricsPresence(t, metricNames, "default_metrics_config.yaml")
}
