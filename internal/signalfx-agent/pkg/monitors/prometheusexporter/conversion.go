package prometheusexporter

import (
	"strconv"

	dto "github.com/prometheus/client_model/go"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type extractor func(m *dto.Metric) float64
type dpFactory func(string, map[string]string, float64) *datapoint.Datapoint

func gaugeExtractor(m *dto.Metric) float64 {
	return m.GetGauge().GetValue()
}

func untypedExtractor(m *dto.Metric) float64 {
	return m.GetUntyped().GetValue()
}

func counterExtractor(m *dto.Metric) float64 {
	return m.GetCounter().GetValue()
}

func convertMetricFamily(mf *dto.MetricFamily) []pmetric.Metric {
	//nolint:protogetter
	if mf.Type == nil || mf.Name == nil {
		return nil
	}
	switch *mf.Type { //nolint:protogetter
	case dto.MetricType_GAUGE:
		return makeSimpleDatapoints(mf.GetName(), mf.GetMetric(), gaugeExtractor)
	case dto.MetricType_COUNTER:
		return makeSimpleCumulativeSum(mf.GetName(), mf.GetMetric(), counterExtractor)
	case dto.MetricType_UNTYPED:
		return makeSimpleDatapoints(mf.GetName(), mf.GetMetric(), untypedExtractor)
	case dto.MetricType_SUMMARY:
		return makeSummaryDatapoints(mf.GetName(), mf.GetMetric())
	// TODO: figure out how to best convert histograms, in particular the
	// upper bound value
	case dto.MetricType_HISTOGRAM:
		return makeHistogramDatapoints(mf.GetName(), mf.GetMetric())
	default:
		return nil
	}
}

func makeSimpleCumulativeSum(name string, ms []*dto.Metric, e extractor) []pmetric.Metric {
	dps := make([]pmetric.Metric, len(ms))
	for i, m := range ms {
		metric := pmetric.NewMetric()
		metric.SetName(name)
		s := metric.SetEmptySum()
		dp := s.DataPoints().AppendEmpty()
		dp.SetDoubleValue(e(m))
		_ = dp.Attributes().FromRaw(labelsToDims(m.GetLabel()))
		dps[i] = metric
	}
	return dps
}

func makeSimpleDatapoints(name string, ms []*dto.Metric, e extractor) []pmetric.Metric {
	dps := make([]pmetric.Metric, len(ms))
	for i, m := range ms {
		metric := pmetric.NewMetric()
		metric.SetName(name)
		g := metric.SetEmptyGauge()
		dp := g.DataPoints().AppendEmpty()
		dp.SetDoubleValue(e(m))
		_ = dp.Attributes().FromRaw(labelsToDims(m.GetLabel()))
		dps[i] = metric
	}
	return dps
}

func makeSummaryDatapoints(name string, ms []*dto.Metric) []pmetric.Metric {
	var dps []pmetric.Metric
	for _, m := range ms {
		dims := labelsToDims(m.GetLabel())
		s := m.GetSummary()
		if s == nil {
			continue
		}

		//nolint:protogetter
		if s.SampleCount != nil {
			metric := pmetric.NewMetric()
			metric.SetName(name + "_count")
			sum := metric.SetEmptySum()
			dp := sum.DataPoints().AppendEmpty()
			dp.SetIntValue(int64(s.GetSampleCount()))
			_ = dp.Attributes().FromRaw(dims)
			dps = append(dps, metric)
		}

		//nolint:protogetter
		if s.SampleSum != nil {
			metric := pmetric.NewMetric()
			metric.SetName(name)
			sum := metric.SetEmptySum()
			dp := sum.DataPoints().AppendEmpty()
			dp.SetIntValue(int64(s.GetSampleSum()))
			_ = dp.Attributes().FromRaw(dims)
			dps = append(dps, metric)
		}

		qs := s.GetQuantile()
		for i := range qs {
			quantileDims := utils.MergeMaps(dims, map[string]any{
				"quantile": strconv.FormatFloat(qs[i].GetQuantile(), 'f', 6, 64),
			})
			metric := pmetric.NewMetric()
			metric.SetName(name + "_quantile")
			sum := metric.SetEmptySum()
			dp := sum.DataPoints().AppendEmpty()
			dp.SetDoubleValue(qs[i].GetValue())
			_ = dp.Attributes().FromRaw(quantileDims)
			dps = append(dps, metric)
		}
	}
	return dps
}

func makeHistogramDatapoints(name string, ms []*dto.Metric) []pmetric.Metric {
	var dps []pmetric.Metric
	for _, m := range ms {
		dims := labelsToDims(m.GetLabel())
		h := m.GetHistogram()
		if h == nil {
			continue
		}

		//nolint:protogetter
		if h.SampleCount != nil {
			metric := pmetric.NewMetric()
			metric.SetName(name + "_count")
			sum := metric.SetEmptySum()
			dp := sum.DataPoints().AppendEmpty()
			dp.SetIntValue(int64(h.GetSampleCount()))
			_ = dp.Attributes().FromRaw(dims)
			dps = append(dps, metric)
		}

		//nolint:protogetter
		if h.SampleSum != nil {
			metric := pmetric.NewMetric()
			metric.SetName(name)
			sum := metric.SetEmptySum()
			dp := sum.DataPoints().AppendEmpty()
			dp.SetIntValue(int64(h.GetSampleSum()))
			_ = dp.Attributes().FromRaw(dims)
			dps = append(dps, metric)
		}

		buckets := h.GetBucket()
		for i := range buckets {
			bucketDims := utils.MergeMaps(dims, map[string]any{
				"upper_bound": strconv.FormatFloat(buckets[i].GetUpperBound(), 'f', 6, 64),
			})
			metric := pmetric.NewMetric()
			metric.SetName(name + "_quantile")
			sum := metric.SetEmptySum()
			dp := sum.DataPoints().AppendEmpty()
			dp.SetIntValue(int64(buckets[i].GetCumulativeCount()))
			_ = dp.Attributes().FromRaw(bucketDims)
			dps = append(dps, metric)
		}
	}
	return dps
}

func labelsToDims(labels []*dto.LabelPair) map[string]any {
	dims := map[string]any{}
	for i := range labels {
		dims[labels[i].GetName()] = labels[i].GetValue()
	}
	return dims
}
