package testutils

import (
	"fmt"

	"go.opentelemetry.io/collector/model/pdata"
)

// Returns a ResourceMetrics item generated from pdata.Metrics content.  At this time histograms and summaries
// aren't supported.
func PDataToResourceMetrics(pdataMetrics ...pdata.Metrics) (ResourceMetrics, error) {
	resourceMetrics := ResourceMetrics{}
	for _, pdataMetric := range pdataMetrics {
		pdataRMs := pdataMetric.ResourceMetrics()
		numRM := pdataRMs.Len()
		for i := 0; i < numRM; i++ {
			rm := ResourceMetric{}
			pdataRM := pdataRMs.At(i)
			pdataRM.Resource().Attributes().Range(
				func(k string, v pdata.AttributeValue) bool {
					addResourceAttribute(&rm, k, v)
					return true
				},
			)
			pdataILMs := pdataRM.InstrumentationLibraryMetrics()
			for j := 0; j < pdataILMs.Len(); j++ {
				ilms := InstrumentationLibraryMetrics{Metrics: []Metric{}}
				pdataILM := pdataILMs.At(j)
				ilms.InstrumentationLibrary = InstrumentationLibrary{
					Name:    pdataILM.InstrumentationLibrary().Name(),
					Version: pdataILM.InstrumentationLibrary().Version(),
				}
				for k := 0; k < pdataILM.Metrics().Len(); k++ {
					pdataMetric := pdataILM.Metrics().At(k)
					switch pdataMetric.DataType() {
					case pdata.MetricDataTypeGauge:
						addGauge(&ilms, pdataMetric)
					case pdata.MetricDataTypeSum:
						addSum(&ilms, pdataMetric)
					case pdata.MetricDataTypeHistogram:
						panic(fmt.Sprintf("%s not yet supported", pdata.MetricDataTypeHistogram))
					case pdata.MetricDataTypeSummary:
						panic(fmt.Sprintf("%s not yet supported", pdata.MetricDataTypeSummary))
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

func addSum(ilms *InstrumentationLibraryMetrics, metric pdata.Metric) {
	sum := metric.Sum()
	doubleMetricType := doubleSumMetricType(sum)
	intMetricType := intSumMetricType(sum)
	for l := 0; l < sum.DataPoints().Len(); l++ {
		dp := sum.DataPoints().At(l)
		var val interface{}
		var metricType MetricType
		switch dp.Type() {
		case pdata.MetricValueTypeInt:
			val = dp.IntVal()
			metricType = intMetricType
		case pdata.MetricValueTypeDouble:
			val = dp.DoubleVal()
			metricType = doubleMetricType
		}
		labels := map[string]string{}
		dp.Attributes().Range(func(k string, v pdata.AttributeValue) bool {
			labels[k] = pdata.AttributeValueToString(v)
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

func addResourceAttribute(resourceMetric *ResourceMetric, name string, value pdata.AttributeValue) {
	var val interface{}
	switch value.Type() {
	case pdata.AttributeValueTypeString:
		val = value.StringVal()
	case pdata.AttributeValueTypeBool:
		val = value.BoolVal()
	case pdata.AttributeValueTypeInt:
		val = value.IntVal()
	case pdata.AttributeValueTypeDouble:
		val = value.DoubleVal()
	case pdata.AttributeValueTypeMap:
		// Coerce to map[string]interface{}
		val = pdata.AttributeMapToMap(value.MapVal())
	case pdata.AttributeValueTypeArray:
		// Coerce to []interface{}
		// Required pdata helper is not exposed so we pass value as a map
		// and use helper that calls it internally.
		toTranslate := pdata.NewAttributeMap()
		toTranslate.Insert(name, value)
		translated := pdata.AttributeMapToMap(toTranslate)
		val = translated[name]
	default:
		val = nil
	}
	if resourceMetric.Resource.Attributes == nil {
		resourceMetric.Resource.Attributes = map[string]interface{}{}
	}
	resourceMetric.Resource.Attributes[name] = val
}

func addGauge(ilms *InstrumentationLibraryMetrics, metric pdata.Metric) {
	doubleGauge := metric.Gauge()
	for l := 0; l < doubleGauge.DataPoints().Len(); l++ {
		dp := doubleGauge.DataPoints().At(l)
		var val interface{}
		var metricType MetricType
		switch dp.Type() {
		case pdata.MetricValueTypeInt:
			val = dp.IntVal()
			metricType = IntGauge
		case pdata.MetricValueTypeDouble:
			val = dp.DoubleVal()
			metricType = DoubleGauge
		}
		labels := map[string]string{}
		dp.Attributes().Range(func(k string, v pdata.AttributeValue) bool {
			labels[k] = pdata.AttributeValueToString(v)
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

func doubleSumMetricType(sum pdata.Sum) MetricType {
	switch sum.AggregationTemporality() {
	case pdata.AggregationTemporalityCumulative:
		if sum.IsMonotonic() {
			return DoubleMonotonicCumulativeSum
		} else {
			return DoubleNonmonotonicCumulativeSum
		}
	case pdata.AggregationTemporalityDelta:
		if sum.IsMonotonic() {
			return DoubleMonotonicDeltaSum
		} else {
			return DoubleNonmonotonicDeltaSum
		}
	case pdata.AggregationTemporalityUnspecified:
		if sum.IsMonotonic() {
			return DoubleMonotonicUnspecifiedSum
		} else {
			return DoubleNonmonotonicUnspecifiedSum
		}
	}
	return "unknown"
}

func intSumMetricType(sum pdata.Sum) MetricType {
	switch sum.AggregationTemporality() {
	case pdata.AggregationTemporalityCumulative:
		if sum.IsMonotonic() {
			return IntMonotonicCumulativeSum
		} else {
			return IntNonmonotonicCumulativeSum
		}
	case pdata.AggregationTemporalityDelta:
		if sum.IsMonotonic() {
			return IntMonotonicDeltaSum
		} else {
			return IntNonmonotonicDeltaSum
		}
	case pdata.AggregationTemporalityUnspecified:
		if sum.IsMonotonic() {
			return IntMonotonicUnspecifiedSum
		} else {
			return IntNonmonotonicUnspecifiedSum
		}
	}
	return "unknown"
}
