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

func TestSAExpandedToCfgInfo(t *testing.T) {
	cfg := yamlToCfgInfo(t, "testdata/sa-complex.yaml")
	assert.Equal(t, "https://ingest.us1.signalfx.com", cfg.ingestURL)
	assert.Equal(t, "https://api.us1.signalfx.com", cfg.APIURL)
	assert.Equal(t, "${include:testdata/token}", cfg.accessToken)
	assert.Equal(t, 3, len(cfg.monitors))
	assert.Equal(t, 2, len(cfg.globalDims))
}

func TestSAExpandedToConfigInfo_SASimple(t *testing.T) {
	cfg := yamlToCfgInfo(t, "testdata/sa-simple.yaml")
	assert.Nil(t, cfg.globalDims)
}

func TestSAExpandedToCfgInfo_Observers(t *testing.T) {
	cfg := yamlToCfgInfo(t, "testdata/sa-observers.yaml")
	assert.Equal(t, map[any]any{
		"type": "k8s-api",
	}, cfg.observers[0])
}

func TestSAExpandedToCfgInfo_ZK(t *testing.T) {
	cfg := yamlToCfgInfo(t, "testdata/sa-zk.yaml")
	zkPwd := cfg.monitors[0].(map[any]any)["port"].(string)
	assert.Equal(t, "${zookeeper:/redis/port}", zkPwd)
}

func TestToString(t *testing.T) {
	assert.Equal(t, "bar", toString(map[any]any{"foo": "bar"}, "foo"))
	assert.Equal(t, "", toString(map[any]any{"foo": "bar"}, "baz"))
	assert.Equal(t, "", toString(map[any]any{"foo": nil}, "foo"))
}

func yamlToCfgInfo(t *testing.T, filename string) saCfgInfo {
	v := fromYAML(t, filename)
	expanded, _, err := expandSA(v, "")
	require.NoError(t, err)
	return saExpandedToCfgInfo(expanded)
}

func fromYAML(t *testing.T, filename string) any {
	yml, err := os.ReadFile(filename)
	require.NoError(t, err)
	var v any
	err = yaml.UnmarshalStrict(yml, &v)
	require.NoError(t, err)
	return v
}
