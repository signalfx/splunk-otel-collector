// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deploymenttest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstrumentationVersionBoundaries(t *testing.T) {
	for _, tc := range []struct {
		version             string
		atLeast87, injector bool
	}{
		{"0.86.0", false, false},
		{"0.87.0", true, false},
		{"0.159.0", true, false},
		{"0.159.0-1", true, true},
		{"0.159.1", true, true},
		{"latest", true, true},
	} {
		t.Run(tc.version, func(t *testing.T) {
			require.Equal(t, tc.atLeast87, VersionAtLeast(tc.version, "0.87.0"))
			require.Equal(t, tc.injector, UsesInjector(tc.version))
		})
	}
}
