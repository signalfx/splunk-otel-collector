// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/prompb"
)

// GetBaseMetricFamilyName uses heuristics to determine the metric family of a given metric, by
// removing known suffixes for Sum/Counter, Histogram and Summary metric types.
// While not strictly enforced in the protobuf, prometheus does not support "colliding"
// "metric family names" in the same write request, so this should be safe
// https://prometheus.io/docs/practices/naming/
// https://prometheus.io/docs/concepts/metric_types/
func GetBaseMetricFamilyName(metricName string) string {
	suffixes := []string{"_count", "_sum", "_bucket", "_created", "_total"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(metricName, suffix) {
			return strings.TrimSuffix(metricName, suffix)
		}
	}

	return metricName
}

// ExtractMetricNameLabel Finds label corresponding to timeseries
func ExtractMetricNameLabel(labels []prompb.Label) (string, error) {

	for _, label := range labels {
		if label.Name == "__name__" {
			return label.Value, nil
		}
	}
	return "", fmt.Errorf("did not find a label with `__name__` as per prometheus spec")
}

// GuessMetricTypeByLabels This is a 'best effort' heuristic applying guidance from the latest PRW Receiver and OpenMetrics specifications
// See: https://prometheus.io/docs/concepts/remote_write_spec/#prometheus-remote-write-specification
// Also see: https://raw.githubusercontent.com/OpenObservability/OpenMetrics/main/specification/OpenMetrics.md
// As this is a heuristic process, the order of operations is SIGNIFICANT.
func GuessMetricTypeByLabels(metricName string, labels []prompb.Label) prompb.MetricMetadata_MetricType {

	var histogramType = prompb.MetricMetadata_HISTOGRAM

	if strings.HasSuffix(metricName, "_gsum") || strings.HasSuffix(metricName, "_gcount") {
		// Should also have an LE
		histogramType = prompb.MetricMetadata_GAUGEHISTOGRAM
	}
	for _, label := range labels {
		if label.Name == "le" {
			return histogramType
		}
		if label.Name == "quantile" {
			return prompb.MetricMetadata_SUMMARY
		}
		if label.Name == metricName {
			// The OpenMetrics spec ABNF examples directly conflict with their own given summary, TODO inform them
			return prompb.MetricMetadata_STATESET
		}
	}
	if strings.HasSuffix(metricName, "_total") || strings.HasSuffix(metricName, "_count") || strings.HasSuffix(metricName, "_counter") || strings.HasSuffix(metricName, "_created") {
		return prompb.MetricMetadata_COUNTER
	}
	if strings.HasSuffix(metricName, "_bucket") {
		// While bucket may exist for a gauge histogram or Summary, we've checked such above
		return prompb.MetricMetadata_HISTOGRAM
	}
	if strings.HasSuffix(metricName, "_info") {
		return prompb.MetricMetadata_INFO
	}
	return prompb.MetricMetadata_GAUGE
}
