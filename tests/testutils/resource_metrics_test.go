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

package testutils

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadedResourceMetrics(t *testing.T) ResourceMetrics {
	resourceMetrics, err := LoadResourceMetrics(path.Join(".", "testdata", "resourceMetrics.yaml"))
	require.NoError(t, err)
	require.NotNil(t, resourceMetrics)
	return *resourceMetrics
}

func TestLoadMetricsHappyPath(t *testing.T) {
	resourceMetrics := loadedResourceMetrics(t)
	assert.Equal(t, 2, len(resourceMetrics.ResourceMetrics))

	firstRM := resourceMetrics.ResourceMetrics[0]
	firstRMAttrs := firstRM.Resource.Attributes
	require.Equal(t, 2, len(firstRMAttrs))
	require.NotNil(t, firstRMAttrs["one_attr"])
	assert.Equal(t, "one_value", firstRMAttrs["one_attr"])
	require.NotNil(t, firstRMAttrs["two_attr"])
	assert.Equal(t, "two_value", firstRMAttrs["two_attr"])

	assert.Equal(t, 2, len(firstRM.ILMs))
	firstRMFirstILM := firstRM.ILMs[0]
	require.NotNil(t, firstRMFirstILM)
	require.NotNil(t, firstRMFirstILM.InstrumentationLibrary)
	assert.Equal(t, "without_metrics", firstRMFirstILM.InstrumentationLibrary.Name)
	assert.Equal(t, "some_version", firstRMFirstILM.InstrumentationLibrary.Version)
	require.Nil(t, firstRMFirstILM.Metrics)

	firstRMSecondILM := firstRM.ILMs[1]
	require.NotNil(t, firstRMSecondILM)
	require.NotNil(t, firstRMSecondILM.InstrumentationLibrary)
	assert.Empty(t, firstRMSecondILM.InstrumentationLibrary.Name)
	assert.Empty(t, firstRMSecondILM.InstrumentationLibrary.Version)
	require.NotNil(t, firstRMSecondILM.Metrics)

	require.Equal(t, 2, len(firstRMSecondILM.Metrics))
	firstRMSecondILMFirstMetric := firstRMSecondILM.Metrics[0]
	require.NotNil(t, firstRMSecondILMFirstMetric)
	assert.Equal(t, "an_int_gauge", firstRMSecondILMFirstMetric.Name)
	assert.Equal(t, IntGauge, firstRMSecondILMFirstMetric.Type)
	assert.Equal(t, "an_int_gauge_description", firstRMSecondILMFirstMetric.Description)
	assert.Equal(t, "an_int_gauge_unit", firstRMSecondILMFirstMetric.Unit)
	assert.Equal(t, 123, firstRMSecondILMFirstMetric.Value)

	firstRMSecondILMSecondMetric := firstRMSecondILM.Metrics[1]
	require.NotNil(t, firstRMSecondILMSecondMetric)
	assert.Equal(t, "a_double_gauge", firstRMSecondILMSecondMetric.Name)
	assert.Equal(t, DoubleGauge, firstRMSecondILMSecondMetric.Type)
	assert.Equal(t, 123.456, firstRMSecondILMSecondMetric.Value)
	assert.Empty(t, firstRMSecondILMSecondMetric.Unit)
	assert.Empty(t, firstRMSecondILMSecondMetric.Description)

	secondRM := resourceMetrics.ResourceMetrics[1]
	require.Zero(t, len(secondRM.Resource.Attributes))

	assert.Equal(t, 1, len(secondRM.ILMs))
	secondRMFirstILM := secondRM.ILMs[0]
	require.NotNil(t, secondRMFirstILM)
	require.NotNil(t, secondRMFirstILM.InstrumentationLibrary)
	assert.Equal(t, "with_metrics", secondRMFirstILM.InstrumentationLibrary.Name)
	assert.Equal(t, "another_version", secondRMFirstILM.InstrumentationLibrary.Version)
	require.NotNil(t, secondRMFirstILM.Metrics)

	require.Equal(t, 2, len(secondRMFirstILM.Metrics))
	secondRMFirstILMFirstMetric := secondRMFirstILM.Metrics[0]
	require.NotNil(t, secondRMFirstILMFirstMetric)
	assert.Equal(t, "another_int_gauge", secondRMFirstILMFirstMetric.Name)
	assert.Equal(t, IntGauge, secondRMFirstILMFirstMetric.Type)
	assert.Empty(t, secondRMFirstILMFirstMetric.Description)
	assert.Empty(t, secondRMFirstILMFirstMetric.Unit)
	assert.Equal(t, 456, secondRMFirstILMFirstMetric.Value)

	secondRMFirstILMSecondMetric := secondRMFirstILM.Metrics[1]
	require.NotNil(t, secondRMFirstILMSecondMetric)
	assert.Equal(t, "another_double_gauge", secondRMFirstILMSecondMetric.Name)
	assert.Equal(t, DoubleGauge, secondRMFirstILMSecondMetric.Type)
	assert.Empty(t, secondRMFirstILMSecondMetric.Description)
	assert.Empty(t, secondRMFirstILMSecondMetric.Unit)
	assert.Equal(t, 567.89, secondRMFirstILMSecondMetric.Value)
}

func TestLoadMetricsNotAValidPath(t *testing.T) {
	resourceMetrics, err := LoadResourceMetrics("notafile")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no such file or directory")
	require.Nil(t, resourceMetrics)
}

func TestLoadMetricsInvalidItems(t *testing.T) {
	resourceMetrics, err := LoadResourceMetrics(path.Join(".", "testdata", "invalidResourceMetrics.yaml"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "field notAttributesOrILMs not found in type testutils.ResourceMetric")
	require.Nil(t, resourceMetrics)
}

func TestLoadMetricsInvalidMetricType(t *testing.T) {
	resourceMetrics, err := LoadResourceMetrics(path.Join(".", "testdata", "invalidMetricType.yaml"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported MetricType for of_an_unsupported_type - NotASupportedType")
	require.Nil(t, resourceMetrics)
}

func TestResourceEquivalence(t *testing.T) {
	resource := func() Resource {
		return Resource{Attributes: map[string]interface{}{
			"one": 1, "two": "two", "three": nil,
			"four": []int{1, 2, 3, 4},
			"five": map[string]interface{}{
				"true": true, "false": false, "nil": nil,
			},
		}}
	}
	rOne := resource()
	assert.True(t, rOne.Equals(rOne))

	rTwo := resource()
	assert.True(t, rOne.Equals(rTwo))
	assert.True(t, rTwo.Equals(rOne))

	rTwo.Attributes["five"].(map[string]interface{})["another"] = "item"
	assert.False(t, rOne.Equals(rTwo))
	assert.False(t, rTwo.Equals(rOne))
	rOne.Attributes["five"].(map[string]interface{})["another"] = "item"
	assert.True(t, rOne.Equals(rTwo))
	assert.True(t, rTwo.Equals(rOne))
}

func TestInstrumentationLibraryEquivalence(t *testing.T) {
	il := func() InstrumentationLibrary {
		return InstrumentationLibrary{
			Name: "an_instrumentation_library", Version: "an_instrumentation_library_version",
		}
	}

	ilOne := il()
	assert.True(t, ilOne.Equals(ilOne))

	ilTwo := il()
	assert.True(t, ilOne.Equals(ilTwo))
	assert.True(t, ilTwo.Equals(ilOne))

	ilTwo.Version = ""
	assert.False(t, ilOne.Equals(ilTwo))
	assert.False(t, ilTwo.Equals(ilOne))
	ilOne.Version = ""
	assert.True(t, ilOne.Equals(ilTwo))
	assert.True(t, ilTwo.Equals(ilOne))

	ilTwo.Name = ""
	assert.False(t, ilOne.Equals(ilTwo))
	assert.False(t, ilTwo.Equals(ilOne))
	ilOne.Name = ""
	assert.True(t, ilOne.Equals(ilTwo))
	assert.True(t, ilTwo.Equals(ilOne))
}

func TestMetricEquivalence(t *testing.T) {
	metric := func() Metric {
		return Metric{
			Name: "a_metric", Description: "a_metric_description",
			Unit: "a_metric_unit", Type: MetricType("a_metric_type"),
			Labels: &map[string]string{
				"one": "one", "two": "two",
			}, Value: 123,
		}
	}

	mOne := metric()
	assert.True(t, mOne.Equals(mOne))
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

	(*mTwo.Labels)["three"] = "three"
	assert.False(t, mOne.Equals(mTwo))
	assert.False(t, mTwo.Equals(mOne))
	(*mOne.Labels)["three"] = "three"
	assert.True(t, mOne.Equals(mTwo))
	assert.True(t, mTwo.Equals(mOne))
}

func TestMetricRelaxedEquivalence(t *testing.T) {
	lacksDescriptionUnitAndType := Metric{
		Name: "a_metric",
		Labels: &map[string]string{
			"one": "one", "two": "two",
		}, Value: 123,
	}

	completeMetric := Metric{
		Name: "a_metric", Description: "a_description",
		Unit: "a_metric_unit", Type: "a_metric_type",
		Labels: &map[string]string{
			"one": "one", "two": "two",
		}, Value: 123,
	}

	require.True(t, lacksDescriptionUnitAndType.RelaxedEquals(completeMetric))
	require.False(t, completeMetric.RelaxedEquals(lacksDescriptionUnitAndType))

	(*lacksDescriptionUnitAndType.Labels)["three"] = "three"
	require.False(t, lacksDescriptionUnitAndType.RelaxedEquals(completeMetric))
	require.False(t, completeMetric.RelaxedEquals(lacksDescriptionUnitAndType))
	(*completeMetric.Labels)["three"] = "three"
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

func TestMetricLabelRelaxedEquivalence(t *testing.T) {
	lackingLabels := Metric{
		Name: "a_metric", Description: "a_description",
		Unit: "a_metric_unit", Value: 123,
	}

	emptyLabels := Metric{
		Name: "a_metric", Description: "a_description",
		Unit: "a_metric_unit", Labels: &map[string]string{},
		Value: 123,
	}

	completeMetric := Metric{
		Name: "a_metric", Description: "a_description",
		Unit: "a_metric_unit", Type: "a_metric_type",
		Labels: &map[string]string{
			"one": "one", "two": "two",
		}, Value: 123,
	}

	require.True(t, lackingLabels.RelaxedEquals(completeMetric))
	require.False(t, emptyLabels.RelaxedEquals(completeMetric))
}

func TestHashFunctionConsistency(t *testing.T) {
	resource := Resource{Attributes: map[string]interface{}{
		"one": "1", "two": 2, "three": 3.000, "four": false, "five": nil,
	}}
	for i := 0; i < 100; i++ {
		require.Equal(t, "1f16e8e05a479e68c4fa5950471169e4", resource.Hash())
	}

	il := InstrumentationLibrary{Name: "some instrumentation library", Version: "some instrumentation version"}
	for i := 0; i < 100; i++ {
		require.Equal(t, "74ad4f0b1d06de1b45484cdbcfdd62db", il.Hash())
	}

	metric := Metric{
		Name: "some metric", Description: "some description",
		Unit: "some unit", Labels: &map[string]string{
			"labelOne": "1", "labelTwo": "two",
		}, Type: MetricType("some metric type"), Value: 123.456,
	}
	for i := 0; i < 100; i++ {
		require.Equal(t, "a481752281903890feb2c149573e0eaa", metric.Hash())
	}
}

func TestFlattenResourceMetricsByResourceIdentity(t *testing.T) {
	resource := Resource{Attributes: map[string]interface{}{"attribute_one": nil, "attribute_two": 123.456}}
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

func TestFlattenResourceMetricsByInstrumentationLibraryMetricsIdentity(t *testing.T) {
	resource := Resource{Attributes: map[string]interface{}{"attribute_three": true, "attribute_four": 23456}}
	ilm := InstrumentationLibraryMetrics{InstrumentationLibrary: InstrumentationLibrary{
		Name: "an instrumentation library", Version: "an instrumentation library version",
	}, Metrics: []Metric{}}
	resourceMetrics := ResourceMetrics{
		ResourceMetrics: []ResourceMetric{
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{}},
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{ilm}},
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{ilm, ilm}},
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{ilm, ilm, ilm}},
		},
	}
	expectedResourceMetrics := ResourceMetrics{
		ResourceMetrics: []ResourceMetric{
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{ilm}},
		},
	}
	require.Equal(t, expectedResourceMetrics, FlattenResourceMetrics(resourceMetrics))
}

func TestFlattenResourceMetricsByMetricsIdentity(t *testing.T) {
	resource := Resource{Attributes: map[string]interface{}{}}
	metrics := []Metric{
		{Name: "a metric", Unit: "a unit", Description: "a description", Value: 123},
		{Name: "another metric", Unit: "another unit", Description: "another description", Value: 234},
		{Name: "yet anothert metric", Unit: "yet anothe unit", Description: "yet anothet description", Value: 345},
	}
	ilm := InstrumentationLibraryMetrics{Metrics: metrics}
	ilmRepeated := InstrumentationLibraryMetrics{Metrics: append(metrics, metrics...)}
	ilmRepeatedTwice := InstrumentationLibraryMetrics{Metrics: append(metrics, append(metrics, metrics...)...)}
	ilmWithoutMetrics := InstrumentationLibraryMetrics{}
	resourceMetrics := ResourceMetrics{
		ResourceMetrics: []ResourceMetric{
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{}},
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{ilm}},
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{ilmRepeated}},
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{ilmRepeatedTwice}},
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{ilmWithoutMetrics}},
		},
	}
	expectedResourceMetrics := ResourceMetrics{
		ResourceMetrics: []ResourceMetric{
			{Resource: resource, ILMs: []InstrumentationLibraryMetrics{ilm}},
		},
	}
	require.Equal(t, expectedResourceMetrics, FlattenResourceMetrics(resourceMetrics))
}

func TestFlattenResourceMetricsConsistency(t *testing.T) {
	resourceMetrics, err := PDataToResourceMetrics(pdataMetrics())
	require.NoError(t, err)
	require.NotNil(t, resourceMetrics)
	require.Equal(t, resourceMetrics, FlattenResourceMetrics(resourceMetrics))
	var rms []ResourceMetrics
	for i := 0; i < 100; i++ {
		rms = append(rms, resourceMetrics)
	}
	for i := 0; i < 100; i++ {
		require.Equal(t, resourceMetrics, FlattenResourceMetrics(rms...))
	}
}

func TestContainsAllSelfCheck(t *testing.T) {
	resourceMetrics := loadedResourceMetrics(t)
	containsAll, err := resourceMetrics.ContainsAll(resourceMetrics)
	require.True(t, containsAll, err)
	require.NoError(t, err)
}

func TestContainsAllNoBijection(t *testing.T) {
	received := loadedResourceMetrics(t)

	expected, err := LoadResourceMetrics(path.Join(".", "testdata", "expectedMetrics.yaml"))
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
		"Missing Metrics: [another_int_gauge:::::IntGauge:456 another_double_gauge:::::DoubleGauge:567.89]",
	)
}

func TestContainsAllValueNeverReceived(t *testing.T) {
	received := loadedResourceMetrics(t)
	expected, err := LoadResourceMetrics(path.Join(".", "testdata", "neverReceivedMetrics.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// neverReceivedMetrics.yaml details a Metric with a value that isn't in resourceMetrics.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing Metrics: [another_int_gauge:::::IntGauge:111]")
}

func TestContainsAllInstrumentationLibraryNeverReceived(t *testing.T) {
	received := loadedResourceMetrics(t)
	expected, err := LoadResourceMetrics(path.Join(".", "testdata", "neverReceivedInstrumentationLibrary.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// neverReceivedMetrics.yaml details an Instrumentation Library that isn't in resourceMetrics.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing InstrumentationLibraries: [unmatched_instrumentation_library:]")
}

func TestContainsAllResourceNeverReceived(t *testing.T) {
	received := loadedResourceMetrics(t)
	expected, err := LoadResourceMetrics(path.Join(".", "testdata", "neverReceivedResource.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// neverReceivedMetrics.yaml details a Resource that isn't in resourceMetrics.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing resources: [not:matched]")
}

func TestContainsAllWithMissingAndEmptyLabels(t *testing.T) {
	received, err := LoadResourceMetrics(path.Join(".", "testdata", "labelValueResourceMetrics.yaml"))
	require.NoError(t, err)
	require.NotNil(t, received)

	unspecified, err := LoadResourceMetrics(path.Join(".", "testdata", "unspecifiedLabelsAllowed.yaml"))
	require.NoError(t, err)
	require.NotNil(t, unspecified)

	empty, err := LoadResourceMetrics(path.Join(".", "testdata", "emptyLabelsRequired.yaml"))
	require.NoError(t, err)
	require.NotNil(t, empty)

	containsAll, err := received.ContainsAll(*unspecified)
	require.True(t, containsAll)
	require.NoError(t, err)

	containsAll, err = received.ContainsAll(*empty)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing Metrics: [another_int_gauge:::::IntGauge:111]")
}
