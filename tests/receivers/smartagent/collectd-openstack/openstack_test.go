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

//go:build smartagent_integration

package tests

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const (
	openstackNetwork = "openstack"
	devstackImage    = "devstack:latest"
)

// newDevstackContainer returns a container builder for a devstack OpenStack instance.
// The devstack:latest image must be pre-built before running this test.
// See testdata/devstack/make-devstack-image.sh for instructions.
func newDevstackContainer() testutils.Container {
	return testutils.NewContainer().
		WithImage(devstackImage).
		WithName("devstack").
		WithNetworks(openstackNetwork).
		WithPriviledged(true).
		WithBinds(
			"/lib/modules:/lib/modules:ro",
			"/sys/fs/cgroup:/sys/fs/cgroup:ro",
		).
		WithEnvVar("container", "docker").
		WithExposedPorts("80:80").
		WithStartupTimeout(30*time.Minute).
		WillWaitForPorts("80").
		// devstack logs this when all OpenStack services are running
		WillWaitForLogs("stack.sh completed")
}

// withOpenstackNetwork is a CollectorBuilder that connects the collector container
// to the openstack Docker network so it can reach the devstack container by hostname.
func withOpenstackNetwork(c testutils.Collector) testutils.Collector {
	if cc, ok := c.(*testutils.CollectorContainer); ok {
		cc.Container = cc.Container.WithNetworks(openstackNetwork)
	}
	return c
}

// TestCollectdOpenstackReceiverProvidesDefaultMetrics tests that the collectd/openstack monitor
// emits all default metrics when connected to a live devstack OpenStack deployment.
//
// Prerequisites:
//   - Build the devstack image once by running:
//     tests/receivers/smartagent/collectd-openstack/testdata/devstack/make-devstack-image.sh
//   - Set SPLUNK_OTEL_COLLECTOR_IMAGE to a valid collector image.
func TestCollectdOpenstackReceiverProvidesDefaultMetrics(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, stop := tc.Containers(newDevstackContainer())
	defer stop()

	_, shutdown := tc.SplunkOtelCollector("default_metrics_config.yaml", withOpenstackNetwork)
	defer shutdown()

	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected", "default.yaml"))
	require.NoError(t, err)

	lastIndex := 0
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		var cmpErr error
		newIndex := len(tc.OTLPReceiverSink.AllMetrics()) - 1
		for i := newIndex; i >= lastIndex; i-- {
			m := tc.OTLPReceiverSink.AllMetrics()[i]
			if m.MetricCount() >= expected.MetricCount() {
				cmpErr = pmetrictest.CompareMetrics(expected, m,
					pmetrictest.IgnoreMetricAttributeValue("host"),
					pmetrictest.IgnoreMetricAttributeValue("plugin_instance"),
					pmetrictest.IgnoreMetricsOrder(),
					pmetrictest.IgnoreMetricValues(),
					pmetrictest.IgnoreTimestamp(),
					pmetrictest.IgnoreStartTimestamp(),
					pmetrictest.IgnoreMetricDataPointsOrder(),
					pmetrictest.IgnoreScopeMetricsOrder(),
					pmetrictest.IgnoreResourceMetricsOrder(),
					pmetrictest.IgnoreScopeVersion(),
					pmetrictest.IgnoreSubsequentDataPoints(),
				)
				if cmpErr == nil {
					return
				}
			}
		}
		lastIndex = newIndex
		assert.NoError(tt, cmpErr)
	}, 2*time.Minute, 5*time.Second)
}

// TestCollectdOpenstackReceiverProvidesAllMetrics tests that the collectd/openstack monitor
// emits all metrics (including non-default ones) when extraMetrics is configured with a wildcard.
//
// Prerequisites:
//   - Build the devstack image once by running:
//     tests/receivers/smartagent/collectd-openstack/testdata/devstack/make-devstack-image.sh
//   - Set SPLUNK_OTEL_COLLECTOR_IMAGE to a valid collector image.
func TestCollectdOpenstackReceiverProvidesAllMetrics(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, stop := tc.Containers(newDevstackContainer())
	defer stop()

	_, shutdown := tc.SplunkOtelCollector("all_metrics_config.yaml", withOpenstackNetwork)
	defer shutdown()

	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected", "all.yaml"))
	require.NoError(t, err)

	lastIndex := 0
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		var cmpErr error
		newIndex := len(tc.OTLPReceiverSink.AllMetrics()) - 1
		for i := newIndex; i >= lastIndex; i-- {
			m := tc.OTLPReceiverSink.AllMetrics()[i]
			if m.MetricCount() >= expected.MetricCount() {
				cmpErr = pmetrictest.CompareMetrics(expected, m,
					pmetrictest.IgnoreMetricAttributeValue("host"),
					pmetrictest.IgnoreMetricAttributeValue("plugin_instance"),
					pmetrictest.IgnoreMetricsOrder(),
					pmetrictest.IgnoreMetricValues(),
					pmetrictest.IgnoreTimestamp(),
					pmetrictest.IgnoreStartTimestamp(),
					pmetrictest.IgnoreMetricDataPointsOrder(),
					pmetrictest.IgnoreScopeMetricsOrder(),
					pmetrictest.IgnoreResourceMetricsOrder(),
					pmetrictest.IgnoreScopeVersion(),
					pmetrictest.IgnoreSubsequentDataPoints(),
				)
				if cmpErr == nil {
					return
				}
			}
		}
		lastIndex = newIndex
		assert.NoError(tt, cmpErr)
	}, 2*time.Minute, 5*time.Second)
}
