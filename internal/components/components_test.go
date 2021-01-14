// Copyright 2020 Splunk, Inc.
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
	"go.opentelemetry.io/collector/config/configmodels"
)

func TestDefaultComponents(t *testing.T) {
	expectedExtensions := []configmodels.Type{
		"health_check",
		"pprof",
		"zpages",
		"fluentbit",
		"host_observer",
		"k8s_observer",
		"http_forwarder",
	}
	expectedReceivers := []configmodels.Type{
		"jaeger",
		"zipkin",
		"prometheus",
		"opencensus",
		"otlp",
		"hostmetrics",
		"fluentforward",
		"carbon",
		"docker_stats",
		"collectd",
		"k8s_cluster",
		"kubeletstats",
		"prometheus_exec",
		"prometheus_simple",
		"receiver_creator",
		"redis",
		"sapm",
		"signalfx",
		"smartagent",
		"splunk_hec",
		"statsd",
	}
	expectedProcessors := []configmodels.Type{
		"attributes",
		"resource",
		"batch",
		"memory_limiter",
		"span",
		"filter",
		"k8s_tagger",
		"metricstransform",
		"resourcedetection",
	}
	expectedExporters := []configmodels.Type{
		"logging",
		"file",
		"otlp",
		"sapm",
		"signalfx",
		"signalfx_correlation",
		"splunk_hec",
	}

	factories, err := Get()
	assert.NoError(t, err)

	exts := factories.Extensions
	assert.Len(t, exts, len(expectedExtensions))
	for _, k := range expectedExtensions {
		v, ok := exts[k]
		assert.True(t, ok)
		assert.Equal(t, k, v.Type())
	}

	recvs := factories.Receivers
	assert.Len(t, recvs, len(expectedReceivers))
	for _, k := range expectedReceivers {
		v, ok := recvs[k]
		require.True(t, ok)
		assert.Equal(t, k, v.Type())
	}

	procs := factories.Processors
	assert.Len(t, procs, len(expectedProcessors))
	for _, k := range expectedProcessors {
		v, ok := procs[k]
		require.True(t, ok)
		assert.Equal(t, k, v.Type())
	}

	exps := factories.Exporters
	assert.Len(t, exps, len(expectedExporters))
	for _, k := range expectedExporters {
		v, ok := exps[k]
		require.True(t, ok)
		assert.Equal(t, k, v.Type())
	}
}
