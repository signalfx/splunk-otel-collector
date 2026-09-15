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
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const gnmiPort = "9339"

// Create containerized gNMI target: the [OpenConfig fake agent][https://github.com/openconfig/gnmi/tree/master/testing/fake],
// which the `openconfig/gnmi` repository ships for testing collector implementations.
//
// The agent streams a fixed set of `SubscribeResponse` messages defined in
// `tests/receivers/gnmi/testdata/server/config.pb.txt`, so the expectations are pinned by
// the fixture rather than by whatever a live device happens to report.
func startTarget(t *testing.T) string {
	target := testutils.NewContainer().
		WithContext(path.Join(".", "testdata", "server")).
		WithExposedPorts(gnmiPort).
		WillWaitForPorts(gnmiPort).
		WillWaitForLogs("listening").
		WithStartupTimeout(5 * time.Minute).
		Build()

	require.NoError(t, target.Start(context.Background()))
	t.Cleanup(func() {
		if err := target.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate gNMI target container: %v", err)
		}
	})

	ctx := context.Background()
	host, err := target.Host(ctx)
	require.NoError(t, err)
	mapped, err := target.MappedPort(ctx, gnmiPort)
	require.NoError(t, err)

	return fmt.Sprintf("%s:%s", host, mapped.Port())
}

func TestGNMIReceiverPipeline(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)

	endpoint := startTarget(t)

	testutils.RunMetricsCollectionTest(t, "pipeline_config.yaml", "pipeline_expected.yaml",
		testutils.WithCollectorEnvVars(map[string]string{"GNMI_ENDPOINT": endpoint}),
		testutils.WithCompareMetricsOptions(
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreMetricsOrder(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
			pmetrictest.IgnoreResourceAttributeValue("server.address"),
		),
	)
}
