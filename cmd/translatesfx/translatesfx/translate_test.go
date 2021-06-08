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
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestExpandedToCfgInfo(t *testing.T) {
	yml, err := ioutil.ReadFile("testdata/sa-complex.yaml")
	require.NoError(t, err)
	var v interface{}
	err = yaml.UnmarshalStrict(yml, &v)
	require.NoError(t, err)
	expanded, _ := expandSA(v, "")
	cfg := saExpandedToCfgInfo(expanded.(map[interface{}]interface{}), "")
	assert.Equal(t, "us1", cfg.realm)
	assert.Equal(t, "${include:testdata/token}", cfg.accessToken)
	assert.Equal(t, 3, len(cfg.monitors))
}
