package prometheusexporter

import (
	"strconv"

	dto "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type extractor func(m *dto.Metric) float64

func gaugeExtractor(m *dto.Metric) float64 {
	return m.GetGauge().GetValue()
}

func untypedExtractor(m *dto.Metric) float64 {
	return m.GetUntyped().GetValue()
}

func counterExtractor(m *dto.Metric) float64 {
	return m.GetCounter().GetValue()
}

func convertMetricFamily(sm pmetric.ScopeMetrics, mf *dto.MetricFamily) {
	if mf.Type == nil || mf.Name == nil {
		return
	}
	switch *mf.Type {
	case dto.MetricType_GAUGE:
		makeGaugeDataPoints(sm, *mf.Name, mf.Metric, gaugeExtractor)
	case dto.MetricType_COUNTER:
		makeSumDataPoints(sm, *mf.Name, mf.Metric, counterExtractor)
	case dto.MetricType_UNTYPED:
		makeGaugeDataPoints(sm, *mf.Name, mf.Metric, untypedExtractor)
	case dto.MetricType_SUMMARY:
		makeSummaryDatapoints(sm, *mf.Name, mf.Metric)
	// TODO: figure out how to best convert histograms, in particular the
	// upper bound value
	case dto.MetricType_HISTOGRAM:
		makeHistogramDatapoints(sm, *mf.Name, mf.Metric)
	default:
	}
}

func makeGaugeDataPoints(sm pmetric.ScopeMetrics, name string, ms []*dto.Metric, e extractor) {
	metric := sm.Metrics().AppendEmpty()
	metric.SetName(name)
	g := metric.SetEmptyGauge()
	for _, m := range ms {
		dp := g.DataPoints().AppendEmpty()
		dp.SetDoubleValue(e(m))
		for i := range m.Label {
			dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
		}
		dp.Attributes().PutStr("system.type", "prometheus-exporter")
	}
}

func makeSumDataPoints(sm pmetric.ScopeMetrics, name string, ms []*dto.Metric, e extractor) {
	metric := sm.Metrics().AppendEmpty()
	metric.SetName(name)
	sum := metric.SetEmptySum()
	sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	sum.SetIsMonotonic(true)
	for _, m := range ms {
		dp := sum.DataPoints().AppendEmpty()
		dp.SetDoubleValue(e(m))
		for i := range m.Label {
			dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
		}
		dp.Attributes().PutStr("system.type", "prometheus-exporter")
	}
}

func makeSummaryDatapoints(sm pmetric.ScopeMetrics, name string, ms []*dto.Metric) {

	for _, m := range ms {
		s := m.GetSummary()
		if s == nil {
			continue
		}

		if s.SampleCount != nil {
			metric := sm.Metrics().AppendEmpty()
			metric.SetName(name + "_count")
			sum := metric.Sum()
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			dp := sum.DataPoints().AppendEmpty()
			for i := range m.Label {
				dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
			}
			dp.Attributes().PutStr("system.type", "prometheus-exporter")
			dp.SetIntValue(int64(s.GetSampleCount()))
		}

		if s.SampleSum != nil {
			metric := sm.Metrics().AppendEmpty()
			metric.SetName(name)
			sum := metric.Sum()
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			dp := sum.DataPoints().AppendEmpty()
			for i := range m.Label {
				dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
			}
			dp.Attributes().PutStr("system.type", "prometheus-exporter")
			dp.SetIntValue(int64(s.GetSampleSum()))
		}

		quantiles := sm.Metrics().AppendEmpty()
		quantiles.SetName(name + "_quantile")
		quantileGauge := quantiles.Gauge()
		qs := s.GetQuantile()
		for i := range qs {
			dp := quantileGauge.DataPoints().AppendEmpty()
			for i := range m.Label {
				dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
			}
			dp.Attributes().PutStr("quantile", strconv.FormatFloat(qs[i].GetQuantile(), 'f', 6, 64))
			dp.Attributes().PutStr("system.type", "prometheus-exporter")
			dp.SetDoubleValue(qs[i].GetValue())
		}
	}
}

func makeHistogramDatapoints(sm pmetric.ScopeMetrics, name string, ms []*dto.Metric) {
	for _, m := range ms {
		dims := labelsToDims(m.Label)
		h := m.GetHistogram()
		if h == nil {
			continue
		}

		if h.SampleCount != nil {
			count := sm.Metrics().AppendEmpty()
			count.SetName(name + "_count")
			sum := count.Sum()
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			dp := sum.DataPoints().AppendEmpty()
			for k, v := range dims {
				dp.Attributes().PutStr(k, v)
			}
			dp.Attributes().PutStr("system.type", "prometheus-exporter")
			dp.SetIntValue(int64(h.GetSampleCount()))
		}

		if h.SampleSum != nil {
			sampleSum := sm.Metrics().AppendEmpty()
			sampleSum.SetName(name)
			sum := sampleSum.Sum()
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			dp := sum.DataPoints().AppendEmpty()
			for k, v := range dims {
				dp.Attributes().PutStr(k, v)
			}
			dp.Attributes().PutStr("system.type", "prometheus-exporter")
			dp.SetIntValue(int64(h.GetSampleSum()))
		}

		b := sm.Metrics().AppendEmpty()
		b.SetName(name + "_bucket")
		bucketsSum := b.Sum()
		bucketsSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		buckets := h.GetBucket()
		for i := range buckets {
			dp := bucketsSum.DataPoints().AppendEmpty()
			for i := range m.Label {
				dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
			}
			dp.Attributes().PutStr("upper_bound", strconv.FormatFloat(buckets[i].GetUpperBound(), 'f', 6, 64))
			dp.Attributes().PutStr("system.type", "prometheus-exporter")
			dp.SetIntValue(int64(buckets[i].GetCumulativeCount()))
		}
	}
}

func labelsToDims(labels []*dto.LabelPair) map[string]string {
	dims := map[string]string{}
	for i := range labels {
		dims[labels[i].GetName()] = labels[i].GetValue()
	}
	return dims
}
