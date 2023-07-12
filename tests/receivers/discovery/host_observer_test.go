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
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDiscoveryReceiverWithHostObserverProvidesEndpointLogs(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	if testutils.CollectorImageIsForArm(t) {
		t.Skip("host_observer missing process info on arm")
	}
	testutils.AssertAllLogsReceived(
		t, "host_observer_endpoints.yaml",
		"host_observer_endpoints_config.yaml", nil, nil,
	)
}

func TestDiscoveryReceiverWithHostObserverAndSimplePrometheusReceiverProvideStatusLogs(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	if testutils.CollectorImageIsForArm(t) {
		t.Skip("host_observer missing process info on arm")
	}
	testutils.AssertAllLogsReceived(
		t, "host_observer_simple_prometheus_statuses.yaml",
		"host_observer_simple_prometheus_config.yaml", nil, nil,
	)
}
