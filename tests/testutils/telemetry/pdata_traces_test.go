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

//go:build testutils

package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPDataToResourceTracesHappyPath(t *testing.T) {
	resourceTraces, err := PDataToResourceTraces(PDataTraces())
	assert.NoError(t, err)
	require.NotNil(t, resourceTraces)

	resourceSpans := resourceTraces.ResourceSpans
	assert.Len(t, resourceSpans, 1)
	fstResourceSpans := resourceSpans[0]
	resourceAttributes := *fstResourceSpans.Resource.Attributes
	assertExpectedAttributes(t, resourceAttributes)

	scopeSpans := fstResourceSpans.ScopeSpans
	assert.Len(t, scopeSpans, 2)

	fstScopeSpans := fstResourceSpans.ScopeSpans[0]
	assert.Equal(t, "instrumentation_scope_one", fstScopeSpans.Scope.Name)
	assert.Equal(t, "instrumentation_scope_one_version", fstScopeSpans.Scope.Version)
	assert.Len(t, fstScopeSpans.Spans, 2)

	fstScopeFstSpan := fstScopeSpans.Spans[0]
	assert.Equal(t, "span_one", fstScopeFstSpan.Name)
	fstScopeFstSpanAttributes := *fstScopeFstSpan.Attributes
	assertExpectedAttributes(t, fstScopeFstSpanAttributes)

	fstScopeSndSpan := fstScopeSpans.Spans[1]
	assert.Equal(t, "span_two", fstScopeSndSpan.Name)

	sndScopeSpans := fstResourceSpans.ScopeSpans[1]
	assert.Equal(t, "instrumentation_scope_two", sndScopeSpans.Scope.Name)
	assert.Equal(t, "instrumentation_scope_two_version", sndScopeSpans.Scope.Version)
	assert.Len(t, sndScopeSpans.Spans, 1)

	sndScopeFstSpan := sndScopeSpans.Spans[0]
	assert.Equal(t, "span_one", sndScopeFstSpan.Name)
}

func assertExpectedAttributes(t *testing.T, attrs map[string]any) {
	assert.True(t, attrs["bool"].(bool))
	assert.Equal(t, "a_string", attrs["string"])
	assert.Equal(t, 123, attrs["int"])
	assert.Equal(t, 123.45, attrs["double"])
	assert.Nil(t, attrs["null"])
}
