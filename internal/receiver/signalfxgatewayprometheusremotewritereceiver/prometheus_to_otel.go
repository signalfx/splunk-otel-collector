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
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/multierr"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/signalfxgatewayprometheusremotewritereceiver/internal"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/signalfxgatewayprometheusremotewritereceiver/internal/metadata"
)

type metricData struct {
	MetricName     string
	Labels         []prompb.Label
	Samples        []prompb.Sample
	Exemplars      []prompb.Exemplar
	Histograms     []prompb.Histogram
	MetricMetadata prompb.MetricMetadata
}

type prometheusRemoteOtelParser struct {
	totalNans            *atomic.Int64
	totalInvalidRequests *atomic.Int64
	totalBadMetrics      *atomic.Int64
}

func newPrometheusRemoteOtelParser() *prometheusRemoteOtelParser {
	return &prometheusRemoteOtelParser{
		totalNans:            &atomic.Int64{},
		totalInvalidRequests: &atomic.Int64{},
		totalBadMetrics:      &atomic.Int64{},
	}
}

func (prwParser *prometheusRemoteOtelParser) fromPrometheusWriteRequestMetrics(request *prompb.WriteRequest) (pmetric.Metrics, error) {
	var otelMetrics pmetric.Metrics
	metricFamiliesAndData, err := prwParser.partitionWriteRequest(request)
	if nil == err {
		otelMetrics = prwParser.transformPrometheusRemoteWriteToOtel(metricFamiliesAndData)
	}
	if otelMetrics == pmetric.NewMetrics() {
		otelMetrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty()
	}
	startTime, endTime := getWriteRequestTimestampBounds(request)
	scope := otelMetrics.ResourceMetrics().At(0).ScopeMetrics().At(0)
	prwParser.addBadRequests(scope, startTime, endTime)
	prwParser.addNanDataPoints(scope, startTime, endTime)
	prwParser.addMetricsWithMissingName(scope, startTime, endTime)
	return otelMetrics, err
}

func (prwParser *prometheusRemoteOtelParser) transformPrometheusRemoteWriteToOtel(parsedPrwMetrics map[prompb.MetricMetadata_MetricType][]metricData) pmetric.Metrics {
	metric := pmetric.NewMetrics()
	rm := metric.ResourceMetrics().AppendEmpty()
	ilm := rm.ScopeMetrics().AppendEmpty()
	ilm.Scope().SetName(metadata.Type)
	ilm.Scope().SetVersion("0.1")
	for metricType, metrics := range parsedPrwMetrics {
		prwParser.addMetrics(ilm, metricType, metrics)
	}
	return metric
}

func (prwParser *prometheusRemoteOtelParser) partitionWriteRequest(writeReq *prompb.WriteRequest) (map[prompb.MetricMetadata_MetricType][]metricData, error) {
	partitions := make(map[prompb.MetricMetadata_MetricType][]metricData)
	var translationErrors error
	for index, ts := range writeReq.Timeseries {
		metricName, err := internal.ExtractMetricNameLabel(ts.Labels)
		if err != nil {
			translationErrors = multierr.Append(translationErrors, err)
		}

		metricType := internal.DetermineMetricTypeByConvention(metricName, ts.Labels)
		metricMetadata := prompb.MetricMetadata{
			Type: metricType,
		}
		md := metricData{
			Labels:         ts.Labels,
			Samples:        writeReq.Timeseries[index].Samples,
			Exemplars:      writeReq.Timeseries[index].Exemplars,
			Histograms:     writeReq.Timeseries[index].Histograms,
			MetricName:     metricName,
			MetricMetadata: metricMetadata,
		}
		if len(md.Samples) < 1 {
			translationErrors = multierr.Append(translationErrors, fmt.Errorf("no samples found for  %s", metricName))
		}
		partitions[metricType] = append(partitions[metricType], md)
	}

	return partitions, translationErrors
}

// This actually converts from a prometheus prompdb.MetaDataType to the closest equivalent otel type
// See https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/13bcae344506fe2169b59d213361d04094c651f6/receiver/prometheusreceiver/internal/util.go#L106
func (prwParser *prometheusRemoteOtelParser) addMetrics(ilm pmetric.ScopeMetrics, metricType prompb.MetricMetadata_MetricType, metrics []metricData) {

	switch metricType {
	case prompb.MetricMetadata_COUNTER, prompb.MetricMetadata_HISTOGRAM, prompb.MetricMetadata_GAUGEHISTOGRAM:
		prwParser.addCounterMetrics(ilm, metrics)
	default:
		prwParser.addGaugeMetrics(ilm, metrics)
	}
}

func (prwParser *prometheusRemoteOtelParser) scaffoldNewMetric(ilm pmetric.ScopeMetrics, name string) pmetric.Metric {
	nm := ilm.Metrics().AppendEmpty()
	nm.SetName(name)
	return nm
}

// addBadRequests is used to report write requests with invalid data
func (prwParser *prometheusRemoteOtelParser) addBadRequests(ilm pmetric.ScopeMetrics, start time.Time, end time.Time) {
	errMetric := ilm.Metrics().AppendEmpty()
	errMetric.SetName("prometheus.invalid_requests")
	errorSum := errMetric.SetEmptySum()
	errorSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	errorSum.SetIsMonotonic(true)
	dp := errorSum.DataPoints().AppendEmpty()
	dp.SetIntValue(prwParser.totalInvalidRequests.Load())
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(start))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(end))
}

// addMetricsWithMissingName is used to report metrics in the remote write request without names
func (prwParser *prometheusRemoteOtelParser) addMetricsWithMissingName(ilm pmetric.ScopeMetrics, start time.Time, end time.Time) {
	errMetric := ilm.Metrics().AppendEmpty()
	errMetric.SetName("prometheus.total_bad_datapoints")
	errorSum := errMetric.SetEmptySum()
	errorSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	errorSum.SetIsMonotonic(true)
	dp := errorSum.DataPoints().AppendEmpty()
	dp.SetIntValue(prwParser.totalBadMetrics.Load())

	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(start))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(end))
}

// addNanDataPoints is an sfx compatibility error metric
func (prwParser *prometheusRemoteOtelParser) addNanDataPoints(ilm pmetric.ScopeMetrics, start time.Time, end time.Time) {
	errMetric := ilm.Metrics().AppendEmpty()
	errMetric.SetName("prometheus.total_NAN_samples")
	errorSum := errMetric.SetEmptySum()
	errorSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	errorSum.SetIsMonotonic(true)
	dp := errorSum.DataPoints().AppendEmpty()
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(start))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(end))
	dp.SetIntValue(prwParser.totalNans.Load())
}

// addGaugeMetrics handles any scalar metric family which can go up or down
func (prwParser *prometheusRemoteOtelParser) addGaugeMetrics(ilm pmetric.ScopeMetrics, metrics []metricData) {
	for _, metricsData := range metrics {
		if metricsData.MetricName == "" {
			prwParser.totalBadMetrics.Add(1)
			continue
		}
		nm := prwParser.scaffoldNewMetric(ilm, metricsData.MetricName)
		nm.SetName(metricsData.MetricName)
		gauge := nm.SetEmptyGauge()
		for _, sample := range metricsData.Samples {
			if math.IsNaN(sample.Value) {
				prwParser.totalNans.Add(1)
				continue
			}
			dp := gauge.DataPoints().AppendEmpty()
			dp.SetTimestamp(prometheusToOtelTimestamp(sample.GetTimestamp()))
			dp.SetStartTimestamp(prometheusToOtelTimestamp(sample.GetTimestamp()))
			prwParser.setFloatOrInt(dp, sample)
			prwParser.setAttributes(dp, metricsData.Labels)
		}
	}
}

// addCounterMetrics handles any scalar metric family which can only goes up, and are cumulative
func (prwParser *prometheusRemoteOtelParser) addCounterMetrics(ilm pmetric.ScopeMetrics, metrics []metricData) {
	for _, metricsData := range metrics {
		if metricsData.MetricName == "" {
			prwParser.totalBadMetrics.Add(1)
			continue
		}
		nm := prwParser.scaffoldNewMetric(ilm, metricsData.MetricName)
		sumMetric := nm.SetEmptySum()
		sumMetric.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		sumMetric.SetIsMonotonic(true)
		for _, sample := range metricsData.Samples {
			if math.IsNaN(sample.Value) {
				prwParser.totalNans.Add(1)
				continue
			}
			dp := nm.Sum().DataPoints().AppendEmpty()
			dp.SetTimestamp(prometheusToOtelTimestamp(sample.GetTimestamp()))
			dp.SetStartTimestamp(prometheusToOtelTimestamp(sample.GetTimestamp()))
			prwParser.setFloatOrInt(dp, sample)
			prwParser.setAttributes(dp, metricsData.Labels)
		}
	}
}

func getSampleTimestampBounds(samples []prompb.Sample) (int64, int64) {
	if len(samples) < 1 {
		return -1, -1
	}
	minTimestamp := int64(math.MaxInt64)
	maxTimestamp := int64(math.MinInt64)
	for _, sample := range samples {
		if minTimestamp > sample.Timestamp {
			minTimestamp = sample.Timestamp
		}
		if maxTimestamp < sample.GetTimestamp() {
			maxTimestamp = sample.GetTimestamp()
		}
	}
	return minTimestamp, maxTimestamp
}

func getWriteRequestTimestampBounds(request *prompb.WriteRequest) (time.Time, time.Time) {
	minTimestamp := int64(math.MaxInt64)
	maxTimestamp := int64(math.MinInt64)
	for _, ts := range request.Timeseries {
		sampleMin, sampleMax := getSampleTimestampBounds(ts.Samples)
		if sampleMin < minTimestamp {
			minTimestamp = sampleMin
		}
		if sampleMax > maxTimestamp {
			maxTimestamp = sampleMax
		}
	}
	return time.UnixMilli(minTimestamp), time.UnixMilli(maxTimestamp)
}

func (prwParser *prometheusRemoteOtelParser) setFloatOrInt(dp pmetric.NumberDataPoint, sample prompb.Sample) error {
	if math.IsNaN(sample.Value) {
		return fmt.Errorf("NAN value found")
	}
	if float64(int64(sample.Value)) == sample.Value {
		dp.SetIntValue(int64(sample.Value))
	} else {
		dp.SetDoubleValue(sample.Value)
	}
	return nil
}

func prometheusToOtelTimestamp(ts int64) pcommon.Timestamp {
	return pcommon.Timestamp(ts * int64(time.Millisecond))
}

func (prwParser *prometheusRemoteOtelParser) setAttributes(dp pmetric.NumberDataPoint, labels []prompb.Label) {
	for _, attr := range labels {
		if attr.Name != "__name__" {
			dp.Attributes().PutStr(attr.Name, attr.Value)
		}
	}
}
