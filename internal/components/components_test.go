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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
)

func TestDefaultComponents(t *testing.T) {
	expectedExtensions := []string{
		"basicauth",
		"ecs_observer",
		"ecs_task_observer",
		"docker_observer",
		"health_check",
		"host_observer",
		"http_forwarder",
		"k8s_observer",
		"pprof",
		"smartagent",
		"zpages",
		"memory_ballast",
		"file_storage",
	}
	expectedReceivers := []string{
		"azureeventhub",
		"carbon",
		"cloudfoundry",
		"collectd",
		"discovery",
		"filelog",
		"fluentforward",
		"hostmetrics",
		"jaeger",
		"jmx",
		"journald",
		"k8s_cluster",
		"k8s_events",
		"k8sobjects",
		"kafka",
		"kafkametrics",
		"kubeletstats",
		"mongodbatlas",
		"mongodb",
		"mysql",
		"lightprometheus",
		"oracledb",
		"otlp",
		"postgresql",
		"prometheus",
		"prometheus_simple",
		"receiver_creator",
		"redis",
		"sapm",
		"scripted_inputs",
		"signalfx",
		"signalfxgatewayprometheusremotewrite",
		"smartagent",
		"solace",
		"splunk_hec",
		"sqlquery",
		"sshcheck",
		"statsd",
		"syslog",
		"tcplog",
		"udplog",
		"vcenter",
		"wavefront",
		"windowseventlog",
		"windowsperfcounters",
		"zipkin",
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
		"metricstransform",
		"probabilistic_sampler",
		"resource",
		"resourcedetection",
		"routing",
		"span",
		"tail_sampling",
		"timestamp",
		"transform",
	}
	expectedExporters := []string{
		"awss3",
		"debug",
		"file",
		"kafka",
		"loadbalancing",
		"logging",
		"otlp",
		"otlphttp",
		"pulsar",
		"sapm",
		"signalfx",
		"splunk_hec",
		"httpsink",
	}
	expectedConnectors := []string{
		"count",
		"routing",
		"spanmetrics",
		"forward",
	}

	factories, err := Get()
	assert.NoError(t, err)

	exts := factories.Extensions
	assert.Len(t, exts, len(expectedExtensions))
	for _, k := range expectedExtensions {
		v, ok := exts[component.MustNewType(k)]
		assert.True(t, ok)
		assert.Equal(t, k, v.Type().String())
	}

	recvs := factories.Receivers
	assert.Len(t, recvs, len(expectedReceivers))
	for _, k := range expectedReceivers {
		v, ok := recvs[component.MustNewType(k)]
		require.True(t, ok)
		assert.Equal(t, k, v.Type().String())
	}

	procs := factories.Processors
	assert.Len(t, procs, len(expectedProcessors))
	for _, k := range expectedProcessors {
		v, ok := procs[component.MustNewType(k)]
		require.True(t, ok, fmt.Sprintf("Missing expected processor %s", k))
		assert.Equal(t, k, v.Type().String())
	}

	exps := factories.Exporters
	assert.Len(t, exps, len(expectedExporters))
	for _, k := range expectedExporters {
		v, ok := exps[component.MustNewType(k)]
		require.True(t, ok)
		assert.Equal(t, k, v.Type().String())
	}

	conns := factories.Connectors
	assert.Len(t, conns, len(expectedConnectors))
	for _, k := range expectedConnectors {
		v, ok := conns[component.MustNewType(k)]
		require.True(t, ok, fmt.Sprintf("Missing expected connector %s", k))
		assert.Equal(t, k, v.Type().String())
	}
}
