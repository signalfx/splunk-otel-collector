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

//go:build openstack_integration

package tests

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

// TestCollectdOpenstackReceiverProvidesDefaultMetrics tests that the collectd/openstack monitor
// emits all default metrics when connected to a live devstack OpenStack deployment running on
// the host machine (deployed via gophercloud/devstack-action in CI).
func TestCollectdOpenstackReceiverProvidesDefaultMetrics(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("default_metrics_config.yaml")
	defer shutdown()

	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected", "default.yaml"))
	require.NoError(t, err)

	lastIndex := 0
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		cmpErr := errors.New("no comparison was made")
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
// Requires a live devstack OpenStack deployment running on the host machine.
func TestCollectdOpenstackReceiverProvidesAllMetrics(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("all_metrics_config.yaml")
	defer shutdown()

	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected", "all.yaml"))
	require.NoError(t, err)

	lastIndex := 0
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		cmpErr := errors.New("no comparison was made")
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
