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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSAToOtelConfig(t *testing.T) {
	expected := map[interface{}]interface{}{
		"type":     "vsphere",
		"host":     "localhost",
		"username": "administrator",
		"password": "abc123",
	}
	otelConfig := saInfoToOtelConfig(saCfgInfo{
		realm:       "us1",
		accessToken: "s3cr3t",
		monitors:    []interface{}{testvSphereMonitorCfg()},
	})
	require.Equal(t, expected, otelConfig.Receivers["smartagent/vsphere"])
}

func TestMonitorToReceiver(t *testing.T) {
	receiver := saMonitorToOtelReceiver(testvSphereMonitorCfg())
	v, ok := receiver["smartagent/vsphere"]
	require.True(t, ok)
	m := v.(map[interface{}]interface{})
	assert.Equal(t, "vsphere", m["type"])
}

func testvSphereMonitorCfg() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"type":     "vsphere",
		"host":     "localhost",
		"username": "administrator",
		"password": "abc123",
	}
}

func TestAPIURLToRealm(t *testing.T) {
	us1, _ := apiURLToRealm("https://api.us1.signalfx.com")
	assert.Equal(t, "us1", us1)

	us0, _ := apiURLToRealm("https://api.signalfx.com")
	assert.Equal(t, "us0", us0)
}

func TestMTOperations(t *testing.T) {
	ops := mtOperations(map[interface{}]interface{}{
		"d": "3",
		"c": "2",
		"b": "1",
		"a": "0",
	})
	a := []string{"a", "b", "c", "d"}
	for i, op := range ops {
		assert.Equal(t, strconv.Itoa(i), op["new_value"])
		assert.Equal(t, a[i], op["new_label"])
	}
}

func TestDimsToMTP(t *testing.T) {
	block := dimsToMetricsTransformProcessor(map[interface{}]interface{}{
		"aaa": "bbb",
		"ccc": "ddd",
	})
	transforms := block["transforms"].([]map[interface{}]interface{})
	transform := transforms[0]
	assert.Equal(t, ".*", transform["include"])
	assert.Equal(t, "regexp", transform["match_type"])
	assert.Equal(t, "update", transform["action"])
	ops := transform["operations"].([]map[interface{}]interface{})
	assert.Equal(t, 2, len(ops))
	assert.Equal(t, map[interface{}]interface{}{
		"action":    "add_label",
		"new_label": "aaa",
		"new_value": "bbb",
	}, ops[0])
	assert.Equal(t, map[interface{}]interface{}{
		"action":    "add_label",
		"new_label": "ccc",
		"new_value": "ddd",
	}, ops[1])
}

func TestMetricsTransform_NoGlobalDims(t *testing.T) {
	cfg := fromYAML(t, "testdata/sa-simple.yaml")
	expanded, err := expandSA(cfg, "")
	require.NoError(t, err)
	info, err := saExpandedToCfgInfo(expanded)
	require.NoError(t, err)
	oc := saInfoToOtelConfig(info)
	_, ok := oc.Processors["metricstransform"]
	assert.False(t, ok)
}

func TestMetricsTransform_GlobalDims(t *testing.T) {
	cfg := fromYAML(t, "testdata/sa-complex.yaml")
	expanded, err := expandSA(cfg, "")
	require.NoError(t, err)
	info, err := saExpandedToCfgInfo(expanded)
	require.NoError(t, err)
	oc := saInfoToOtelConfig(info)
	_, ok := oc.Processors["metricstransform"]
	assert.True(t, ok)
	pipelines := oc.Service["pipelines"].(map[string]interface{})
	metrics := pipelines["metrics"].(rpe)
	assert.NotNil(t, metrics.Processors)
}

func TestMetricsTransform_CollectD(t *testing.T) {
	cfg := fromYAML(t, "testdata/sa-collectd.yaml")
	expanded, err := expandSA(cfg, "")
	require.NoError(t, err)
	info, err := saExpandedToCfgInfo(expanded)
	require.NoError(t, err)
	oc := saInfoToOtelConfig(info)
	v, ok := oc.Extensions["smartagent"]
	require.True(t, ok)
	saExt, ok := v.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 7, len(saExt))
	v = saExt["collectd"]
	collectd, ok := v.(map[interface{}]interface{})
	require.True(t, ok)
	assert.Equal(t, 4, len(collectd))
	v = oc.Service["extensions"]
	serviceExt, ok := v.([]string)
	require.True(t, ok)
	assert.Equal(t, 1, len(serviceExt))
}

func TestMetricsTransform_DeleteDiscoverRules(t *testing.T) {
	cfg := fromYAML(t, "testdata/sa-discoveryrules.yaml")
	expanded, err := expandSA(cfg, "")
	require.NoError(t, err)
	info, err := saExpandedToCfgInfo(expanded)
	require.NoError(t, err)
	oc := saInfoToOtelConfig(info)
	for k, v := range oc.Receivers {
		receiver := v.(map[interface{}]interface{})
		_, found := receiver["discoveryRule"]
		assert.False(t, found, "discoveryRule not deleted for %s", k)
	}
}
