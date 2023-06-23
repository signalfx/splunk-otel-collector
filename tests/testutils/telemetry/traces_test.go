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

func TestSaveAndLoadCycle(t *testing.T) {
	yamlFile := filepath.Join(".", "testdata", "traces", "tmp-save-and-load-cycle.yaml")

	toSaveResourceTraces, err := PDataToResourceTraces(PDataTraces())
	require.NoError(t, err)
	require.NotNil(t, toSaveResourceTraces)
	require.NoError(t, toSaveResourceTraces.SaveResourceTraces(yamlFile))
	defer os.Remove(yamlFile)

	loadedResourceTraces, err := LoadResourceTraces(yamlFile)
	require.NoError(t, err)
	require.NotNil(t, loadedResourceTraces)

	containsAll, err := loadedResourceTraces.ContainsAll(toSaveResourceTraces)
	require.NoError(t, err)
	require.True(t, containsAll)
}

func loadedResourceTraces(t *testing.T) ResourceTraces {
	yamlFile := filepath.Join(".", "testdata", "traces", "resource-traces.yaml")
	resourceTraces, err := LoadResourceTraces(yamlFile)
	require.NoError(t, err)
	require.NotNil(t, resourceTraces)
	return *resourceTraces
}

func TestResourceTracesYamlStringRep(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(".", "testdata", "traces", "resource-traces.yaml"))
	require.NoError(t, err)
	resourceTraces := loadedResourceTraces(t)
	require.Equal(t, string(b), fmt.Sprintf("%v", resourceTraces))
}

func TestLoadTracesHappyPath(t *testing.T) {
	resourceTraces := loadedResourceTraces(t)

	assert.Equal(t, 2, len(resourceTraces.ResourceSpans))

	firstRSs := resourceTraces.ResourceSpans[0]
	firstRSsAttrs := *firstRSs.Resource.Attributes
	require.Equal(t, 2, len(firstRSsAttrs))
	require.NotNil(t, firstRSsAttrs["one_attr"])
	assert.Equal(t, "one_value", firstRSsAttrs["one_attr"])
	require.NotNil(t, firstRSsAttrs["two_attr"])
	assert.Equal(t, "two_value", firstRSsAttrs["two_attr"])

	assert.Equal(t, 2, len(firstRSs.ScopeSpans))
	firstRSsFirstSSs := firstRSs.ScopeSpans[0]
	require.NotNil(t, firstRSsFirstSSs)
	require.NotNil(t, firstRSsFirstSSs.Scope)
	assert.Equal(t, "scope_one", firstRSsFirstSSs.Scope.Name)
	assert.Equal(t, "scope_one_version", firstRSsFirstSSs.Scope.Version)
	firstRSsFirstSSsScopeAttrs := *firstRSsFirstSSs.Scope.Attributes
	require.Equal(t, 1, len(firstRSsFirstSSsScopeAttrs))
	require.NotNil(t, firstRSsAttrs["one_attr"])
	assert.Equal(t, "one_value", firstRSsAttrs["one_attr"])

	spans := firstRSsFirstSSs.Spans
	assert.Equal(t, 2, len(spans))
	require.NotNil(t, spans[0])
	assert.Equal(t, "span_one", spans[0].Name)
	spanAttrs := *spans[0].Attributes
	require.Equal(t, 2, len(spanAttrs))
	require.NotNil(t, spanAttrs["one_attr"])
	assert.Equal(t, "one_value", spanAttrs["one_attr"])
	require.NotNil(t, spanAttrs["two_attr"])
	assert.Equal(t, "two_value", spanAttrs["two_attr"])

	require.NotNil(t, spans[1])
	assert.Equal(t, "span_two", spans[1].Name)
	spanAttrs = *spans[1].Attributes
	assert.NotNil(t, spanAttrs)
	assert.Equal(t, 0, len(spanAttrs))

	firstRSsSecondSSs := firstRSs.ScopeSpans[1]
	require.NotNil(t, firstRSsSecondSSs)
	require.NotNil(t, firstRSsSecondSSs.Scope)
	assert.Equal(t, "instrumentation_scope_two", firstRSsSecondSSs.Scope.Name)
	assert.Equal(t, "instrumentation_scope_two_version", firstRSsSecondSSs.Scope.Version)
	firstRSsSecondSSsScopeAttrs := *firstRSsSecondSSs.Scope.Attributes
	assert.NotNil(t, firstRSsSecondSSsScopeAttrs)
	assert.Equal(t, 0, len(firstRSsSecondSSsScopeAttrs))

	spans = firstRSsSecondSSs.Spans
	assert.Equal(t, 1, len(spans))
	require.NotNil(t, spans[0])
	assert.Equal(t, "span_one", spans[0].Name)
	spanAttrs = *spans[0].Attributes
	assert.NotNil(t, spanAttrs)
	assert.Equal(t, 0, len(spanAttrs))

	secondRSs := resourceTraces.ResourceSpans[1]
	secondRSsAttrs := *secondRSs.Resource.Attributes
	require.Equal(t, 1, len(secondRSsAttrs))
	require.NotNil(t, secondRSsAttrs["some_attr"])
	assert.Equal(t, "some_value", secondRSsAttrs["some_attr"])

	assert.Equal(t, 1, len(secondRSs.ScopeSpans))
	secondRSsFirstSSs := secondRSs.ScopeSpans[0]
	require.NotNil(t, secondRSsFirstSSs)
	require.NotNil(t, secondRSsFirstSSs.Scope)
	assert.Equal(t, "scope_one", secondRSsFirstSSs.Scope.Name)
	assert.Equal(t, "scope_one_version", secondRSsFirstSSs.Scope.Version)
	assert.Nil(t, secondRSsFirstSSs.Scope.Attributes)

	spans = secondRSsFirstSSs.Spans
	assert.Equal(t, 1, len(spans))
	require.NotNil(t, spans[0])
	assert.Equal(t, "last_span", spans[0].Name)
	assert.Nil(t, spans[0].Attributes)
}

func TestLoadResourcesNotAValidPath(t *testing.T) {
	resourceTraces, err := LoadResourceTraces("notafile")
	require.Error(t, err)
	require.Contains(t, err.Error(), invalidPathErrorMsg())
	require.Nil(t, resourceTraces)
}

func TestLoadTracesInvalidItems(t *testing.T) {
	resourceTraces, err := LoadResourceTraces(filepath.Join(".", "testdata", "traces", "invalid-resource-traces.yaml"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "field notAttributesOrScopeSpans not found in type telemetry.ResourceSpans")
	require.Nil(t, resourceTraces)
}

func TestFlattenResourceTracesByResourceIdentity(t *testing.T) {
	resource := Resource{Attributes: &map[string]any{"attribute_one": nil, "attribute_two": 123.456}}
	resourceTraces := ResourceTraces{
		ResourceSpans: []ResourceSpans{
			{Resource: resource},
			{Resource: resource},
			{Resource: resource},
		},
	}
	expectedResourceTraces := ResourceTraces{ResourceSpans: []ResourceSpans{{Resource: resource}}}
	require.Equal(t, expectedResourceTraces, FlattenResourceTraces(resourceTraces))
}

func TestFlattenResourceTracesByScopeSpansIdentity(t *testing.T) {
	resource := Resource{Attributes: &map[string]any{"attribute_three": true, "attribute_four": 23456}}
	ss := ScopeSpans{Scope: InstrumentationScope{
		Name: "an instrumentation library", Version: "an instrumentation library version",
	}, Spans: []Span{}}
	resourceTraces := ResourceTraces{
		ResourceSpans: []ResourceSpans{
			{Resource: resource, ScopeSpans: []ScopeSpans{}},
			{Resource: resource, ScopeSpans: []ScopeSpans{ss}},
			{Resource: resource, ScopeSpans: []ScopeSpans{ss, ss}},
			{Resource: resource, ScopeSpans: []ScopeSpans{ss, ss, ss}},
		},
	}
	expectedResourceTraces := ResourceTraces{
		ResourceSpans: []ResourceSpans{
			{Resource: resource, ScopeSpans: []ScopeSpans{ss}},
		},
	}
	require.Equal(t, expectedResourceTraces, FlattenResourceTraces(resourceTraces))
}

func TestFlattenResourceTracesBySpansIdentity(t *testing.T) {
	resource := Resource{}
	spans := []Span{
		{Name: "a span"},
		{Name: "another span", Attributes: &map[string]any{}},
		{Name: "yet another span", Attributes: &map[string]any{"attr": "value"}},
	}
	ss := ScopeSpans{Spans: spans}
	resourceTraces := ResourceTraces{
		ResourceSpans: []ResourceSpans{
			{Resource: resource, ScopeSpans: []ScopeSpans{}},
			{Resource: resource, ScopeSpans: []ScopeSpans{{Spans: spans[0:1]}}},
			{Resource: resource, ScopeSpans: []ScopeSpans{{Spans: spans[1:3]}}},
		},
	}
	expectedResourceTraces := ResourceTraces{
		ResourceSpans: []ResourceSpans{
			{Resource: resource, ScopeSpans: []ScopeSpans{ss}},
		},
	}
	require.Equal(t, expectedResourceTraces, FlattenResourceTraces(resourceTraces))
}

func TestTraceContainsAllSelfCheck(t *testing.T) {
	resourceTraces := loadedResourceTraces(t)
	containsAll, err := resourceTraces.ContainsAll(resourceTraces)
	require.True(t, containsAll)
	require.NoError(t, err)
}

func TestTraceContainsAllNoBijection(t *testing.T) {
	received := loadedResourceTraces(t)

	expected, err := LoadResourceTraces(filepath.Join(".", "testdata", "traces", "expected-traces.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	containsAll, err := received.ContainsAll(*expected)
	require.True(t, containsAll, err)
	require.NoError(t, err)

	// Since expected-traces.yaml specify a single resource spans, they will never find resource spans from received.
	containsAll, err = expected.ContainsAll(received)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing Spans: [attributes:\n  one_attr: one_value\n  two_attr: two_value\nname: span_one\n]")
}

func TestTraceContainsAllSpansNeverReceived(t *testing.T) {
	received := loadedResourceTraces(t)
	expected, err := LoadResourceTraces(filepath.Join(".", "testdata", "traces", "never-received-spans.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// never-received-spans.yaml contains a span that isn't in resource-metrics.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing Spans: [name: missing_span\n]")
}

func TestTraceContainsAllInstrumentationScopeNeverReceived(t *testing.T) {
	received := loadedResourceTraces(t)
	expected, err := LoadResourceTraces(filepath.Join(".", "testdata", "traces", "never-received-instrumentation-scope.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// never-received-instrumentation-scope.yaml details an InstrumentationScope that isn't in resource-traces.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing InstrumentationScopes: [name: unmatched_instrumentation_scope\nversion: scope_one_version\n]")
}

func TestTraceContainsAllResourceNeverReceived(t *testing.T) {
	received := loadedResourceTraces(t)
	expected, err := LoadResourceTraces(filepath.Join(".", "testdata", "traces", "never-received-resource.yaml"))
	require.NoError(t, err)
	require.NotNil(t, expected)

	// never-received-resource.yaml details a Resource that isn't in resource-traces.yaml
	containsAll, err := received.ContainsAll(*expected)
	require.False(t, containsAll)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Missing resources: [attributes:\n  some_attr: different_value\n]")
}
