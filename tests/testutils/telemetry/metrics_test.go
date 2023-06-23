// Copyright Splunk, Inc.
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

//go:build testutils

package telemetry

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadedResourceMetrics(t *testing.T) ResourceMetrics {
	resourceMetrics, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "resource-metrics.yaml"))
	require.NoError(t, err)
	require.NotNil(t, resourceMetrics)
	return *resourceMetrics
}

func TestResourceMetricsYamlStringRep(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(".", "testdata", "metrics", "resource-metrics.yaml"))
	require.NoError(t, err)
	resourceMetrics := loadedResourceMetrics(t)
	require.Equal(t, string(b), fmt.Sprintf("%v", resourceMetrics))
}

func TestLoadMetricsHappyPath(t *testing.T) {
	resourceMetrics := loadedResourceMetrics(t)
	assert.Equal(t, 2, len(resourceMetrics.ResourceMetrics))

	firstRM := resourceMetrics.ResourceMetrics[0]
	firstRMAttrs := *firstRM.Resource.Attributes
	require.Equal(t, 2, len(firstRMAttrs))
	require.NotNil(t, firstRMAttrs["one_attr"])
	assert.Equal(t, "one_value", firstRMAttrs["one_attr"])
	require.NotNil(t, firstRMAttrs["two_attr"])
	assert.Equal(t, "two_value", firstRMAttrs["two_attr"])

	assert.Equal(t, 2, len(firstRM.ScopeMetrics))
	firstRMFirstSM := firstRM.ScopeMetrics[0]
	require.NotNil(t, firstRMFirstSM)
	require.NotNil(t, firstRMFirstSM.Scope)
	assert.Equal(t, "without_metrics", firstRMFirstSM.Scope.Name)
	assert.Equal(t, "some_version", firstRMFirstSM.Scope.Version)
	require.Nil(t, firstRMFirstSM.Metrics)

	firstRMSecondSM := firstRM.ScopeMetrics[1]
	require.NotNil(t, firstRMSecondSM)
	require.NotNil(t, firstRMSecondSM.Scope)
	assert.Empty(t, firstRMSecondSM.Scope.Name)
	assert.Empty(t, firstRMSecondSM.Scope.Version)
	require.NotNil(t, firstRMSecondSM.Metrics)

	require.Equal(t, 2, len(firstRMSecondSM.Metrics))
	firstRMSecondSMFirstMetric := firstRMSecondSM.Metrics[0]
	require.NotNil(t, firstRMSecondSMFirstMetric)
	assert.Equal(t, "an_int_gauge", firstRMSecondSMFirstMetric.Name)
	assert.Equal(t, IntGauge, firstRMSecondSMFirstMetric.Type)
	assert.Equal(t, "an_int_gauge_description", firstRMSecondSMFirstMetric.Description)
	assert.Equal(t, "an_int_gauge_unit", firstRMSecondSMFirstMetric.Unit)
	assert.Equal(t, 123, firstRMSecondSMFirstMetric.Value)

	firstRMSecondScopeMetricsecondMetric := firstRMSecondSM.Metrics[1]
	require.NotNil(t, firstRMSecondScopeMetricsecondMetric)
	assert.Equal(t, "a_double_gauge", firstRMSecondScopeMetricsecondMetric.Name)
	assert.Equal(t, DoubleGauge, firstRMSecondScopeMetricsecondMetric.Type)
	assert.Equal(t, 123.456, firstRMSecondScopeMetricsecondMetric.Value)
	assert.Empty(t, firstRMSecondScopeMetricsecondMetric.Unit)
	assert.Empty(t, firstRMSecondScopeMetricsecondMetric.Description)

	secondRM := resourceMetrics.ResourceMetrics[1]
	require.Nil(t, secondRM.Resource.Attributes)

	assert.Equal(t, 1, len(secondRM.ScopeMetrics))
	secondRMFirstSM := secondRM.ScopeMetrics[0]
	require.NotNil(t, secondRMFirstSM)
	require.NotNil(t, secondRMFirstSM.Scope)
	assert.Equal(t, "with_metrics", secondRMFirstSM.Scope.Name)
	assert.Equal(t, "another_version", secondRMFirstSM.Scope.Version)
	require.NotNil(t, secondRMFirstSM.Metrics)

	require.Equal(t, 2, len(secondRMFirstSM.Metrics))
	secondRMFirstSMFirstMetric := secondRMFirstSM.Metrics[0]
	require.NotNil(t, secondRMFirstSMFirstMetric)
	assert.Equal(t, "another_int_gauge", secondRMFirstSMFirstMetric.Name)
	assert.Equal(t, IntGauge, secondRMFirstSMFirstMetric.Type)
	assert.Empty(t, secondRMFirstSMFirstMetric.Description)
	assert.Empty(t, secondRMFirstSMFirstMetric.Unit)
	assert.Equal(t, 456, secondRMFirstSMFirstMetric.Value)

	secondRMFirstScopeMetricsecondMetric := secondRMFirstSM.Metrics[1]
	require.NotNil(t, secondRMFirstScopeMetricsecondMetric)
	assert.Equal(t, "another_double_gauge", secondRMFirstScopeMetricsecondMetric.Name)
	assert.Equal(t, DoubleGauge, secondRMFirstScopeMetricsecondMetric.Type)
	assert.Empty(t, secondRMFirstScopeMetricsecondMetric.Description)
	assert.Empty(t, secondRMFirstScopeMetricsecondMetric.Unit)
	assert.Equal(t, 567.89, secondRMFirstScopeMetricsecondMetric.Value)
}

func TestLoadMetricsNotAValidPath(t *testing.T) {
	resourceMetrics, err := LoadResourceMetrics("notafile")
	require.Error(t, err)
	require.Contains(t, err.Error(), invalidPathErrorMsg())
	require.Nil(t, resourceMetrics)
}

func TestLoadMetricsInvalidItems(t *testing.T) {
	resourceMetrics, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "invalid-resource-metrics.yaml"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "field notAttributesOrScopeMetrics not found in type telemetry.ResourceMetric")
	require.Nil(t, resourceMetrics)
}

func TestLoadMetricsInvalidMetricType(t *testing.T) {
	resourceMetrics, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "invalid-metric-type.yaml"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported MetricType for of_an_unsupported_type - NotASupportedType")
	require.Nil(t, resourceMetrics)
}

func TestResourceMatchesWithAny(t *testing.T) {
	rReference := Resource{Attributes: &map[string]any{
		"one": 1, "two": "<ANY>", "three": nil,
		"four": []int{1, 2, 3, 4},
		"five": map[string]any{
			"true": true, "false": false, "nil": nil,
		},
	}}
	rShouldEqual := Resource{Attributes: &map[string]any{
		"one": 1, "two": "two", "three": nil,
		"four": []int{1, 2, 3, 4},
		"five": map[string]any{
			"true": true, "false": false, "nil": nil,
		},
	}}
	rMissingTwo := Resource{Attributes: &map[string]any{
		"one": 1, "three": nil,
		"four": []int{1, 2, 3, 4},
		"five": map[string]any{
			"true": true, "false": false, "nil": nil,
		},
	}}

	assert.True(t, rReference.Equals(rShouldEqual))
	assert.False(t, rReference.Equals(rMissingTwo))
}

func TestMetricEquivalence(t *testing.T) {
	metric := func() Metric {
		return Metric{
			Name: "a_metric", Description: "a_metric_description",
			Unit: "a_metric_unit", Type: MetricType("a_metric_type"),
			Attributes: &map[string]any{
				"one": "one", "two": "two",
			}, Value: 123,
		}
	}

	mOne := metric()
	mOneSelf := metric()
	assert.True(t, mOne.Equals(mOneSelf))

	mTwo := metric()
	assert.True(t, mOne.Equals(mTwo))
	assert.True(t, mTwo.Equals(mOne))

	mTwo.Name = ""
	assert.False(t, mOne.Equals(mTwo))
	assert.False(t, mTwo.Equals(mOne))
	mOne.Name = ""
	assert.True(t, mOne.Equals(mTwo))
	assert.True(t, mTwo.Equals(mOne))

	mTwo.Description = ""
	assert.False(t, mOne.Equals(mTwo))
	assert.False(t, mTwo.Equals(mOne))
	mOne.Description = ""
	assert.True(t, mOne.Equals(mTwo))
	assert.True(t, mTwo.Equals(mOne))

	mTwo.Unit = ""
	assert.False(t, mOne.Equals(mTwo))
	assert.False(t, mTwo.Equals(mOne))
	mOne.Unit = ""
	assert.True(t, mOne.Equals(mTwo))
	assert.True(t, mTwo.Equals(mOne))

	mTwo.Type = ""
	assert.False(t, mOne.Equals(mTwo))
	assert.False(t, mTwo.Equals(mOne))
	mOne.Type = ""
	assert.True(t, mOne.Equals(mTwo))
	assert.True(t, mTwo.Equals(mOne))

	mTwo.Value = 0
	assert.False(t, mOne.Equals(mTwo))
	assert.False(t, mTwo.Equals(mOne))
	mOne.Value = 0
	assert.True(t, mOne.Equals(mTwo))
	assert.True(t, mTwo.Equals(mOne))

	(*mTwo.Attributes)["three"] = "three"
	assert.False(t, mOne.Equals(mTwo))
	assert.False(t, mTwo.Equals(mOne))
	(*mOne.Attributes)["three"] = "three"
	assert.True(t, mOne.Equals(mTwo))
	assert.True(t, mTwo.Equals(mOne))
}

func TestMetricRelaxedEquivalence(t *testing.T) {
	lacksDescriptionUnitAndType := Metric{
		Name: "a_metric",
		Attributes: &map[string]any{
			"one": "one", "two": "two",
		}, Value: 123,
	}

	completeMetric := Metric{
		Name: "a_metric", Description: "a_description",
		Unit: "a_metric_unit", Type: "a_metric_type",
		Attributes: &map[string]any{
			"one": "one", "two": "two",
		}, Value: 123,
	}

	require.True(t, lacksDescriptionUnitAndType.RelaxedEquals(completeMetric))
	require.False(t, completeMetric.RelaxedEquals(lacksDescriptionUnitAndType))

	(*lacksDescriptionUnitAndType.Attributes)["three"] = "three"
	require.False(t, lacksDescriptionUnitAndType.RelaxedEquals(completeMetric))
	require.False(t, completeMetric.RelaxedEquals(lacksDescriptionUnitAndType))
	(*completeMetric.Attributes)["three"] = "three"
	require.True(t, lacksDescriptionUnitAndType.RelaxedEquals(completeMetric))
	require.False(t, completeMetric.RelaxedEquals(lacksDescriptionUnitAndType))

	lacksDescriptionUnitAndType.Value = 234
	require.False(t, lacksDescriptionUnitAndType.RelaxedEquals(completeMetric))
	require.False(t, completeMetric.RelaxedEquals(lacksDescriptionUnitAndType))
	completeMetric.Value = 234
	require.True(t, lacksDescriptionUnitAndType.RelaxedEquals(completeMetric))
	require.False(t, completeMetric.RelaxedEquals(lacksDescriptionUnitAndType))

	lacksDescriptionUnitAndType.Value = nil
	require.True(t, lacksDescriptionUnitAndType.RelaxedEquals(completeMetric))
	require.False(t, completeMetric.RelaxedEquals(lacksDescriptionUnitAndType))

	lacksDescriptionUnitAndType.Value = 234
	completeMetric.Description = ""
	completeMetric.Unit = ""
	completeMetric.Type = ""
	require.True(t, lacksDescriptionUnitAndType.RelaxedEquals(completeMetric))
	require.True(t, completeMetric.RelaxedEquals(lacksDescriptionUnitAndType))
}

func TestMetricAttributeRelaxedEquivalence(t *testing.T) {
	lackingAttributes := Metric{
		Name: "a_metric", Description: "a_description",
		Unit: "a_metric_unit", Value: 123,
	}

	emptyAttributes := Metric{
		Name: "a_metric", Description: "a_description",
		Unit: "a_metric_unit", Attributes: &map[string]any{},
		Value: 123,
	}

	completeMetric := Metric{
		Name: "a_metric", Description: "a_description",
		Unit: "a_metric_unit", Type: "a_metric_type",
		Attributes: &map[string]any{
			"one": "one", "two": "two",
		}, Value: 123,
	}

	require.True(t, lackingAttributes.RelaxedEquals(completeMetric))
	require.False(t, emptyAttributes.RelaxedEquals(completeMetric))
}

func TestMetricHashFunctionConsistency(t *testing.T) {
	metric := Metric{
		Name: "some metric", Description: "some description",
		Unit: "some unit", Attributes: &map[string]any{
			"attributeOne": "1", "attributeTwo": "two",
		}, Type: MetricType("some metric type"), Value: 123.456,
	}
	for i := 0; i < 100; i++ {
		require.Equal(t, "7fb66e09a072a06173f4cd1f2d63bf03", metric.Hash())
	}
}

func TestFlattenResourceMetricsByResourceIdentity(t *testing.T) {
	resource := Resource{Attributes: &map[string]any{"attribute_one": nil, "attribute_two": 123.456}}
	resourceMetrics := ResourceMetrics{
		ResourceMetrics: []ResourceMetric{
			{Resource: resource},
			{Resource: resource},
			{Resource: resource},
		},
	}
	expectedResourceMetrics := ResourceMetrics{ResourceMetrics: []ResourceMetric{{Resource: resource}}}
	require.Equal(t, expectedResourceMetrics, FlattenResourceMetrics(resourceMetrics))
}

func TestFlattenResourceMetricsByScopeMetricsIdentity(t *testing.T) {
	resource := Resource{Attributes: &map[string]any{"attribute_three": true, "attribute_four": 23456}}
	sm := ScopeMetrics{Scope: InstrumentationScope{
		Name: "an instrumentation library", Version: "an instrumentation library version",
	}, Metrics: []Metric{}}
	resourceMetrics := ResourceMetrics{
		ResourceMetrics: []ResourceMetric{
			{Resource: resource, ScopeMetrics: []ScopeMetrics{}},
			{Resource: resource, ScopeMetrics: []ScopeMetrics{sm}},
			{Resource: resource, ScopeMetrics: []ScopeMetrics{sm, sm}},
			{Resource: resource, ScopeMetrics: []ScopeMetrics{sm, sm, sm}},
		},
	}
	expectedResourceMetrics := ResourceMetrics{
		ResourceMetrics: []ResourceMetric{
			{Resource: resource, ScopeMetrics: []ScopeMetrics{sm}},
		},
	}
	require.Equal(t, expectedResourceMetrics, FlattenResourceMetrics(resourceMetrics))
}

func TestFlattenResourceMetricsByMetricsIdentity(t *testing.T) {
	resource := Resource{}
	metrics := []Metric{
		{Name: "a metric", Unit: "a unit", Description: "a description", Value: 123},
		{Name: "another metric", Unit: "another unit", Description: "another description", Value: 234},
		{Name: "yet another metric", Unit: "yet anothe unit", Description: "yet anothet description", Value: 345},
	}
	sm := ScopeMetrics{Metrics: metrics}
	smRepeated := ScopeMetrics{Metrics: append(metrics, metrics...)}
	smRepeatedTwice := ScopeMetrics{Metrics: append(metrics, append(metrics, metrics...)...)}
	smWithoutMetrics := ScopeMetrics{}
	resourceMetrics := ResourceMetrics{
		ResourceMetrics: []ResourceMetric{
			{Resource: resource, ScopeMetrics: []ScopeMetrics{}},
			{Resource: resource, ScopeMetrics: []ScopeMetrics{sm}},
			{Resource: resource, ScopeMetrics: []ScopeMetrics{smRepeated}},
			{Resource: resource, ScopeMetrics: []ScopeMetrics{smRepeatedTwice}},
			{Resource: resource, ScopeMetrics: []ScopeMetrics{smWithoutMetrics}},
		},
	}
	expectedResourceMetrics := ResourceMetrics{
		ResourceMetrics: []ResourceMetric{
			{Resource: resource, ScopeMetrics: []ScopeMetrics{sm}},
		},
	}
	require.Equal(t, expectedResourceMetrics, FlattenResourceMetrics(resourceMetrics))
}

func TestFlattenResourceMetricsConsistency(t *testing.T) {
	resourceMetrics, err := PDataToResourceMetrics(PDataMetrics())
	require.NoError(t, err)
	require.NotNil(t, resourceMetrics)
	require.Equal(t, resourceMetrics, FlattenResourceMetrics(resourceMetrics))
	var rms []ResourceMetrics
	for i := 0; i < 50; i++ {
		rms = append(rms, resourceMetrics)
	}
	for i := 0; i < 50; i++ {
		require.Equal(t, resourceMetrics, FlattenResourceMetrics(rms...))
	}
}

func TestMetricContainsAllSelfCheck(t *testing.T) {
	resourceMetrics := loadedResourceMetrics(t)
	containsAll, err := resourceMetrics.ContainsAll(resourceMetrics)
	require.True(t, containsAll, err)
	require.NoError(t, err)
}

func TestMetricContainsOnlySelfCheck(t *testing.T) {
	resourceMetrics := loadedResourceMetrics(t)
	containsAll, err := resourceMetrics.ContainsOnly(resourceMetrics)
	require.True(t, containsAll, err)
	require.NoError(t, err)
}

func TestMetricContainsAllNoBijection(t *testing.T) {
	received := loadedResourceMetrics(t)

	expected, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "expected-metrics.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	containsAll, err := received.ContainsAll(*expected)
	require.True(t, containsAll, err)
	require.NoError(t, err)

	// Since expectedMetrics specify no values, they will never find matches with metrics w/ them.
	containsAll, err = expected.ContainsAll(received)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(),
		"Missing Metrics: [name: another_int_gauge\ntype: IntGauge\nvalue: 456\n name: another_double_gauge\ntype: DoubleGauge\nvalue: 567.89\n]",
	)
}

func TestMetricContainsAllValueNeverReceived(t *testing.T) {
	received := loadedResourceMetrics(t)
	expected, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "never-received-metrics.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// neverReceivedMetrics.yaml details a Metric with a value that isn't in resourceMetrics.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing Metrics: [name: another_int_gauge\ntype: IntGauge\nvalue: 111\n]")
}

func TestMetricContainsAllInstrumentationScopeNeverReceived(t *testing.T) {
	received := loadedResourceMetrics(t)
	expected, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "never-received-instrumentation-scope.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// neverReceivedMetrics.yaml details an InstrumentationScope  that isn't in resourceMetrics.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing InstrumentationScopes: [name: unmatched_instrumentation_scope\n]")
}

func TestMetricContainsAllResourceNeverReceived(t *testing.T) {
	received := loadedResourceMetrics(t)
	expected, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "never-received-resource.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// neverReceivedMetrics.yaml details a Resource that isn't in resourceMetrics.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing resources: [attributes:\n  not: matched\n]")
}

func TestMetricContainsAllWithMissingAndEmptyAttributes(t *testing.T) {
	received, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "attribute-value-resource-metrics.yaml"))
	require.NoError(t, err)
	require.NotNil(t, received)

	unspecified, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "unspecified-attributes-allowed.yaml"))
	require.NoError(t, err)
	require.NotNil(t, unspecified)

	empty, err := LoadResourceMetrics(filepath.Join(".", "testdata", "metrics", "empty-attributes-required.yaml"))
	require.NoError(t, err)
	require.NotNil(t, empty)

	containsAll, err := received.ContainsAll(*unspecified)
	require.True(t, containsAll)
	require.NoError(t, err)

	containsAll, err = received.ContainsAll(*empty)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing Metrics: [name: another_int_gauge\ntype: IntGauge\nvalue: 111\n]")
}

func TestMetricContainsOnlyDetectsUnexpectedMetric(t *testing.T) {
	resourceMetrics := loadedResourceMetrics(t)
	sm := resourceMetrics.ResourceMetrics[0].ScopeMetrics
	sm[0].Metrics = append(sm[0].Metrics, Metric{Name: "unexpected_metric"})
	containsAll, err := resourceMetrics.ContainsOnly(loadedResourceMetrics(t))
	require.False(t, containsAll, err)
	require.EqualError(t, err, fmt.Sprintf("%v contains unexpected metrics unexpected_metric", resourceMetrics))
}
