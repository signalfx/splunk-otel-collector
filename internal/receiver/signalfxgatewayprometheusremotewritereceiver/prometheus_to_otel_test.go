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

package signalfxgatewayprometheusremotewritereceiver

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"golang.org/x/exp/maps"
)

func TestParseAndPartitionPrometheusRemoteWriteRequest(t *testing.T) {
	reporter := newMockReporter()
	require.NotNil(t, reporter)
	parser := newPrometheusRemoteOtelParser()

	sampleWriteRequests := FlattenWriteRequests(GetWriteRequestsOfAllTypesWithoutMetadata())
	noMdPartitions, err := parser.partitionWriteRequest(sampleWriteRequests)
	require.NoError(t, err)
	require.Empty(t, sampleWriteRequests.Metadata, "NoMetadata (heuristical) portion of test contains metadata")

	noMdMap := make(map[prompb.MetricMetadata_MetricType]map[string][]metricData)
	for key, partition := range noMdPartitions {
		require.Nil(t, noMdMap[key])
		noMdMap[key] = make(map[string][]metricData)

		for _, md := range partition {
			noMdMap[key][md.MetricName] = append(noMdMap[key][md.MetricName], md)

			assert.NotEmpty(t, md.MetricMetadata.Type)
		}
	}

	results := parser.transformPrometheusRemoteWriteToOtel(noMdPartitions)

	typesSeen := make(map[pmetric.MetricType][]string)
	for resourceMetricsIndex := 0; resourceMetricsIndex < results.ResourceMetrics().Len(); resourceMetricsIndex++ {
		rm := results.ResourceMetrics().At(resourceMetricsIndex)
		for scopeMetricsIndex := 0; scopeMetricsIndex < rm.ScopeMetrics().Len(); scopeMetricsIndex++ {
			sm := rm.ScopeMetrics().At(scopeMetricsIndex)
			for metricsIndex := 0; metricsIndex < sm.Metrics().Len(); metricsIndex++ {
				metric := sm.Metrics().At(metricsIndex)
				typesSeen[metric.Type()] = append(typesSeen[metric.Type()], metric.Name())
			}
		}
	}
	expectedTypesSeen := map[pmetric.MetricType][]string{
		pmetric.MetricTypeSum:   {"http_requests_total", "api_request_duration_seconds_bucket", "api_request_duration_seconds_bucket", "api_request_duration_seconds_count", "request_duration_seconds_count"},
		pmetric.MetricTypeGauge: {"i_am_a_gauge", "request_duration_seconds", "request_duration_seconds", "request_duration_seconds_sum", "api_request_duration_seconds_sum"},
	}
	require.ElementsMatch(t, maps.Keys(expectedTypesSeen), maps.Keys(typesSeen))
	for key, values := range typesSeen {
		require.ElementsMatch(t, expectedTypesSeen[key], values)
	}

}

func TestAddMetricsHappyPath(t *testing.T) {

	testCases := []struct {
		Sample   *prompb.WriteRequest
		Expected pmetric.Metrics
		Name     string
	}{
		{
			Name:     "test counters",
			Sample:   SampleCounterWq(),
			Expected: AddSfxCompatibilityMetrics(ExpectedCounter(), 0, 0, 0),
		},
		{
			Name:     "test gauges",
			Sample:   SampleGaugeWq(),
			Expected: AddSfxCompatibilityMetrics(ExpectedGauge(), 0, 0, 0),
		},
		{
			Name:     "test histograms",
			Sample:   SampleHistogramWq(),
			Expected: AddSfxCompatibilityMetrics(ExpectedSfxCompatibleHistogram(), 0, 0, 0),
		},
		{
			Name:     "test quantiles",
			Sample:   SampleSummaryWq(),
			Expected: AddSfxCompatibilityMetrics(ExpectedSfxCompatibleQuantile(), 0, 0, 0),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			reporter := newMockReporter()
			require.NotNil(t, reporter)
			parser := newPrometheusRemoteOtelParser()
			actual, err := parser.fromPrometheusWriteRequestMetrics(tc.Sample)
			assert.NoError(t, err)

			require.NoError(t, pmetrictest.CompareMetrics(tc.Expected, actual,
				pmetrictest.IgnoreMetricDataPointsOrder(),
				pmetrictest.IgnoreMetricsOrder()))
		})

	}
}
