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
			pdataRM.Resource().Attributes().ForEach(
				func(k string, v pdata.AttributeValue) {
					var val interface{}
					switch v.Type() {
					case pdata.AttributeValueSTRING:
						val = v.StringVal()
					case pdata.AttributeValueBOOL:
						val = v.BoolVal()
					case pdata.AttributeValueINT:
						val = v.IntVal()
					case pdata.AttributeValueDOUBLE:
						val = v.DoubleVal()
					case pdata.AttributeValueMAP:
						// Coerce to map[string]interface{}
						val = tracetranslator.AttributeMapToMap(v.MapVal())
					case pdata.AttributeValueARRAY:
						// Coerce to []interface{}
						val = tracetranslator.AttributeArrayToSlice(v.ArrayVal())
					default:
						val = nil
					}
					if rm.Resource.Attributes == nil {
						rm.Resource.Attributes = map[string]interface{}{}
					}
					rm.Resource.Attributes[k] = val
				})
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
						intGauge := pdataMetric.IntGauge()
						for l := 0; l < intGauge.DataPoints().Len(); l++ {
							dp := intGauge.DataPoints().At(l)
							val := dp.Value()
							labels := map[string]string{}
							dp.LabelsMap().ForEach(func(k, v string) {
								labels[k] = v

							})
							metric := Metric{
								Name:        pdataMetric.Name(),
								Description: pdataMetric.Description(),
								Unit:        pdataMetric.Unit(),
								Labels:      labels,
								Type:        IntGauge,
								Value:       val,
							}
							ilms.Metrics = append(ilms.Metrics, metric)
						}
					case pdata.MetricDataTypeDoubleGauge:
						doubleGauge := pdataMetric.DoubleGauge()
						for l := 0; l < doubleGauge.DataPoints().Len(); l++ {
							dp := doubleGauge.DataPoints().At(l)
							val := dp.Value()
							labels := map[string]string{}
							dp.LabelsMap().ForEach(func(k, v string) {
								labels[k] = v

							})
							metric := Metric{
								Name:        pdataMetric.Name(),
								Description: pdataMetric.Description(),
								Unit:        pdataMetric.Unit(),
								Labels:      labels,
								Type:        DoubleGauge,
								Value:       val,
							}
							ilms.Metrics = append(ilms.Metrics, metric)
						}
					case pdata.MetricDataTypeIntSum:
						intSum := pdataMetric.IntSum()
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
							dp.LabelsMap().ForEach(func(k, v string) {
								labels[k] = v

							})
							metric := Metric{
								Name:        pdataMetric.Name(),
								Description: pdataMetric.Description(),
								Unit:        pdataMetric.Unit(),
								Labels:      labels,
								Type:        metricType,
								Value:       val,
							}
							ilms.Metrics = append(ilms.Metrics, metric)
						}
					case pdata.MetricDataTypeDoubleSum:
						doubleSum := pdataMetric.DoubleSum()
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
							dp.LabelsMap().ForEach(func(k, v string) {
								labels[k] = v

							})
							metric := Metric{
								Name:        pdataMetric.Name(),
								Description: pdataMetric.Description(),
								Unit:        pdataMetric.Unit(),
								Labels:      labels,
								Type:        metricType,
								Value:       val,
							}
							ilms.Metrics = append(ilms.Metrics, metric)
						}
					case pdata.MetricDataTypeIntHistogram:
						panic(fmt.Sprintf("%s not yet supported", pdata.MetricDataTypeIntHistogram))
					case pdata.MetricDataTypeDoubleHistogram:
						panic(fmt.Sprintf("%s not yet supported", pdata.MetricDataTypeDoubleHistogram))
					case pdata.MetricDataTypeDoubleSummary:
						panic(fmt.Sprintf("%s not yet supported", pdata.MetricDataTypeDoubleSummary))
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
