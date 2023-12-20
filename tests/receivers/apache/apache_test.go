// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

//go:build integration

package tests

import (
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

var apache = []testutils.Container{testutils.NewContainer().WithContext(
	path.Join(".", "testdata", "server"),
).WithName("apache").WithExposedPorts("8000:80").WillWaitForLogs("Command line: 'httpd -D FOREGROUND'")}

// This test ensures the collector can connect to an Apache Web Server, and properly get metrics. It's not intended to
// test the receiver itself.
func TestApacheIntegration(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()
	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected.yaml"))
	require.NoError(t, err)

	_, stop := tc.Containers(apache...)
	defer stop()

	_, shutdown := tc.SplunkOtelCollector("all_metrics_config.yaml")
	defer shutdown()

	require.Eventually(t, func() bool {
		if len(tc.OTLPReceiverSink.AllMetrics()) == 0 {
			return false
		}
		return true
		//return expected.DataPointCount() == tc.OTLPReceiverSink.AllMetrics()[len(tc.OTLPReceiverSink.AllMetrics())-1].DataPointCount()
	}, 1*time.Minute, 1*time.Second)

	selected := tc.OTLPReceiverSink.AllMetrics()[len(tc.OTLPReceiverSink.AllMetrics())-1]

	metricNames := []string{"apache_requests", "apache_bytes", "apache_connections", "apache_idle_workers", "apache_scoreboard.open"}
	require.NoError(t, pmetrictest.CompareMetrics(expected, selected,
		pmetrictest.IgnoreMetricDataPointsOrder(),
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreMetricValues(metricNames...),
		pmetrictest.IgnoreMetricAttributeValue("host", metricNames...)))
}
