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
	"time"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/signalfxgatewayprometheusremotewritereceiver/internal/metadata"
)

var jan20 = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

func sampleCounterTs() []prompb.TimeSeries {
	return []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "http_requests_total"},
				{Name: "method", Value: "GET"},
				{Name: "status", Value: "200"},
			},
			Samples: []prompb.Sample{
				{Value: 1024, Timestamp: jan20.UnixMilli()},
			},
		},
	}
}

func sampleCounterWq() *prompb.WriteRequest {
	return &prompb.WriteRequest{Timeseries: sampleCounterTs()}
}

func sampleGaugeTs() []prompb.TimeSeries {
	return []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "i_am_a_gauge"},
			},
			Samples: []prompb.Sample{
				{Value: 42, Timestamp: jan20.UnixMilli()},
			},
		},
	}
}

func sampleGaugeWq() *prompb.WriteRequest { return &prompb.WriteRequest{Timeseries: sampleGaugeTs()} }

func sampleHistogramTs() []prompb.TimeSeries {
	return []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "api_request_duration_seconds_bucket"},
				{Name: "le", Value: "0.1"},
			},
			Samples: []prompb.Sample{
				{Value: 500, Timestamp: jan20.UnixMilli()},
			},
		},
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "api_request_duration_seconds_bucket"},
				{Name: "le", Value: "0.2"},
			},
			Samples: []prompb.Sample{
				{Value: 1500, Timestamp: jan20.UnixMilli()},
			},
		},
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "api_request_duration_seconds_count"},
			},
			Samples: []prompb.Sample{
				{Value: 2500, Timestamp: jan20.UnixMilli()},
			},
		},
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "api_request_duration_seconds_sum"},
			},
			Samples: []prompb.Sample{
				{Value: 350, Timestamp: jan20.UnixMilli()},
			},
		},
	}
}

func sampleHistogramWq() *prompb.WriteRequest {
	return &prompb.WriteRequest{
		Timeseries: sampleHistogramTs(),
	}
}

func sampleSummaryTs() []prompb.TimeSeries {
	return []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "request_duration_seconds"},
				{Name: "quantile", Value: "0.5"},
			},
			Samples: []prompb.Sample{
				{Value: 0.25, Timestamp: jan20.UnixMilli()},
			},
		},
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "request_duration_seconds"},
				{Name: "quantile", Value: "0.9"},
			},
			Samples: []prompb.Sample{
				{Value: 0.35, Timestamp: jan20.UnixMilli()},
			},
		},
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "request_duration_seconds_sum"},
			},
			Samples: []prompb.Sample{
				{Value: 123.5, Timestamp: jan20.UnixMilli()},
			},
		},
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "request_duration_seconds_count"},
			},
			Samples: []prompb.Sample{
				{Value: 1500, Timestamp: jan20.UnixMilli()},
			},
		},
	}
}

func sampleSummaryWq() *prompb.WriteRequest {
	return &prompb.WriteRequest{
		Timeseries: sampleSummaryTs(),
	}
}

func expectedCounter() pmetric.Metrics {
	result := pmetric.NewMetrics()
	resourceMetrics := result.ResourceMetrics().AppendEmpty()
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	scopeMetrics.Scope().SetName(metadata.ScopeName)
	scopeMetrics.Scope().SetVersion("0.1")
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName("http_requests_total")
	counter := metric.SetEmptySum()
	counter.SetIsMonotonic(true)
	counter.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dp := counter.DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(jan20))
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(jan20))
	dp.SetIntValue(1024)
	dp.Attributes().PutStr("method", "GET")
	dp.Attributes().PutStr("status", "200")

	return result
}

func expectedGauge() pmetric.Metrics {
	result := pmetric.NewMetrics()
	resourceMetrics := result.ResourceMetrics().AppendEmpty()
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	scopeMetrics.Scope().SetName(metadata.ScopeName)
	scopeMetrics.Scope().SetVersion("0.1")
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName("i_am_a_gauge")
	counter := metric.SetEmptyGauge()
	dp := counter.DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(jan20))
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(jan20))
	dp.SetIntValue(42)

	return result
}

func expectedSfxCompatibleHistogram() pmetric.Metrics {
	result := pmetric.NewMetrics()
	resourceMetrics := result.ResourceMetrics().AppendEmpty()
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	scopeMetrics.Scope().SetName(metadata.ScopeName)
	scopeMetrics.Scope().SetVersion("0.1")

	// set bucket sizes
	pairs := []struct {
		bucket    string
		value     int64
		timestamp int64
	}{
		{
			bucket:    "0.1",
			value:     500,
			timestamp: jan20.UnixNano(),
		},
		{
			bucket:    "0.2",
			value:     1500,
			timestamp: jan20.UnixNano(),
		},
	}
	for _, values := range pairs {
		metric := scopeMetrics.Metrics().AppendEmpty()
		metric.SetName("api_request_duration_seconds_bucket")
		counter := metric.SetEmptySum()
		counter.SetIsMonotonic(true)
		counter.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		dp := counter.DataPoints().AppendEmpty()
		dp.Attributes().PutStr("le", values.bucket)
		dp.SetTimestamp(pcommon.Timestamp(values.timestamp))      //nolint:gosec
		dp.SetStartTimestamp(pcommon.Timestamp(values.timestamp)) //nolint:gosec
		dp.SetIntValue(values.value)
	}

	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName("api_request_duration_seconds_count")
	counter := metric.SetEmptySum()
	counter.SetIsMonotonic(true)
	counter.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dp := counter.DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.Timestamp(jan20.UnixNano()))      //nolint:gosec
	dp.SetStartTimestamp(pcommon.Timestamp(jan20.UnixNano())) //nolint:gosec
	dp.SetIntValue(2500)

	metric = scopeMetrics.Metrics().AppendEmpty()
	metric.SetName("api_request_duration_seconds_sum")
	gauge := metric.SetEmptyGauge()
	dp = gauge.DataPoints().AppendEmpty()

	dp.SetTimestamp(pcommon.Timestamp(jan20.UnixNano()))      //nolint:gosec
	dp.SetStartTimestamp(pcommon.Timestamp(jan20.UnixNano())) //nolint:gosec
	dp.SetIntValue(350)

	return result
}

func expectedSfxCompatibleQuantile() pmetric.Metrics {
	result := pmetric.NewMetrics()
	resourceMetrics := result.ResourceMetrics().AppendEmpty()
	scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
	scopeMetrics.Scope().SetName(metadata.ScopeName)
	scopeMetrics.Scope().SetVersion("0.1")

	// set bucket sizes
	pairs := []struct {
		bucket    string
		value     float64
		timestamp int64
	}{
		{
			bucket:    "0.5",
			value:     .25,
			timestamp: jan20.UnixNano(),
		},
		{
			bucket:    "0.9",
			value:     .35,
			timestamp: jan20.UnixNano(),
		},
	}
	for _, values := range pairs {
		metric := scopeMetrics.Metrics().AppendEmpty()
		metric.SetName("request_duration_seconds")
		gauge := metric.SetEmptyGauge()
		dp := gauge.DataPoints().AppendEmpty()
		dp.Attributes().PutStr("quantile", values.bucket)
		dp.SetTimestamp(pcommon.Timestamp(values.timestamp))      //nolint:gosec
		dp.SetStartTimestamp(pcommon.Timestamp(values.timestamp)) //nolint:gosec
		dp.SetDoubleValue(values.value)
	}

	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName("request_duration_seconds_count")
	sum := metric.SetEmptySum()
	sum.SetIsMonotonic(true)
	sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dp := sum.DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.Timestamp(jan20.UnixNano()))      //nolint:gosec
	dp.SetStartTimestamp(pcommon.Timestamp(jan20.UnixNano())) //nolint:gosec
	dp.SetIntValue(1500)

	metric = scopeMetrics.Metrics().AppendEmpty()
	metric.SetName("request_duration_seconds_sum")
	gauge := metric.SetEmptyGauge()
	dp = gauge.DataPoints().AppendEmpty()

	dp.SetTimestamp(pcommon.Timestamp(jan20.UnixNano()))      //nolint:gosec
	dp.SetStartTimestamp(pcommon.Timestamp(jan20.UnixNano())) //nolint:gosec
	dp.SetDoubleValue(123.5)

	return result
}

func getWriteRequestsOfAllTypesWithoutMetadata() []*prompb.WriteRequest {
	sampleWriteRequestsNoMetadata := []*prompb.WriteRequest{
		// Counter
		sampleCounterWq(),
		// Gauge
		sampleGaugeWq(),
		// Histogram
		sampleHistogramWq(),
		// Summary
		sampleSummaryWq(),
	}
	return sampleWriteRequestsNoMetadata
}

func addSfxCompatibilityMetrics(metrics pmetric.Metrics, expectedInvalid, expectedNans, expectedMissing int64) pmetric.Metrics {
	if metrics.ResourceMetrics().Len() == 0 {
		metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty()
	}
	scope := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0)
	addSfxCompatibilityInvalidRequestMetrics(scope, expectedInvalid)
	addSfxCompatibilityMissingNameMetrics(scope, expectedMissing)
	addSfxCompatibilityNanMetrics(scope, expectedNans)
	return metrics
}

// addSfxCompatibilityInvalidRequestMetrics adds the meta-metrics to a given scope, but won't set values
// See https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#L188
func addSfxCompatibilityInvalidRequestMetrics(scopeMetrics pmetric.ScopeMetrics, value int64) pmetric.Metric {
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName("prometheus.invalid_requests")
	counter := metric.SetEmptySum()
	counter.SetIsMonotonic(true)
	counter.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dp := counter.DataPoints().AppendEmpty()
	dp.SetIntValue(value)
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(jan20))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(jan20))
	return metric
}

// addSfxCompatibilityMissingNameMetrics adds the meta-metrics to a given scope, but won't set values
// See https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#L188
func addSfxCompatibilityMissingNameMetrics(scopeMetrics pmetric.ScopeMetrics, value int64) pmetric.Metric {
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName("prometheus.total_bad_datapoints")
	counter := metric.SetEmptySum()
	counter.SetIsMonotonic(true)
	counter.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dp := counter.DataPoints().AppendEmpty()
	dp.SetIntValue(value)
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(jan20))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(jan20))
	return metric
}

// addSfxCompatibilityNanMetrics adds the meta-metrics to a given scope, but won't set values
// See https://github.com/signalfx/gateway/blob/main/protocol/prometheus/prometheuslistener.go#L188
func addSfxCompatibilityNanMetrics(scopeMetrics pmetric.ScopeMetrics, value int64) pmetric.Metric {
	metric := scopeMetrics.Metrics().AppendEmpty()
	metric.SetName("prometheus.total_NAN_samples")
	counter := metric.SetEmptySum()
	counter.SetIsMonotonic(true)
	counter.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dp := counter.DataPoints().AppendEmpty()
	dp.SetIntValue(value)
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(jan20))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(jan20))
	return metric
}

func flattenWriteRequests(request []*prompb.WriteRequest) *prompb.WriteRequest {
	var ts []prompb.TimeSeries
	for _, req := range request {
		ts = append(ts, req.Timeseries...)
	}
	return &prompb.WriteRequest{
		Timeseries: ts,
	}
}

func TestBasicNoMd(t *testing.T) {
	wqs := getWriteRequestsOfAllTypesWithoutMetadata()
	require.NotNil(t, wqs)
	for _, wq := range wqs {
		for _, ts := range wq.Timeseries {
			require.NotNil(t, ts)
			assert.NotEmpty(t, ts.Labels)
		}
		require.Empty(t, wq.Metadata)
	}
}
