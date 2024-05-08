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
	"reflect"
	"strings"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	"go.opentelemetry.io/collector/confmap"
)

// promJobNamePrefix is the name prefix of the prometheus jobs that scrapes internal otel-collector metrics.
const promJobNamePrefix = "otel-"

// promScrapeConfigsKeys are possible keys to get the config of a prometheus receiver scrapings internal collector metrics.
var promScrapeConfigsKeys = []string{
	"receivers::prometheus/internal::config::scrape_configs",
	"receivers::prometheus/agent::config::scrape_configs",
	"receivers::prometheus/k8s_cluster_receiver::config::scrape_configs",
	"receivers::prometheus/collector::config::scrape_configs",
}

// The metric_relabel_configs prometheus config section to replace.
var metricRelabelConfigsToReplace = []any{
	map[string]any{
		"source_labels": []any{model.MetricNameLabel},
		"regex":         ".*grpc_io.*",
		"action":        string(relabel.Drop),
	},
}

var metricRelabelConfigsToSet = []any{
	map[string]any{
		"source_labels": []any{model.MetricNameLabel},
		"regex":         "otelcol_rpc_.*",
		"action":        string(relabel.Drop),
	},
	map[string]any{
		"source_labels": []any{model.MetricNameLabel},
		"regex":         "otelcol_http_.*",
		"action":        string(relabel.Drop),
	},
	map[string]any{
		"source_labels": []any{model.MetricNameLabel},
		"regex":         "otelcol_processor_batch_.*",
		"action":        string(relabel.Drop),
	},
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
				continue // Keep unset metric_relabel_configs as is.
			}
			mrcs, ok := metricRelabelConfigs.([]any)
			if !ok {
				continue // Ignore invalid metric_relabel_configs, as they will be caught by the config validation.
			}

			// Replace the metric_relabel_configs only if it's set to the old default value.
			if len(mrcs) == 1 && reflect.DeepEqual(mrcs[0], metricRelabelConfigsToReplace[0]) {
				sc["metric_relabel_configs"] = metricRelabelConfigsToSet
			}
		}

		// Update the config with the new scrape_configs.
		if err := cfgMap.Merge(confmap.NewFromStringMap(map[string]any{promScrapeConfigsKey: scrapeConfigs})); err != nil {
			return err
		}
	}

	return nil
}
