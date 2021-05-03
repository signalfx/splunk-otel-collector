package testutils

import (
	"fmt"

	"go.opentelemetry.io/collector/consumer/pdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
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
					case pdata.MetricDataTypeIntGauge:
						addIntGauge(&ilms, pdataMetric)
					case pdata.MetricDataTypeDoubleGauge:
						addDoubleGauge(&ilms, pdataMetric)
					case pdata.MetricDataTypeIntSum:
						addIntSum(&ilms, pdataMetric)
					case pdata.MetricDataTypeDoubleSum:
						addDoubleSum(&ilms, pdataMetric)
					case pdata.MetricDataTypeIntHistogram:
						panic(fmt.Sprintf("%s not yet supported", pdata.MetricDataTypeIntHistogram))
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

func addDoubleSum(ilms *InstrumentationLibraryMetrics, metric pdata.Metric) {
	doubleSum := metric.DoubleSum()
	var metricType MetricType
	switch doubleSum.AggregationTemporality() {
	case pdata.AggregationTemporalityCumulative:
		if doubleSum.IsMonotonic() {
			metricType = DoubleMonotonicCumulativeSum
		} else {
			metricType = DoubleNonmonotonicCumulativeSum
		}
	case pdata.AggregationTemporalityDelta:
		if doubleSum.IsMonotonic() {
			metricType = DoubleMonotonicDeltaSum
		} else {
			metricType = DoubleNonmonotonicDeltaSum
		}
	case pdata.AggregationTemporalityUnspecified:
		if doubleSum.IsMonotonic() {
			metricType = DoubleMonotonicUnspecifiedSum
		} else {
			metricType = DoubleNonmonotonicUnspecifiedSum
		}
	}
	for l := 0; l < doubleSum.DataPoints().Len(); l++ {
		dp := doubleSum.DataPoints().At(l)
		val := dp.Value()
		labels := map[string]string{}
		dp.LabelsMap().Range(func(k, v string) bool {
			labels[k] = v
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
	case pdata.AttributeValueSTRING:
		val = value.StringVal()
	case pdata.AttributeValueBOOL:
		val = value.BoolVal()
	case pdata.AttributeValueINT:
		val = value.IntVal()
	case pdata.AttributeValueDOUBLE:
		val = value.DoubleVal()
	case pdata.AttributeValueMAP:
		// Coerce to map[string]interface{}
		val = tracetranslator.AttributeMapToMap(value.MapVal())
	case pdata.AttributeValueARRAY:
		// Coerce to []interface{}
		val = tracetranslator.AttributeArrayToSlice(value.ArrayVal())
	default:
		val = nil
	}
	if resourceMetric.Resource.Attributes == nil {
		resourceMetric.Resource.Attributes = map[string]interface{}{}
	}
	resourceMetric.Resource.Attributes[name] = val
}

func addIntSum(ilms *InstrumentationLibraryMetrics, metric pdata.Metric) {
	intSum := metric.IntSum()
	var metricType MetricType
	switch intSum.AggregationTemporality() {
	case pdata.AggregationTemporalityCumulative:
		if intSum.IsMonotonic() {
			metricType = IntMonotonicCumulativeSum
		} else {
			metricType = IntNonmonotonicCumulativeSum
		}
	case pdata.AggregationTemporalityDelta:
		if intSum.IsMonotonic() {
			metricType = IntMonotonicDeltaSum
		} else {
			metricType = IntNonmonotonicDeltaSum
		}
	case pdata.AggregationTemporalityUnspecified:
		if intSum.IsMonotonic() {
			metricType = IntMonotonicUnspecifiedSum
		} else {
			metricType = IntNonmonotonicUnspecifiedSum
		}
	}
	for l := 0; l < intSum.DataPoints().Len(); l++ {
		dp := intSum.DataPoints().At(l)
		val := dp.Value()
		labels := map[string]string{}
		dp.LabelsMap().Range(func(k, v string) bool {
			labels[k] = v
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

func addDoubleGauge(ilms *InstrumentationLibraryMetrics, metric pdata.Metric) {
	doubleGauge := metric.DoubleGauge()
	for l := 0; l < doubleGauge.DataPoints().Len(); l++ {
		dp := doubleGauge.DataPoints().At(l)
		val := dp.Value()
		labels := map[string]string{}
		dp.LabelsMap().Range(func(k, v string) bool {
			labels[k] = v
			return true
		})
		metric := Metric{
			Name:        metric.Name(),
			Description: metric.Description(),
			Unit:        metric.Unit(),
			Labels:      &labels,
			Type:        DoubleGauge,
			Value:       val,
		}
		ilms.Metrics = append(ilms.Metrics, metric)
	}
}

func addIntGauge(ilms *InstrumentationLibraryMetrics, metric pdata.Metric) {
	intGauge := metric.IntGauge()
	for l := 0; l < intGauge.DataPoints().Len(); l++ {
		dp := intGauge.DataPoints().At(l)
		val := dp.Value()
		labels := map[string]string{}
		dp.LabelsMap().Range(func(k, v string) bool {
			labels[k] = v
			return true
		})
		metric := Metric{
			Name:        metric.Name(),
			Description: metric.Description(),
			Unit:        metric.Unit(),
			Labels:      &labels,
			Type:        IntGauge,
			Value:       val,
		}
		ilms.Metrics = append(ilms.Metrics, metric)
	}
}
