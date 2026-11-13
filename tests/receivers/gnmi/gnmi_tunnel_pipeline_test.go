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
	"context"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func createTarget(port string) *testutils.Container {
	target := testutils.NewContainer().
		WithNetworkMode("host").
		WithEnv(map[string]string{"TUNNEL_ADDR": "localhost:" + port}).
		WillWaitForLogs("listening").
		WithContext(path.Join(".", "testdata", "tunnel_server")).
		WithStartupTimeout(5 * time.Minute).
		Build()

	return target
}

func TestGNMIReceiverTunnelPipeline(t *testing.T) {
	port := strconv.FormatUint(uint64(testutils.GetAvailablePort(t)), 10)
	target := createTarget(port)

	testutils.RunMetricsCollectionTest(t,
		"tunnel_pipeline_config.yaml", "tunnel_pipeline_expected.yaml",
		testutils.WithCollectorEnvVars(map[string]string{"GNMI_TUNNEL_PORT": port}),
		testutils.WithAfterCollectorStart(func(t *testing.T) {
			require.NoError(t, target.Start(context.Background()))
			t.Cleanup(func() {
				if err := target.Terminate(context.Background()); err != nil {
					t.Logf("failed to terminate gNMI target container: %v", err)
				}
			})
		}),
		testutils.WithCompareMetricsOptions(
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreMetricsOrder(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
		),
	)
}
