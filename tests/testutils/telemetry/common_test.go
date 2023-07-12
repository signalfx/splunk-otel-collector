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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceHashFunctionConsistency(t *testing.T) {
	resource := Resource{Attributes: &map[string]any{
		"one": "1", "two": 2, "three": 3.000, "four": false, "five": nil,
	}}
	for i := 0; i < 100; i++ {
		require.Equal(t, "8c2edf4b5b71836ef95c2d64c200f30c", resource.Hash())
	}

	il := InstrumentationScope{Name: "some instrumentation library", Version: "some instrumentation version"}
	for i := 0; i < 100; i++ {
		require.Equal(t, "aa00805240d9717e6db7a0d88cf5e2ba", il.Hash())
	}
}

func TestEmptyResourcesAreEqual(t *testing.T) {
	rOne := Resource{Attributes: &map[string]any{}}
	rTwo := Resource{Attributes: &map[string]any{}}
	rThree := Resource{}

	require.True(t, rOne.Equals(rTwo))
	require.True(t, rTwo.Equals(rOne))
	require.True(t, rThree.Equals(rOne))
	// nil attrs aren't equal to empty map
	require.False(t, rTwo.Equals(rThree))

	for i := 0; i < 100; i++ {
		require.Equal(t, rOne.Hash(), rTwo.Hash())
	}
}

func TestResourceEquivalence(t *testing.T) {
	resource := func() Resource {
		return Resource{Attributes: &map[string]any{
			"one": 1, "two": "two", "three": nil,
			"four": []int{1, 2, 3, 4},
			"five": map[string]any{
				"true": true, "false": false, "nil": nil,
			},
		}}
	}
	rOne := resource()
	rOneSelf := rOne
	assert.True(t, rOne.Equals(rOneSelf))

	rTwo := resource()
	assert.True(t, rOne.Equals(rTwo))
	assert.True(t, rTwo.Equals(rOne))

	(*rTwo.Attributes)["five"].(map[string]any)["another"] = "item"
	assert.False(t, rOne.Equals(rTwo))
	assert.False(t, rTwo.Equals(rOne))
	(*rOne.Attributes)["five"].(map[string]any)["another"] = "item"
	assert.True(t, rOne.Equals(rTwo))
	assert.True(t, rTwo.Equals(rOne))
}

func TestInstrumentationScopeEquivalence(t *testing.T) {
	il := func() InstrumentationScope {
		return InstrumentationScope{
			Name: "an_instrumentation_scope", Version: "an_instrumentation_scope_version",
		}
	}

	ilOne := il()
	ilOneSelf := ilOne
	assert.True(t, ilOne.Equals(ilOneSelf))

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
