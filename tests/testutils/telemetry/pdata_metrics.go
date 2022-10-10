// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telemetry

import (
	"fmt"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

// PDataToResourceMetrics returns a ResourceMetrics item generated from pmetric.Metrics content.
// At this time histograms and summaries aren't supported.
func PDataToResourceMetrics(pdataMetrics ...pmetric.Metrics) (ResourceMetrics, error) {
	resourceMetrics := ResourceMetrics{}
	for _, pdataMetric := range pdataMetrics {
		pdataRMs := pdataMetric.ResourceMetrics()
		numRM := pdataRMs.Len()
		for i := 0; i < numRM; i++ {
			rm := ResourceMetric{}
			pdataRM := pdataRMs.At(i)
			rm.Resource.Attributes = sanitizeAttributes(pdataRM.Resource().Attributes().AsRaw())
			pdataSMs := pdataRM.ScopeMetrics()
			for j := 0; j < pdataSMs.Len(); j++ {
				ISMs := ScopeMetrics{Metrics: []Metric{}}
				pdataSM := pdataSMs.At(j)
				ISMs.Scope = InstrumentationScope{
					Name:    pdataSM.Scope().Name(),
					Version: pdataSM.Scope().Version(),
				}
				for k := 0; k < pdataSM.Metrics().Len(); k++ {
					pdMetric := pdataSM.Metrics().At(k)
					switch t := pdMetric.Type(); t {
					case pmetric.MetricTypeGauge:
						addGauge(&ISMs, pdMetric)
					case pmetric.MetricTypeSum:
						addSum(&ISMs, pdMetric)
					case pmetric.MetricTypeHistogram:
						panic(fmt.Sprintf("%s not yet supported", pmetric.MetricTypeHistogram))
					case pmetric.MetricTypeSummary:
						panic(fmt.Sprintf("%s not yet supported", pmetric.MetricTypeSummary))
					default:
						panic(fmt.Sprintf("unexpected data type: %s", t))
					}
				}
				rm.ScopeMetrics = append(rm.ScopeMetrics, ISMs)
			}
			resourceMetrics.ResourceMetrics = append(resourceMetrics.ResourceMetrics, rm)
		}
	}
	return resourceMetrics, nil
}

func addSum(sms *ScopeMetrics, metric pmetric.Metric) {
	sum := metric.Sum()
	doubleMetricType := doubleSumMetricType(sum)
	intMetricType := intSumMetricType(sum)
	for l := 0; l < sum.DataPoints().Len(); l++ {
		dp := sum.DataPoints().At(l)
		var val any
		var metricType MetricType
		switch dp.ValueType() {
		case pmetric.NumberDataPointValueTypeInt:
			val = dp.IntValue()
			metricType = intMetricType
		case pmetric.NumberDataPointValueTypeDouble:
			val = dp.DoubleValue()
			metricType = doubleMetricType
		}
		attributes := sanitizeAttributes(dp.Attributes().AsRaw())
		m := Metric{
			Name:        metric.Name(),
			Description: metric.Description(),
			Unit:        metric.Unit(),
			Attributes:  &attributes,
			Type:        metricType,
			Value:       val,
		}
		sms.Metrics = append(sms.Metrics, m)
	}
}

func addGauge(sms *ScopeMetrics, metric pmetric.Metric) {
	doubleGauge := metric.Gauge()
	for l := 0; l < doubleGauge.DataPoints().Len(); l++ {
		dp := doubleGauge.DataPoints().At(l)
		var val any
		var metricType MetricType
		switch dp.ValueType() {
		case pmetric.NumberDataPointValueTypeInt:
			val = dp.IntValue()
			metricType = IntGauge
		case pmetric.NumberDataPointValueTypeDouble:
			val = dp.DoubleValue()
			metricType = DoubleGauge
		}
		attributes := sanitizeAttributes(dp.Attributes().AsRaw())
		m := Metric{
			Name:        metric.Name(),
			Description: metric.Description(),
			Unit:        metric.Unit(),
			Attributes:  &attributes,
			Type:        metricType,
			Value:       val,
		}
		sms.Metrics = append(sms.Metrics, m)
	}
}

func doubleSumMetricType(sum pmetric.Sum) MetricType {
	switch sum.AggregationTemporality() {
	case pmetric.MetricAggregationTemporalityCumulative:
		if sum.IsMonotonic() {
			return DoubleMonotonicCumulativeSum
		}
		return DoubleNonmonotonicCumulativeSum
	case pmetric.MetricAggregationTemporalityDelta:
		if sum.IsMonotonic() {
			return DoubleMonotonicDeltaSum
		}
		return DoubleNonmonotonicDeltaSum
	case pmetric.MetricAggregationTemporalityUnspecified:
		if sum.IsMonotonic() {
			return DoubleMonotonicUnspecifiedSum
		}
		return DoubleNonmonotonicUnspecifiedSum
	}
	return "unknown"
}

func intSumMetricType(sum pmetric.Sum) MetricType {
	switch sum.AggregationTemporality() {
	case pmetric.MetricAggregationTemporalityCumulative:
		if sum.IsMonotonic() {
			return IntMonotonicCumulativeSum
		}
		return IntNonmonotonicCumulativeSum
	case pmetric.MetricAggregationTemporalityDelta:
		if sum.IsMonotonic() {
			return IntMonotonicDeltaSum
		}
		return IntNonmonotonicDeltaSum
	case pmetric.MetricAggregationTemporalityUnspecified:
		if sum.IsMonotonic() {
			return IntMonotonicUnspecifiedSum
		}
		return IntNonmonotonicUnspecifiedSum
	}
	return "unknown"
}

func PDataMetrics() pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
	attrs := resourceMetrics.Resource().Attributes()
	attrs.PutBool("bool", true)
	attrs.PutString("string", "a_string")
	attrs.PutInt("int", 123)
	attrs.PutDouble("double", 123.45)
	attrs.PutEmpty("null")

	scopeMetrics := resourceMetrics.ScopeMetrics()
	smOne := scopeMetrics.AppendEmpty()
	smOne.Scope().SetName("an_instrumentation_scope_name")
	smOne.Scope().SetVersion("an_instrumentation_scope_version")
	smOneMetrics := smOne.Metrics()
	smOneMetricOne := smOneMetrics.AppendEmpty()
	smOneMetricOne.SetName("an_int_gauge")
	smOneMetricOne.SetDescription("an_int_gauge_description")
	smOneMetricOne.SetUnit("an_int_gauge_unit")
	smOneMetricOne.SetEmptyGauge()
	smOneMetricOneDps := smOneMetricOne.Gauge().DataPoints()
	smOneMetricOneDps.AppendEmpty().SetIntValue(12345)
	smOneMetricOneDps.At(0).Attributes().PutString("attribute_name_1", "attribute_value_1")
	smOneMetricOneDps.AppendEmpty().SetIntValue(23456)
	smOneMetricOneDps.At(1).Attributes().PutString("attribute_name_2", "attribute_value_2")

	smOneMetricTwo := smOneMetrics.AppendEmpty()
	smOneMetricTwo.SetName("a_double_gauge")
	smOneMetricTwo.SetDescription("a_double_gauge_description")
	smOneMetricTwo.SetUnit("a_double_gauge_unit")
	smOneMetricTwo.SetEmptyGauge()
	smOneMetricTwoDps := smOneMetricTwo.Gauge().DataPoints()
	smOneMetricTwoDps.AppendEmpty().SetDoubleValue(234.56)
	smOneMetricTwoDps.At(0).Attributes().PutString("attribute_name_3", "attribute_value_3")
	smOneMetricTwoDps.AppendEmpty().SetDoubleValue(345.67)
	smOneMetricTwoDps.At(1).Attributes().PutString("attribute_name_4", "attribute_value_4")

	scopeMetrics.AppendEmpty().Scope().SetName("an_instrumentation_scope_without_version_or_metrics")

	smThreeMetrics := scopeMetrics.AppendEmpty().Metrics()

	smThreeMetricOne := smThreeMetrics.AppendEmpty()
	smThreeMetricOne.SetName("a_monotonic_cumulative_int_sum")
	smThreeMetricOne.SetDescription("a_monotonic_cumulative_int_sum_description")
	smThreeMetricOne.SetUnit("a_monotonic_cumulative_int_sum_unit")
	smThreeMetricOne.SetEmptySum()
	smThreeMetricOne.Sum().SetIsMonotonic(true)
	smThreeMetricOne.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
	smThreeMetricOneDps := smThreeMetricOne.Sum().DataPoints()
	smThreeMetricOneDps.AppendEmpty().SetIntValue(34567)
	smThreeMetricOneDps.At(0).Attributes().PutString("attribute_name_5", "attribute_value_5")
	smThreeMetricOneDps.AppendEmpty().SetIntValue(45678)
	smThreeMetricOneDps.At(1).Attributes().PutString("attribute_name_6", "attribute_value_6")

	smThreeMetricTwo := smThreeMetrics.AppendEmpty()
	smThreeMetricTwo.SetName("a_monotonic_delta_int_sum")
	smThreeMetricTwo.SetDescription("a_monotonic_delta_int_sum_description")
	smThreeMetricTwo.SetUnit("a_monotonic_delta_int_sum_unit")
	smThreeMetricTwo.SetEmptySum()
	smThreeMetricTwo.Sum().SetIsMonotonic(true)
	smThreeMetricTwo.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityDelta)
	smThreeMetricTwoDps := smThreeMetricTwo.Sum().DataPoints()
	smThreeMetricTwoDps.AppendEmpty().SetIntValue(56789)
	smThreeMetricTwoDps.At(0).Attributes().PutString("attribute_name_7", "attribute_value_7")
	smThreeMetricTwoDps.AppendEmpty().SetIntValue(67890)
	smThreeMetricTwoDps.At(1).Attributes().PutString("attribute_name_8", "attribute_value_8")

	smThreeMetricThree := smThreeMetrics.AppendEmpty()
	smThreeMetricThree.SetName("a_monotonic_unspecified_int_sum")
	smThreeMetricThree.SetDescription("a_monotonic_unspecified_int_sum_description")
	smThreeMetricThree.SetUnit("a_monotonic_unspecified_int_sum_unit")
	smThreeMetricThree.SetEmptySum()
	smThreeMetricThree.Sum().SetIsMonotonic(true)
	smThreeMetricThree.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityUnspecified)
	smThreeMetricThreeDps := smThreeMetricThree.Sum().DataPoints()
	smThreeMetricThreeDps.AppendEmpty().SetIntValue(78901)
	smThreeMetricThreeDps.At(0).Attributes().PutString("attribute_name_9", "attribute_value_9")
	smThreeMetricThreeDps.AppendEmpty().SetIntValue(89012)
	smThreeMetricThreeDps.At(1).Attributes().PutString("attribute_name_10", "attribute_value_10")

	smThreeMetricFour := smThreeMetrics.AppendEmpty()
	smThreeMetricFour.SetName("a_monotonic_cumulative_double_sum")
	smThreeMetricFour.SetDescription("a_monotonic_cumulative_double_sum_description")
	smThreeMetricFour.SetUnit("a_monotonic_cumulative_double_sum_unit")
	smThreeMetricFour.SetEmptySum()
	smThreeMetricFour.Sum().SetIsMonotonic(true)
	smThreeMetricFour.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
	smThreeMetricFourDps := smThreeMetricFour.Sum().DataPoints()
	smThreeMetricFourDps.AppendEmpty().SetDoubleValue(456.78)
	smThreeMetricFourDps.At(0).Attributes().PutString("attribute_name_11", "attribute_value_11")
	smThreeMetricFourDps.AppendEmpty().SetDoubleValue(567.89)
	smThreeMetricFourDps.At(1).Attributes().PutString("attribute_name_12", "attribute_value_12")

	smThreeMetricFive := smThreeMetrics.AppendEmpty()
	smThreeMetricFive.SetName("a_monotonic_delta_double_sum")
	smThreeMetricFive.SetDescription("a_monotonic_delta_double_sum_description")
	smThreeMetricFive.SetUnit("a_monotonic_delta_double_sum_unit")
	smThreeMetricFive.SetEmptySum()
	smThreeMetricFive.Sum().SetIsMonotonic(true)
	smThreeMetricFive.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityDelta)
	smThreeMetricFiveDps := smThreeMetricFive.Sum().DataPoints()
	smThreeMetricFiveDps.AppendEmpty().SetDoubleValue(678.90)
	smThreeMetricFiveDps.At(0).Attributes().PutString("attribute_name_13", "attribute_value_13")
	smThreeMetricFiveDps.AppendEmpty().SetDoubleValue(789.01)
	smThreeMetricFiveDps.At(1).Attributes().PutString("attribute_name_14", "attribute_value_14")

	smThreeMetricSix := smThreeMetrics.AppendEmpty()
	smThreeMetricSix.SetName("a_monotonic_unspecified_double_sum")
	smThreeMetricSix.SetDescription("a_monotonic_unspecified_double_sum_description")
	smThreeMetricSix.SetUnit("a_monotonic_unspecified_double_sum_unit")
	smThreeMetricSix.SetEmptySum()
	smThreeMetricSix.Sum().SetIsMonotonic(true)
	smThreeMetricSix.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityUnspecified)
	smThreeMetricSixDps := smThreeMetricSix.Sum().DataPoints()
	smThreeMetricSixDps.AppendEmpty().SetDoubleValue(890.12)
	smThreeMetricSixDps.At(0).Attributes().PutString("attribute_name_15", "attribute_value_15")
	smThreeMetricSixDps.AppendEmpty().SetDoubleValue(901.23)
	smThreeMetricSixDps.At(1).Attributes().PutString("attribute_name_16", "attribute_value_16")

	smThreeMetricSeven := smThreeMetrics.AppendEmpty()
	smThreeMetricSeven.SetName("a_nonmonotonic_cumulative_int_sum")
	smThreeMetricSeven.SetDescription("a_nonmonotonic_cumulative_int_sum_description")
	smThreeMetricSeven.SetUnit("a_nonmonotonic_cumulative_int_sum_unit")
	smThreeMetricSeven.SetEmptySum()
	smThreeMetricSeven.Sum().SetIsMonotonic(false)
	smThreeMetricSeven.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
	smThreeMetricSevenDps := smThreeMetricSeven.Sum().DataPoints()
	smThreeMetricSevenDps.AppendEmpty().SetIntValue(90123)
	smThreeMetricSevenDps.At(0).Attributes().PutString("attribute_name_17", "attribute_value_17")
	smThreeMetricSevenDps.AppendEmpty().SetIntValue(123456)
	smThreeMetricSevenDps.At(1).Attributes().PutString("attribute_name_18", "attribute_value_18")

	smThreeMetricEight := smThreeMetrics.AppendEmpty()
	smThreeMetricEight.SetName("a_nonmonotonic_delta_int_sum")
	smThreeMetricEight.SetDescription("a_nonmonotonic_delta_int_sum_description")
	smThreeMetricEight.SetUnit("a_nonmonotonic_delta_int_sum_unit")
	smThreeMetricEight.SetEmptySum()
	smThreeMetricEight.Sum().SetIsMonotonic(false)
	smThreeMetricEight.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityDelta)
	smThreeMetricEightDps := smThreeMetricEight.Sum().DataPoints()
	smThreeMetricEightDps.AppendEmpty().SetIntValue(234567)
	smThreeMetricEightDps.At(0).Attributes().PutString("attribute_name_19", "attribute_value_19")
	smThreeMetricEightDps.AppendEmpty().SetIntValue(345678)
	smThreeMetricEightDps.At(1).Attributes().PutString("attribute_name_20", "attribute_value_20")

	smThreeMetricNine := smThreeMetrics.AppendEmpty()
	smThreeMetricNine.SetName("a_nonmonotonic_unspecified_int_sum")
	smThreeMetricNine.SetDescription("a_nonmonotonic_unspecified_int_sum_description")
	smThreeMetricNine.SetUnit("a_nonmonotonic_unspecified_int_sum_unit")
	smThreeMetricNine.SetEmptySum()
	smThreeMetricNine.Sum().SetIsMonotonic(false)
	smThreeMetricNine.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityUnspecified)
	smThreeMetricNineDps := smThreeMetricNine.Sum().DataPoints()
	smThreeMetricNineDps.AppendEmpty().SetIntValue(456789)
	smThreeMetricNineDps.At(0).Attributes().PutString("attribute_name_21", "attribute_value_21")
	smThreeMetricNineDps.AppendEmpty().SetIntValue(567890)
	smThreeMetricNineDps.At(1).Attributes().PutString("attribute_name_22", "attribute_value_22")

	smThreeMetricTen := smThreeMetrics.AppendEmpty()
	smThreeMetricTen.SetName("a_nonmonotonic_cumulative_double_sum")
	smThreeMetricTen.SetDescription("a_nonmonotonic_cumulative_double_sum_description")
	smThreeMetricTen.SetUnit("a_nonmonotonic_cumulative_double_sum_unit")
	smThreeMetricTen.SetEmptySum()
	smThreeMetricTen.Sum().SetIsMonotonic(false)
	smThreeMetricTen.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityCumulative)
	smThreeMetricTenDps := smThreeMetricTen.Sum().DataPoints()
	smThreeMetricTenDps.AppendEmpty().SetDoubleValue(1234.56)
	smThreeMetricTenDps.At(0).Attributes().PutString("attribute_name_23", "attribute_value_23")
	smThreeMetricTenDps.AppendEmpty().SetDoubleValue(2345.67)
	smThreeMetricTenDps.At(1).Attributes().PutString("attribute_name_24", "attribute_value_24")

	smThreeMetricEleven := smThreeMetrics.AppendEmpty()
	smThreeMetricEleven.SetName("a_nonmonotonic_delta_double_sum")
	smThreeMetricEleven.SetDescription("a_nonmonotonic_delta_double_sum_description")
	smThreeMetricEleven.SetUnit("a_nonmonotonic_delta_double_sum_unit")
	smThreeMetricEleven.SetEmptySum()
	smThreeMetricEleven.Sum().SetIsMonotonic(false)
	smThreeMetricEleven.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityDelta)
	smThreeMetricElevenDps := smThreeMetricEleven.Sum().DataPoints()
	smThreeMetricElevenDps.AppendEmpty().SetDoubleValue(3456.78)
	smThreeMetricElevenDps.At(0).Attributes().PutString("attribute_name_25", "attribute_value_25")
	smThreeMetricElevenDps.AppendEmpty().SetDoubleValue(4567.89)
	smThreeMetricElevenDps.At(1).Attributes().PutString("attribute_name_26", "attribute_value_26")

	smThreeMetricTwelve := smThreeMetrics.AppendEmpty()
	smThreeMetricTwelve.SetName("a_nonmonotonic_unspecified_double_sum")
	smThreeMetricTwelve.SetDescription("a_nonmonotonic_unspecified_double_sum_description")
	smThreeMetricTwelve.SetUnit("a_nonmonotonic_unspecified_double_sum_unit")
	smThreeMetricTwelve.SetEmptySum()
	smThreeMetricTwelve.Sum().SetIsMonotonic(false)
	smThreeMetricTwelve.Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityUnspecified)
	smThreeMetricTwelveDps := smThreeMetricTwelve.Sum().DataPoints()
	smThreeMetricTwelveDps.AppendEmpty().SetDoubleValue(5678.90)
	smThreeMetricTwelveDps.At(0).Attributes().PutString("attribute_name_27", "attribute_value_27")
	smThreeMetricTwelveDps.AppendEmpty().SetDoubleValue(6789.01)
	smThreeMetricTwelveDps.At(1).Attributes().PutString("attribute_name_28", "attribute_value_28")
	return metrics
}
