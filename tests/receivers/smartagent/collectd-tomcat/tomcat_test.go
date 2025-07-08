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

func TestCollectdTomcatReceiverProvidesDefaultMetrics(t *testing.T) {
	testutils.CheckMetricsPresence(t, []string{
		"counter.tomcat.GlobalRequestProcessor.bytesReceived",
		"counter.tomcat.GlobalRequestProcessor.bytesSent",
		"counter.tomcat.GlobalRequestProcessor.errorCount",
		"counter.tomcat.GlobalRequestProcessor.processingTime",
		"counter.tomcat.GlobalRequestProcessor.requestCount",
		"gauge.tomcat.GlobalRequestProcessor.maxTime",
		"gauge.tomcat.ThreadPool.currentThreadsBusy",
		"gauge.tomcat.ThreadPool.maxThreads",
		"gauge.loaded_classes",
		"jmx_memory.init",
		"jmx_memory.max",
		"jmx_memory.used",
		"gauge.jvm.threads.count",
		"invocations",
		"jmx_memory.committed",
		"total_time_in_ms.collection_time",
	}, "default_metrics_config.yaml")
}
