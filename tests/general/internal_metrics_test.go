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
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDefaultInternalMetrics(t *testing.T) {
	testutils.RunMetricsCollectionTest(t, "internal_metrics_config.yaml", "internal_metrics_expected.yaml",
		testutils.WithCompareMetricsOptions(
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreMetricsOrder(),
			pmetrictest.IgnoreScopeVersion(),
			pmetrictest.IgnoreMetricValues(
				"otelcol_exporter_sent_metric_points",
				"otelcol_process_cpu_seconds",
				"otelcol_process_memory_rss",
				"otelcol_process_runtime_heap_alloc_bytes",
				"otelcol_process_runtime_total_alloc_bytes",
				"otelcol_process_runtime_total_sys_memory_bytes",
				"otelcol_process_uptime",
				"otelcol_receiver_accepted_metric_points",
				"otelcol_receiver_failed_metric_points",
				"scrape_duration_seconds",
				"scrape_samples_post_metric_relabeling",
				"scrape_samples_scraped",
				"scrape_series_added",
				"up",
			),
			pmetrictest.IgnoreResourceAttributeValue("service.instance.id"),
			pmetrictest.IgnoreResourceAttributeValue("service.version"),
			pmetrictest.IgnoreMetricAttributeValue("service.instance.id"),
			pmetrictest.IgnoreMetricAttributeValue("service.version"),
		),
	)
}
