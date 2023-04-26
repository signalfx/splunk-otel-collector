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
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/multierr"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/prometheusremotewritereceiver/internal"
)

func (prwParser *PrometheusRemoteOtelParser) FromPrometheusWriteRequestMetrics(request *prompb.WriteRequest) (pmetric.Metrics, error) {
	var otelMetrics pmetric.Metrics
	metricFamiliesAndData, err := prwParser.partitionWriteRequest(request)
	if nil == err {
		otelMetrics, err = prwParser.TransformPrometheusRemoteWriteToOtel(metricFamiliesAndData)
	}
	return otelMetrics, err
}

type MetricData struct {
	MetricName     string
	Labels         []prompb.Label
	Samples        []prompb.Sample
	Exemplars      []prompb.Exemplar
	Histograms     []prompb.Histogram
	MetricMetadata prompb.MetricMetadata
}

type PrometheusRemoteOtelParser struct {
	SfxGatewayCompatability bool
}

func (prwParser *PrometheusRemoteOtelParser) partitionWriteRequest(writeReq *prompb.WriteRequest) (map[string][]MetricData, error) {
	partitions := make(map[string][]MetricData)
	var translationErrors error
	for index, ts := range writeReq.Timeseries {
		metricName, err := internal.ExtractMetricNameLabel(ts.Labels)
		if err != nil {
			translationErrors = multierr.Append(translationErrors, err)
		}
		metricFamilyName := internal.GetBaseMetricFamilyName(metricName)
		if metricFamilyName == "" {
			translationErrors = multierr.Append(translationErrors, fmt.Errorf("metric family name missing: %s", metricName))
		}

		metricType := internal.GuessMetricTypeByLabels(metricName, ts.Labels)
		metricMetadata := prompb.MetricMetadata{
			MetricFamilyName: metricFamilyName,
			Type:             metricType,
		}
		metricData := MetricData{
			Labels:         ts.Labels,
			Samples:        writeReq.Timeseries[index].Samples,
			Exemplars:      writeReq.Timeseries[index].Exemplars,
			Histograms:     writeReq.Timeseries[index].Histograms,
			MetricName:     metricName,
			MetricMetadata: metricMetadata,
		}
		if len(metricData.Samples) < 1 {
			translationErrors = multierr.Append(translationErrors, fmt.Errorf("no samples found for  %s", metricName))
		}
		partitions[metricData.MetricMetadata.MetricFamilyName] = append(partitions[metricData.MetricMetadata.MetricFamilyName], metricData)
	}

	return partitions, translationErrors
}

func (prwParser *PrometheusRemoteOtelParser) TransformPrometheusRemoteWriteToOtel(parsedPrwMetrics map[string][]MetricData) (pmetric.Metrics, error) {
	metric := pmetric.NewMetrics()
	var translationErrors error
	for metricFamily, metrics := range parsedPrwMetrics {
		rm := metric.ResourceMetrics().AppendEmpty()
		err := prwParser.addMetrics(rm, metricFamily, metrics)
		if err != nil {
			translationErrors = multierr.Append(translationErrors, err)
		}
	}
	return metric, translationErrors
}

// This actually converts from a prometheus prompdb.MetaDataType to the closest equivalent otel type
// See https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/13bcae344506fe2169b59d213361d04094c651f6/receiver/prometheusreceiver/internal/util.go#L106
func (prwParser *PrometheusRemoteOtelParser) addMetrics(rm pmetric.ResourceMetrics, family string, metrics []MetricData) error {

	// TODO hughesjj cast to int if essentially int... maybe?  idk they do it in sfx.gateway
	ilm := rm.ScopeMetrics().AppendEmpty()
	ilm.Scope().SetName(internal.TypeStr)
	ilm.Scope().SetVersion("0.1")

	if family == "" {
		prwParser.addBadDataPoints(ilm, metrics)
		return errors.New("missing name")
	}

	// TODO error handling
	// badrequest if empty samples/metrics right?
	// if family empty then the other dealio for sfx compatibility mode

	metricsMetadata := metrics[0].MetricMetadata

	var err error
	switch metricsMetadata.Type {
	case prompb.MetricMetadata_GAUGE, prompb.MetricMetadata_UNKNOWN:
		err = prwParser.addGaugeMetrics(ilm, family, metrics, metricsMetadata)
	case prompb.MetricMetadata_COUNTER:
		err = prwParser.addCounterMetrics(ilm, family, metrics, metricsMetadata)
	case prompb.MetricMetadata_HISTOGRAM, prompb.MetricMetadata_GAUGEHISTOGRAM:
		if prwParser.SfxGatewayCompatability {
			err = prwParser.addHistogramCounterMetrics(ilm, family, metrics, metricsMetadata)
		} else {
			err = fmt.Errorf("this version of the prometheus remote write receiver only supports SfxGatewayCompatability mode")
		}
	case prompb.MetricMetadata_SUMMARY:
		if prwParser.SfxGatewayCompatability {
			err = prwParser.addQuantileCounterMetrics(ilm, family, metrics, metricsMetadata)
		} else {
			err = fmt.Errorf("this version of the prometheus remote write receiver only supports SfxGatewayCompatability mode")
		}
	case prompb.MetricMetadata_INFO, prompb.MetricMetadata_STATESET:
		err = prwParser.addInfoStateset(ilm, family, metrics, metricsMetadata)
	default:
		err = fmt.Errorf("unsupported type %s for metric family %s", metricsMetadata.Type, family)
	}
	return err
}

func (prwParser *PrometheusRemoteOtelParser) scaffoldNewMetric(ilm pmetric.ScopeMetrics, family string, metricsMetadata prompb.MetricMetadata) pmetric.Metric {
	nm := ilm.Metrics().AppendEmpty()
	nm.SetUnit(metricsMetadata.Unit)
	nm.SetDescription(metricsMetadata.GetHelp())
	nm.SetName(family)
	return nm
}

func (prwParser *PrometheusRemoteOtelParser) addBadDataPoints(ilm pmetric.ScopeMetrics, metrics []MetricData) {
	errMetric := ilm.Metrics().AppendEmpty()
	errMetric.SetName("prometheus.total_bad_datapoints")
	errorSum := errMetric.SetEmptySum()
	errorSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	errorSum.SetIsMonotonic(true)
	for _, metric := range metrics {
		dp := errorSum.DataPoints().AppendEmpty()
		dp.SetIntValue(int64(len(metric.Samples)))
	}
}

func (prwParser *PrometheusRemoteOtelParser) addGaugeMetrics(ilm pmetric.ScopeMetrics, family string, metrics []MetricData, metadata prompb.MetricMetadata) error {
	var translationErrors []error
	if nil == metrics {
		translationErrors = append(translationErrors, fmt.Errorf("Nil metricsdata pointer! %s", family))
	}
	if translationErrors != nil {
		return multierr.Combine(translationErrors...)
	}
	nm := prwParser.scaffoldNewMetric(ilm, family, metadata)
	for _, metricsData := range metrics {
		if metricsData.MetricName != "" {
			nm.SetName(metricsData.MetricName)
		}
		gauge := nm.SetEmptyGauge()
		for _, sample := range metricsData.Samples {
			dp := gauge.DataPoints().AppendEmpty()
			dp.SetDoubleValue(sample.Value)
			dp.SetTimestamp(pcommon.Timestamp(sample.Timestamp * int64(time.Millisecond)))
			for _, attr := range metricsData.Labels {
				dp.Attributes().PutStr(attr.Name, attr.Value)
			}
		}
	}
	return nil
}

func (prwParser *PrometheusRemoteOtelParser) addCounterMetrics(ilm pmetric.ScopeMetrics, family string, metrics []MetricData, metadata prompb.MetricMetadata) error {
	var translationErrors []error
	if nil == metrics {
		translationErrors = append(translationErrors, fmt.Errorf("Nil metricsdata pointer! %s", family))
	}
	if translationErrors != nil {
		return multierr.Combine(translationErrors...)
	}
	nm := prwParser.scaffoldNewMetric(ilm, family, metadata)
	for _, metricsData := range metrics {
		// TODO yeah no
		if metricsData.MetricName != "" {
			nm.SetName(metricsData.MetricName)
		}
		sumMetric := nm.SetEmptySum()
		// TODO hughesjj No idea how correct this is, but scraper always sets this way.  could totally see PRW being different
		sumMetric.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		// TODO hughesjj No idea how correct this is, but scraper always sets this way.  could totally see PRW being different
		sumMetric.SetIsMonotonic(true)
		for _, sample := range metricsData.Samples {
			counter := nm.Sum().DataPoints().AppendEmpty()
			counter.SetDoubleValue(sample.Value)
			counter.SetTimestamp(pcommon.Timestamp(sample.Timestamp * int64(time.Millisecond)))
			for _, attr := range metricsData.Labels {
				counter.Attributes().PutStr(attr.Name, attr.Value)
			}
			// Fairly certain counter is byref here
		}
	}
	return nil
}

func (prwParser *PrometheusRemoteOtelParser) addInfoStateset(ilm pmetric.ScopeMetrics, family string, metrics []MetricData, metadata prompb.MetricMetadata) error {
	var translationErrors []error
	if nil == metrics {
		translationErrors = append(translationErrors, fmt.Errorf("Nil metricsdata pointer! %s", family))
	}
	if translationErrors != nil {
		return multierr.Combine(translationErrors...)
	}
	nm := prwParser.scaffoldNewMetric(ilm, family, metadata)
	for _, metricsData := range metrics {
		// TODO better way to do this
		if metricsData.MetricName != "" {
			nm.SetName(metricsData.MetricName)
		}

		// set as SUM but non-monotonic
		sumMetric := nm.SetEmptySum()
		sumMetric.SetIsMonotonic(false)
		sumMetric.SetAggregationTemporality(pmetric.AggregationTemporalityUnspecified)

		for _, sample := range metricsData.Samples {
			dp := sumMetric.DataPoints().AppendEmpty()
			dp.SetTimestamp(pcommon.Timestamp(sample.GetTimestamp() * int64(time.Millisecond)))
			dp.SetDoubleValue(sample.GetValue()) // TODO hughesjj maybe see if can be intvalue
			for _, attr := range metricsData.Labels {
				dp.Attributes().PutStr(attr.Name, attr.Value)
			}
		}
	}
	return nil
}

// addQuantileCounterMetrics supports the legacy signalfx format of simply reporting all histograms as counters
func (prwParser *PrometheusRemoteOtelParser) addQuantileCounterMetrics(ilm pmetric.ScopeMetrics, family string, metrics []MetricData, metadata prompb.MetricMetadata) error {
	var translationErrors []error
	if nil == metrics {
		translationErrors = append(translationErrors, fmt.Errorf("Nil metricsdata pointer! %s", family))
	}
	if translationErrors != nil {
		return multierr.Combine(translationErrors...)
	}
	for index, metric := range metrics {
		bucket := "bucket" // TODO extract quantile from label
		metrics[index].MetricName = metric.MetricName + bucket
	}
	return prwParser.addCounterMetrics(ilm, family, metrics, metadata)
}

// addHistogramCounterMetrics supports the legacy signalfx format of simply reporting all histograms as counters
func (prwParser *PrometheusRemoteOtelParser) addHistogramCounterMetrics(ilm pmetric.ScopeMetrics, family string, metrics []MetricData, metadata prompb.MetricMetadata) error {
	var translationErrors []error
	if nil == metrics {
		translationErrors = append(translationErrors, fmt.Errorf("Nil metricsdata pointer! %s", family))
	}
	if translationErrors != nil {
		return multierr.Combine(translationErrors...)
	}
	for index, metric := range metrics {
		bucket := "bucket" // TODO extract LE from label
		metrics[index].MetricName = metric.MetricName + bucket
	}
	return prwParser.addCounterMetrics(ilm, family, metrics, metadata)
}
