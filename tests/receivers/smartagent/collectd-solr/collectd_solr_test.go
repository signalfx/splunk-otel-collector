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

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCollectdSolrReceiverProvidesAllMetrics(t *testing.T) {
	testutils.RunMetricsCollectionTest(t, "all_metrics_config.yaml", "all_expected.yaml",
		testutils.WithCompareMetricsOptions(
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreMetricValues(
				"counter.solr.http_2xx_responses",
				"counter.solr.http_requests",
				"counter.solr.node_collections_requests",
				"counter.solr.node_metrics_requests",
				"counter.solr.update_handler_requests",
				"gauge.solr.core_index_size",
				"gauge.solr.core_max_docs",
				"gauge.solr.core_num_docs",
				"gauge.solr.core_totalspace",
				"gauge.solr.core_usablespace",
				"gauge.solr.jetty_request_latency",
				"gauge.solr.jvm_heap_usage",
				"gauge.solr.jvm_memory_pools_Metaspace_usage",
				"gauge.solr.jvm_total_memory_used",
				"gauge.solr.search_query_response",
				"gauge.solr.searcher_warmup",
				"gauge.solr.update_request_handler_response",
			),
			pmetrictest.IgnoreResourceAttributeValue("node"),
		),
	)
}
