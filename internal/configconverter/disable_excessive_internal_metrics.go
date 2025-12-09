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
	"errors"
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

// metricRelabelConfigsV1 is the first version of prometheus metric_relabel_configs used in the Splunk distribution.
var metricRelabelConfigsV1 = []any{
	map[string]any{
		"source_labels": []any{model.MetricNameLabel},
		"regex":         ".*grpc_io.*",
		"action":        string(relabel.Drop),
	},
}

// metricRelabelConfigsV2 is the second version of prometheus metric_relabel_configs used in the Splunk distribution.
var metricRelabelConfigsV2 = []any{
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

// metricRelabelConfigsCurrent is the current version of prometheus metric_relabel_configs used in the Splunk distribution.
var metricRelabelConfigsCurrent = []any{
	map[string]any{
		"source_labels": []any{model.MetricNameLabel},
		"regex":         "promhttp_metric_handler_errors.*",
		"action":        string(relabel.Drop),
	},
	map[string]any{
		"source_labels": []any{model.MetricNameLabel},
		"regex":         "otelcol_processor_batch_.*",
		"action":        string(relabel.Drop),
	},
}

// DisableExcessiveInternalMetrics updates config of the prometheus receiver scraping internal
// collector metrics to drop excessive internal metrics matching the following patterns:
// - "promhttp_metric_handler_errors.*"
// - "otelcol_processor_batch_.*"
func DisableExcessiveInternalMetrics(_ context.Context, cfgMap *confmap.Conf) error {
	if cfgMap == nil {
		return errors.New("cannot DisableExcessiveInternalMetrics on nil *confmap.Conf")
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

			// Replace the metric_relabel_configs only if it's set to the old default values.
			if reflect.DeepEqual(mrcs, metricRelabelConfigsV1) || reflect.DeepEqual(mrcs, metricRelabelConfigsV2) {
				sc["metric_relabel_configs"] = metricRelabelConfigsCurrent
			}
		}

		// Update the config with the new scrape_configs.
		if err := cfgMap.Merge(confmap.NewFromStringMap(map[string]any{promScrapeConfigsKey: scrapeConfigs})); err != nil {
			return err
		}
	}

	return nil
}
