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
	"errors"
	"strings"

	"github.com/prometheus/prometheus/prompb"
)

// ExtractMetricNameLabel Finds label corresponding to metric name
func ExtractMetricNameLabel(labels []prompb.Label) (string, error) {
	metricName, ok := getLabelValue(labels, "__name__")
	if !ok {
		return "", errors.New("did not find a label with `__name__` as per prometheus spec")
	}
	return metricName, nil
}

// DetermineMetricTypeByConvention This is a 'best effort' heuristic applying guidance from the latest PRW receiver and OpenMetrics specifications
// See: https://prometheus.io/docs/concepts/remote_write_spec/#prometheus-remote-write-specification
// Also see: https://raw.githubusercontent.com/OpenObservability/OpenMetrics/main/specification/OpenMetrics.md
// As this is a heuristic process, the order of operations is significant.
func DetermineMetricTypeByConvention(metricName string, labels []prompb.Label) prompb.MetricMetadata_MetricType {

	_, hasLe := getLabelValue(labels, "le")
	_, hasQuantile := getLabelValue(labels, "quantile")
	_, hasMetricName := getLabelValue(labels, metricName)

	switch {
	case hasLe && (strings.HasSuffix(metricName, "_gsum") || strings.HasSuffix(metricName, "_gcount")):
		return prompb.MetricMetadata_GAUGEHISTOGRAM
	case hasLe:
		return prompb.MetricMetadata_HISTOGRAM
	case hasQuantile:
		return prompb.MetricMetadata_SUMMARY
	case hasMetricName:
		return prompb.MetricMetadata_STATESET
	case strings.HasSuffix(metricName, "_total") || strings.HasSuffix(metricName, "_count") || strings.HasSuffix(metricName, "_counter") || strings.HasSuffix(metricName, "_created"):
		return prompb.MetricMetadata_COUNTER
	case strings.HasSuffix(metricName, "_bucket"):
		// While bucket may exist for a gauge histogram or Summary, we've checked such above
		return prompb.MetricMetadata_HISTOGRAM
	case strings.HasSuffix(metricName, "_info"):
		return prompb.MetricMetadata_INFO
	}
	return prompb.MetricMetadata_GAUGE
}

// getLabelValue will return the first label matching the provided name (if present), and an "ok" flag denoting whether
// the name was actually found within the provided labels
func getLabelValue(labels []prompb.Label, name string) (string, bool) {
	for _, label := range labels {
		if label.Name == name {
			return label.Value, true
		}
	}
	return "", false
}
