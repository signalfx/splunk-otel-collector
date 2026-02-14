// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
)

func TestDefaultComponents(t *testing.T) {
	expectedExtensions := []string{
		"ack",
		"basicauth",
		"bearertokenauth",
		"docker_observer",
		"ecs_observer",
		"file_storage",
		"googlecloudlogentry_encoding",
		"headers_setter",
		"health_check",
		"host_observer",
		"http_forwarder",
		"k8s_leader_elector",
		"k8s_observer",
		"oauth2client",
		"opamp",
		"pprof",
		"smartagent",
		"text_encoding",
		"zpages",
	}
	expectedReceivers := []string{
		"active_directory_ds",
		"apache",
		"apachespark",
		"awscloudwatch",
		"awscontainerinsightreceiver",
		"awsecscontainermetrics",
		"azureblob",
		"azure_event_hub",
		"azuremonitor",
		"carbon",
		"chrony",
		"ciscoos",
		"cloudfoundry",
		"collectd",
		"discovery",
		"docker_stats",
		"elasticsearch",
		"filelog",
		"filestats",
		"fluentforward",
		"googlecloudpubsub",
		"haproxy",
		"hostmetrics",
		"httpcheck",
		"icmpcheckreceiver",
		"iis",
		"influxdb",
		"jaeger",
		"jmx",
		"journald",
		"k8s_cluster",
		"k8s_events",
		"k8sobjects",
		"kafka",
		"kafkametrics",
		"kubeletstats",
		"lightprometheus",
		"memcached",
		"mongodb",
		"mongodb_atlas",
		"mysql",
		"nginx",
		"nop",
		"ntp",
		"oracledb",
		"otlp",
		"postgresql",
		"prometheus",
		"prometheusremotewrite",
		"prometheus_simple",
		"purefa",
		"rabbitmq",
		"receiver_creator",
		"redis",
		"saphana",
		"scripted_inputs",
		"signalfx",
		"signalfxgatewayprometheusremotewrite",
		"smartagent",
		"snmp",
		"snowflake",
		"solace",
		"splunkenterprise",
		"splunk_hec",
		"sqlquery",
		"sqlserver",
		"sshcheck",
		"statsd",
		"syslog",
		"systemd",
		"tcpcheck",
		"tcplog",
		"tlscheck",
		"udplog",
		"vcenter",
		"wavefront",
		"windowseventlog",
		"windowsperfcounters",
		"yanggrpc",
		"zipkin",
		"zookeeper",
	}
	expectedReceiverAliases := map[string]string{
		"azureeventhub": "azure_event_hub",
		"mongodbatlas":  "mongodb_atlas",
	}
	expectedProcessors := []string{
		"attributes",
		"batch",
		"cumulativetodelta",
		"filter",
		"groupbyattrs",
		"k8sattributes",
		"logstransform",
		"memory_limiter",
		"metricsgeneration",
		"metricstransform",
		"probabilistic_sampler",
		"redaction",
		"resource",
		"resourcedetection",
		"span",
		"tail_sampling",
		"timestamp",
		"transform",
	}
	expectedExporters := []string{
		"awss3",
		"debug",
		"file",
		"googlecloudstorage",
		"kafka",
		"loadbalancing",
		"nop",
		"otlp_grpc",
		"otlp_http",
		"prometheusremotewrite",
		"pulsar",
		"sapm",
		"signalfx",
		"splunk_hec",
	}
	expectedExporterAliases := map[string]string{
		"otlp":     "otlp_grpc",
		"otlphttp": "otlp_http",
	}
	expectedConnectors := []string{
		"count",
		"forward",
		"routing",
		"spanmetrics",
		"sum",
	}

	factories, err := Get()
	require.NoError(t, err)

	exts := factories.Extensions
	assert.Len(t, exts, len(expectedExtensions))
	for _, k := range expectedExtensions {
		v, ok := exts[component.MustNewType(k)]
		assert.True(t, ok)
		assert.Equal(t, k, v.Type().String())
	}

	recvs := factories.Receivers
	assert.Len(t, recvs, len(expectedReceivers)+len(expectedReceiverAliases))

	for _, k := range expectedReceivers {
		v, ok := recvs[component.MustNewType(k)]
		require.True(t, ok, k)
		assert.Equal(t, k, v.Type().String())
	}
	for alias, actual := range expectedReceiverAliases {
		v, ok := recvs[component.MustNewType(alias)]
		require.True(t, ok, "Missing expected exporter alias "+alias)
		assert.Equal(t, actual, v.Type().String())
	}

	procs := factories.Processors
	assert.Len(t, procs, len(expectedProcessors))
	for _, k := range expectedProcessors {
		v, ok := procs[component.MustNewType(k)]
		require.True(t, ok, "Missing expected processor "+k)
		assert.Equal(t, k, v.Type().String())
	}

	exps := factories.Exporters
	assert.Len(t, exps, len(expectedExporters)+len(expectedExporterAliases))
	for _, k := range expectedExporters {
		v, ok := exps[component.MustNewType(k)]
		require.True(t, ok)
		assert.Equal(t, k, v.Type().String())
	}
	for alias, actual := range expectedExporterAliases {
		v, ok := exps[component.MustNewType(alias)]
		require.True(t, ok, "Missing expected exporter alias "+alias)
		assert.Equal(t, actual, v.Type().String())
	}

	conns := factories.Connectors
	assert.Len(t, conns, len(expectedConnectors))
	for _, k := range expectedConnectors {
		v, ok := conns[component.MustNewType(k)]
		require.True(t, ok, "Missing expected connector "+k)
		assert.Equal(t, k, v.Type().String())
	}
}
