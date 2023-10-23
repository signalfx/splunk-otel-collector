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
	"sort"
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
	expected := map[string]any{
		"type":     "vsphere",
		"host":     "localhost",
		"username": "administrator",
		"password": "abc123",
	}
	otelConfig, w := saInfoToOtelConfig(saCfgInfo{
		realm:       "us1",
		accessToken: "s3cr3t",
		monitors:    []any{testvSphereMonitorCfg()},
	}, nil)
	assert.Nil(t, w)
	require.Equal(t, expected, otelConfig.Receivers["smartagent/vsphere"])
}

func TestMonitorToReceiver(t *testing.T) {
	cmp, w, isRC := saMonitorToOtelReceiver(testvSphereMonitorCfg(), nil)
	assert.Nil(t, w)
	assert.False(t, isRC)
	assert.Equal(t, "smartagent/vsphere", cmp.baseName)
	assert.Equal(t, "vsphere", cmp.attrs["type"])
}

func testvSphereMonitorCfg() map[any]any {
	return map[any]any{
		"type":     "vsphere",
		"host":     "localhost",
		"username": "administrator",
		"password": "abc123",
	}
}

func TestMonitorToReceiver_Rule(t *testing.T) {
	cmp, w, isRC := saMonitorToOtelReceiver(map[any]any{
		"type":          "redis",
		"discoveryRule": `target == "hostport" && container_image =~ "redis" && port == 6379`,
	}, nil)
	assert.Nil(t, w)
	assert.True(t, isRC)
	assert.Equal(t, "smartagent/redis", cmp.baseName)
	_, ok := cmp.attrs["rule"]
	require.True(t, ok)
	_, ok = cmp.attrs["config"]
	require.True(t, ok)
}

func TestMTOperations(t *testing.T) {
	ops := mtOperations(map[any]any{
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
	block := dimsToMetricsTransformProcessor(map[any]any{
		"aaa": "bbb",
		"ccc": "ddd",
	})
	transforms := block["transforms"].([]map[any]any)
	transform := transforms[0]
	assert.Equal(t, ".*", transform["include"])
	assert.Equal(t, "regexp", transform["match_type"])
	assert.Equal(t, "update", transform["action"])
	ops := transform["operations"].([]map[any]any)
	assert.Equal(t, 2, len(ops))
	assert.Equal(t, map[any]any{
		"action":    "add_label",
		"new_label": "aaa",
		"new_value": "bbb",
	}, ops[0])
	assert.Equal(t, map[any]any{
		"action":    "add_label",
		"new_label": "ccc",
		"new_value": "ddd",
	}, ops[1])
}

func TestInfoToOtelConfig_NoGlobalDims(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-simple.yaml")
	_, ok := oc.Processors["metricstransform"]
	assert.False(t, ok)
}

func TestInfoToOtelConfig_GlobalDims(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-complex.yaml")
	_, ok := oc.Processors["metricstransform"]
	assert.True(t, ok)
	mp := oc.Service.Pipelines["metrics"]
	assert.NotNil(t, mp.Processors)
}

func TestInfoToOtelConfig_CollectD(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-collectd.yaml")
	saExt, ok := oc.Extensions["smartagent"]
	require.True(t, ok)
	assert.Equal(t, 7, len(saExt))
	v := saExt["collectd"]
	collectd, ok := v.(map[any]any) // FIXME?
	require.True(t, ok)
	assert.Equal(t, 4, len(collectd))
	serviceExt := oc.Service.Extensions
	assert.Equal(t, 1, len(serviceExt))
}

func TestInfoToOtelConfig_ResourceDetectionProcessor(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-simple.yaml")
	assert.NotNil(t, oc.Processors)
	rdProc := oc.Processors["resourcedetection"]
	assert.Equal(t, map[string]any{
		"detectors": []string{"system", "env", "gcp", "ecs", "ec2", "azure"},
	}, rdProc)
	assert.Equal(t, []string{"resourcedetection"}, oc.Service.Pipelines["metrics"].Processors)
}

func TestInfoToOtelConfig_SFxForwarder(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-forwarder.yaml")
	receiverName := "smartagent/signalfx-forwarder"
	rcvr := oc.Receivers[receiverName]
	assert.Equal(t, map[string]any{
		"type":          "signalfx-forwarder",
		"listenAddress": "0.0.0.0:9080",
	}, rcvr)
	pl := oc.Service.Pipelines["metrics"]
	assert.Contains(t, pl.Receivers, receiverName)
	tp := oc.Service.Pipelines["traces"]
	assert.Equal(t, []string{receiverName}, tp.Receivers)
	assert.Equal(t, []string{"resourcedetection"}, tp.Processors)
	assert.Equal(t, []string{"sapm", "signalfx"}, tp.Exporters)
}

func TestInfoToOtelConfig_ProcessList(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-processlist.yaml")
	processListReceiverName := "smartagent/processlist"
	rcvr := oc.Receivers[processListReceiverName]
	assert.Equal(t, map[string]any{
		"type": "processlist",
	}, rcvr)

	kubernetesEventsReceiverName := "smartagent/kubernetes-events"
	rcvr = oc.Receivers[kubernetesEventsReceiverName]
	assert.Equal(t, map[string]any{
		"type": "kubernetes-events",
	}, rcvr)

	pl := oc.Service.Pipelines["logs"]
	assert.Equal(t, &rpe{
		Receivers:  []string{kubernetesEventsReceiverName, processListReceiverName},
		Processors: []string{"resourcedetection"},
		Exporters:  []string{"signalfx"},
	}, pl)

	_, ok := oc.Service.Pipelines["metrics"]
	assert.False(t, ok)
}

func TestInfoToOtelConfig_Observers(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-observers.yaml")
	obs, ok := oc.Extensions["k8s_observer"]
	require.True(t, ok)
	assert.Equal(t, map[string]any{
		"auth_type": "serviceAccount",
		"node":      "${K8S_NODE_NAME}",
	}, obs)
	assert.Equal(t, []string{"k8s_observer"}, oc.Service.Extensions)
	v := oc.Receivers["receiver_creator"]["receivers"]
	m := v.(map[string]map[string]any)
	rr := m["smartagent/collectd/redis"]
	assert.Equal(t, `type == "port" && pod.name == "redis" && port == 6379`, rr["rule"])
}

func TestInfoToOtelConfig_ZK(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-zk.yaml")
	v, ok := oc.ConfigSources["zookeeper"]
	require.True(t, ok)
	zk, ok := v.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []any{"127.0.0.1:2181"}, zk["endpoints"])
	assert.Equal(t, "10s", zk["timeout"])
}

func TestInfoToOtelConfig_Etcd(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-etcd.yaml")
	v, ok := oc.ConfigSources["etcd2"]
	require.True(t, ok)
	etcd, ok := v.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []any{"http://127.0.0.1:2379"}, etcd["endpoints"])
	auth, ok := etcd["auth"]
	require.True(t, ok)
	assert.Equal(t, map[string]any{
		"username": "foo",
		"password": "bar",
	}, auth)
	r := oc.Receivers["smartagent/collectd/redis"]
	assert.Equal(t, "${etcd2:/redishost}", r["host"])
}

func TestInfoToOtelConfig_Vault(t *testing.T) {
	oc, _ := yamlToOtelConfig(t, "testdata/sa-vault.yaml")
	v, ok := oc.ConfigSources["vault/0"]
	require.True(t, ok)
	vault, ok := v.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "http://127.0.0.1:8200", vault["endpoint"])
}

func TestInfoToOtelConfig_MetricsToExclude_Simple(t *testing.T) {
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-metrics-to-exclude-simple.yaml")
	fp := cfg.Processors["filter"]
	require.NotNil(t, fp)
	metrics := fp["metrics"].(map[string]any)
	exclude := metrics["exclude"].(map[string]any)
	assert.Equal(t, "expr", exclude["match_type"])
	expressions := exclude["expressions"].([]string)
	assert.Equal(t, []string{
		`MetricName matches "^node_filesystem_.*$"`,
	}, expressions)
}

func TestInfoToOtelConfig_MetricsToExclude(t *testing.T) {
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-metrics-to-exclude.yaml")
	fp := cfg.Processors["filter"]
	require.NotNil(t, fp)
	metrics := fp["metrics"].(map[string]any)
	exclude := metrics["exclude"].(map[string]any)
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

func TestInfoToOtelConfig_MetricsToExclude_Regex(t *testing.T) {
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-metrics-to-exclude-regex.yaml")
	expression := cfg.Processors["filter"]["metrics"].(map[string]any)["exclude"].(map[string]any)["expressions"].([]string)[0]
	assert.Equal(t, `MetricName matches "vsphere\\.cpu_\\w*_percent"`, expression)
}

func TestInfoToOtelConfig_MetricsToExclude_Monitor(t *testing.T) {
	// This metricsToExclude is attached to the monitor, not the agent, which is
	// apparently an invalid configuration according to the docs, but it actually
	// works and some customers are using it. If we just rename it to the supported
	// datapointsToExclude, things work fine in the translated config.
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-metrics-to-exclude-monitor.yaml")

	_, ok := cfg.Processors["filter"]
	assert.False(t, ok)

	saReceiver := cfg.Receivers["smartagent/cpu"]
	_, ok = saReceiver["metricsToExclude"]
	assert.False(t, ok)

	ex, ok := saReceiver["datapointsToExclude"]
	assert.True(t, ok)
	assert.Equal(t, []any{
		map[any]any{
			"metricNames": []any{"foo*"},
		},
	}, ex)
}

func TestComponentCollection_Single(t *testing.T) {
	cc := componentCollection{{
		baseName: "mycomponent",
		attrs:    map[string]any{"foo": "bar"},
	}}
	componentMap := cc.toComponentMap()
	var keys []string
	for k := range componentMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	assert.Equal(t, []string{"mycomponent"}, keys)
}

func TestComponentCollection_Multiple(t *testing.T) {
	cc := componentCollection{{
		baseName: "mycomponent",
		attrs:    map[string]any{"foo": "bar"},
	}, {
		baseName: "mycomponent",
		attrs:    map[string]any{"foo": "bar"},
	}}
	componentMap := cc.toComponentMap()
	var keys []string
	for k := range componentMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	assert.Equal(t, []string{"mycomponent/0", "mycomponent/1"}, keys)
}

func TestInfoToOtelConfig_DuplicateMonitors(t *testing.T) {
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-duplicate-monitors.yaml")
	assert.Equal(t, 2, len(cfg.Receivers))
	metrics := cfg.Service.Pipelines["metrics"]
	const sa0 = "smartagent/sql/0"
	const sa1 = "smartagent/sql/1"
	assert.Equal(t, []string{sa0, sa1}, metrics.Receivers)
	receiver0 := cfg.Receivers[sa0]
	_, found := receiver0["connectionString"]
	assert.True(t, found)
	assert.Equal(t, 7, len(receiver0))
	receiver1 := cfg.Receivers[sa1]
	assert.Equal(t, map[any]any{"user": "postgres", "password": "s3cr3t"}, receiver1["params"])
}

func yamlToOtelConfig(t *testing.T, filename string) (out *otelCfg, warnings []error) {
	cfg := fromYAML(t, filename)
	expanded, vaultPaths, err := expandSA(cfg, "")
	require.NoError(t, err)
	info := saExpandedToCfgInfo(expanded)
	require.NoError(t, err)
	return saInfoToOtelConfig(info, vaultPaths)
}

func TestSAExcludesToExpr_Simple(t *testing.T) {
	ex := saExcludesToExpr([]any{
		map[any]any{
			"metricNames": []any{
				"aaa",
			},
		},
	}, nil, false)
	assert.Equal(t, []string{`MetricName matches "^aaa$"`}, ex)
}

func TestSAExcludesToExpr_MetricNamesAndDimensions(t *testing.T) {
	ex := saExcludesToExpr([]any{
		map[any]any{
			"metricNames": []any{
				"aaa",
			},
			"dimensions": map[any]any{
				"foo": "bar",
			},
		},
	}, nil, false)
	expected := `MetricName matches "^aaa$" and (Label("foo") matches "^bar$")`
	assert.Equal(t, expected, ex[0])
}

func TestMetricNamesToExpr(t *testing.T) {
	assert.Equal(
		t,
		`MetricName matches "^aaaa$"`,
		metricNamesToExpr([]any{"aaaa"}, false),
	)
	assert.Equal(
		t,
		`MetricName matches "^aaaa$" or MetricName matches "^bbbb$"`,
		metricNamesToExpr([]any{"aaaa", "bbbb"}, false),
	)
}

func TestDimsToExpr_OneDim(t *testing.T) {
	assert.Equal(
		t,
		`Label("interfaces") matches "^eth0$"`,
		dimsToExpr(map[any]any{
			"interfaces": []any{"eth0"},
		}),
	)
}

func TestDimsToExpr_TwoDims(t *testing.T) {
	ex := dimsToExpr(map[any]any{
		"interfaces": []any{"eth*", "!eth1"},
	})
	assert.Equal(
		t,
		`Label("interfaces") matches "^eth.*$" and not (Label("interfaces") matches "^eth1$")`,
		ex,
	)
}

func TestDimsToExpr_ThreeDims(t *testing.T) {
	ex := dimsToExpr(map[any]any{
		"interfaces": []any{"eth*", "!eth1", "!eth2"},
	})
	assert.Equal(
		t,
		`Label("interfaces") matches "^eth.*$" and not (Label("interfaces") matches "^eth1$") and not (Label("interfaces") matches "^eth2$")`,
		ex,
	)
}

func TestSAExcludesToExpr_MetricNamesOnly(t *testing.T) {
	ex := saExcludesToExpr([]any{
		map[any]any{
			"metricNames": []any{
				"node_filesystem_*",
				"!node_filesystem_free_bytes",
				"!node_filesystem_readonly",
			},
		},
	}, nil, false)
	assert.Equal(t, []string{`MetricName matches "^node_filesystem_.*$" and not (MetricName matches "^node_filesystem_free_bytes$") and not (MetricName matches "^node_filesystem_readonly$")`}, ex)
}

func TestSAExcludesToExpr_MetricNameAndDims(t *testing.T) {
	ex := saExcludesToExpr([]any{
		map[any]any{
			"metricName": "node_network_*",
			"dimensions": map[any]any{
				"interface": []any{"*", "!eth0"},
			},
		},
	}, nil, false)
	assert.Equal(
		t,
		`MetricName matches "^node_network_.*$" and (Label("interface") matches "^.*$" and not (Label("interface") matches "^eth0$"))`,
		ex[0],
	)
}

func TestOptionalForwarder(t *testing.T) {
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-simple.yaml")
	_, ok := cfg.Service.Pipelines["traces"]
	assert.False(t, ok)
	_, ok = cfg.Receivers["smartagent/signalfx-forwarder"]
	assert.False(t, ok)
}

func TestSFxExporterUrls(t *testing.T) {
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-collectd.yaml")
	exporter := cfg.Exporters["signalfx"]
	assert.Equal(t, "https://ingest.us1.signalfx.com", exporter["ingest_url"])
	assert.Equal(t, "https://api.us1.signalfx.com", exporter["api_url"])
}

func TestNoProcessList(t *testing.T) {
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-simple.yaml")
	_, ok := cfg.Service.Pipelines["logs"]
	assert.False(t, ok)
}

func TestSAToRegexpStr(t *testing.T) {
	regexStr, b := saToRegexpStr("foo")
	assert.False(t, b)
	assert.Equal(t, "^foo$", regexStr)

	regexStr, b = saToRegexpStr("bar_*")
	assert.False(t, b)
	assert.Equal(t, "^bar_.*$", regexStr)

	regexStr, b = saToRegexpStr("/baz_.*/")
	assert.False(t, b)
	assert.Equal(t, "baz_.*", regexStr)
}

func TestIsRegexFilter(t *testing.T) {
	assert.False(t, isRegexFilter("fooby"))
	assert.False(t, isRegexFilter("/fooby"))
	assert.False(t, isRegexFilter("fooby/"))
	assert.False(t, isRegexFilter("/"))
	assert.True(t, isRegexFilter("/fooby/"))
}

func TestSAIncludesToExpr(t *testing.T) {
	expression := saIncludesToExpr([]any{
		map[any]any{
			"metricName": "foo",
		},
	})
	assert.Equal(t, `not (MetricName matches "^foo$")`, expression)
}

func TestHostObs(t *testing.T) {
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-host-obs.yaml")
	rc := cfg.Receivers["receiver_creator"]
	wo := rc["watch_observers"].([]string)
	assert.Equal(t, "host_observer", wo[0])
}

func TestSapmEndpoint(t *testing.T) {
	endpt := sapmEndpoint(saCfgInfo{
		ingestURL: "https://ingest.lab0.signalfx.com",
	})
	assert.Equal(t, "https://ingest.lab0.signalfx.com/v2/trace", endpt)
	endpt = sapmEndpoint(saCfgInfo{
		realm: "lab0",
	})
	assert.Equal(t, "https://ingest.lab0.signalfx.com/v2/trace", endpt)
}

func TestNoCorrelationMetrics(t *testing.T) {
	cfg, _ := yamlToOtelConfig(t, "testdata/sa-no-trace-correlation.yaml")
	ex := cfg.Service.Pipelines["traces"].Exporters
	assert.Equal(t, []string{"sapm"}, ex)
}
