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
	"errors"
	"strconv"
	"testing"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
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

func TestInfoToOtelConfig_MetricsToExclude_Simple(t *testing.T) {
	cfg := yamlToOtelConfig(t, "testdata/sa-dp-to-exclude-simple.yaml")
	fp := cfg.Processors["filter"]
	require.NotNil(t, fp)
	metrics := fp["metrics"].(map[string]interface{})
	exclude := metrics["exclude"].(map[string]interface{})
	assert.Equal(t, "expr", exclude["match_type"])
	expressions := exclude["expressions"].([]string)
	assert.Equal(t, []string{
		`MetricName matches "^node_filesystem_.*$"`,
	}, expressions)
}

func TestInfoToOtelConfig_MetricsToExclude(t *testing.T) {
	cfg := yamlToOtelConfig(t, "testdata/sa-dp-to-exclude.yaml")
	fp := cfg.Processors["filter"]
	require.NotNil(t, fp)
	metrics := fp["metrics"].(map[string]interface{})
	exclude := metrics["exclude"].(map[string]interface{})
	assert.Equal(t, "expr", exclude["match_type"])
	expressions := exclude["expressions"].([]string)
	assert.Equal(t, 3, len(expressions))
	assert.Equal(t, `MetricName matches "^node_filesystem_.*$"`+
		` and not (MetricName matches "^node_filesystem_free_bytes$")`+
		` and not (MetricName matches "^node_filesystem_readonly$")`, expressions[0])
	assert.Equal(
		t,
		`MetricName matches "^node_network_.*$" and (Label("interface") matches "^.*$" and not (Label("interface") matches "^eth0$"))`,
		expressions[1],
	)
	assert.Equal(t, `MetricName matches "^node_disk_.*$" and (Label("device") matches "^sr.*$")`, expressions[2])
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

func TestSAFiltersToExpr_Simple(t *testing.T) {
	ex := saFiltersToExpr([]interface{}{
		map[interface{}]interface{}{
			"metricNames": []interface{}{
				"aaa",
			},
		},
	})
	assert.Equal(t, []string{`MetricName matches "^aaa$"`}, ex)
}

func TestSAFiltersToExpr_MetricNamesAndDimensions(t *testing.T) {
	ex := saFiltersToExpr([]interface{}{
		map[interface{}]interface{}{
			"metricNames": []interface{}{
				"aaa",
			},
			"dimensions": map[interface{}]interface{}{
				"foo": "bar",
			},
		},
	})
	expected := `MetricName matches "^aaa$" and (Label("foo") matches "^bar$")`
	assert.Equal(t, expected, ex[0])
}

func TestFilterTranslation(t *testing.T) {
	tests := []struct {
		name                        string
		metricFilters               []config.MetricFilter
		matchingDPs, nonMatchingDPs []*datapoint.Datapoint
	}{
		{
			name: "simple glob",
			metricFilters: []config.MetricFilter{
				{MetricName: "cpu.*"},
			},
			matchingDPs: []*datapoint.Datapoint{
				{Metric: "cpu.utilization"},
			},
			nonMatchingDPs: []*datapoint.Datapoint{
				{Metric: "foo"},
			},
		},
		{
			name: "glob with single override",
			metricFilters: []config.MetricFilter{{
				MetricNames: []string{
					"cpu.*", "!cpu.utilization",
				},
			}},
			matchingDPs: []*datapoint.Datapoint{
				{Metric: "cpu.user"},
			},
			nonMatchingDPs: []*datapoint.Datapoint{
				{Metric: "foo"},
				{Metric: "cpu.utilization"},
			},
		},
		{
			name: "glob with two overrides",
			metricFilters: []config.MetricFilter{{
				MetricNames: []string{
					"cpu.*", "!cpu.utilization", "!cpu.user",
				},
			}},
			matchingDPs: []*datapoint.Datapoint{
				{Metric: "cpu.sys"},
			},
			nonMatchingDPs: []*datapoint.Datapoint{
				{Metric: "cpu.user"},
				{Metric: "cpu.utilization"},
			},
		},
		{
			name: "multi filters",
			metricFilters: []config.MetricFilter{{
				MetricNames: []string{
					"cpu.*", "disk.*",
				},
			}},
			matchingDPs: []*datapoint.Datapoint{
				{Metric: "cpu.sys"},
				{Metric: "disk.reads"},
			},
		},
		{
			name: "dimension filter",
			metricFilters: []config.MetricFilter{{
				MetricNames: []string{"cpu.*"},
				Dimensions: map[string]interface{}{
					"host": "aaa",
				},
			}},
			matchingDPs: []*datapoint.Datapoint{
				{
					Metric: "cpu.sys",
					Dimensions: map[string]string{
						"host": "aaa",
					},
				},
			},
			nonMatchingDPs: []*datapoint.Datapoint{
				{
					Metric: "cpu.sys",
					Dimensions: map[string]string{
						"host": "bbb",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			saFilterSet, err := newSAFilterSet(test.metricFilters)
			require.NoError(t, err)

			mapRepr := metricFiltersToMapRepresentation(test.metricFilters)
			exprs := saFiltersToExpr(mapRepr)
			var programs []*vm.Program
			for _, ex := range exprs {
				program, err := expr.Compile(ex)
				require.NoError(t, err)
				programs = append(programs, program)
			}

			for _, dp := range test.matchingDPs {
				testFilter(t, assert.True, saFilterSet, dp, programs)
			}
			for _, dp := range test.nonMatchingDPs {
				testFilter(t, assert.False, saFilterSet, dp, programs)
			}
		})
	}
}

func metricFiltersToMapRepresentation(metricFilters []config.MetricFilter) []interface{} {
	var out []interface{}
	for _, filter := range metricFilters {
		names := filter.MetricNames
		if names == nil {
			names = []string{filter.MetricName}
		}
		out = append(out, map[interface{}]interface{}{
			"metricNames": func(strings []string) []interface{} {
				var v []interface{}
				for _, s := range strings {
					v = append(v, s)
				}
				return v
			}(names),
			"dimensions": stringMapToInterfaceMap(filter.Dimensions),
		})
	}
	return out
}

func TestMetricNamesToExpr(t *testing.T) {
	assert.Equal(
		t,
		`MetricName matches "^aaaa$"`,
		metricNamesToExpr([]interface{}{"aaaa"}),
	)
	assert.Equal(
		t,
		`MetricName matches "^aaaa$" or MetricName matches "^bbbb$"`,
		metricNamesToExpr([]interface{}{"aaaa", "bbbb"}),
	)
}

type assertionFunc func(t assert.TestingT, value bool, msgAndArgs ...interface{}) bool

func testFilter(
	t *testing.T,
	f assertionFunc,
	saFilterSet *dpfilters.FilterSet,
	dp *datapoint.Datapoint,
	programs []*vm.Program,
) {
	saMatched := saFilterSet.Matches(dp)
	f(t, saMatched)

	exprResult := testExpr(t, programs, dp)
	f(t, exprResult)
}

func testExpr(t *testing.T, programs []*vm.Program, dp *datapoint.Datapoint) bool {
	for _, program := range programs {
		v, err := expr.Run(program, map[string]interface{}{
			"MetricName": dp.Metric,
			"Label": func(key string) string {
				return dp.Dimensions[key]
			},
		})
		require.NoError(t, err)
		if v.(bool) {
			return true
		}
	}
	return false
}

// from SA codebase
func newSAFilterSet(excludes []config.MetricFilter) (*dpfilters.FilterSet, error) {
	var excludeSet []dpfilters.DatapointFilter
	for _, f := range excludes {
		if f.Negated {
			return nil, errors.New("new filters can't be negated")
		}
		dimSet, err := f.Normalize()
		if err != nil {
			return nil, err
		}

		dpf, err := dpfilters.NewOverridable(f.MetricNames, dimSet)
		if err != nil {
			return nil, err
		}

		excludeSet = append(excludeSet, dpf)
	}
	return &dpfilters.FilterSet{
		ExcludeFilters: excludeSet,
	}, nil
}

func TestDimsToExpr_OneDim(t *testing.T) {
	assert.Equal(
		t,
		`Label("interfaces") matches "^eth0$"`,
		dimsToExpr(map[interface{}]interface{}{
			"interfaces": []interface{}{"eth0"},
		}),
	)
}

func TestDimsToExpr_TwoDims(t *testing.T) {
	ex := dimsToExpr(map[interface{}]interface{}{
		"interfaces": []interface{}{"eth*", "!eth1"},
	})
	assert.Equal(
		t,
		`Label("interfaces") matches "^eth.*$" and not (Label("interfaces") matches "^eth1$")`,
		ex,
	)
}

func TestDimsToExpr_ThreeDims(t *testing.T) {
	ex := dimsToExpr(map[interface{}]interface{}{
		"interfaces": []interface{}{"eth*", "!eth1", "!eth2"},
	})
	assert.Equal(
		t,
		`Label("interfaces") matches "^eth.*$" and not (Label("interfaces") matches "^eth1$") and not (Label("interfaces") matches "^eth2$")`,
		ex,
	)
}

func TestSAFiltersToExpr_MetricNamesOnly(t *testing.T) {
	ex := saFiltersToExpr([]interface{}{
		map[interface{}]interface{}{
			"metricNames": []interface{}{
				"node_filesystem_*",
				"!node_filesystem_free_bytes",
				"!node_filesystem_readonly",
			},
		},
	})
	assert.Equal(t, []string{`MetricName matches "^node_filesystem_.*$" and not (MetricName matches "^node_filesystem_free_bytes$") and not (MetricName matches "^node_filesystem_readonly$")`}, ex)
}

func TestSAFiltersToExpr_MetricNameAndDims(t *testing.T) {
	ex := saFiltersToExpr([]interface{}{
		map[interface{}]interface{}{
			"metricName": "node_network_*",
			"dimensions": map[interface{}]interface{}{
				"interface": []interface{}{"*", "!eth0"},
			},
		},
	})
	assert.Equal(
		t,
		`MetricName matches "^node_network_.*$" and (Label("interface") matches "^.*$" and not (Label("interface") matches "^eth0$"))`,
		ex[0],
	)
}
