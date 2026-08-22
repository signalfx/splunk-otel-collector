// Copyright Splunk, Inc.
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

//go:build discovery_integration_rabbitmq

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestRabbitMQDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	dockerSocketProxy, err := testutils.CreateDockerSocketProxy(t)
	require.NoError(t, err)

	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer("otlp_exporter.yaml", func(c testutils.Collector) testutils.Collector {
		cc := c.(*testutils.CollectorContainer)
		cc.Container = cc.Container.WillWaitForLogs("Everything is ready")
		return cc.WithEnv(map[string]string{
			"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
		}).WithArgs(
			"--discovery",
			"--set", `splunk.discovery.extensions.k8s_observer.enabled=false`,
			"--set", `splunk.discovery.extensions.host_observer.enabled=false`,
			"--set", fmt.Sprintf("splunk.discovery.extensions.docker_observer.config.endpoint=tcp://%s", dockerSocketProxy.ContainerEndpoint),
		)
	})
	defer shutdown()

	expectedMetricNames := []string{
		"rabbitmq.node.disk_free",
		"rabbitmq.node.mem_used",
		"rabbitmq.node.uptime",
	}
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		foundMetricNames := metricNames(tc.OTLPReceiverSink.AllMetrics())
		for _, metricName := range expectedMetricNames {
			assert.Contains(tt, foundMetricNames, metricName)
		}
	}, 120*time.Second, time.Second)
}

func metricNames(metrics []pmetric.Metrics) map[string]struct{} {
	names := map[string]struct{}{}
	for _, md := range metrics {
		for i := 0; i < md.ResourceMetrics().Len(); i++ {
			rm := md.ResourceMetrics().At(i)
			for j := 0; j < rm.ScopeMetrics().Len(); j++ {
				sm := rm.ScopeMetrics().At(j)
				for k := 0; k < sm.Metrics().Len(); k++ {
					names[sm.Metrics().At(k).Name()] = struct{}{}
				}
			}
		}
	}
	return names
}
