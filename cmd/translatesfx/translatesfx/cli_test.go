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

package translatesfx

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestTranslateConfig(t *testing.T) {
	translated := translateConfig("testdata/sa-e2e-simple-input.yaml", "")
	expected, err := os.ReadFile("testdata/otel-e2e-simple-expected.yaml")
	require.NoError(t, err)
	var translatedV, expectedV interface{}
	err = yaml.Unmarshal([]byte(translated), &translatedV)
	require.NoError(t, err)
	err = yaml.Unmarshal(expected, &expectedV)
	require.NoError(t, err)
	assert.Equal(t, expectedV, translatedV)
}
