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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPDataToResourceMetricsHappyPath(t *testing.T) {
	resourceMetrics, err := PDataToResourceMetrics(PDataMetrics())
	assert.NoError(t, err)
	require.NotNil(t, resourceMetrics)

	rms := resourceMetrics.ResourceMetrics
	assert.Len(t, rms, 1)
	rm := rms[0]
	attrs := *rm.Resource.Attributes
	assert.True(t, attrs["bool"].(bool))
	assert.Equal(t, "a_string", attrs["string"])
	assert.Equal(t, 123, attrs["int"])
	assert.Equal(t, 123.45, attrs["double"])
	assert.Nil(t, attrs["null"])

	scopeMetrics := rm.ScopeMetrics
	assert.Len(t, scopeMetrics, 3)
	assert.Equal(t, "an_instrumentation_scope_name", scopeMetrics[0].Scope.Name)
	assert.Equal(t, "an_instrumentation_scope_version", scopeMetrics[0].Scope.Version)

	require.Len(t, scopeMetrics[0].Metrics, 4)

	smOneMetricOne := scopeMetrics[0].Metrics[0]
	assert.Equal(t, "an_int_gauge", smOneMetricOne.Name)
	assert.Equal(t, "an_int_gauge_description", smOneMetricOne.Description)
	assert.Equal(t, "an_int_gauge_unit", smOneMetricOne.Unit)
	assert.Equal(t, IntGauge, smOneMetricOne.Type)
	assert.Equal(t, map[string]any{"attribute_name_1": "attribute_value_1"}, *smOneMetricOne.Attributes)
	assert.EqualValues(t, 12345, smOneMetricOne.Value)

	smOneMetricTwo := scopeMetrics[0].Metrics[1]
	assert.Equal(t, "an_int_gauge", smOneMetricTwo.Name)
	assert.Equal(t, "an_int_gauge_description", smOneMetricTwo.Description)
	assert.Equal(t, "an_int_gauge_unit", smOneMetricTwo.Unit)
	assert.Equal(t, IntGauge, smOneMetricTwo.Type)
	assert.Equal(t, map[string]any{"attribute_name_2": "attribute_value_2"}, *smOneMetricTwo.Attributes)
	assert.EqualValues(t, 23456, smOneMetricTwo.Value)

	smOneMetricThree := scopeMetrics[0].Metrics[2]
	assert.Equal(t, "a_double_gauge", smOneMetricThree.Name)
	assert.Equal(t, "a_double_gauge_description", smOneMetricThree.Description)
	assert.Equal(t, "a_double_gauge_unit", smOneMetricThree.Unit)
	assert.Equal(t, DoubleGauge, smOneMetricThree.Type)
	assert.Equal(t, map[string]any{"attribute_name_3": "attribute_value_3"}, *smOneMetricThree.Attributes)
	assert.EqualValues(t, 234.56, smOneMetricThree.Value)

	smOneMetricFour := scopeMetrics[0].Metrics[3]
	assert.Equal(t, "a_double_gauge", smOneMetricFour.Name)
	assert.Equal(t, "a_double_gauge_description", smOneMetricFour.Description)
	assert.Equal(t, "a_double_gauge_unit", smOneMetricFour.Unit)
	assert.Equal(t, DoubleGauge, smOneMetricFour.Type)
	assert.Equal(t, map[string]any{"attribute_name_4": "attribute_value_4"}, *smOneMetricFour.Attributes)
	assert.EqualValues(t, 345.67, smOneMetricFour.Value)

	assert.Equal(t, "an_instrumentation_scope_without_version_or_metrics", scopeMetrics[1].Scope.Name)
	assert.Empty(t, scopeMetrics[1].Scope.Version)
	assert.Empty(t, scopeMetrics[1].Metrics)

	require.Len(t, scopeMetrics[2].Metrics, 24)

	smThreeMetricOne := scopeMetrics[2].Metrics[0]
	assert.Equal(t, "a_monotonic_cumulative_int_sum", smThreeMetricOne.Name)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_description", smThreeMetricOne.Description)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_unit", smThreeMetricOne.Unit)
	assert.Equal(t, IntMonotonicCumulativeSum, smThreeMetricOne.Type)
	assert.Equal(t, map[string]any{"attribute_name_5": "attribute_value_5"}, *smThreeMetricOne.Attributes)
	assert.EqualValues(t, 34567, smThreeMetricOne.Value)

	smThreeMetricTwo := scopeMetrics[2].Metrics[1]
	assert.Equal(t, "a_monotonic_cumulative_int_sum", smThreeMetricTwo.Name)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_description", smThreeMetricTwo.Description)
	assert.Equal(t, "a_monotonic_cumulative_int_sum_unit", smThreeMetricTwo.Unit)
	assert.Equal(t, IntMonotonicCumulativeSum, smThreeMetricTwo.Type)
	assert.Equal(t, map[string]any{"attribute_name_6": "attribute_value_6"}, *smThreeMetricTwo.Attributes)
	assert.EqualValues(t, 45678, smThreeMetricTwo.Value)

	smThreeMetricThree := scopeMetrics[2].Metrics[2]
	assert.Equal(t, "a_monotonic_delta_int_sum", smThreeMetricThree.Name)
	assert.Equal(t, "a_monotonic_delta_int_sum_description", smThreeMetricThree.Description)
	assert.Equal(t, "a_monotonic_delta_int_sum_unit", smThreeMetricThree.Unit)
	assert.Equal(t, IntMonotonicDeltaSum, smThreeMetricThree.Type)
	assert.Equal(t, map[string]any{"attribute_name_7": "attribute_value_7"}, *smThreeMetricThree.Attributes)
	assert.EqualValues(t, 56789, smThreeMetricThree.Value)

	smThreeMetricFour := scopeMetrics[2].Metrics[3]
	assert.Equal(t, "a_monotonic_delta_int_sum", smThreeMetricFour.Name)
	assert.Equal(t, "a_monotonic_delta_int_sum_description", smThreeMetricFour.Description)
	assert.Equal(t, "a_monotonic_delta_int_sum_unit", smThreeMetricFour.Unit)
	assert.Equal(t, IntMonotonicDeltaSum, smThreeMetricFour.Type)
	assert.Equal(t, map[string]any{"attribute_name_8": "attribute_value_8"}, *smThreeMetricFour.Attributes)
	assert.EqualValues(t, 67890, smThreeMetricFour.Value)

	smThreeMetricFive := scopeMetrics[2].Metrics[4]
	assert.Equal(t, "a_monotonic_unspecified_int_sum", smThreeMetricFive.Name)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_description", smThreeMetricFive.Description)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_unit", smThreeMetricFive.Unit)
	assert.Equal(t, IntMonotonicUnspecifiedSum, smThreeMetricFive.Type)
	assert.Equal(t, map[string]any{"attribute_name_9": "attribute_value_9"}, *smThreeMetricFive.Attributes)
	assert.EqualValues(t, 78901, smThreeMetricFive.Value)

	smThreeMetricSix := scopeMetrics[2].Metrics[5]
	assert.Equal(t, "a_monotonic_unspecified_int_sum", smThreeMetricSix.Name)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_description", smThreeMetricSix.Description)
	assert.Equal(t, "a_monotonic_unspecified_int_sum_unit", smThreeMetricSix.Unit)
	assert.Equal(t, IntMonotonicUnspecifiedSum, smThreeMetricSix.Type)
	assert.Equal(t, map[string]any{"attribute_name_10": "attribute_value_10"}, *smThreeMetricSix.Attributes)
	assert.EqualValues(t, 89012, smThreeMetricSix.Value)

	smThreeMetricSeven := scopeMetrics[2].Metrics[6]
	assert.Equal(t, "a_monotonic_cumulative_double_sum", smThreeMetricSeven.Name)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_description", smThreeMetricSeven.Description)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_unit", smThreeMetricSeven.Unit)
	assert.Equal(t, DoubleMonotonicCumulativeSum, smThreeMetricSeven.Type)
	assert.Equal(t, map[string]any{"attribute_name_11": "attribute_value_11"}, *smThreeMetricSeven.Attributes)
	assert.EqualValues(t, 456.78, smThreeMetricSeven.Value)

	smThreeMetricEight := scopeMetrics[2].Metrics[7]
	assert.Equal(t, "a_monotonic_cumulative_double_sum", smThreeMetricEight.Name)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_description", smThreeMetricEight.Description)
	assert.Equal(t, "a_monotonic_cumulative_double_sum_unit", smThreeMetricEight.Unit)
	assert.Equal(t, DoubleMonotonicCumulativeSum, smThreeMetricEight.Type)
	assert.Equal(t, map[string]any{"attribute_name_12": "attribute_value_12"}, *smThreeMetricEight.Attributes)
	assert.EqualValues(t, 567.89, smThreeMetricEight.Value)

	smThreeMetricNine := scopeMetrics[2].Metrics[8]
	assert.Equal(t, "a_monotonic_delta_double_sum", smThreeMetricNine.Name)
	assert.Equal(t, "a_monotonic_delta_double_sum_description", smThreeMetricNine.Description)
	assert.Equal(t, "a_monotonic_delta_double_sum_unit", smThreeMetricNine.Unit)
	assert.Equal(t, DoubleMonotonicDeltaSum, smThreeMetricNine.Type)
	assert.Equal(t, map[string]any{"attribute_name_13": "attribute_value_13"}, *smThreeMetricNine.Attributes)
	assert.EqualValues(t, 678.90, smThreeMetricNine.Value)

	smThreeMetricTen := scopeMetrics[2].Metrics[9]
	assert.Equal(t, "a_monotonic_delta_double_sum", smThreeMetricTen.Name)
	assert.Equal(t, "a_monotonic_delta_double_sum_description", smThreeMetricTen.Description)
	assert.Equal(t, "a_monotonic_delta_double_sum_unit", smThreeMetricTen.Unit)
	assert.Equal(t, DoubleMonotonicDeltaSum, smThreeMetricTen.Type)
	assert.Equal(t, map[string]any{"attribute_name_14": "attribute_value_14"}, *smThreeMetricTen.Attributes)
	assert.EqualValues(t, 789.01, smThreeMetricTen.Value)

	smThreeMetricEleven := scopeMetrics[2].Metrics[10]
	assert.Equal(t, "a_monotonic_unspecified_double_sum", smThreeMetricEleven.Name)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_description", smThreeMetricEleven.Description)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_unit", smThreeMetricEleven.Unit)
	assert.Equal(t, DoubleMonotonicUnspecifiedSum, smThreeMetricEleven.Type)
	assert.Equal(t, map[string]any{"attribute_name_15": "attribute_value_15"}, *smThreeMetricEleven.Attributes)
	assert.EqualValues(t, 890.12, smThreeMetricEleven.Value)

	smThreeMetricTwelve := scopeMetrics[2].Metrics[11]
	assert.Equal(t, "a_monotonic_unspecified_double_sum", smThreeMetricTwelve.Name)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_description", smThreeMetricTwelve.Description)
	assert.Equal(t, "a_monotonic_unspecified_double_sum_unit", smThreeMetricTwelve.Unit)
	assert.Equal(t, DoubleMonotonicUnspecifiedSum, smThreeMetricTwelve.Type)
	assert.Equal(t, map[string]any{"attribute_name_16": "attribute_value_16"}, *smThreeMetricTwelve.Attributes)
	assert.EqualValues(t, 901.23, smThreeMetricTwelve.Value)

	smThreeMetricThirteen := scopeMetrics[2].Metrics[12]
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum", smThreeMetricThirteen.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_description", smThreeMetricThirteen.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_unit", smThreeMetricThirteen.Unit)
	assert.Equal(t, IntNonmonotonicCumulativeSum, smThreeMetricThirteen.Type)
	assert.Equal(t, map[string]any{"attribute_name_17": "attribute_value_17"}, *smThreeMetricThirteen.Attributes)
	assert.EqualValues(t, 90123, smThreeMetricThirteen.Value)

	smThreeMetricFourteen := scopeMetrics[2].Metrics[13]
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum", smThreeMetricFourteen.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_description", smThreeMetricFourteen.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_int_sum_unit", smThreeMetricFourteen.Unit)
	assert.Equal(t, IntNonmonotonicCumulativeSum, smThreeMetricFourteen.Type)
	assert.Equal(t, map[string]any{"attribute_name_18": "attribute_value_18"}, *smThreeMetricFourteen.Attributes)
	assert.EqualValues(t, 123456, smThreeMetricFourteen.Value)

	smThreeMetricFifteen := scopeMetrics[2].Metrics[14]
	assert.Equal(t, "a_nonmonotonic_delta_int_sum", smThreeMetricFifteen.Name)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_description", smThreeMetricFifteen.Description)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_unit", smThreeMetricFifteen.Unit)
	assert.Equal(t, IntNonmonotonicDeltaSum, smThreeMetricFifteen.Type)
	assert.Equal(t, map[string]any{"attribute_name_19": "attribute_value_19"}, *smThreeMetricFifteen.Attributes)
	assert.EqualValues(t, 234567, smThreeMetricFifteen.Value)

	smThreeMetricSixteen := scopeMetrics[2].Metrics[15]
	assert.Equal(t, "a_nonmonotonic_delta_int_sum", smThreeMetricSixteen.Name)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_description", smThreeMetricSixteen.Description)
	assert.Equal(t, "a_nonmonotonic_delta_int_sum_unit", smThreeMetricSixteen.Unit)
	assert.Equal(t, IntNonmonotonicDeltaSum, smThreeMetricSixteen.Type)
	assert.Equal(t, map[string]any{"attribute_name_20": "attribute_value_20"}, *smThreeMetricSixteen.Attributes)
	assert.EqualValues(t, 345678, smThreeMetricSixteen.Value)

	smThreeMetricSeventeen := scopeMetrics[2].Metrics[16]
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum", smThreeMetricSeventeen.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_description", smThreeMetricSeventeen.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_unit", smThreeMetricSeventeen.Unit)
	assert.Equal(t, IntNonmonotonicUnspecifiedSum, smThreeMetricSeventeen.Type)
	assert.Equal(t, map[string]any{"attribute_name_21": "attribute_value_21"}, *smThreeMetricSeventeen.Attributes)
	assert.EqualValues(t, 456789, smThreeMetricSeventeen.Value)

	smThreeMetricEighteen := scopeMetrics[2].Metrics[17]
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum", smThreeMetricEighteen.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_description", smThreeMetricEighteen.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_int_sum_unit", smThreeMetricEighteen.Unit)
	assert.Equal(t, IntNonmonotonicUnspecifiedSum, smThreeMetricEighteen.Type)
	assert.Equal(t, map[string]any{"attribute_name_22": "attribute_value_22"}, *smThreeMetricEighteen.Attributes)
	assert.EqualValues(t, 567890, smThreeMetricEighteen.Value)

	smThreeMetricNineteen := scopeMetrics[2].Metrics[18]
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum", smThreeMetricNineteen.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_description", smThreeMetricNineteen.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_unit", smThreeMetricNineteen.Unit)
	assert.Equal(t, DoubleNonmonotonicCumulativeSum, smThreeMetricNineteen.Type)
	assert.Equal(t, map[string]any{"attribute_name_23": "attribute_value_23"}, *smThreeMetricNineteen.Attributes)
	assert.EqualValues(t, 1234.56, smThreeMetricNineteen.Value)

	smThreeMetricTwenty := scopeMetrics[2].Metrics[19]
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum", smThreeMetricTwenty.Name)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_description", smThreeMetricTwenty.Description)
	assert.Equal(t, "a_nonmonotonic_cumulative_double_sum_unit", smThreeMetricTwenty.Unit)
	assert.Equal(t, DoubleNonmonotonicCumulativeSum, smThreeMetricTwenty.Type)
	assert.Equal(t, map[string]any{"attribute_name_24": "attribute_value_24"}, *smThreeMetricTwenty.Attributes)
	assert.EqualValues(t, 2345.67, smThreeMetricTwenty.Value)

	smThreeMetricTwentyOne := scopeMetrics[2].Metrics[20]
	assert.Equal(t, "a_nonmonotonic_delta_double_sum", smThreeMetricTwentyOne.Name)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_description", smThreeMetricTwentyOne.Description)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_unit", smThreeMetricTwentyOne.Unit)
	assert.Equal(t, DoubleNonmonotonicDeltaSum, smThreeMetricTwentyOne.Type)
	assert.Equal(t, map[string]any{"attribute_name_25": "attribute_value_25"}, *smThreeMetricTwentyOne.Attributes)
	assert.EqualValues(t, 3456.78, smThreeMetricTwentyOne.Value)

	smThreeMetricTwentyTwo := scopeMetrics[2].Metrics[21]
	assert.Equal(t, "a_nonmonotonic_delta_double_sum", smThreeMetricTwentyTwo.Name)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_description", smThreeMetricTwentyTwo.Description)
	assert.Equal(t, "a_nonmonotonic_delta_double_sum_unit", smThreeMetricTwentyTwo.Unit)
	assert.Equal(t, DoubleNonmonotonicDeltaSum, smThreeMetricTwentyTwo.Type)
	assert.Equal(t, map[string]any{"attribute_name_26": "attribute_value_26"}, *smThreeMetricTwentyTwo.Attributes)
	assert.EqualValues(t, 4567.89, smThreeMetricTwentyTwo.Value)

	smThreeMetricTwentyThree := scopeMetrics[2].Metrics[22]
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum", smThreeMetricTwentyThree.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_description", smThreeMetricTwentyThree.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_unit", smThreeMetricTwentyThree.Unit)
	assert.Equal(t, DoubleNonmonotonicUnspecifiedSum, smThreeMetricTwentyThree.Type)
	assert.Equal(t, map[string]any{"attribute_name_27": "attribute_value_27"}, *smThreeMetricTwentyThree.Attributes)
	assert.EqualValues(t, 5678.90, smThreeMetricTwentyThree.Value)

	smThreeMetricTwentyFour := scopeMetrics[2].Metrics[23]
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum", smThreeMetricTwentyFour.Name)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_description", smThreeMetricTwentyFour.Description)
	assert.Equal(t, "a_nonmonotonic_unspecified_double_sum_unit", smThreeMetricTwentyFour.Unit)
	assert.Equal(t, DoubleNonmonotonicUnspecifiedSum, smThreeMetricTwentyFour.Type)
	assert.Equal(t, map[string]any{"attribute_name_28": "attribute_value_28"}, *smThreeMetricTwentyFour.Attributes)
	assert.EqualValues(t, 6789.01, smThreeMetricTwentyFour.Value)
}
