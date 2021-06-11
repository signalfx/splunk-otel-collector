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

func TestSAExpandedToCfgInfo(t *testing.T) {
	v := fromYAML(t, "testdata/sa-complex.yaml")
	expanded, err := expandSA(v, "")
	require.NoError(t, err)
	cfg, err := saExpandedToCfgInfo(expanded)
	require.NoError(t, err)
	require.NoError(t, err)
	assert.Equal(t, "us1", cfg.realm)
	assert.Equal(t, "${include:testdata/token}", cfg.accessToken)
	assert.Equal(t, 3, len(cfg.monitors))
	assert.Equal(t, 2, len(cfg.globalDims))
}

func TestSAExpandedToConfigInfo_SASimple(t *testing.T) {
	v := fromYAML(t, "testdata/sa-simple.yaml")
	expanded, err := expandSA(v, "")
	require.NoError(t, err)
	cfg, err := saExpandedToCfgInfo(expanded)
	require.NoError(t, err)
	assert.Nil(t, cfg.globalDims)
}

func fromYAML(t *testing.T, filename string) interface{} {
	yml, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	var v interface{}
	err = yaml.UnmarshalStrict(yml, &v)
	require.NoError(t, err)
	return v
}
