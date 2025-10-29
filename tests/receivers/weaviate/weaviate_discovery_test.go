// Copyright Splunk, Inc.
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

//go:build discovery_integration_weaviate

package test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const (
	testDataDir = "testdata"
	nodeID      = "5245ra617433bd"
)

func TestWeaviateDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	dockerSocketProxy, err := testutils.CreateDockerSocketProxy(t)
	require.NoError(t, err)

	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()
	_, shutdown := tc.SplunkOtelCollectorContainer("otlp_exporter.yaml", func(collector testutils.Collector) testutils.Collector {
		cc := collector.(*testutils.CollectorContainer)
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

	var actualMetrics pmetric.Metrics
	expectedMetricsFile := filepath.Join(testDataDir, "expected.yaml")
	expected, err := golden.ReadMetrics(expectedMetricsFile)
	require.NoError(t, err)
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		actualMetrics = tc.OTLPReceiverSink.AllMetrics()[len(tc.OTLPReceiverSink.AllMetrics())-1]
		require.NotNil(t, actualMetrics)
	}, 60*time.Second, 1*time.Second)

	testutils.MaybeUpdateExpectedMetricsResults(t, expectedMetricsFile, &actualMetrics)
	err = pmetrictest.CompareMetrics(expected, actualMetrics,
		pmetrictest.IgnoreResourceAttributeValue("service.instance.id"),
		pmetrictest.IgnoreResourceAttributeValue("net.host.port"),
		pmetrictest.IgnoreResourceAttributeValue("net.host.name"),
		pmetrictest.IgnoreResourceAttributeValue("server.address"),
		pmetrictest.IgnoreResourceAttributeValue("container.name"),
		pmetrictest.IgnoreResourceAttributeValue("server.port"),
		pmetrictest.IgnoreResourceAttributeValue("service.name"),
		pmetrictest.IgnoreResourceAttributeValue("service_instance_id"),
		pmetrictest.IgnoreResourceAttributeValue("service_version"),
		pmetrictest.IgnoreResourceAttributeValue("discovery.endpoint.id"),
		pmetrictest.IgnoreMetricAttributeValue("service_version"),
		pmetrictest.IgnoreMetricAttributeValue("service_instance_id"),
		pmetrictest.IgnoreResourceAttributeValue("server.address"),
		pmetrictest.ChangeDatapointAttributeValue("nodeID", func(string) string { return nodeID }),
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreStartTimestamp(),
		pmetrictest.IgnoreMetricDataPointsOrder(),
		pmetrictest.IgnoreScopeMetricsOrder(),
		pmetrictest.IgnoreScopeVersion(),
		pmetrictest.IgnoreResourceMetricsOrder(),
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreMetricValues(),
	)
	assert.NoError(t, err)
}
