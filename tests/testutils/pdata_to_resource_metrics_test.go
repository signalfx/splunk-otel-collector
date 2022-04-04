package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
)

func pdataMetrics() pdata.Metrics {
	metrics := pdata.NewMetrics()
	resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
	attrs := resourceMetrics.Resource().Attributes()
	attrs.InsertBool("bool", true)
	attrs.InsertString("string", "a_string")
	attrs.InsertInt("int", 123)
	attrs.InsertDouble("double", 123.45)
	attrs.InsertNull("null")

	ilms := resourceMetrics.ScopeMetrics()
	ilmOne := ilms.AppendEmpty()
	ilmOne.InstrumentationLibrary().SetName("an_instrumentation_library_name")
	ilmOne.InstrumentationLibrary().SetVersion("an_instrumentation_library_version")
	ilmOneMetrics := ilmOne.Metrics()
	ilmOneMetricOne := ilmOneMetrics.AppendEmpty()
	ilmOneMetricOne.SetName("an_int_gauge")
	ilmOneMetricOne.SetDescription("an_int_gauge_description")
	ilmOneMetricOne.SetUnit("an_int_gauge_unit")
	ilmOneMetricOne.SetDataType(pdata.MetricDataTypeGauge)
	ilmOneMetricOneDps := ilmOneMetricOne.Gauge().DataPoints()
	ilmOneMetricOneDps.AppendEmpty().SetIntVal(12345)
	ilmOneMetricOneDps.At(0).Attributes().Insert("label_name_1", pdata.NewValueString("label_value_1"))
	ilmOneMetricOneDps.AppendEmpty().SetIntVal(23456)
	ilmOneMetricOneDps.At(1).Attributes().Insert("label_name_2", pdata.NewValueString("label_value_2"))

	ilmOneMetricTwo := ilmOneMetrics.AppendEmpty()
	ilmOneMetricTwo.SetName("a_double_gauge")
	ilmOneMetricTwo.SetDescription("a_double_gauge_description")
	ilmOneMetricTwo.SetUnit("a_double_gauge_unit")
	ilmOneMetricTwo.SetDataType(pdata.MetricDataTypeGauge)
	ilmOneMetricTwoDps := ilmOneMetricTwo.Gauge().DataPoints()
	ilmOneMetricTwoDps.AppendEmpty().SetDoubleVal(234.56)
	ilmOneMetricTwoDps.At(0).Attributes().Insert("label_name_3", pdata.NewValueString("label_value_3"))
	ilmOneMetricTwoDps.AppendEmpty().SetDoubleVal(345.67)
	ilmOneMetricTwoDps.At(1).Attributes().Insert("label_name_4", pdata.NewValueString("label_value_4"))

	ilms.AppendEmpty().InstrumentationLibrary().SetName("an_instrumentation_library_without_version_or_metrics")

	ilmThreeMetrics := ilms.AppendEmpty().Metrics()

	ilmThreeMetricOne := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricOne.SetName("a_monotonic_cumulative_int_sum")
	ilmThreeMetricOne.SetDescription("a_monotonic_cumulative_int_sum_description")
	ilmThreeMetricOne.SetUnit("a_monotonic_cumulative_int_sum_unit")
	ilmThreeMetricOne.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricOne.Sum().SetIsMonotonic(true)
	ilmThreeMetricOne.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
	ilmThreeMetricOneDps := ilmThreeMetricOne.Sum().DataPoints()
	ilmThreeMetricOneDps.AppendEmpty().SetIntVal(34567)
	ilmThreeMetricOneDps.At(0).Attributes().Insert("label_name_5", pdata.NewValueString("label_value_5"))
	ilmThreeMetricOneDps.AppendEmpty().SetIntVal(45678)
	ilmThreeMetricOneDps.At(1).Attributes().Insert("label_name_6", pdata.NewValueString("label_value_6"))

	ilmThreeMetricTwo := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricTwo.SetName("a_monotonic_delta_int_sum")
	ilmThreeMetricTwo.SetDescription("a_monotonic_delta_int_sum_description")
	ilmThreeMetricTwo.SetUnit("a_monotonic_delta_int_sum_unit")
	ilmThreeMetricTwo.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricTwo.Sum().SetIsMonotonic(true)
	ilmThreeMetricTwo.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityDelta)
	ilmThreeMetricTwoDps := ilmThreeMetricTwo.Sum().DataPoints()
	ilmThreeMetricTwoDps.AppendEmpty().SetIntVal(56789)
	ilmThreeMetricTwoDps.At(0).Attributes().Insert("label_name_7", pdata.NewValueString("label_value_7"))
	ilmThreeMetricTwoDps.AppendEmpty().SetIntVal(67890)
	ilmThreeMetricTwoDps.At(1).Attributes().Insert("label_name_8", pdata.NewValueString("label_value_8"))

	ilmThreeMetricThree := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricThree.SetName("a_monotonic_unspecified_int_sum")
	ilmThreeMetricThree.SetDescription("a_monotonic_unspecified_int_sum_description")
	ilmThreeMetricThree.SetUnit("a_monotonic_unspecified_int_sum_unit")
	ilmThreeMetricThree.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricThree.Sum().SetIsMonotonic(true)
	ilmThreeMetricThree.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityUnspecified)
	ilmThreeMetricThreeDps := ilmThreeMetricThree.Sum().DataPoints()
	ilmThreeMetricThreeDps.AppendEmpty().SetIntVal(78901)
	ilmThreeMetricThreeDps.At(0).Attributes().Insert("label_name_9", pdata.NewValueString("label_value_9"))
	ilmThreeMetricThreeDps.AppendEmpty().SetIntVal(89012)
	ilmThreeMetricThreeDps.At(1).Attributes().Insert("label_name_10", pdata.NewValueString("label_value_10"))

	ilmThreeMetricFour := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricFour.SetName("a_monotonic_cumulative_double_sum")
	ilmThreeMetricFour.SetDescription("a_monotonic_cumulative_double_sum_description")
	ilmThreeMetricFour.SetUnit("a_monotonic_cumulative_double_sum_unit")
	ilmThreeMetricFour.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricFour.Sum().SetIsMonotonic(true)
	ilmThreeMetricFour.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
	ilmThreeMetricFourDps := ilmThreeMetricFour.Sum().DataPoints()
	ilmThreeMetricFourDps.AppendEmpty().SetDoubleVal(456.78)
	ilmThreeMetricFourDps.At(0).Attributes().Insert("label_name_11", pdata.NewValueString("label_value_11"))
	ilmThreeMetricFourDps.AppendEmpty().SetDoubleVal(567.89)
	ilmThreeMetricFourDps.At(1).Attributes().Insert("label_name_12", pdata.NewValueString("label_value_12"))

	ilmThreeMetricFive := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricFive.SetName("a_monotonic_delta_double_sum")
	ilmThreeMetricFive.SetDescription("a_monotonic_delta_double_sum_description")
	ilmThreeMetricFive.SetUnit("a_monotonic_delta_double_sum_unit")
	ilmThreeMetricFive.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricFive.Sum().SetIsMonotonic(true)
	ilmThreeMetricFive.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityDelta)
	ilmThreeMetricFiveDps := ilmThreeMetricFive.Sum().DataPoints()
	ilmThreeMetricFiveDps.AppendEmpty().SetDoubleVal(678.90)
	ilmThreeMetricFiveDps.At(0).Attributes().Insert("label_name_13", pdata.NewValueString("label_value_13"))
	ilmThreeMetricFiveDps.AppendEmpty().SetDoubleVal(789.01)
	ilmThreeMetricFiveDps.At(1).Attributes().Insert("label_name_14", pdata.NewValueString("label_value_14"))

	ilmThreeMetricSix := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricSix.SetName("a_monotonic_unspecified_double_sum")
	ilmThreeMetricSix.SetDescription("a_monotonic_unspecified_double_sum_description")
	ilmThreeMetricSix.SetUnit("a_monotonic_unspecified_double_sum_unit")
	ilmThreeMetricSix.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricSix.Sum().SetIsMonotonic(true)
	ilmThreeMetricSix.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityUnspecified)
	ilmThreeMetricSixDps := ilmThreeMetricSix.Sum().DataPoints()
	ilmThreeMetricSixDps.AppendEmpty().SetDoubleVal(890.12)
	ilmThreeMetricSixDps.At(0).Attributes().Insert("label_name_15", pdata.NewValueString("label_value_15"))
	ilmThreeMetricSixDps.AppendEmpty().SetDoubleVal(901.23)
	ilmThreeMetricSixDps.At(1).Attributes().Insert("label_name_16", pdata.NewValueString("label_value_16"))

	ilmThreeMetricSeven := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricSeven.SetName("a_nonmonotonic_cumulative_int_sum")
	ilmThreeMetricSeven.SetDescription("a_nonmonotonic_cumulative_int_sum_description")
	ilmThreeMetricSeven.SetUnit("a_nonmonotonic_cumulative_int_sum_unit")
	ilmThreeMetricSeven.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricSeven.Sum().SetIsMonotonic(false)
	ilmThreeMetricSeven.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
	ilmThreeMetricSevenDps := ilmThreeMetricSeven.Sum().DataPoints()
	ilmThreeMetricSevenDps.AppendEmpty().SetIntVal(90123)
	ilmThreeMetricSevenDps.At(0).Attributes().Insert("label_name_17", pdata.NewValueString("label_value_17"))
	ilmThreeMetricSevenDps.AppendEmpty().SetIntVal(123456)
	ilmThreeMetricSevenDps.At(1).Attributes().Insert("label_name_18", pdata.NewValueString("label_value_18"))

	ilmThreeMetricEight := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricEight.SetName("a_nonmonotonic_delta_int_sum")
	ilmThreeMetricEight.SetDescription("a_nonmonotonic_delta_int_sum_description")
	ilmThreeMetricEight.SetUnit("a_nonmonotonic_delta_int_sum_unit")
	ilmThreeMetricEight.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricEight.Sum().SetIsMonotonic(false)
	ilmThreeMetricEight.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityDelta)
	ilmThreeMetricEightDps := ilmThreeMetricEight.Sum().DataPoints()
	ilmThreeMetricEightDps.AppendEmpty().SetIntVal(234567)
	ilmThreeMetricEightDps.At(0).Attributes().Insert("label_name_19", pdata.NewValueString("label_value_19"))
	ilmThreeMetricEightDps.AppendEmpty().SetIntVal(345678)
	ilmThreeMetricEightDps.At(1).Attributes().Insert("label_name_20", pdata.NewValueString("label_value_20"))

	ilmThreeMetricNine := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricNine.SetName("a_nonmonotonic_unspecified_int_sum")
	ilmThreeMetricNine.SetDescription("a_nonmonotonic_unspecified_int_sum_description")
	ilmThreeMetricNine.SetUnit("a_nonmonotonic_unspecified_int_sum_unit")
	ilmThreeMetricNine.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricNine.Sum().SetIsMonotonic(false)
	ilmThreeMetricNine.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityUnspecified)
	ilmThreeMetricNineDps := ilmThreeMetricNine.Sum().DataPoints()
	ilmThreeMetricNineDps.AppendEmpty().SetIntVal(456789)
	ilmThreeMetricNineDps.At(0).Attributes().Insert("label_name_21", pdata.NewValueString("label_value_21"))
	ilmThreeMetricNineDps.AppendEmpty().SetIntVal(567890)
	ilmThreeMetricNineDps.At(1).Attributes().Insert("label_name_22", pdata.NewValueString("label_value_22"))

	ilmThreeMetricTen := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricTen.SetName("a_nonmonotonic_cumulative_double_sum")
	ilmThreeMetricTen.SetDescription("a_nonmonotonic_cumulative_double_sum_description")
	ilmThreeMetricTen.SetUnit("a_nonmonotonic_cumulative_double_sum_unit")
	ilmThreeMetricTen.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricTen.Sum().SetIsMonotonic(false)
	ilmThreeMetricTen.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
	ilmThreeMetricTenDps := ilmThreeMetricTen.Sum().DataPoints()
	ilmThreeMetricTenDps.AppendEmpty().SetDoubleVal(1234.56)
	ilmThreeMetricTenDps.At(0).Attributes().Insert("label_name_23", pdata.NewValueString("label_value_23"))
	ilmThreeMetricTenDps.AppendEmpty().SetDoubleVal(2345.67)
	ilmThreeMetricTenDps.At(1).Attributes().Insert("label_name_24", pdata.NewValueString("label_value_24"))

	ilmThreeMetricEleven := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricEleven.SetName("a_nonmonotonic_delta_double_sum")
	ilmThreeMetricEleven.SetDescription("a_nonmonotonic_delta_double_sum_description")
	ilmThreeMetricEleven.SetUnit("a_nonmonotonic_delta_double_sum_unit")
	ilmThreeMetricEleven.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricEleven.Sum().SetIsMonotonic(false)
	ilmThreeMetricEleven.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityDelta)
	ilmThreeMetricElevenDps := ilmThreeMetricEleven.Sum().DataPoints()
	ilmThreeMetricElevenDps.AppendEmpty().SetDoubleVal(3456.78)
	ilmThreeMetricElevenDps.At(0).Attributes().Insert("label_name_25", pdata.NewValueString("label_value_25"))
	ilmThreeMetricElevenDps.AppendEmpty().SetDoubleVal(4567.89)
	ilmThreeMetricElevenDps.At(1).Attributes().Insert("label_name_26", pdata.NewValueString("label_value_26"))

	ilmThreeMetricTwelve := ilmThreeMetrics.AppendEmpty()
	ilmThreeMetricTwelve.SetName("a_nonmonotonic_unspecified_double_sum")
	ilmThreeMetricTwelve.SetDescription("a_nonmonotonic_unspecified_double_sum_description")
	ilmThreeMetricTwelve.SetUnit("a_nonmonotonic_unspecified_double_sum_unit")
	ilmThreeMetricTwelve.SetDataType(pdata.MetricDataTypeSum)
	ilmThreeMetricTwelve.Sum().SetIsMonotonic(false)
	ilmThreeMetricTwelve.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityUnspecified)
	ilmThreeMetricTwelveDps := ilmThreeMetricTwelve.Sum().DataPoints()
	ilmThreeMetricTwelveDps.AppendEmpty().SetDoubleVal(5678.90)
	ilmThreeMetricTwelveDps.At(0).Attributes().Insert("label_name_27", pdata.NewValueString("label_value_27"))
	ilmThreeMetricTwelveDps.AppendEmpty().SetDoubleVal(6789.01)
	ilmThreeMetricTwelveDps.At(1).Attributes().Insert("label_name_28", pdata.NewValueString("label_value_28"))
	return metrics
}

func TestPDataToResourceMetricsHappyPath(t *testing.T) {
	resourceMetrics, err := PDataToResourceMetrics(pdataMetrics())
	assert.NoError(t, err)
	require.NotNil(t, resourceMetrics)

	rms := resourceMetrics.ResourceMetrics
	assert.Len(t, rms, 1)
	rm := rms[0]
	attrs := rm.Resource.Attributes
	assert.True(t, attrs["bool"].(bool))
	assert.Equal(t, "a_string", attrs["string"].(string))
	assert.Equal(t, 123, int(attrs["int"].(int64)))
	assert.Equal(t, 123.45, attrs["double"].(float64))
	assert.Nil(t, attrs["null"])

	ilms := rm.ILMs
	assert.Len(t, ilms, 3)
	assert.Equal(t, "an_instrumentation_library_name", ilms[0].InstrumentationLibrary.Name)
	assert.Equal(t, "an_instrumentation_library_version", ilms[0].InstrumentationLibrary.Version)

	require.Len(t, ilms[0].Metrics, 4)

	ilmOneMetricOne := ilms[0].Metrics[0]
	assert.Equal(t, "an_int_gauge", ilmOneMetricOne.Name)
	assert.Equal(t, "an_int_gauge_description", ilmOneMetricOne.Description)
	assert.Equal(t, "an_int_gauge_unit", ilmOneMetricOne.Unit)
	assert.Equal(t, IntGauge, ilmOneMetricOne.Type)
	assert.Equal(t, map[string]string{"label_name_1": "label_value_1"}, *ilmOneMetricOne.Labels)
	assert.EqualValues(t, 12345, ilmOneMetricOne.Value)

	ilmOneMetricTwo := ilms[0].Metrics[1]
	assert.Equal(t, "an_int_gauge", ilmOneMetricTwo.Name)
	assert.Equal(t, "an_int_gauge_description", ilmOneMetricTwo.Description)
	assert.Equal(t, "an_int_gauge_unit", ilmOneMetricTwo.Unit)
	assert.Equal(t, IntGauge, ilmOneMetricTwo.Type)
	assert.Equal(t, map[string]string{"label_name_2": "label_value_2"}, *ilmOneMetricTwo.Labels)
	assert.EqualValues(t, 23456, ilmOneMetricTwo.Value)

	ilmOneMetricThree := ilms[0].Metrics[2]
	assert.Equal(t, "a_double_gauge", ilmOneMetricThree.Name)
	assert.Equal(t, "a_double_gauge_description", ilmOneMetricThree.Description)
	assert.Equal(t, "a_double_gauge_unit", ilmOneMetricThree.Unit)
	assert.Equal(t, DoubleGauge, ilmOneMetricThree.Type)
	assert.Equal(t, map[string]string{"label_name_3": "label_value_3"}, *ilmOneMetricThree.Labels)
	assert.EqualValues(t, 234.56, ilmOneMetricThree.Value)

	ilmOneMetricFour := ilms[0].Metrics[3]
	assert.Equal(t, "a_double_gauge", ilmOneMetricFour.Name)
	assert.Equal(t, "a_double_gauge_description", ilmOneMetricFour.Description)
	assert.Equal(t, "a_double_gauge_unit", ilmOneMetricFour.Unit)
	assert.Equal(t, DoubleGauge, ilmOneMetricFour.Type)
	assert.Equal(t, map[string]string{"label_name_4": "label_value_4"}, *ilmOneMetricFour.Labels)
	assert.EqualValues(t, 345.67, ilmOneMetricFour.Value)

	assert.Equal(t, "an_instrumentation_library_without_version_or_metrics", ilms[1].InstrumentationLibrary.Name)
	assert.Empty(t, ilms[1].InstrumentationLibrary.Version)
	assert.Empty(t, ilms[1].Metrics)

	require.Len(t, ilms[2].Metrics, 24)

	ilmThreeMetricOne := ilms[2].Metrics[0]
	assert.Equal(t, "a_monotonic_cumulative_int_sum", ilmThreeMetricOne.Name)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_description", ilmThreeMetricOne.Description)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_unit", ilmThreeMetricOne.Unit)
	assert.Equal(t, IntMonotonicCumulativeSum, ilmThreeMetricOne.Type)
	assert.Equal(t, map[string]string{"label_name_5": "label_value_5"}, *ilmThreeMetricOne.Labels)
	assert.EqualValues(t, 34567, ilmThreeMetricOne.Value)

	ilmThreeMetricTwo := ilms[2].Metrics[1]
	assert.Equal(t, "a_monotonic_cumulative_int_sum", ilmThreeMetricTwo.Name)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_description", ilmThreeMetricTwo.Description)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_unit", ilmThreeMetricTwo.Unit)
	assert.Equal(t, IntMonotonicCumulativeSum, ilmThreeMetricTwo.Type)
	assert.Equal(t, map[string]string{"label_name_6": "label_value_6"}, *ilmThreeMetricTwo.Labels)
	assert.EqualValues(t, 45678, ilmThreeMetricTwo.Value)

	ilmThreeMetricThree := ilms[2].Metrics[2]
	assert.Equal(t, "a_monotonic_delta_int_sum", ilmThreeMetricThree.Name)
	assert.Equal(t, "a_monotonic_delta_int_sum_description", ilmThreeMetricThree.Description)
	assert.Equal(t, "a_monotonic_delta_int_sum_unit", ilmThreeMetricThree.Unit)
	assert.Equal(t, IntMonotonicDeltaSum, ilmThreeMetricThree.Type)
	assert.Equal(t, map[string]string{"label_name_7": "label_value_7"}, *ilmThreeMetricThree.Labels)
	assert.EqualValues(t, 56789, ilmThreeMetricThree.Value)

	ilmThreeMetricFour := ilms[2].Metrics[3]
	assert.Equal(t, "a_monotonic_delta_int_sum", ilmThreeMetricFour.Name)
	assert.Equal(t, "a_monotonic_delta_int_sum_description", ilmThreeMetricFour.Description)
	assert.Equal(t, "a_monotonic_delta_int_sum_unit", ilmThreeMetricFour.Unit)
	assert.Equal(t, IntMonotonicDeltaSum, ilmThreeMetricFour.Type)
	assert.Equal(t, map[string]string{"label_name_8": "label_value_8"}, *ilmThreeMetricFour.Labels)
	assert.EqualValues(t, 67890, ilmThreeMetricFour.Value)

	ilmThreeMetricFive := ilms[2].Metrics[4]
	assert.Equal(t, "a_monotonic_unspecified_int_sum", ilmThreeMetricFive.Name)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_description", ilmThreeMetricFive.Description)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_unit", ilmThreeMetricFive.Unit)
	assert.Equal(t, IntMonotonicUnspecifiedSum, ilmThreeMetricFive.Type)
	assert.Equal(t, map[string]string{"label_name_9": "label_value_9"}, *ilmThreeMetricFive.Labels)
	assert.EqualValues(t, 78901, ilmThreeMetricFive.Value)

	ilmThreeMetricSix := ilms[2].Metrics[5]
	assert.Equal(t, "a_monotonic_unspecified_int_sum", ilmThreeMetricSix.Name)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_description", ilmThreeMetricSix.Description)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_unit", ilmThreeMetricSix.Unit)
	assert.Equal(t, IntMonotonicUnspecifiedSum, ilmThreeMetricSix.Type)
	assert.Equal(t, map[string]string{"label_name_10": "label_value_10"}, *ilmThreeMetricSix.Labels)
	assert.EqualValues(t, 89012, ilmThreeMetricSix.Value)

	ilmThreeMetricSeven := ilms[2].Metrics[6]
	assert.Equal(t, "a_monotonic_cumulative_double_sum", ilmThreeMetricSeven.Name)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_description", ilmThreeMetricSeven.Description)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_unit", ilmThreeMetricSeven.Unit)
	assert.Equal(t, DoubleMonotonicCumulativeSum, ilmThreeMetricSeven.Type)
	assert.Equal(t, map[string]string{"label_name_11": "label_value_11"}, *ilmThreeMetricSeven.Labels)
	assert.EqualValues(t, 456.78, ilmThreeMetricSeven.Value)

	ilmThreeMetricEight := ilms[2].Metrics[7]
	assert.Equal(t, "a_monotonic_cumulative_double_sum", ilmThreeMetricEight.Name)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_description", ilmThreeMetricEight.Description)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_unit", ilmThreeMetricEight.Unit)
	assert.Equal(t, DoubleMonotonicCumulativeSum, ilmThreeMetricEight.Type)
	assert.Equal(t, map[string]string{"label_name_12": "label_value_12"}, *ilmThreeMetricEight.Labels)
	assert.EqualValues(t, 567.89, ilmThreeMetricEight.Value)

	ilmThreeMetricNine := ilms[2].Metrics[8]
	assert.Equal(t, "a_monotonic_delta_double_sum", ilmThreeMetricNine.Name)
	assert.Equal(t, "a_monotonic_delta_double_sum_description", ilmThreeMetricNine.Description)
	assert.Equal(t, "a_monotonic_delta_double_sum_unit", ilmThreeMetricNine.Unit)
	assert.Equal(t, DoubleMonotonicDeltaSum, ilmThreeMetricNine.Type)
	assert.Equal(t, map[string]string{"label_name_13": "label_value_13"}, *ilmThreeMetricNine.Labels)
	assert.EqualValues(t, 678.90, ilmThreeMetricNine.Value)

	ilmThreeMetricTen := ilms[2].Metrics[9]
	assert.Equal(t, "a_monotonic_delta_double_sum", ilmThreeMetricTen.Name)
	assert.Equal(t, "a_monotonic_delta_double_sum_description", ilmThreeMetricTen.Description)
	assert.Equal(t, "a_monotonic_delta_double_sum_unit", ilmThreeMetricTen.Unit)
	assert.Equal(t, DoubleMonotonicDeltaSum, ilmThreeMetricTen.Type)
	assert.Equal(t, map[string]string{"label_name_14": "label_value_14"}, *ilmThreeMetricTen.Labels)
	assert.EqualValues(t, 789.01, ilmThreeMetricTen.Value)

	ilmThreeMetricEleven := ilms[2].Metrics[10]
	assert.Equal(t, "a_monotonic_unspecified_double_sum", ilmThreeMetricEleven.Name)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_description", ilmThreeMetricEleven.Description)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_unit", ilmThreeMetricEleven.Unit)
	assert.Equal(t, DoubleMonotonicUnspecifiedSum, ilmThreeMetricEleven.Type)
	assert.Equal(t, map[string]string{"label_name_15": "label_value_15"}, *ilmThreeMetricEleven.Labels)
	assert.EqualValues(t, 890.12, ilmThreeMetricEleven.Value)

	ilmThreeMetricTwelve := ilms[2].Metrics[11]
	assert.Equal(t, "a_monotonic_unspecified_double_sum", ilmThreeMetricTwelve.Name)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_description", ilmThreeMetricTwelve.Description)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_unit", ilmThreeMetricTwelve.Unit)
	assert.Equal(t, DoubleMonotonicUnspecifiedSum, ilmThreeMetricTwelve.Type)
	assert.Equal(t, map[string]string{"label_name_16": "label_value_16"}, *ilmThreeMetricTwelve.Labels)
	assert.EqualValues(t, 901.23, ilmThreeMetricTwelve.Value)

	ilmThreeMetricThirteen := ilms[2].Metrics[12]
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum", ilmThreeMetricThirteen.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_description", ilmThreeMetricThirteen.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_unit", ilmThreeMetricThirteen.Unit)
	assert.Equal(t, IntNonmonotonicCumulativeSum, ilmThreeMetricThirteen.Type)
	assert.Equal(t, map[string]string{"label_name_17": "label_value_17"}, *ilmThreeMetricThirteen.Labels)
	assert.EqualValues(t, 90123, ilmThreeMetricThirteen.Value)

	ilmThreeMetricFourteen := ilms[2].Metrics[13]
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum", ilmThreeMetricFourteen.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_description", ilmThreeMetricFourteen.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_unit", ilmThreeMetricFourteen.Unit)
	assert.Equal(t, IntNonmonotonicCumulativeSum, ilmThreeMetricFourteen.Type)
	assert.Equal(t, map[string]string{"label_name_18": "label_value_18"}, *ilmThreeMetricFourteen.Labels)
	assert.EqualValues(t, 123456, ilmThreeMetricFourteen.Value)

	ilmThreeMetricFifteen := ilms[2].Metrics[14]
	assert.Equal(t, "a_nonmonotonic_delta_int_sum", ilmThreeMetricFifteen.Name)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_description", ilmThreeMetricFifteen.Description)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_unit", ilmThreeMetricFifteen.Unit)
	assert.Equal(t, IntNonmonotonicDeltaSum, ilmThreeMetricFifteen.Type)
	assert.Equal(t, map[string]string{"label_name_19": "label_value_19"}, *ilmThreeMetricFifteen.Labels)
	assert.EqualValues(t, 234567, ilmThreeMetricFifteen.Value)

	ilmThreeMetricSixteen := ilms[2].Metrics[15]
	assert.Equal(t, "a_nonmonotonic_delta_int_sum", ilmThreeMetricSixteen.Name)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_description", ilmThreeMetricSixteen.Description)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_unit", ilmThreeMetricSixteen.Unit)
	assert.Equal(t, IntNonmonotonicDeltaSum, ilmThreeMetricSixteen.Type)
	assert.Equal(t, map[string]string{"label_name_20": "label_value_20"}, *ilmThreeMetricSixteen.Labels)
	assert.EqualValues(t, 345678, ilmThreeMetricSixteen.Value)

	ilmThreeMetricSeventeen := ilms[2].Metrics[16]
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum", ilmThreeMetricSeventeen.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_description", ilmThreeMetricSeventeen.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_unit", ilmThreeMetricSeventeen.Unit)
	assert.Equal(t, IntNonmonotonicUnspecifiedSum, ilmThreeMetricSeventeen.Type)
	assert.Equal(t, map[string]string{"label_name_21": "label_value_21"}, *ilmThreeMetricSeventeen.Labels)
	assert.EqualValues(t, 456789, ilmThreeMetricSeventeen.Value)

	ilmThreeMetricEighteen := ilms[2].Metrics[17]
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum", ilmThreeMetricEighteen.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_description", ilmThreeMetricEighteen.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_unit", ilmThreeMetricEighteen.Unit)
	assert.Equal(t, IntNonmonotonicUnspecifiedSum, ilmThreeMetricEighteen.Type)
	assert.Equal(t, map[string]string{"label_name_22": "label_value_22"}, *ilmThreeMetricEighteen.Labels)
	assert.EqualValues(t, 567890, ilmThreeMetricEighteen.Value)

	ilmThreeMetricNineteen := ilms[2].Metrics[18]
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum", ilmThreeMetricNineteen.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_description", ilmThreeMetricNineteen.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_unit", ilmThreeMetricNineteen.Unit)
	assert.Equal(t, DoubleNonmonotonicCumulativeSum, ilmThreeMetricNineteen.Type)
	assert.Equal(t, map[string]string{"label_name_23": "label_value_23"}, *ilmThreeMetricNineteen.Labels)
	assert.EqualValues(t, 1234.56, ilmThreeMetricNineteen.Value)

	ilmThreeMetricTwenty := ilms[2].Metrics[19]
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum", ilmThreeMetricTwenty.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_description", ilmThreeMetricTwenty.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_unit", ilmThreeMetricTwenty.Unit)
	assert.Equal(t, DoubleNonmonotonicCumulativeSum, ilmThreeMetricTwenty.Type)
	assert.Equal(t, map[string]string{"label_name_24": "label_value_24"}, *ilmThreeMetricTwenty.Labels)
	assert.EqualValues(t, 2345.67, ilmThreeMetricTwenty.Value)

	ilmThreeMetricTwentyOne := ilms[2].Metrics[20]
	assert.Equal(t, "a_nonmonotonic_delta_double_sum", ilmThreeMetricTwentyOne.Name)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_description", ilmThreeMetricTwentyOne.Description)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_unit", ilmThreeMetricTwentyOne.Unit)
	assert.Equal(t, DoubleNonmonotonicDeltaSum, ilmThreeMetricTwentyOne.Type)
	assert.Equal(t, map[string]string{"label_name_25": "label_value_25"}, *ilmThreeMetricTwentyOne.Labels)
	assert.EqualValues(t, 3456.78, ilmThreeMetricTwentyOne.Value)

	ilmThreeMetricTwentyTwo := ilms[2].Metrics[21]
	assert.Equal(t, "a_nonmonotonic_delta_double_sum", ilmThreeMetricTwentyTwo.Name)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_description", ilmThreeMetricTwentyTwo.Description)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_unit", ilmThreeMetricTwentyTwo.Unit)
	assert.Equal(t, DoubleNonmonotonicDeltaSum, ilmThreeMetricTwentyTwo.Type)
	assert.Equal(t, map[string]string{"label_name_26": "label_value_26"}, *ilmThreeMetricTwentyTwo.Labels)
	assert.EqualValues(t, 4567.89, ilmThreeMetricTwentyTwo.Value)

	ilmThreeMetricTwentyThree := ilms[2].Metrics[22]
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum", ilmThreeMetricTwentyThree.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_description", ilmThreeMetricTwentyThree.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_unit", ilmThreeMetricTwentyThree.Unit)
	assert.Equal(t, DoubleNonmonotonicUnspecifiedSum, ilmThreeMetricTwentyThree.Type)
	assert.Equal(t, map[string]string{"label_name_27": "label_value_27"}, *ilmThreeMetricTwentyThree.Labels)
	assert.EqualValues(t, 5678.90, ilmThreeMetricTwentyThree.Value)

	ilmThreeMetricTwentyFour := ilms[2].Metrics[23]
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum", ilmThreeMetricTwentyFour.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_description", ilmThreeMetricTwentyFour.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_unit", ilmThreeMetricTwentyFour.Unit)
	assert.Equal(t, DoubleNonmonotonicUnspecifiedSum, ilmThreeMetricTwentyFour.Type)
	assert.Equal(t, map[string]string{"label_name_28": "label_value_28"}, *ilmThreeMetricTwentyFour.Labels)
	assert.EqualValues(t, 6789.01, ilmThreeMetricTwentyFour.Value)
}
