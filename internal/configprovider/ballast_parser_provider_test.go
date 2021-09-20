// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBallastParserProvider(t *testing.T) {
	tests := []struct {
		name        string
		fname       string
		cfgKey      string
		expectedVal int
	}{
		{
			name:        "percentage",
			fname:       "testdata/ballast_percentage.yaml",
			cfgKey:      "extensions::memory_ballast::size_in_percentage",
			expectedVal: 20,
		},
		{
			name:        "mib",
			fname:       "testdata/ballast_mib.yaml",
			cfgKey:      "extensions::memory_ballast::size_mib",
			expectedVal: 64,
		},
		{
			name:        "custom",
			fname:       "testdata/ballast_custom.yaml",
			cfgKey:      "extensions::memory_ballast/foo::size_in_percentage",
			expectedVal: 20,
		},
		{
			name:        "missing",
			fname:       "testdata/ballast_missing.yaml",
			cfgKey:      "extensions::memory_ballast::size_mib",
			expectedVal: 128,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bpp := &ballastParserProvider{
				pp:      &fileParserProvider{FileName: test.fname},
				sizeMib: 128,
			}
			cfgMap, err := bpp.Get()
			require.NoError(t, err)
			hasExt := hasBallastExtension(cfgMap)
			assert.True(t, hasExt)
			assert.Equal(t, test.expectedVal, cfgMap.Get(test.cfgKey))
		})
	}
}
