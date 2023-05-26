package prometheusexporter

import (
	"strconv"

	dto "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/collector/pdata/pcommon"
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

func convertMetricFamily(sm pmetric.ScopeMetrics, mf *dto.MetricFamily, applyDimensions func(attrs pcommon.Map)) {
	if mf.Type == nil || mf.Name == nil {
		return
	}
	switch *mf.Type {
	case dto.MetricType_GAUGE:
		makeGaugeDataPoints(sm, *mf.Name, mf.Metric, gaugeExtractor, applyDimensions)
	case dto.MetricType_COUNTER:
		makeSumDataPoints(sm, *mf.Name, mf.Metric, counterExtractor, applyDimensions)
	case dto.MetricType_UNTYPED:
		makeGaugeDataPoints(sm, *mf.Name, mf.Metric, untypedExtractor, applyDimensions)
	case dto.MetricType_SUMMARY:
		makeSummaryDatapoints(sm, *mf.Name, mf.Metric, applyDimensions)
	// TODO: figure out how to best convert histograms, in particular the
	// upper bound value
	case dto.MetricType_HISTOGRAM:
		makeHistogramDatapoints(sm, *mf.Name, mf.Metric, applyDimensions)
	default:
	}
}

func makeGaugeDataPoints(sm pmetric.ScopeMetrics, name string, ms []*dto.Metric, e extractor, applyDimensions func(attrs pcommon.Map)) {
	metric := sm.Metrics().AppendEmpty()
	metric.SetName(name)
	g := metric.SetEmptyGauge()
	for _, m := range ms {
		dp := g.DataPoints().AppendEmpty()
		dp.SetDoubleValue(e(m))
		for i := range m.Label {
			dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
		}
		applyDimensions(dp.Attributes())
	}
}

func makeSumDataPoints(sm pmetric.ScopeMetrics, name string, ms []*dto.Metric, e extractor, applyDimensions func(attrs pcommon.Map)) {
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
		applyDimensions(dp.Attributes())
	}
}

func makeSummaryDatapoints(sm pmetric.ScopeMetrics, name string, ms []*dto.Metric, applyDimensions func(attrs pcommon.Map)) {

	for _, m := range ms {
		s := m.GetSummary()
		if s == nil {
			continue
		}

		if s.SampleCount != nil {
			metric := sm.Metrics().AppendEmpty()
			metric.SetName(name + "_count")
			sum := metric.SetEmptySum()
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			sum.SetIsMonotonic(true)
			dp := sum.DataPoints().AppendEmpty()
			for i := range m.Label {
				dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
			}
			applyDimensions(dp.Attributes())
			dp.SetIntValue(int64(s.GetSampleCount()))
		}

		if s.SampleSum != nil {
			metric := sm.Metrics().AppendEmpty()
			metric.SetName(name)
			sum := metric.SetEmptySum()
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			sum.SetIsMonotonic(true)
			dp := sum.DataPoints().AppendEmpty()
			for i := range m.Label {
				dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
			}
			applyDimensions(dp.Attributes())
			dp.SetDoubleValue(s.GetSampleSum())
		}

		quantiles := sm.Metrics().AppendEmpty()
		quantiles.SetName(name + "_quantile")
		quantileGauge := quantiles.SetEmptyGauge()
		qs := s.GetQuantile()
		for i := range qs {
			dp := quantileGauge.DataPoints().AppendEmpty()
			for i := range m.Label {
				dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
			}
			dp.Attributes().PutStr("quantile", strconv.FormatFloat(qs[i].GetQuantile(), 'f', 6, 64))
			applyDimensions(dp.Attributes())
			dp.SetDoubleValue(qs[i].GetValue())
		}
	}
}

func makeHistogramDatapoints(sm pmetric.ScopeMetrics, name string, ms []*dto.Metric, applyDimensions func(attrs pcommon.Map)) {
	for _, m := range ms {
		dims := labelsToDims(m.Label)
		h := m.GetHistogram()
		if h == nil {
			continue
		}

		if h.SampleCount != nil {
			count := sm.Metrics().AppendEmpty()
			count.SetName(name + "_count")
			sum := count.SetEmptySum()
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			sum.SetIsMonotonic(true)
			dp := sum.DataPoints().AppendEmpty()
			for k, v := range dims {
				dp.Attributes().PutStr(k, v)
			}
			applyDimensions(dp.Attributes())
			dp.SetIntValue(int64(h.GetSampleCount()))
		}

		if h.SampleSum != nil {
			sampleSum := sm.Metrics().AppendEmpty()
			sampleSum.SetName(name)
			sum := sampleSum.SetEmptySum()
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
			sum.SetIsMonotonic(true)
			dp := sum.DataPoints().AppendEmpty()
			for k, v := range dims {
				dp.Attributes().PutStr(k, v)
			}
			applyDimensions(dp.Attributes())
			dp.SetDoubleValue(h.GetSampleSum())
		}

		b := sm.Metrics().AppendEmpty()
		b.SetName(name + "_bucket")
		bucketsSum := b.SetEmptySum()
		bucketsSum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		bucketsSum.SetIsMonotonic(true)
		buckets := h.GetBucket()
		for i := range buckets {
			dp := bucketsSum.DataPoints().AppendEmpty()
			for i := range m.Label {
				dp.Attributes().PutStr(m.Label[i].GetName(), m.Label[i].GetValue())
			}
			dp.Attributes().PutStr("upper_bound", strconv.FormatFloat(buckets[i].GetUpperBound(), 'f', 6, 64))
			applyDimensions(dp.Attributes())
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
