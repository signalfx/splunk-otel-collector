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

//go:build openstack_receiver_integration

package tests

import (
	"errors"
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

// TestOpenstackReceiver tests that the openstack receiver emits the expected
// metrics when connected to a live devstack OpenStack deployment running on
// the host machine (deployed via gophercloud/devstack-action in CI).
//
// The test shares expected golden files with the collectd/openstack monitor test
// under tests/receivers/smartagent/collectd-openstack/testdata/expected/ so that
// both implementations are held to the same metric contract.
func TestOpenstackReceiver(t *testing.T) {
	// Path to the shared expected-metrics golden files.
	sharedExpected := filepath.Join(
		"..", "smartagent", "collectd-openstack", "testdata", "expected",
	)

	testcase := testutils.NewTestcase(t)
	defer testcase.PrintLogsOnFailure()
	defer testcase.ShutdownOTLPReceiverSink()

	_, shutdown := testcase.SplunkOtelCollectorProcess("testdata/config.yaml")
	defer shutdown()

	expected, err := golden.ReadMetrics(filepath.Join(sharedExpected, "default.yaml"))
	require.NoError(t, err)

	lastIndex := 0
	ok := assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(testcase.OTLPReceiverSink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		cmpErr := errors.New("no comparison was made")
		newIndex := len(testcase.OTLPReceiverSink.AllMetrics()) - 1
		for i := newIndex; i >= lastIndex; i-- {
			m := testcase.OTLPReceiverSink.AllMetrics()[i]
			if m.MetricCount() >= expected.MetricCount() {
				cmpErr = pmetrictest.CompareMetrics(expected, m,
					pmetrictest.IgnoreMetricAttributeValue("host"),
					pmetrictest.IgnoreMetricAttributeValue("hypervisor_hostname"),
					pmetrictest.IgnoreMetricAttributeValue("plugin_instance"),
					pmetrictest.IgnoreMetricAttributeValue("project_id"),
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
				t.Logf("Batch %d had %d metrics but did not match expected: %v", i, m.MetricCount(), cmpErr)
				filename := fmt.Sprintf("testdata/received_failed_batch_%s.yaml", time.Now().Format("20060102_150405"))
				if writeErr := golden.WriteMetricsToFile(filename, m); writeErr != nil {
					t.Logf("failed to write %s: %v", filename, writeErr)
				}
			}
		}
		lastIndex = newIndex
		assert.NoError(tt, cmpErr)
	}, 2*time.Minute, 5*time.Second)

	if !ok {
		t.Logf("expected %d metrics, largest batch had %d", expected.MetricCount(), largestBatchMetricCount(testcase))
		dumpReceivedMetrics(t, testcase, fmt.Sprintf("testdata/received_%s.yaml", t.Name()))
	}
}

func largestBatchMetricCount(tc *testutils.Testcase) int {
	max := 0
	for _, m := range tc.OTLPReceiverSink.AllMetrics() {
		if c := m.MetricCount(); c > max {
			max = c
		}
	}
	return max
}

func dumpReceivedMetrics(t *testing.T, tc *testutils.Testcase, filename string) {
	t.Helper()
	all := tc.OTLPReceiverSink.AllMetrics()
	if len(all) == 0 {
		t.Logf("dumpReceivedMetrics: no metrics were received at all")
		return
	}
	merged := pmetric.NewMetrics()
	for _, m := range all {
		m.ResourceMetrics().MoveAndAppendTo(merged.ResourceMetrics())
	}
	t.Logf("dumpReceivedMetrics: writing %d metric(s) from %d batch(es) to %s",
		merged.MetricCount(), len(all), filename)
	if err := golden.WriteMetricsToFile(filename, merged); err != nil {
		t.Logf("dumpReceivedMetrics: failed to write %s: %v", filename, err)
	}
}
