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

func TestService_AppendExtension(t *testing.T) {
	svc := service{}
	svc.appendExtension("bbb")
	svc.appendExtension("aaa")
	assert.Equal(t, []string{"aaa", "bbb"}, svc.Extensions)
}

func TestRPE_Append(t *testing.T) {
	r := rpe{}
	r.appendReceiver("bbb")
	r.appendReceiver("aaa")
	assert.Equal(t, []string{"aaa", "bbb"}, r.Receivers)
}

func TestSAToOtelConfig(t *testing.T) {
	expected := map[string]interface{}{
		"type":     "vsphere",
		"host":     "localhost",
		"username": "administrator",
		"password": "abc123",
	}
	otelConfig := saInfoToOtelConfig(saCfgInfo{
		realm:       "us1",
		accessToken: "s3cr3t",
		monitors:    []interface{}{testvSphereMonitorCfg()},
	}, nil)
	require.Equal(t, expected, otelConfig.Receivers["smartagent/vsphere"])
}

func TestMonitorToReceiver(t *testing.T) {
	receiver, _ := saMonitorToOtelReceiver(testvSphereMonitorCfg())
	v, ok := receiver["smartagent/vsphere"]
	require.True(t, ok)
	assert.Equal(t, "vsphere", v["type"])
}

func TestMonitorToReceiver_Rule(t *testing.T) {
	otel, _ := saMonitorToOtelReceiver(map[interface{}]interface{}{
		"type":          "redis",
		"discoveryRule": `container_image =~ "redis" && port == 6379`,
	})
	redis := otel["smartagent/redis"]
	_, ok := redis["rule"]
	require.True(t, ok)
	_, ok = redis["config"]
	require.True(t, ok)
}

func TestAPIURLToRealm(t *testing.T) {
	us0, _ := apiURLToRealm(map[interface{}]interface{}{
		"apiUrl": "https://api.signalfx.com",
	})
	assert.Equal(t, "us0", us0)

	us1, _ := apiURLToRealm(map[interface{}]interface{}{
		"apiUrl": "https://api.us1.signalfx.com",
	})
	assert.Equal(t, "us1", us1)

	us2, _ := apiURLToRealm(map[interface{}]interface{}{
		"signalFxRealm": "us2",
	})
	assert.Equal(t, "us2", us2)
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

func TestInfoToOtelConfig_NoGlobalDims(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-simple.yaml")
	_, ok := oc.Processors["metricstransform"]
	assert.False(t, ok)
}

func TestInfoToOtelConfig_GlobalDims(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-complex.yaml")
	_, ok := oc.Processors["metricstransform"]
	assert.True(t, ok)
	mp := oc.Service.Pipelines["metrics"]
	assert.NotNil(t, mp.Processors)
}

func TestInfoToOtelConfig_CollectD(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-collectd.yaml")
	saExt, ok := oc.Extensions["smartagent"]
	require.True(t, ok)
	assert.Equal(t, 7, len(saExt))
	v := saExt["collectd"]
	collectd, ok := v.(map[interface{}]interface{}) // FIXME?
	require.True(t, ok)
	assert.Equal(t, 4, len(collectd))
	serviceExt := oc.Service.Extensions
	assert.Equal(t, 1, len(serviceExt))
}

func TestInfoToOtelConfig_ResourceDetectionProcessor(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-simple.yaml")
	assert.NotNil(t, oc.Processors)
	rdProc := oc.Processors["resourcedetection"]
	assert.Equal(t, map[string]interface{}{
		"detectors": []string{"system", "env", "gce", "ecs", "ec2", "azure"},
	}, rdProc)
	assert.Equal(t, []string{"resourcedetection"}, oc.Service.Pipelines["metrics"].Processors)
}

func TestInfoToOtelConfig_SFxForwarder(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-forwarder.yaml")
	receiverName := "smartagent/signalfx-forwarder"
	rcvr := oc.Receivers[receiverName]
	assert.Equal(t, map[string]interface{}{
		"type":          "signalfx-forwarder",
		"listenAddress": "0.0.0.0:9080",
	}, rcvr)
	pl := oc.Service.Pipelines["metrics"]
	assert.Contains(t, pl.Receivers, receiverName)
	tp := oc.Service.Pipelines["traces"]
	assert.Equal(t, []string{receiverName}, tp.Receivers)
	assert.Equal(t, []string{"resourcedetection"}, tp.Processors)
	assert.Equal(t, []string{"signalfx"}, tp.Exporters)
}

func TestInfoToOtelConfig_ProcessList(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-processlist.yaml")
	receiverName := "smartagent/processlist"
	rcvr := oc.Receivers[receiverName]
	assert.Equal(t, map[string]interface{}{
		"type": "processlist",
	}, rcvr)
	pl := oc.Service.Pipelines["logs"]
	assert.Equal(t, &rpe{
		Receivers:  []string{receiverName},
		Processors: []string{"resourcedetection"},
		Exporters:  []string{"signalfx"},
	}, pl)
}

func TestInfoToOtelConfig_Observers(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-observers.yaml")
	obs, ok := oc.Extensions["k8s_observer"]
	require.True(t, ok)
	assert.Equal(t, map[string]interface{}{
		"auth_type": "serviceAccount",
		"node":      "${K8S_NODE_NAME}",
	}, obs)
	assert.Equal(t, []string{"k8s_observer"}, oc.Service.Extensions)
}

func TestDiscoveryRuleToRCRule(t *testing.T) {
	rcr := discoveryRuleToRCRule(`container_image =~ "redis" && port == 6379`)
	assert.Equal(t, `type == "port" && pod.name matches "redis" && port == 6379`, rcr)
}

func TestInfoToOtelConfig_ZK(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-zk.yaml")
	v, ok := oc.ConfigSources["zookeeper"]
	require.True(t, ok)
	zk, ok := v.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, []interface{}{"127.0.0.1:2181"}, zk["endpoints"])
	assert.Equal(t, "10s", zk["timeout"])
}

func TestInfoToOtelConfig_Etcd(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-etcd.yaml")
	v, ok := oc.ConfigSources["etcd2"]
	require.True(t, ok)
	etcd, ok := v.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, []interface{}{"http://127.0.0.1:2379"}, etcd["endpoints"])
	auth, ok := etcd["auth"]
	require.True(t, ok)
	assert.Equal(t, map[string]interface{}{
		"username": "foo",
		"password": "bar",
	}, auth)
	r := oc.Receivers["smartagent/collectd/redis"]
	assert.Equal(t, "${etcd2:/redishost}", r["host"])
}

func TestInfoToOtelConfig_Vault(t *testing.T) {
	oc := yamlToOtelConfig(t, "testdata/sa-vault.yaml")
	v, ok := oc.ConfigSources["vault/0"]
	require.True(t, ok)
	vault, ok := v.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "http://127.0.0.1:8200", vault["endpoint"])
}

func testvSphereMonitorCfg() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"type":     "vsphere",
		"host":     "localhost",
		"username": "administrator",
		"password": "abc123",
	}
}

func yamlToOtelConfig(t *testing.T, filename string) *otelCfg {
	cfg := fromYAML(t, filename)
	expanded, vaultPaths, err := expandSA(cfg, "")
	require.NoError(t, err)
	info, err := saExpandedToCfgInfo(expanded)
	require.NoError(t, err)
	return saInfoToOtelConfig(info, vaultPaths)
}
