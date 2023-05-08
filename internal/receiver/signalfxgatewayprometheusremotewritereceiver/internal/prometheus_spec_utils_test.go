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
	"testing"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"
)

func TestGetMetricTypeByLabels(t *testing.T) {
	testCases := []struct {
		labels     []prompb.Label
		metricType prompb.MetricMetadata_MetricType
	}{
		{
			labels: []prompb.Label{
				{Name: "__name__", Value: "http_requests_count"},
			},
			metricType: prompb.MetricMetadata_COUNTER,
		},
		{
			labels: []prompb.Label{
				{Name: "__name__", Value: "http_requests_total"},
				{Name: "method", Value: "GET"},
				{Name: "status", Value: "200"},
			},
			metricType: prompb.MetricMetadata_COUNTER,
		},
		{
			labels: []prompb.Label{
				{Name: "__name__", Value: "go_goroutines"},
			},
			metricType: prompb.MetricMetadata_GAUGE,
		},
		{
			labels: []prompb.Label{
				{Name: "__name__", Value: "api_request_duration_seconds_bucket"},
				{Name: "le", Value: "0.1"},
			},
			metricType: prompb.MetricMetadata_HISTOGRAM,
		},
		{
			labels: []prompb.Label{
				{Name: "__name__", Value: "rpc_duration_seconds_total"},
				{Name: "le", Value: "0.1"},
			},
			metricType: prompb.MetricMetadata_HISTOGRAM,
		},
		{
			labels: []prompb.Label{
				{Name: "__name__", Value: "rpc_duration_seconds_total"},
				{Name: "quantile", Value: "0.5"},
			},
			metricType: prompb.MetricMetadata_SUMMARY,
		},
		{
			labels: []prompb.Label{
				{Name: "__name__", Value: "rpc_duration_total"},
				{Name: "quantile", Value: "0.5"},
			},
			metricType: prompb.MetricMetadata_SUMMARY,
		},
	}

	for _, tc := range testCases {
		metricName, err := ExtractMetricNameLabel(tc.labels)
		require.NoError(t, err)
		metricType := DetermineMetricTypeByConvention(metricName, tc.labels)
		if metricType != tc.metricType {
			t.Errorf("DetermineMetricTypeByConvention(%v) = %v; want %v", tc.labels, metricType, tc.metricType)
		}
	}
}
