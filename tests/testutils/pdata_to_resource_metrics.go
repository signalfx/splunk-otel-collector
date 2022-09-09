package testutils

import (
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// Returns a ResourceMetrics item generated from pmetric.Metrics content.  At this time histograms and summaries
// aren't supported.
func PDataToResourceMetrics(pdataMetrics ...pmetric.Metrics) (ResourceMetrics, error) {
	resourceMetrics := ResourceMetrics{}
	for _, pdataMetric := range pdataMetrics {
		pdataRMs := pdataMetric.ResourceMetrics()
		numRM := pdataRMs.Len()
		for i := 0; i < numRM; i++ {
			rm := ResourceMetric{}
			pdataRM := pdataRMs.At(i)
			pdataRM.Resource().Attributes().Range(
				func(k string, v pcommon.Value) bool {
					addResourceAttribute(&rm, k, v)
					return true
				},
			)
			pdataILMs := pdataRM.ScopeMetrics()
			for j := 0; j < pdataILMs.Len(); j++ {
				ilms := ScopeMetrics{Metrics: []Metric{}}
				pdataILM := pdataILMs.At(j)
				ilms.InstrumentationLibrary = InstrumentationLibrary{
					Name:    pdataILM.Scope().Name(),
					Version: pdataILM.Scope().Version(),
				}
				for k := 0; k < pdataILM.Metrics().Len(); k++ {
					pdataMetric := pdataILM.Metrics().At(k)
					switch pdataMetric.DataType() {
					case pmetric.MetricDataTypeGauge:
						addGauge(&ilms, pdataMetric)
					case pmetric.MetricDataTypeSum:
						addSum(&ilms, pdataMetric)
					case pmetric.MetricDataTypeHistogram:
						panic(fmt.Sprintf("%s not yet supported", pmetric.MetricDataTypeHistogram))
					case pmetric.MetricDataTypeSummary:
						panic(fmt.Sprintf("%s not yet supported", pmetric.MetricDataTypeSummary))
					default:
						panic(fmt.Sprintf("unexpected data type: %s", pdataMetric.DataType()))
					}
				}
				rm.ILMs = append(rm.ILMs, ilms)
			}
			resourceMetrics.ResourceMetrics = append(resourceMetrics.ResourceMetrics, rm)
		}
	}
	return resourceMetrics, nil
}

func addSum(ilms *ScopeMetrics, metric pmetric.Metric) {
	sum := metric.Sum()
	doubleMetricType := doubleSumMetricType(sum)
	intMetricType := intSumMetricType(sum)
	for l := 0; l < sum.DataPoints().Len(); l++ {
		dp := sum.DataPoints().At(l)
		var val any
		var metricType MetricType
		switch dp.ValueType() {
		case pmetric.NumberDataPointValueTypeInt:
			val = dp.IntVal()
			metricType = intMetricType
		case pmetric.NumberDataPointValueTypeDouble:
			val = dp.DoubleVal()
			metricType = doubleMetricType
		}
		labels := map[string]string{}
		dp.Attributes().Range(func(k string, v pcommon.Value) bool {
			labels[k] = v.AsString()
			return true
		})
		metric := Metric{
			Name:        metric.Name(),
			Description: metric.Description(),
			Unit:        metric.Unit(),
			Labels:      &labels,
			Type:        metricType,
			Value:       val,
		}
		ilms.Metrics = append(ilms.Metrics, metric)
	}
}

func addResourceAttribute(resourceMetric *ResourceMetric, name string, value pcommon.Value) {
	var val any
	switch value.Type() {
	case pcommon.ValueTypeString:
		val = value.StringVal()
	case pcommon.ValueTypeBool:
		val = value.BoolVal()
	case pcommon.ValueTypeInt:
		val = value.IntVal()
	case pcommon.ValueTypeDouble:
		val = value.DoubleVal()
	case pcommon.ValueTypeMap:
		val = value.MapVal().AsRaw()
	case pcommon.ValueTypeSlice:
		val = value.SliceVal().AsRaw()
	default:
		val = nil
	}
	if resourceMetric.Resource.Attributes == nil {
		resourceMetric.Resource.Attributes = map[string]any{}
	}
	resourceMetric.Resource.Attributes[name] = val
}

func addGauge(ilms *ScopeMetrics, metric pmetric.Metric) {
	doubleGauge := metric.Gauge()
	for l := 0; l < doubleGauge.DataPoints().Len(); l++ {
		dp := doubleGauge.DataPoints().At(l)
		var val any
		var metricType MetricType
		switch dp.ValueType() {
		case pmetric.NumberDataPointValueTypeInt:
			val = dp.IntVal()
			metricType = IntGauge
		case pmetric.NumberDataPointValueTypeDouble:
			val = dp.DoubleVal()
			metricType = DoubleGauge
		}
		labels := map[string]string{}
		dp.Attributes().Range(func(k string, v pcommon.Value) bool {
			labels[k] = v.AsString()
			return true
		})
		metric := Metric{
			Name:        metric.Name(),
			Description: metric.Description(),
			Unit:        metric.Unit(),
			Labels:      &labels,
			Type:        metricType,
			Value:       val,
		}
		ilms.Metrics = append(ilms.Metrics, metric)
	}
}

func doubleSumMetricType(sum pmetric.Sum) MetricType {
	switch sum.AggregationTemporality() {
	case pmetric.MetricAggregationTemporalityCumulative:
		if sum.IsMonotonic() {
			return DoubleMonotonicCumulativeSum
		} else {
			return DoubleNonmonotonicCumulativeSum
		}
	case pmetric.MetricAggregationTemporalityDelta:
		if sum.IsMonotonic() {
			return DoubleMonotonicDeltaSum
		} else {
			return DoubleNonmonotonicDeltaSum
		}
	case pmetric.MetricAggregationTemporalityUnspecified:
		if sum.IsMonotonic() {
			return DoubleMonotonicUnspecifiedSum
		} else {
			return DoubleNonmonotonicUnspecifiedSum
		}
	}
	return "unknown"
}

func intSumMetricType(sum pmetric.Sum) MetricType {
	switch sum.AggregationTemporality() {
	case pmetric.MetricAggregationTemporalityCumulative:
		if sum.IsMonotonic() {
			return IntMonotonicCumulativeSum
		} else {
			return IntNonmonotonicCumulativeSum
		}
	case pmetric.MetricAggregationTemporalityDelta:
		if sum.IsMonotonic() {
			return IntMonotonicDeltaSum
		} else {
			return IntNonmonotonicDeltaSum
		}
	case pmetric.MetricAggregationTemporalityUnspecified:
		if sum.IsMonotonic() {
			return IntMonotonicUnspecifiedSum
		} else {
			return IntNonmonotonicUnspecifiedSum
		}
	}
	return "unknown"
}
