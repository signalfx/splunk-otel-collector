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

package tests

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

// The Oracle DB container takes close to 10 minutes on a local machine to do the default setup, so the best way to
// account for startup time is to wait for the container to be healthy before continuing test.
var oracledb = []testutils.Container{testutils.NewContainer().WithContext(
	path.Join(".", "testdata", "server"),
).WithName("oracledb").WithExposedPorts("1521:1521").WillWaitForHealth(15 * time.Minute)}

// This test ensures the collector can connect to an Oracle DB, and properly get metrics. It's not intended to
// test the receiver itself.
func TestOracleDBIntegration(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPMetricsReceiverSink()

	expectedResourceMetrics := tc.ResourceMetrics("all.yaml")

	_, stop := tc.Containers(oracledb...)
	defer stop()
	env := map[string]string{}
	if image := os.Getenv("SPLUNK_OTEL_COLLECTOR_IMAGE"); strings.TrimSpace(image) != "" {
		env["ORACLEDB_URL"] = "oracle://otel:password@oracledb:1521/XE"
	} else {
		env["ORACLEDB_URL"] = "oracle://otel:password@localhost:1521/XE"
	}
	_, shutdown := tc.SplunkOtelCollectorWithEnv("all_metrics_config.yaml", env)
	defer shutdown()
	receivedMetrics := testutils.ResourceMetrics{}
	var err error
	assert.Eventually(t, func() bool {
		if tc.OTLPMetricsReceiverSink.DataPointCount() == 0 {
			if err == nil {
				err = fmt.Errorf("no metrics received")
			}
			return false
		}
		receivedOTLPMetrics := tc.OTLPMetricsReceiverSink.AllMetrics()
		tc.OTLPMetricsReceiverSink.Reset()

		receivedResourceMetrics, e := testutils.PDataToResourceMetrics(receivedOTLPMetrics...)
		require.NoError(t, e)
		require.NotNil(t, receivedResourceMetrics)
		receivedMetrics = testutils.FlattenResourceMetrics(receivedMetrics, receivedResourceMetrics)

		var containsAll bool
		containsAll, err = receivedMetrics.ContainsAll(*expectedResourceMetrics, false)
		return containsAll
	}, 30*time.Second, 10*time.Millisecond, "Failed to receive expected metrics")

	require.NoError(t, err)
}
