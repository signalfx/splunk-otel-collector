// Copyright  Splunk, Inc.
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

package configconverter

import (
	"context"
	"fmt"
	"strings"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	"go.opentelemetry.io/collector/confmap"
)

const (
	// promJobNamePrefix is the name prefix of the prometheus jobs that scrapes internal otel-collector metrics.
	promJobNamePrefix = "otel-"

	// Metric patterns to drop.
	rpcMetricPattern   = "otelcol_rpc_.*"
	httpMetricPattern  = "otelcol_http_.*"
	batchMetricPattern = "otelcol_processor_batch_.*"
)

// promScrapeConfigsKeys are possible keys to get the config of a prometheus receiver scrapings internal collector metrics.
var promScrapeConfigsKeys = []string{
	"receivers::prometheus/internal::config::scrape_configs",
	"receivers::prometheus/agent::config::scrape_configs",
	"receivers::prometheus/k8s_cluster_receiver::config::scrape_configs",
	"receivers::prometheus/collector::config::scrape_configs",
}

// DisableExcessiveInternalMetrics is a MapConverter that updates config of the prometheus receiver scraping internal
// collector metrics to drop excessive internal metrics matching the following patterns:
// - "otelcol_rpc_.*"
// - "otelcol_http_.*"
// - "otelcol_processor_batch_.*"
type DisableExcessiveInternalMetrics struct{}

func (DisableExcessiveInternalMetrics) Convert(_ context.Context, cfgMap *confmap.Conf) error {
	if cfgMap == nil {
		return fmt.Errorf("cannot DisableExcessiveInternalMetrics on nil *confmap.Conf")
	}

	for _, promScrapeConfigsKey := range promScrapeConfigsKeys {
		scrapeConfMap := cfgMap.Get(promScrapeConfigsKey)
		if scrapeConfMap == nil {
			continue
		}

		scrapeConfigs, ok := scrapeConfMap.([]any)
		if !ok {
			continue // Ignore invalid scrape_configs, as they will be caught by the config validation.
		}

		for _, scrapeConfig := range scrapeConfigs {
			sc, ok := scrapeConfig.(map[string]any)
			if !ok {
				continue // Ignore Ignore invalid metric_relabel_configs, as they will be caught by the config validation.
			}
			jobName, ok := sc["job_name"]
			if !ok || !strings.HasPrefix(jobName.(string), promJobNamePrefix) {
				continue
			}

			metricRelabelConfigs := sc["metric_relabel_configs"]
			if metricRelabelConfigs == nil {
				metricRelabelConfigs = make([]any, 0, 3)
			}
			mrcs, ok := metricRelabelConfigs.([]any)
			if !ok {
				continue // Ignore invalid metric_relabel_configs, as they will be caught by the config validation.
			}

			foundRegexPatterns := make(map[string]bool)
			for _, metricRelabelConfig := range mrcs {
				mrc, ok := metricRelabelConfig.(map[string]any)
				if !ok {
					continue // Ignore invalid metric_relabel_config, as they will be caught by the config validation.
				}
				sourceLabels, ok := mrc["source_labels"].([]any)
				if !ok || len(sourceLabels) != 1 || sourceLabels[0] != model.MetricNameLabel {
					continue
				}
				regex, ok := mrc["regex"].(string)
				if !ok {
					continue
				}
				foundRegexPatterns[regex] = true
			}

			for _, pattern := range []string{rpcMetricPattern, httpMetricPattern, batchMetricPattern} {
				if !foundRegexPatterns[pattern] {
					mrcs = append(mrcs, map[string]any{
						"source_labels": []any{model.MetricNameLabel},
						"regex":         pattern,
						"action":        string(relabel.Drop),
					})
				}
			}

			sc["metric_relabel_configs"] = mrcs
		}

		// Update the config with the new scrape_configs.
		if err := cfgMap.Merge(confmap.NewFromStringMap(map[string]any{promScrapeConfigsKey: scrapeConfigs})); err != nil {
			return err
		}
	}

	return nil
}
