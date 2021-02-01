package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/pdata"
)

func pdataMetrics() pdata.Metrics {
	metrics := pdata.NewMetrics()
	metrics.ResourceMetrics().Resize(1)
	resourceMetrics := metrics.ResourceMetrics().At(0)
	attrs := resourceMetrics.Resource().Attributes()
	attrs.InsertBool("bool", true)
	attrs.InsertString("string", "a_string")
	attrs.InsertInt("int", 123)
	attrs.InsertDouble("double", 123.45)
	attrs.InsertNull("null")

	ilms := resourceMetrics.InstrumentationLibraryMetrics()
	ilms.Resize(3)

	ilms.At(0).InstrumentationLibrary().SetName("an_instrumentation_library_name")
	ilms.At(0).InstrumentationLibrary().SetVersion("an_instrumentation_library_version")
	ilmOneMetrics := ilms.At(0).Metrics()
	ilmOneMetrics.Resize(2)
	ilmOneMetricOne := ilmOneMetrics.At(0)
	ilmOneMetricOne.SetName("an_int_gauge")
	ilmOneMetricOne.SetDescription("an_int_gauge_description")
	ilmOneMetricOne.SetUnit("an_int_gauge_unit")
	ilmOneMetricOne.SetDataType(pdata.MetricDataTypeIntGauge)
	ilmOneMetricOneDps := ilmOneMetricOne.IntGauge().DataPoints()
	ilmOneMetricOneDps.Resize(2)
	ilmOneMetricOneDps.At(0).SetValue(12345)
	ilmOneMetricOneDps.At(0).LabelsMap().Insert("label_name_1", "label_value_1")
	ilmOneMetricOneDps.At(1).SetValue(23456)
	ilmOneMetricOneDps.At(1).LabelsMap().Insert("label_name_2", "label_value_2")

	ilmOneMetricTwo := ilmOneMetrics.At(1)
	ilmOneMetricTwo.SetName("a_double_gauge")
	ilmOneMetricTwo.SetDescription("a_double_gauge_description")
	ilmOneMetricTwo.SetUnit("a_double_gauge_unit")
	ilmOneMetricTwo.SetDataType(pdata.MetricDataTypeDoubleGauge)
	ilmOneMetricTwoDps := ilmOneMetricTwo.DoubleGauge().DataPoints()
	ilmOneMetricTwoDps.Resize(2)
	ilmOneMetricTwoDps.At(0).SetValue(234.56)
	ilmOneMetricTwoDps.At(0).LabelsMap().Insert("label_name_3", "label_value_3")
	ilmOneMetricTwoDps.At(1).SetValue(345.67)
	ilmOneMetricTwoDps.At(1).LabelsMap().Insert("label_name_4", "label_value_4")

	ilms.At(1).InstrumentationLibrary().SetName("an_instrumentation_library_without_version_or_metrics")

	ilmThreeMetrics := ilms.At(2).Metrics()
	ilmThreeMetrics.Resize(12)

	ilmThreeMetricOne := ilmThreeMetrics.At(0)
	ilmThreeMetricOne.SetName("a_monotonic_cumulative_int_sum")
	ilmThreeMetricOne.SetDescription("a_monotonic_cumulative_int_sum_description")
	ilmThreeMetricOne.SetUnit("a_monotonic_cumulative_int_sum_unit")
	ilmThreeMetricOne.SetDataType(pdata.MetricDataTypeIntSum)
	ilmThreeMetricOne.IntSum().SetIsMonotonic(true)
	ilmThreeMetricOne.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	ilmThreeMetricOneDps := ilmThreeMetricOne.IntSum().DataPoints()
	ilmThreeMetricOneDps.Resize(2)
	ilmThreeMetricOneDps.At(0).SetValue(34567)
	ilmThreeMetricOneDps.At(0).LabelsMap().Insert("label_name_5", "label_value_5")
	ilmThreeMetricOneDps.At(1).SetValue(45678)
	ilmThreeMetricOneDps.At(1).LabelsMap().Insert("label_name_6", "label_value_6")

	ilmThreeMetricTwo := ilmThreeMetrics.At(1)
	ilmThreeMetricTwo.SetName("a_monotonic_delta_int_sum")
	ilmThreeMetricTwo.SetDescription("a_monotonic_delta_int_sum_description")
	ilmThreeMetricTwo.SetUnit("a_monotonic_delta_int_sum_unit")
	ilmThreeMetricTwo.SetDataType(pdata.MetricDataTypeIntSum)
	ilmThreeMetricTwo.IntSum().SetIsMonotonic(true)
	ilmThreeMetricTwo.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
	ilmThreeMetricTwoDps := ilmThreeMetricTwo.IntSum().DataPoints()
	ilmThreeMetricTwoDps.Resize(2)
	ilmThreeMetricTwoDps.At(0).SetValue(56789)
	ilmThreeMetricTwoDps.At(0).LabelsMap().Insert("label_name_7", "label_value_7")
	ilmThreeMetricTwoDps.At(1).SetValue(67890)
	ilmThreeMetricTwoDps.At(1).LabelsMap().Insert("label_name_8", "label_value_8")

	ilmThreeMetricThree := ilmThreeMetrics.At(2)
	ilmThreeMetricThree.SetName("a_monotonic_unspecified_int_sum")
	ilmThreeMetricThree.SetDescription("a_monotonic_unspecified_int_sum_description")
	ilmThreeMetricThree.SetUnit("a_monotonic_unspecified_int_sum_unit")
	ilmThreeMetricThree.SetDataType(pdata.MetricDataTypeIntSum)
	ilmThreeMetricThree.IntSum().SetIsMonotonic(true)
	ilmThreeMetricThree.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityUnspecified)
	ilmThreeMetricThreeDps := ilmThreeMetricThree.IntSum().DataPoints()
	ilmThreeMetricThreeDps.Resize(2)
	ilmThreeMetricThreeDps.At(0).SetValue(78901)
	ilmThreeMetricThreeDps.At(0).LabelsMap().Insert("label_name_9", "label_value_9")
	ilmThreeMetricThreeDps.At(1).SetValue(89012)
	ilmThreeMetricThreeDps.At(1).LabelsMap().Insert("label_name_10", "label_value_10")

	ilmThreeMetricFour := ilmThreeMetrics.At(3)
	ilmThreeMetricFour.SetName("a_monotonic_cumulative_double_sum")
	ilmThreeMetricFour.SetDescription("a_monotonic_cumulative_double_sum_description")
	ilmThreeMetricFour.SetUnit("a_monotonic_cumulative_double_sum_unit")
	ilmThreeMetricFour.SetDataType(pdata.MetricDataTypeDoubleSum)
	ilmThreeMetricFour.DoubleSum().SetIsMonotonic(true)
	ilmThreeMetricFour.DoubleSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	ilmThreeMetricFourDps := ilmThreeMetricFour.DoubleSum().DataPoints()
	ilmThreeMetricFourDps.Resize(2)
	ilmThreeMetricFourDps.At(0).SetValue(456.78)
	ilmThreeMetricFourDps.At(0).LabelsMap().Insert("label_name_11", "label_value_11")
	ilmThreeMetricFourDps.At(1).SetValue(567.89)
	ilmThreeMetricFourDps.At(1).LabelsMap().Insert("label_name_12", "label_value_12")

	ilmThreeMetricFive := ilmThreeMetrics.At(4)
	ilmThreeMetricFive.SetName("a_monotonic_delta_double_sum")
	ilmThreeMetricFive.SetDescription("a_monotonic_delta_double_sum_description")
	ilmThreeMetricFive.SetUnit("a_monotonic_delta_double_sum_unit")
	ilmThreeMetricFive.SetDataType(pdata.MetricDataTypeDoubleSum)
	ilmThreeMetricFive.DoubleSum().SetIsMonotonic(true)
	ilmThreeMetricFive.DoubleSum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
	ilmThreeMetricFiveDps := ilmThreeMetricFive.DoubleSum().DataPoints()
	ilmThreeMetricFiveDps.Resize(2)
	ilmThreeMetricFiveDps.At(0).SetValue(678.90)
	ilmThreeMetricFiveDps.At(0).LabelsMap().Insert("label_name_13", "label_value_13")
	ilmThreeMetricFiveDps.At(1).SetValue(789.01)
	ilmThreeMetricFiveDps.At(1).LabelsMap().Insert("label_name_14", "label_value_14")

	ilmThreeMetricSix := ilmThreeMetrics.At(5)
	ilmThreeMetricSix.SetName("a_monotonic_unspecified_double_sum")
	ilmThreeMetricSix.SetDescription("a_monotonic_unspecified_double_sum_description")
	ilmThreeMetricSix.SetUnit("a_monotonic_unspecified_double_sum_unit")
	ilmThreeMetricSix.SetDataType(pdata.MetricDataTypeDoubleSum)
	ilmThreeMetricSix.DoubleSum().SetIsMonotonic(true)
	ilmThreeMetricSix.DoubleSum().SetAggregationTemporality(pdata.AggregationTemporalityUnspecified)
	ilmThreeMetricSixDps := ilmThreeMetricSix.DoubleSum().DataPoints()
	ilmThreeMetricSixDps.Resize(2)
	ilmThreeMetricSixDps.At(0).SetValue(890.12)
	ilmThreeMetricSixDps.At(0).LabelsMap().Insert("label_name_15", "label_value_15")
	ilmThreeMetricSixDps.At(1).SetValue(901.23)
	ilmThreeMetricSixDps.At(1).LabelsMap().Insert("label_name_16", "label_value_16")

	ilmThreeMetricSeven := ilmThreeMetrics.At(6)
	ilmThreeMetricSeven.SetName("a_nonmonotonic_cumulative_int_sum")
	ilmThreeMetricSeven.SetDescription("a_nonmonotonic_cumulative_int_sum_description")
	ilmThreeMetricSeven.SetUnit("a_nonmonotonic_cumulative_int_sum_unit")
	ilmThreeMetricSeven.SetDataType(pdata.MetricDataTypeIntSum)
	ilmThreeMetricSeven.IntSum().SetIsMonotonic(false)
	ilmThreeMetricSeven.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	ilmThreeMetricSevenDps := ilmThreeMetricSeven.IntSum().DataPoints()
	ilmThreeMetricSevenDps.Resize(2)
	ilmThreeMetricSevenDps.At(0).SetValue(90123)
	ilmThreeMetricSevenDps.At(0).LabelsMap().Insert("label_name_17", "label_value_17")
	ilmThreeMetricSevenDps.At(1).SetValue(123456)
	ilmThreeMetricSevenDps.At(1).LabelsMap().Insert("label_name_18", "label_value_18")

	ilmThreeMetricEight := ilmThreeMetrics.At(7)
	ilmThreeMetricEight.SetName("a_nonmonotonic_delta_int_sum")
	ilmThreeMetricEight.SetDescription("a_nonmonotonic_delta_int_sum_description")
	ilmThreeMetricEight.SetUnit("a_nonmonotonic_delta_int_sum_unit")
	ilmThreeMetricEight.SetDataType(pdata.MetricDataTypeIntSum)
	ilmThreeMetricEight.IntSum().SetIsMonotonic(false)
	ilmThreeMetricEight.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
	ilmThreeMetricEightDps := ilmThreeMetricEight.IntSum().DataPoints()
	ilmThreeMetricEightDps.Resize(2)
	ilmThreeMetricEightDps.At(0).SetValue(234567)
	ilmThreeMetricEightDps.At(0).LabelsMap().Insert("label_name_19", "label_value_19")
	ilmThreeMetricEightDps.At(1).SetValue(345678)
	ilmThreeMetricEightDps.At(1).LabelsMap().Insert("label_name_20", "label_value_20")

	ilmThreeMetricNine := ilmThreeMetrics.At(8)
	ilmThreeMetricNine.SetName("a_nonmonotonic_unspecified_int_sum")
	ilmThreeMetricNine.SetDescription("a_nonmonotonic_unspecified_int_sum_description")
	ilmThreeMetricNine.SetUnit("a_nonmonotonic_unspecified_int_sum_unit")
	ilmThreeMetricNine.SetDataType(pdata.MetricDataTypeIntSum)
	ilmThreeMetricNine.IntSum().SetIsMonotonic(false)
	ilmThreeMetricNine.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityUnspecified)
	ilmThreeMetricNineDps := ilmThreeMetricNine.IntSum().DataPoints()
	ilmThreeMetricNineDps.Resize(2)
	ilmThreeMetricNineDps.At(0).SetValue(456789)
	ilmThreeMetricNineDps.At(0).LabelsMap().Insert("label_name_21", "label_value_21")
	ilmThreeMetricNineDps.At(1).SetValue(567890)
	ilmThreeMetricNineDps.At(1).LabelsMap().Insert("label_name_22", "label_value_22")

	ilmThreeMetricTen := ilmThreeMetrics.At(9)
	ilmThreeMetricTen.SetName("a_nonmonotonic_cumulative_double_sum")
	ilmThreeMetricTen.SetDescription("a_nonmonotonic_cumulative_double_sum_description")
	ilmThreeMetricTen.SetUnit("a_nonmonotonic_cumulative_double_sum_unit")
	ilmThreeMetricTen.SetDataType(pdata.MetricDataTypeDoubleSum)
	ilmThreeMetricTen.DoubleSum().SetIsMonotonic(false)
	ilmThreeMetricTen.DoubleSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	ilmThreeMetricTenDps := ilmThreeMetricTen.DoubleSum().DataPoints()
	ilmThreeMetricTenDps.Resize(2)
	ilmThreeMetricTenDps.At(0).SetValue(1234.56)
	ilmThreeMetricTenDps.At(0).LabelsMap().Insert("label_name_23", "label_value_23")
	ilmThreeMetricTenDps.At(1).SetValue(2345.67)
	ilmThreeMetricTenDps.At(1).LabelsMap().Insert("label_name_24", "label_value_24")

	ilmThreeMetricEleven := ilmThreeMetrics.At(10)
	ilmThreeMetricEleven.SetName("a_nonmonotonic_delta_double_sum")
	ilmThreeMetricEleven.SetDescription("a_nonmonotonic_delta_double_sum_description")
	ilmThreeMetricEleven.SetUnit("a_nonmonotonic_delta_double_sum_unit")
	ilmThreeMetricEleven.SetDataType(pdata.MetricDataTypeDoubleSum)
	ilmThreeMetricEleven.DoubleSum().SetIsMonotonic(false)
	ilmThreeMetricEleven.DoubleSum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
	ilmThreeMetricElevenDps := ilmThreeMetricEleven.DoubleSum().DataPoints()
	ilmThreeMetricElevenDps.Resize(2)
	ilmThreeMetricElevenDps.At(0).SetValue(3456.78)
	ilmThreeMetricElevenDps.At(0).LabelsMap().Insert("label_name_25", "label_value_25")
	ilmThreeMetricElevenDps.At(1).SetValue(4567.89)
	ilmThreeMetricElevenDps.At(1).LabelsMap().Insert("label_name_26", "label_value_26")

	ilmThreeMetricTwelve := ilmThreeMetrics.At(11)
	ilmThreeMetricTwelve.SetName("a_nonmonotonic_unspecified_double_sum")
	ilmThreeMetricTwelve.SetDescription("a_nonmonotonic_unspecified_double_sum_description")
	ilmThreeMetricTwelve.SetUnit("a_nonmonotonic_unspecified_double_sum_unit")
	ilmThreeMetricTwelve.SetDataType(pdata.MetricDataTypeDoubleSum)
	ilmThreeMetricTwelve.DoubleSum().SetIsMonotonic(false)
	ilmThreeMetricTwelve.DoubleSum().SetAggregationTemporality(pdata.AggregationTemporalityUnspecified)
	ilmThreeMetricTwelveDps := ilmThreeMetricTwelve.DoubleSum().DataPoints()
	ilmThreeMetricTwelveDps.Resize(2)
	ilmThreeMetricTwelveDps.At(0).SetValue(5678.90)
	ilmThreeMetricTwelveDps.At(0).LabelsMap().Insert("label_name_27", "label_value_27")
	ilmThreeMetricTwelveDps.At(1).SetValue(6789.01)
	ilmThreeMetricTwelveDps.At(1).LabelsMap().Insert("label_name_28", "label_value_28")
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
	assert.Equal(t, map[string]string{"label_name_1": "label_value_1"}, ilmOneMetricOne.Labels)
	assert.EqualValues(t, 12345, ilmOneMetricOne.Value)

	ilmOneMetricTwo := ilms[0].Metrics[1]
	assert.Equal(t, "an_int_gauge", ilmOneMetricTwo.Name)
	assert.Equal(t, "an_int_gauge_description", ilmOneMetricTwo.Description)
	assert.Equal(t, "an_int_gauge_unit", ilmOneMetricTwo.Unit)
	assert.Equal(t, IntGauge, ilmOneMetricTwo.Type)
	assert.Equal(t, map[string]string{"label_name_2": "label_value_2"}, ilmOneMetricTwo.Labels)
	assert.EqualValues(t, 23456, ilmOneMetricTwo.Value)

	ilmOneMetricThree := ilms[0].Metrics[2]
	assert.Equal(t, "a_double_gauge", ilmOneMetricThree.Name)
	assert.Equal(t, "a_double_gauge_description", ilmOneMetricThree.Description)
	assert.Equal(t, "a_double_gauge_unit", ilmOneMetricThree.Unit)
	assert.Equal(t, DoubleGauge, ilmOneMetricThree.Type)
	assert.Equal(t, map[string]string{"label_name_3": "label_value_3"}, ilmOneMetricThree.Labels)
	assert.EqualValues(t, 234.56, ilmOneMetricThree.Value)

	ilmOneMetricFour := ilms[0].Metrics[3]
	assert.Equal(t, "a_double_gauge", ilmOneMetricFour.Name)
	assert.Equal(t, "a_double_gauge_description", ilmOneMetricFour.Description)
	assert.Equal(t, "a_double_gauge_unit", ilmOneMetricFour.Unit)
	assert.Equal(t, DoubleGauge, ilmOneMetricFour.Type)
	assert.Equal(t, map[string]string{"label_name_4": "label_value_4"}, ilmOneMetricFour.Labels)
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
	assert.Equal(t, map[string]string{"label_name_5": "label_value_5"}, ilmThreeMetricOne.Labels)
	assert.EqualValues(t, 34567, ilmThreeMetricOne.Value)

	ilmThreeMetricTwo := ilms[2].Metrics[1]
	assert.Equal(t, "a_monotonic_cumulative_int_sum", ilmThreeMetricTwo.Name)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_description", ilmThreeMetricTwo.Description)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_unit", ilmThreeMetricTwo.Unit)
	assert.Equal(t, IntMonotonicCumulativeSum, ilmThreeMetricTwo.Type)
	assert.Equal(t, map[string]string{"label_name_6": "label_value_6"}, ilmThreeMetricTwo.Labels)
	assert.EqualValues(t, 45678, ilmThreeMetricTwo.Value)

	ilmThreeMetricThree := ilms[2].Metrics[2]
	assert.Equal(t, "a_monotonic_delta_int_sum", ilmThreeMetricThree.Name)
	assert.Equal(t, "a_monotonic_delta_int_sum_description", ilmThreeMetricThree.Description)
	assert.Equal(t, "a_monotonic_delta_int_sum_unit", ilmThreeMetricThree.Unit)
	assert.Equal(t, IntMonotonicDeltaSum, ilmThreeMetricThree.Type)
	assert.Equal(t, map[string]string{"label_name_7": "label_value_7"}, ilmThreeMetricThree.Labels)
	assert.EqualValues(t, 56789, ilmThreeMetricThree.Value)

	ilmThreeMetricFour := ilms[2].Metrics[3]
	assert.Equal(t, "a_monotonic_delta_int_sum", ilmThreeMetricFour.Name)
	assert.Equal(t, "a_monotonic_delta_int_sum_description", ilmThreeMetricFour.Description)
	assert.Equal(t, "a_monotonic_delta_int_sum_unit", ilmThreeMetricFour.Unit)
	assert.Equal(t, IntMonotonicDeltaSum, ilmThreeMetricFour.Type)
	assert.Equal(t, map[string]string{"label_name_8": "label_value_8"}, ilmThreeMetricFour.Labels)
	assert.EqualValues(t, 67890, ilmThreeMetricFour.Value)

	ilmThreeMetricFive := ilms[2].Metrics[4]
	assert.Equal(t, "a_monotonic_unspecified_int_sum", ilmThreeMetricFive.Name)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_description", ilmThreeMetricFive.Description)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_unit", ilmThreeMetricFive.Unit)
	assert.Equal(t, IntMonotonicUnspecifiedSum, ilmThreeMetricFive.Type)
	assert.Equal(t, map[string]string{"label_name_9": "label_value_9"}, ilmThreeMetricFive.Labels)
	assert.EqualValues(t, 78901, ilmThreeMetricFive.Value)

	ilmThreeMetricSix := ilms[2].Metrics[5]
	assert.Equal(t, "a_monotonic_unspecified_int_sum", ilmThreeMetricSix.Name)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_description", ilmThreeMetricSix.Description)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_unit", ilmThreeMetricSix.Unit)
	assert.Equal(t, IntMonotonicUnspecifiedSum, ilmThreeMetricSix.Type)
	assert.Equal(t, map[string]string{"label_name_10": "label_value_10"}, ilmThreeMetricSix.Labels)
	assert.EqualValues(t, 89012, ilmThreeMetricSix.Value)

	ilmThreeMetricSeven := ilms[2].Metrics[6]
	assert.Equal(t, "a_monotonic_cumulative_double_sum", ilmThreeMetricSeven.Name)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_description", ilmThreeMetricSeven.Description)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_unit", ilmThreeMetricSeven.Unit)
	assert.Equal(t, DoubleMonotonicCumulativeSum, ilmThreeMetricSeven.Type)
	assert.Equal(t, map[string]string{"label_name_11": "label_value_11"}, ilmThreeMetricSeven.Labels)
	assert.EqualValues(t, 456.78, ilmThreeMetricSeven.Value)

	ilmThreeMetricEight := ilms[2].Metrics[7]
	assert.Equal(t, "a_monotonic_cumulative_double_sum", ilmThreeMetricEight.Name)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_description", ilmThreeMetricEight.Description)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_unit", ilmThreeMetricEight.Unit)
	assert.Equal(t, DoubleMonotonicCumulativeSum, ilmThreeMetricEight.Type)
	assert.Equal(t, map[string]string{"label_name_12": "label_value_12"}, ilmThreeMetricEight.Labels)
	assert.EqualValues(t, 567.89, ilmThreeMetricEight.Value)

	ilmThreeMetricNine := ilms[2].Metrics[8]
	assert.Equal(t, "a_monotonic_delta_double_sum", ilmThreeMetricNine.Name)
	assert.Equal(t, "a_monotonic_delta_double_sum_description", ilmThreeMetricNine.Description)
	assert.Equal(t, "a_monotonic_delta_double_sum_unit", ilmThreeMetricNine.Unit)
	assert.Equal(t, DoubleMonotonicDeltaSum, ilmThreeMetricNine.Type)
	assert.Equal(t, map[string]string{"label_name_13": "label_value_13"}, ilmThreeMetricNine.Labels)
	assert.EqualValues(t, 678.90, ilmThreeMetricNine.Value)

	ilmThreeMetricTen := ilms[2].Metrics[9]
	assert.Equal(t, "a_monotonic_delta_double_sum", ilmThreeMetricTen.Name)
	assert.Equal(t, "a_monotonic_delta_double_sum_description", ilmThreeMetricTen.Description)
	assert.Equal(t, "a_monotonic_delta_double_sum_unit", ilmThreeMetricTen.Unit)
	assert.Equal(t, DoubleMonotonicDeltaSum, ilmThreeMetricTen.Type)
	assert.Equal(t, map[string]string{"label_name_14": "label_value_14"}, ilmThreeMetricTen.Labels)
	assert.EqualValues(t, 789.01, ilmThreeMetricTen.Value)

	ilmThreeMetricEleven := ilms[2].Metrics[10]
	assert.Equal(t, "a_monotonic_unspecified_double_sum", ilmThreeMetricEleven.Name)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_description", ilmThreeMetricEleven.Description)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_unit", ilmThreeMetricEleven.Unit)
	assert.Equal(t, DoubleMonotonicUnspecifiedSum, ilmThreeMetricEleven.Type)
	assert.Equal(t, map[string]string{"label_name_15": "label_value_15"}, ilmThreeMetricEleven.Labels)
	assert.EqualValues(t, 890.12, ilmThreeMetricEleven.Value)

	ilmThreeMetricTwelve := ilms[2].Metrics[11]
	assert.Equal(t, "a_monotonic_unspecified_double_sum", ilmThreeMetricTwelve.Name)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_description", ilmThreeMetricTwelve.Description)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_unit", ilmThreeMetricTwelve.Unit)
	assert.Equal(t, DoubleMonotonicUnspecifiedSum, ilmThreeMetricTwelve.Type)
	assert.Equal(t, map[string]string{"label_name_16": "label_value_16"}, ilmThreeMetricTwelve.Labels)
	assert.EqualValues(t, 901.23, ilmThreeMetricTwelve.Value)

	ilmThreeMetricThirteen := ilms[2].Metrics[12]
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum", ilmThreeMetricThirteen.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_description", ilmThreeMetricThirteen.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_unit", ilmThreeMetricThirteen.Unit)
	assert.Equal(t, IntNonmonotonicCumulativeSum, ilmThreeMetricThirteen.Type)
	assert.Equal(t, map[string]string{"label_name_17": "label_value_17"}, ilmThreeMetricThirteen.Labels)
	assert.EqualValues(t, 90123, ilmThreeMetricThirteen.Value)

	ilmThreeMetricFourteen := ilms[2].Metrics[13]
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum", ilmThreeMetricFourteen.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_description", ilmThreeMetricFourteen.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_unit", ilmThreeMetricFourteen.Unit)
	assert.Equal(t, IntNonmonotonicCumulativeSum, ilmThreeMetricFourteen.Type)
	assert.Equal(t, map[string]string{"label_name_18": "label_value_18"}, ilmThreeMetricFourteen.Labels)
	assert.EqualValues(t, 123456, ilmThreeMetricFourteen.Value)

	ilmThreeMetricFifteen := ilms[2].Metrics[14]
	assert.Equal(t, "a_nonmonotonic_delta_int_sum", ilmThreeMetricFifteen.Name)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_description", ilmThreeMetricFifteen.Description)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_unit", ilmThreeMetricFifteen.Unit)
	assert.Equal(t, IntNonmonotonicDeltaSum, ilmThreeMetricFifteen.Type)
	assert.Equal(t, map[string]string{"label_name_19": "label_value_19"}, ilmThreeMetricFifteen.Labels)
	assert.EqualValues(t, 234567, ilmThreeMetricFifteen.Value)

	ilmThreeMetricSixteen := ilms[2].Metrics[15]
	assert.Equal(t, "a_nonmonotonic_delta_int_sum", ilmThreeMetricSixteen.Name)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_description", ilmThreeMetricSixteen.Description)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_unit", ilmThreeMetricSixteen.Unit)
	assert.Equal(t, IntNonmonotonicDeltaSum, ilmThreeMetricSixteen.Type)
	assert.Equal(t, map[string]string{"label_name_20": "label_value_20"}, ilmThreeMetricSixteen.Labels)
	assert.EqualValues(t, 345678, ilmThreeMetricSixteen.Value)

	ilmThreeMetricSeventeen := ilms[2].Metrics[16]
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum", ilmThreeMetricSeventeen.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_description", ilmThreeMetricSeventeen.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_unit", ilmThreeMetricSeventeen.Unit)
	assert.Equal(t, IntNonmonotonicUnspecifiedSum, ilmThreeMetricSeventeen.Type)
	assert.Equal(t, map[string]string{"label_name_21": "label_value_21"}, ilmThreeMetricSeventeen.Labels)
	assert.EqualValues(t, 456789, ilmThreeMetricSeventeen.Value)

	ilmThreeMetricEighteen := ilms[2].Metrics[17]
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum", ilmThreeMetricEighteen.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_description", ilmThreeMetricEighteen.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_unit", ilmThreeMetricEighteen.Unit)
	assert.Equal(t, IntNonmonotonicUnspecifiedSum, ilmThreeMetricEighteen.Type)
	assert.Equal(t, map[string]string{"label_name_22": "label_value_22"}, ilmThreeMetricEighteen.Labels)
	assert.EqualValues(t, 567890, ilmThreeMetricEighteen.Value)

	ilmThreeMetricNineteen := ilms[2].Metrics[18]
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum", ilmThreeMetricNineteen.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_description", ilmThreeMetricNineteen.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_unit", ilmThreeMetricNineteen.Unit)
	assert.Equal(t, DoubleNonmonotonicCumulativeSum, ilmThreeMetricNineteen.Type)
	assert.Equal(t, map[string]string{"label_name_23": "label_value_23"}, ilmThreeMetricNineteen.Labels)
	assert.EqualValues(t, 1234.56, ilmThreeMetricNineteen.Value)

	ilmThreeMetricTwenty := ilms[2].Metrics[19]
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum", ilmThreeMetricTwenty.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_description", ilmThreeMetricTwenty.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_unit", ilmThreeMetricTwenty.Unit)
	assert.Equal(t, DoubleNonmonotonicCumulativeSum, ilmThreeMetricTwenty.Type)
	assert.Equal(t, map[string]string{"label_name_24": "label_value_24"}, ilmThreeMetricTwenty.Labels)
	assert.EqualValues(t, 2345.67, ilmThreeMetricTwenty.Value)

	ilmThreeMetricTwentyOne := ilms[2].Metrics[20]
	assert.Equal(t, "a_nonmonotonic_delta_double_sum", ilmThreeMetricTwentyOne.Name)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_description", ilmThreeMetricTwentyOne.Description)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_unit", ilmThreeMetricTwentyOne.Unit)
	assert.Equal(t, DoubleNonmonotonicDeltaSum, ilmThreeMetricTwentyOne.Type)
	assert.Equal(t, map[string]string{"label_name_25": "label_value_25"}, ilmThreeMetricTwentyOne.Labels)
	assert.EqualValues(t, 3456.78, ilmThreeMetricTwentyOne.Value)

	ilmThreeMetricTwentyTwo := ilms[2].Metrics[21]
	assert.Equal(t, "a_nonmonotonic_delta_double_sum", ilmThreeMetricTwentyTwo.Name)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_description", ilmThreeMetricTwentyTwo.Description)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_unit", ilmThreeMetricTwentyTwo.Unit)
	assert.Equal(t, DoubleNonmonotonicDeltaSum, ilmThreeMetricTwentyTwo.Type)
	assert.Equal(t, map[string]string{"label_name_26": "label_value_26"}, ilmThreeMetricTwentyTwo.Labels)
	assert.EqualValues(t, 4567.89, ilmThreeMetricTwentyTwo.Value)

	ilmThreeMetricTwentyThree := ilms[2].Metrics[22]
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum", ilmThreeMetricTwentyThree.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_description", ilmThreeMetricTwentyThree.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_unit", ilmThreeMetricTwentyThree.Unit)
	assert.Equal(t, DoubleNonmonotonicUnspecifiedSum, ilmThreeMetricTwentyThree.Type)
	assert.Equal(t, map[string]string{"label_name_27": "label_value_27"}, ilmThreeMetricTwentyThree.Labels)
	assert.EqualValues(t, 5678.90, ilmThreeMetricTwentyThree.Value)

	ilmThreeMetricTwentyFour := ilms[2].Metrics[23]
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum", ilmThreeMetricTwentyFour.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_description", ilmThreeMetricTwentyFour.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_unit", ilmThreeMetricTwentyFour.Unit)
	assert.Equal(t, DoubleNonmonotonicUnspecifiedSum, ilmThreeMetricTwentyFour.Type)
	assert.Equal(t, map[string]string{"label_name_28": "label_value_28"}, ilmThreeMetricTwentyFour.Labels)
	assert.EqualValues(t, 6789.01, ilmThreeMetricTwentyFour.Value)
}
