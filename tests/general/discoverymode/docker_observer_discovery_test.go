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

//go:build integration

package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

// TestDockerObserver verifies basic discovery mode functionality within the collector container by
// starting a collector with the daemon domain socket mounted and the container running with its group id
// to detect a prometheus container with a test.id label the receiver creator rule matches against.
func TestDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	dockerSocketProxy := testutils.CreateDockerSocketProxy(t)
	require.NoError(t, dockerSocketProxy.Start())
	t.Cleanup(func() {
		dockerSocketProxy.Stop()
	})

	_, shutdownPrometheus := tc.Containers(
		testutils.NewContainer().WithImage("bitnami/prometheus").WithLabel("test.id", tc.ID).WillWaitForLogs("Server is ready to receive web requests."),
	)
	defer shutdownPrometheus()

	cc, shutdown := tc.SplunkOtelCollectorContainer(
		"docker-otlp-exporter-no-internal-prometheus.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			configd, err := filepath.Abs(filepath.Join(".", "testdata", "docker-observer-config.d"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(configd, "/opt/config.d"))
			properties, err := filepath.Abs(filepath.Join(".", "testdata", "docker-observer-properties.yaml"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(properties, "/opt/properties.yaml"))
			// uid check is for basic collector functionality not using the splunk-otel-collector user
			// but the docker gid is required to reach the daemon
			cc.Container = cc.Container.WithUser(fmt.Sprintf("%d:%d", os.Getuid(), testutils.GetDockerGID(t)))
			return cc
		},
		func(c testutils.Collector) testutils.Collector {
			return c.WithEnv(map[string]string{
				// confirm that debug logging doesn't affect runtime
				"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
				"DOCKER_DOMAIN_SOCKET":       fmt.Sprintf("tcp://%s", dockerSocketProxy.ContainerEndpoint),
				"LABEL_ONE_VALUE":            "actual.label.one.value",
				"LABEL_TWO_VALUE":            "actual.label.two.value",
				"SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_CONFIG_labels_x3a__x3a_label_x5f_three": "overwritten by --set property",
				"SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_CONFIG_labels_x3a__x3a_label_x5f_four":  "actual.label.four.value",
			}).WithArgs(
				"--discovery", "--config-dir", "/opt/config.d",
				"--set", `splunk.discovery.extensions.k8s_observer.enabled=false`,
				"--set", `splunk.discovery.extensions.docker_observer.enabled=true`,
				"--set", `splunk.discovery.extensions.docker_observer.config.endpoint=${DOCKER_DOMAIN_SOCKET}`,
				"--set", `splunk.discovery.receivers.prometheus_simple.enabled=true`,
				"--set", `splunk.discovery.receivers.prometheus_simple.config.labels::label_three=actual.label.three.value`,
				"--discovery-properties", "/opt/properties.yaml",
			)
		},
	)
	defer shutdown()

	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected", "docker-observer-internal-prometheus-expected.yaml"))
	require.NoError(t, err)
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		err := pmetrictest.CompareMetrics(expected, tc.OTLPReceiverSink.AllMetrics()[len(tc.OTLPReceiverSink.AllMetrics())-1],
			pmetrictest.IgnoreResourceAttributeValue("service.instance.id"),
			pmetrictest.IgnoreResourceAttributeValue("server.address"),
			pmetrictest.IgnoreResourceAttributeValue("container.name"),
			pmetrictest.IgnoreResourceAttributeValue("server.port"),
			pmetrictest.IgnoreResourceAttributeValue("discovery.endpoint.id"),
			pmetrictest.IgnoreResourceAttributeValue("service.name"),
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
			pmetrictest.IgnoreScopeMetricsOrder(),
			pmetrictest.IgnoreScopeVersion(),
			pmetrictest.IgnoreResourceMetricsOrder(),
			pmetrictest.IgnoreMetricValues(),
		)
		assert.NoError(tt, err)
	}, 30*time.Second, 1*time.Second)

	expectedInitialFile := filepath.Join("testdata", "expected", "docker-observer-initial-config-expected.yaml")
	expectedInitial := readConfigFromYamlTmplFile(t, expectedInitialFile, nil)
	gotInitial := cc.InitialConfig(t, 55554)
	assert.NotZero(t, removeBundledReceivers(gotInitial["splunk.discovery"].(map[string]any)["receivers"].(map[string]any)["discovery/docker_observer"]))
	assert.Equal(t, expectedInitial, gotInitial)

	expectedEffectiveFile := filepath.Join("testdata", "expected", "docker-observer-actual-config-expected.yaml")
	expectedEffective := readConfigFromYamlTmplFile(t, expectedEffectiveFile, map[string]any{
		"OTLPEndpoint":   tc.OTLPEndpointForCollector,
		"DockerEndpoint": fmt.Sprintf("tcp://%s", dockerSocketProxy.ContainerEndpoint),
		"TestID":         tc.ID,
	})
	gotEffective := cc.EffectiveConfig(t, 55554)
	assert.NotZero(t, removeBundledReceivers(gotEffective["receivers"].(map[string]any)["discovery/docker_observer"]))
	require.Equal(t, expectedEffective, gotEffective)

	expectedDryRunFile := filepath.Join("testdata", "expected", "docker-observer-dry-run-config-expected.yaml")
	expectedDryRun := readConfigFromYamlTmplFile(t, expectedDryRunFile, nil)
	_, out, _ := cc.Container.AssertExec(t, 3*time.Minute,
		"sh", "-c", `SPLUNK_DISCOVERY_LOG_LEVEL=error SPLUNK_DEBUG_CONFIG_SERVER=false \
SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED=false \
SPLUNK_DISCOVERY_EXTENSIONS_docker_observer_ENABLED=true \
SPLUNK_DISCOVERY_EXTENSIONS_docker_observer_CONFIG_endpoint=\${DOCKER_DOMAIN_SOCKET} \
SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_ENABLED=true \
SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_CONFIG_labels_x3a__x3a_label_x5f_three=="overwritten by --set property" \
SPLUNK_DISCOVERY_RECEIVERS_prometheus_x5f_simple_CONFIG_labels_x3a__x3a_label_x5f_four="actual.label.four.value" \
/otelcol --config-dir /opt/config.d --discovery --dry-run \
--set splunk.discovery.receivers.prometheus_simple.config.labels::label_three=actual.label.three.value \
--discovery-properties /opt/properties.yaml
`)
	cm, err := confmap.NewRetrievedFromYAML([]byte(out))
	require.NoError(t, err)
	cmr, err := cm.AsRaw()
	require.NoError(t, err)
	gotDryRun := cmr.(map[string]any)
	assert.NotZero(t, removeBundledReceivers(gotDryRun["receivers"].(map[string]any)["discovery/docker_observer"]))
	require.Equal(t, expectedDryRun, gotDryRun)
}
