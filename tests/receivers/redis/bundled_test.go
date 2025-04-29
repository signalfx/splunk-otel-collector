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

//go:build discovery_integration_redis

package tests

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestRedisDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	dockerSocket := testutils.CreateDockerSocketProxy(t)
	require.NoError(t, dockerSocket.Start())
	t.Cleanup(func() {
		dockerSocket.Stop()
	})

	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer("otlp_exporter.yaml", func(c testutils.Collector) testutils.Collector {
		cc := c.(*testutils.CollectorContainer)
		cc.Container = cc.Container.WillWaitForLogs("Everything is ready")
		return cc
	},
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(map[string]string{
				"REDIS_PASSWORD": "securepassword",
				"REDIS_USERNAME": "otel",
				// confirm that debug logging doesn't affect runtime
				"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
			}).WithArgs(
				"--discovery",
				"--set", "splunk.discovery.receivers.redis.config.password=${REDIS_PASSWORD}",
				"--set", "splunk.discovery.receivers.redis.config.username=${REDIS_USERNAME}",
				"--set", `splunk.discovery.extensions.k8s_observer.enabled=false`,
				"--set", `splunk.discovery.extensions.host_observer.enabled=false`,
				"--set", fmt.Sprintf("splunk.discovery.extensions.docker_observer.config.endpoint=tcp://%s", dockerSocket.ContainerEndpoint),
			)
		})
	defer shutdown()
	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected.yaml"))
	require.NoError(t, err)
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		err := pmetrictest.CompareMetrics(expected, tc.OTLPReceiverSink.AllMetrics()[len(tc.OTLPReceiverSink.AllMetrics())-1],
			pmetrictest.IgnoreResourceAttributeValue("container.name"),
			pmetrictest.IgnoreResourceAttributeValue("discovery.endpoint.id"),
			pmetrictest.IgnoreResourceAttributeValue("net.host.name"),
			pmetrictest.IgnoreResourceAttributeValue("net.host.port"),
			pmetrictest.IgnoreResourceAttributeValue("redis.version"),
			pmetrictest.IgnoreResourceAttributeValue("server.address"),
			pmetrictest.IgnoreResourceAttributeValue("server.port"),
			pmetrictest.IgnoreResourceAttributeValue("service.name"),
			pmetrictest.IgnoreResourceAttributeValue("service.instance.id"),
			pmetrictest.IgnoreResourceAttributeValue("service_instance_id"),
			pmetrictest.IgnoreResourceAttributeValue("service_version"),
			pmetrictest.IgnoreMetricAttributeValue("service_instance_id"),
			pmetrictest.IgnoreMetricAttributeValue("service_version"),
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
			pmetrictest.IgnoreScopeMetricsOrder(),
			pmetrictest.IgnoreScopeVersion(),
			pmetrictest.IgnoreResourceMetricsOrder(),
			pmetrictest.IgnoreMetricValues(),
		)
		assert.NoError(tt, err)
	}, 60*time.Second, 1*time.Second)
}
