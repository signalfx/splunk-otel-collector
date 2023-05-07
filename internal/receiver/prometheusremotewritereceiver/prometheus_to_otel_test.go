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

package prometheusremotewritereceiver

import (
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// for now, we only support sfx compatibility
func TestParseNoSfxCompat(t *testing.T) {
	reporter := newMockReporter()
	require.NotEmpty(t, reporter)
	parser := &prometheusRemoteOtelParser{}

	require.False(t, parser.SfxGatewayCompatability)

	shouldFailTestCases := []struct {
		Sample *prompb.WriteRequest
		Name   string
	}{
		{
			Name:   "quantile",
			Sample: SampleSummaryWq(),
		},
		{
			Name:   "histogram",
			Sample: SampleHistogramWq(),
		},
	}

	for _, test := range shouldFailTestCases {
		t.Run(test.Name, func(tt *testing.T) {
			metrics, err := parser.fromPrometheusWriteRequestMetrics(test.Sample)
			assert.ErrorContains(t, err, "support")
			assert.NotNil(t, metrics)
			assert.Empty(t, metrics.DataPointCount())
		})
	}

	shouldBeTransparentTestCases := []struct {
		Sample *prompb.WriteRequest
		Name   string
	}{
		{
			Name:   "counter",
			Sample: SampleCounterWq(),
		},
		{
			Name:   "gauge",
			Sample: SampleGaugeWq(),
		},
	}

	for _, test := range shouldBeTransparentTestCases {
		t.Run(test.Name, func(tt *testing.T) {
			metrics, err := parser.fromPrometheusWriteRequestMetrics(test.Sample)
			assert.NoError(t, err)
			assert.NotNil(t, metrics)
			assert.NotEmpty(t, metrics.DataPointCount())
		})
	}

}

func TestParseAndPartitionPrometheusRemoteWriteRequest(t *testing.T) {
	reporter := newMockReporter()
	require.NotNil(t, reporter)
	parser := &prometheusRemoteOtelParser{SfxGatewayCompatability: true}

	sampleWriteRequests := FlattenWriteRequests(GetWriteRequestsOfAllTypesWithoutMetadata())
	noMdPartitions, err := parser.partitionWriteRequest(sampleWriteRequests)
	require.NoError(t, err)
	require.Empty(t, sampleWriteRequests.Metadata, "NoMetadata (heuristical) portion of test contains metadata")

	noMdMap := make(map[string]map[string][]metricData)
	for key, partition := range noMdPartitions {
		require.Nil(t, noMdMap[key])
		noMdMap[key] = make(map[string][]metricData)

		for _, md := range partition {
			assert.Equal(t, key, md.MetricMetadata.MetricFamilyName)

			noMdMap[key][md.MetricName] = append(noMdMap[key][md.MetricName], md)

			assert.Equal(t, md.MetricMetadata.MetricFamilyName, key)
			assert.NotEmpty(t, md.MetricMetadata.Type)
			assert.NotEmpty(t, md.MetricMetadata.MetricFamilyName)

			// Help and Unit should only exist for things with metadata
			assert.Empty(t, md.MetricMetadata.Unit)
			assert.Empty(t, md.MetricMetadata.Help)
		}
	}

	results, err := parser.transformPrometheusRemoteWriteToOtel(noMdPartitions)
	require.NoError(t, err)

	typesSeen := mapset.NewSet[pmetric.MetricType]()
	for resourceMetricsIndex := 0; resourceMetricsIndex < results.ResourceMetrics().Len(); resourceMetricsIndex++ {
		rm := results.ResourceMetrics().At(resourceMetricsIndex)
		for scopeMetricsIndex := 0; scopeMetricsIndex < rm.ScopeMetrics().Len(); scopeMetricsIndex++ {
			sm := rm.ScopeMetrics().At(scopeMetricsIndex)
			for metricsIndex := 0; metricsIndex < sm.Metrics().Len(); metricsIndex++ {
				metric := sm.Metrics().At(metricsIndex)
				typesSeen.Add(metric.Type())
			}
		}
	}
	expectedTypesSeen := mapset.NewSet(pmetric.MetricTypeSum, pmetric.MetricTypeGauge)
	require.Equal(t, expectedTypesSeen, typesSeen)

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
			parser := &prometheusRemoteOtelParser{SfxGatewayCompatability: true}
			actual, err := parser.fromPrometheusWriteRequestMetrics(tc.Sample)
			assert.NoError(t, err)

			require.NoError(t, pmetrictest.CompareMetrics(tc.Expected, actual,
				pmetrictest.IgnoreMetricDataPointsOrder(),
				pmetrictest.IgnoreMetricsOrder()))
		})

	}
}
