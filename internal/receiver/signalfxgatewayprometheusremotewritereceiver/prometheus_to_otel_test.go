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
	"math"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"golang.org/x/exp/maps"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/signalfxgatewayprometheusremotewritereceiver/internal/metadata"
)

func TestParseAndPartitionPrometheusRemoteWriteRequest(t *testing.T) {
	reporter := newMockReporter()
	require.NotNil(t, reporter)
	parser := newPrometheusRemoteOtelParser()

	sampleWriteRequests := flattenWriteRequests(getWriteRequestsOfAllTypesWithoutMetadata())
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

func TestAddMetrics(t *testing.T) {
	testCases := []struct {
		sample    *prompb.WriteRequest
		expected  pmetric.Metrics
		name      string
		errWanted bool
	}{
		{
			name:     "test counters",
			sample:   sampleCounterWq(),
			expected: addSfxCompatibilityMetrics(expectedCounter(), 0, 0, 0),
		},
		{
			name:     "test gauges",
			sample:   sampleGaugeWq(),
			expected: addSfxCompatibilityMetrics(expectedGauge(), 0, 0, 0),
		},
		{
			name:     "test histograms",
			sample:   sampleHistogramWq(),
			expected: addSfxCompatibilityMetrics(expectedSfxCompatibleHistogram(), 0, 0, 0),
		},
		{
			name:     "test quantiles",
			sample:   sampleSummaryWq(),
			expected: addSfxCompatibilityMetrics(expectedSfxCompatibleQuantile(), 0, 0, 0),
		},
		{
			name: "test missing",
			sample: &prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{
					{
						Labels: []prompb.Label{
							{Name: "quantile", Value: "-0.5"},
						},
						Samples: []prompb.Sample{
							{Value: math.NaN(), Timestamp: jan20.UnixMilli()},
						},
					},
				},
			},
			expected: addSfxCompatibilityMetrics(func() pmetric.Metrics {
				result := pmetric.NewMetrics()
				resourceMetrics := result.ResourceMetrics().AppendEmpty()
				scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
				scopeMetrics.Scope().SetName(metadata.ScopeName)
				scopeMetrics.Scope().SetVersion("0.1")
				return result
			}(), 0, 0, 1),
			errWanted: true,
		},
		{
			name: "test nan",
			sample: &prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{
					{
						Labels: []prompb.Label{
							{Name: "__name__", Value: "foo"},
						},
						Samples: []prompb.Sample{
							{Value: math.NaN(), Timestamp: jan20.UnixMilli()},
						},
					},
				},
			},
			expected: addSfxCompatibilityMetrics(func() pmetric.Metrics {
				result := pmetric.NewMetrics()
				resourceMetrics := result.ResourceMetrics().AppendEmpty()
				scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
				scopeMetrics.Scope().SetName(metadata.ScopeName)
				scopeMetrics.Scope().SetVersion("0.1")
				m := scopeMetrics.Metrics().AppendEmpty()
				m.SetName("foo")
				m.SetEmptyGauge()
				return result
			}(), 0, 1, 0),
		},
		{
			name: "test invalid",
			sample: &prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{
					{
						Labels: []prompb.Label{
							{Name: "__name__", Value: "foo"},
						},
					},
				},
			},
			expected: addSfxCompatibilityMetrics(func() pmetric.Metrics {
				result := pmetric.NewMetrics()
				resourceMetrics := result.ResourceMetrics().AppendEmpty()
				scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
				scopeMetrics.Scope().SetName(metadata.ScopeName)
				scopeMetrics.Scope().SetVersion("0.1")
				m := scopeMetrics.Metrics().AppendEmpty()
				m.SetName("foo")
				m.SetEmptyGauge()
				return result
			}(), 1, 0, 0),
			errWanted: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter()
			require.NotNil(t, reporter)
			parser := newPrometheusRemoteOtelParser()
			actual, err := parser.fromPrometheusWriteRequestMetrics(tc.sample)
			if tc.errWanted {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			require.NoError(t, pmetrictest.CompareMetrics(tc.expected, actual,
				pmetrictest.IgnoreMetricDataPointsOrder(),
				pmetrictest.IgnoreMetricsOrder(),
				pmetrictest.IgnoreTimestamp(),
				pmetrictest.IgnoreStartTimestamp()))
		})
	}
}
