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
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/plogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDiscoveryReceiverWithHostObserverAndSimplePrometheusReceiverProvideStatusLogs(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollector("host_observer_simple_prometheus_config.yaml")
	defer shutdown()

	failure, err := golden.ReadLogs(filepath.Join("testdata", "expected", "host_observer.yaml"))
	require.NoError(t, err)
	success, err := golden.ReadLogs(filepath.Join("testdata", "expected", "host_observer-connected.yaml"))
	require.NoError(t, err)

	foundFailure := false
	foundSuccess := false

	assert.Eventually(t, func() bool {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			return false
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		if !foundFailure {
			for _, received := range receivedOTLPLogs {
				entityAttributes, ok := received.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes().Get("otel.entity.attributes")
				if !ok {
					continue
				}
				if v, ok := entityAttributes.Map().Get("discovery.status"); ok && v.Str() == "failed" {
					err := plogtest.CompareLogs(failure, received,
						plogtest.IgnoreResourceAttributeValue("service_instance_id"),
						plogtest.IgnoreResourceAttributeValue("service_version"),
						plogtest.IgnoreTimestamp(),
						plogtest.IgnoreObservedTimestamp(),
					)
					if err == nil {
						foundFailure = true
						break
					} else {
						t.Log(err)
					}
				}
			}
		}

		if !foundSuccess {
			for _, received := range receivedOTLPLogs {
				entityAttributes, ok := received.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes().Get("otel.entity.attributes")
				if !ok {
					continue
				}
				if v, ok := entityAttributes.Map().Get("discovery.status"); ok && v.Str() == "successful" {
					entityAttributes.Map().PutStr("service_instance_id", "")
					entityAttributes.Map().PutStr("service_version", "")
					entityAttributes.Map().PutStr("service.instance.id", "")
					entityAttributes.Map().PutStr("endpoint", "")
					entityAttributes.Map().PutBool("is_ipv6", false)
					entityAttributes.Map().PutStr("discovery.endpoint.id", "")
					entityAttributes.Map().PutStr("discovery.receiver.config", "")
					entityAttributes.Map().PutStr("service.name", "")

					err := plogtest.CompareLogs(success, received,
						plogtest.IgnoreResourceAttributeValue("service_instance_id"),
						plogtest.IgnoreResourceAttributeValue("service_version"),
						plogtest.IgnoreTimestamp(),
						plogtest.IgnoreObservedTimestamp(),
					)
					if err == nil {
						foundSuccess = true
						break
					} else {
						t.Log(err)
					}
				}
			}
		}

		t.Logf("Foundfailure: %t, foundSuccess: %t, received log count: %d", foundFailure, foundSuccess, tc.OTLPReceiverSink.LogRecordCount())
		return foundSuccess && foundFailure
	}, 30*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}
