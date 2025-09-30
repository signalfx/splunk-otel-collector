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

// TestHostObserver verifies basic discovery mode functionality within the collector container by
// discovering collector's own internal prometheus metrics with a config dir containing a host discovery observer and
// simple prometheus discovery config.
func TestHostObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	promPort := testutils.GetAvailablePort(t)

	cc, shutdown := tc.SplunkOtelCollectorContainer(
		"host-otlp-exporter-no-internal-prometheus.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			configd, err := filepath.Abs(filepath.Join(".", "testdata", "host-observer-config.d"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(configd, "/opt/config.d"))
			return cc
		},
		func(c testutils.Collector) testutils.Collector {
			return c.WithEnv(map[string]string{
				"INTERNAL_PROMETHEUS_PORT": fmt.Sprintf("%d", promPort),
				// confirm that debug logging doesn't affect runtime
				"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
				"LABEL_ONE_VALUE":            "actual.label.one.value.from.env.var",
				"LABEL_TWO_VALUE":            "actual.label.two.value.from.env.var",
			}).WithArgs(
				"--discovery", "--config-dir", "/opt/config.d",
				"--set", "splunk.discovery.receivers.prometheus_simple.config.labels::label_three=actual.label.three.value.from.cmdline.property",
				"--set", "splunk.discovery.extensions.k8s_observer.enabled=false",
				"--set", "splunk.discovery.extensions.host_observer.config.refresh_interval=1s",
			)
		},
	)
	defer shutdown()

	// verify collector emits discovered prometheus metrics
	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected", "host-observer-internal-prometheus-expected.yaml"))
	require.NoError(t, err)
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		err := pmetrictest.CompareMetrics(expected, tc.OTLPReceiverSink.AllMetrics()[len(tc.OTLPReceiverSink.AllMetrics())-1],
			pmetrictest.IgnoreResourceAttributeValue("service.instance.id"),
			pmetrictest.IgnoreResourceAttributeValue("server.port"),
			pmetrictest.IgnoreResourceAttributeValue("discovery.endpoint.id"),
			pmetrictest.IgnoreResourceAttributeValue("service.version"),
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

	// verify collector's initial config
	expectedInitialFile := filepath.Join("testdata", "expected", "host-observer-initial-config-expected.yaml")
	expectedInitial := readConfigFromYamlTmplFile(t, expectedInitialFile, nil)
	gotInitial := cc.InitialConfig(t, 55554)
	assert.NotZero(t, removeBundledReceivers(gotInitial["splunk.discovery"].(map[string]any)["receivers"].(map[string]any)["discovery/host_observer"]))
	assert.Equal(t, expectedInitial, gotInitial)

	// verify collector's effective config
	expectedEffectiveFile := filepath.Join("testdata", "expected", "host-observer-actual-config-expected.yaml")
	expectedEffective := readConfigFromYamlTmplFile(t, expectedEffectiveFile, map[string]any{
		"OTLPEndpoint": tc.OTLPEndpointForCollector,
		"PromPort":     promPort,
	})
	gotEffective := cc.EffectiveConfig(t, "localhost")
	assert.NotZero(t, removeBundledReceivers(gotEffective["receivers"].(map[string]any)["discovery/host_observer"]))
	require.Equal(t, expectedEffective, gotEffective)

	// verify collector's dry-run config
	expectedDryRunFile := filepath.Join("testdata", "expected", "host-observer-dry-run-config-expected.yaml")
	expectedDryRun := readConfigFromYamlTmplFile(t, expectedDryRunFile, nil)
	_, out, _ := cc.Container.AssertExec(t, 15*time.Second,
		"sh", "-c", `SPLUNK_DISCOVERY_LOG_LEVEL=error SPLUNK_DEBUG_CONFIG_SERVER=false \
REFRESH_INTERVAL=1s \
SPLUNK_DISCOVERY_RECEIVERS_prometheus_simple_CONFIG_labels_x3a__x3a_label_three=actual.label.three.value.from.env.var.property \
SPLUNK_DISCOVERY_EXTENSIONS_k8s_observer_ENABLED=false \
SPLUNK_DISCOVERY_EXTENSIONS_host_observer_CONFIG_refresh_interval=\${REFRESH_INTERVAL} \
/otelcol --config-dir /opt/config.d --discovery --dry-run`)
	cm, err := confmap.NewRetrievedFromYAML([]byte(out))
	require.NoError(t, err)
	cmr, err := cm.AsRaw()
	require.NoError(t, err)
	gotDryRun := cmr.(map[string]any)
	assert.NotZero(t, removeBundledReceivers(gotDryRun["receivers"].(map[string]any)["discovery/host_observer"]))
	require.Equal(t, expectedDryRun, gotDryRun)
}
