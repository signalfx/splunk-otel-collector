package prometheusexporter

import (
	"strconv"

	dto "github.com/prometheus/client_model/go"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/signalfx-agent/pkg/utils"
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

func convertMetricFamily(mf *dto.MetricFamily) []*datapoint.Datapoint {
	if mf.Type == nil || mf.Name == nil {
		return nil
	}
	switch *mf.Type {
	case dto.MetricType_GAUGE:
		return makeSimpleDatapoints(*mf.Name, mf.Metric, sfxclient.GaugeF, gaugeExtractor)
	case dto.MetricType_COUNTER:
		return makeSimpleDatapoints(*mf.Name, mf.Metric, sfxclient.CumulativeF, counterExtractor)
	case dto.MetricType_UNTYPED:
		return makeSimpleDatapoints(*mf.Name, mf.Metric, sfxclient.GaugeF, untypedExtractor)
	case dto.MetricType_SUMMARY:
		return makeSummaryDatapoints(*mf.Name, mf.Metric)
	// TODO: figure out how to best convert histograms, in particular the
	// upper bound value
	case dto.MetricType_HISTOGRAM:
		return makeHistogramDatapoints(*mf.Name, mf.Metric)
	default:
		return nil
	}
}

func makeSimpleDatapoints(name string, ms []*dto.Metric, dpf dpFactory, e extractor) []*datapoint.Datapoint {
	dps := make([]*datapoint.Datapoint, len(ms))
	for i, m := range ms {
		dps[i] = dpf(name, labelsToDims(m.Label), e(m))
	}
	return dps
}

func makeSummaryDatapoints(name string, ms []*dto.Metric) []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint
	for _, m := range ms {
		dims := labelsToDims(m.Label)
		s := m.GetSummary()
		if s == nil {
			continue
		}

		if s.SampleCount != nil {
			dps = append(dps, sfxclient.Cumulative(name+"_count", dims, int64(s.GetSampleCount())))
		}

		if s.SampleSum != nil {
			dps = append(dps, sfxclient.CumulativeF(name, dims, s.GetSampleSum()))
		}

		qs := s.GetQuantile()
		for i := range qs {
			quantileDims := utils.MergeStringMaps(dims, map[string]string{
				"quantile": strconv.FormatFloat(qs[i].GetQuantile(), 'f', 6, 64),
			})
			dps = append(dps, sfxclient.GaugeF(name+"_quantile", quantileDims, qs[i].GetValue()))
		}
	}
	return dps
}

func makeHistogramDatapoints(name string, ms []*dto.Metric) []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint
	for _, m := range ms {
		dims := labelsToDims(m.Label)
		h := m.GetHistogram()
		if h == nil {
			continue
		}

		if h.SampleCount != nil {
			dps = append(dps, sfxclient.Cumulative(name+"_count", dims, int64(h.GetSampleCount())))
		}

		if h.SampleSum != nil {
			dps = append(dps, sfxclient.CumulativeF(name, dims, h.GetSampleSum()))
		}

		buckets := h.GetBucket()
		for i := range buckets {
			bucketDims := utils.MergeStringMaps(dims, map[string]string{
				"upper_bound": strconv.FormatFloat(buckets[i].GetUpperBound(), 'f', 6, 64),
			})
			dps = append(dps, sfxclient.Cumulative(name+"_bucket", bucketDims, int64(buckets[i].GetCumulativeCount())))
		}
	}
	return dps
}

func labelsToDims(labels []*dto.LabelPair) map[string]string {
	dims := map[string]string{}
	for i := range labels {
		dims[labels[i].GetName()] = labels[i].GetValue()
	}
	return dims
}
