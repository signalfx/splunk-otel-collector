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
		"config_source_telemetry",
		"docker_observer",
		"ecs_observer",
		"file_storage",
		"google_cloud_logentry_encoding",
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
	expectedExtensionAliases := map[string]string{
		"googlecloudlogentry_encoding": "google_cloud_logentry_encoding",
	}
	expectedReceivers := []string{
		"active_directory_ds",
		"apache",
		"apachespark",
		"awscloudwatch",
		"awscontainerinsightreceiver",
		"awsecscontainermetrics",
		"azure_blob",
		"azure_event_hub",
		"azure_monitor",
		"carbon",
		"chrony",
		"cisco_os",
		"cloud_foundry",
		"collectd",
		"discovery",
		"docker_stats",
		"elasticsearch",
		"file_log",
		"file_stats",
		"fluent_forward",
		"googlecloudpubsub",
		"haproxy",
		"host_metrics",
		"http_check",
		"icmpcheckreceiver",
		"iis",
		"influxdb",
		"jaeger",
		"jmx",
		"journald",
		"k8s_cluster",
		"k8s_events",
		"k8s_objects",
		"kafka",
		"kafka_metrics",
		"kubelet_stats",
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
		"prometheus_remote_write",
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
		"ssh_check",
		"statsd",
		"syslog",
		"systemd",
		"tcp_check",
		"tcp_log",
		"tls_check",
		"udp_log",
		"vcenter",
		"wavefront",
		"windows_event_log",
		"windowsperfcounters",
		"windowsservice",
		"yang_grpc",
		"zipkin",
		"zookeeper",
	}
	expectedReceiverAliases := map[string]string{
		"azureblob":             "azure_blob",
		"azureeventhub":         "azure_event_hub",
		"azuremonitor":          "azure_monitor",
		"ciscoos":               "cisco_os",
		"cloudfoundry":          "cloud_foundry",
		"filelog":               "file_log",
		"filestats":             "file_stats",
		"fluentforward":         "fluent_forward",
		"hostmetrics":           "host_metrics",
		"httpcheck":             "http_check",
		"k8sobjects":            "k8s_objects",
		"kafkametrics":          "kafka_metrics",
		"kubeletstats":          "kubelet_stats",
		"mongodbatlas":          "mongodb_atlas",
		"prometheusremotewrite": "prometheus_remote_write",
		"sshcheck":              "ssh_check",
		"tcpcheck":              "tcp_check",
		"tcplog":                "tcp_log",
		"tlscheck":              "tls_check",
		"udplog":                "udp_log",
		"windowseventlog":       "windows_event_log",
		"yanggrpc":              "yang_grpc",
	}
	expectedProcessors := []string{
		"attributes",
		"batch",
		"cumulativetodelta",
		"filter",
		"groupbyattrs",
		"k8s_attributes",
		"logstransform",
		"memory_limiter",
		"metricsgeneration",
		"metrics_transform",
		"probabilistic_sampler",
		"redaction",
		"resource",
		"resourcedetection",
		"span",
		"tail_sampling",
		"timestamp",
		"transform",
	}
	expectedProcessorAliases := map[string]string{
		"k8sattributes":    "k8s_attributes",
		"metricstransform": "metrics_transform",
	}
	expectedExporters := []string{
		"awss3",
		"debug",
		"file",
		"google_cloud_storage",
		"kafka",
		"loadbalancing",
		"nop",
		"otlp_grpc",
		"otlp_http",
		"prometheusremotewrite",
		"pulsar",
		"signalfx",
		"splunk_hec",
	}
	expectedExporterAliases := map[string]string{
		"otlp":               "otlp_grpc",
		"otlphttp":           "otlp_http",
		"googlecloudstorage": "google_cloud_storage",
	}
	expectedConnectors := []string{
		"count",
		"forward",
		"routing",
		"span_metrics",
		"sum",
	}
	expectedConnectorAliases := map[string]string{
		"spanmetrics": "span_metrics",
	}

	factories, err := Get()
	require.NoError(t, err)

	exts := factories.Extensions
	assert.Len(t, exts, len(expectedExtensions)+len(expectedExtensionAliases))
	for _, k := range expectedExtensions {
		v, ok := exts[component.MustNewType(k)]
		assert.True(t, ok)
		assert.Equal(t, k, v.Type().String())
	}
	for alias, actual := range expectedExtensionAliases {
		v, ok := exts[component.MustNewType(alias)]
		require.True(t, ok, "Missing expected extension alias "+alias)
		assert.Equal(t, actual, v.Type().String())
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
	assert.Len(t, procs, len(expectedProcessors)+len(expectedProcessorAliases))
	for _, k := range expectedProcessors {
		v, ok := procs[component.MustNewType(k)]
		require.True(t, ok, "Missing expected processor "+k)
		assert.Equal(t, k, v.Type().String())
	}
	for alias, actual := range expectedProcessorAliases {
		v, ok := procs[component.MustNewType(alias)]
		require.True(t, ok, "Missing expected processor alias "+alias)
		assert.Equal(t, actual, v.Type().String())
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
	assert.Len(t, conns, len(expectedConnectors)+len(expectedConnectorAliases))
	for _, k := range expectedConnectors {
		v, ok := conns[component.MustNewType(k)]
		require.True(t, ok, "Missing expected connector "+k)
		assert.Equal(t, k, v.Type().String())
	}
	for alias, actual := range expectedConnectorAliases {
		v, ok := conns[component.MustNewType(alias)]
		require.True(t, ok, "Missing expected connector alias "+alias)
		assert.Equal(t, actual, v.Type().String())
	}
}
