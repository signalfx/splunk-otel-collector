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

package telemetry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResourceHashFunctionConsistency(t *testing.T) {
	resource := Resource{Attributes: map[string]any{
		"one": "1", "two": 2, "three": 3.000, "four": false, "five": nil,
	}}
	for i := 0; i < 100; i++ {
		require.Equal(t, "d3b92e5ff5847c43f397d5856f14c607", resource.Hash())
	}

	il := InstrumentationScope{Name: "some instrumentation library", Version: "some instrumentation version"}
	for i := 0; i < 100; i++ {
		require.Equal(t, "aa00805240d9717e6db7a0d88cf5e2ba", il.Hash())
	}
}
