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
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	if runtime.GOOS == "darwin" {
		t.Skip("unable to share sockets between mac and d4m vm: https://github.com/docker/for-mac/issues/483#issuecomment-758836836")
	}
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, stop := tc.Containers(
		testutils.NewContainer().WithContext(
			path.Join(".", "testdata", "server"),
		).WithExposedPorts(
			fmt.Sprintf("%d:80", testutils.GetAvailablePort(t)),
		).WithName("nginx").WillWaitForPorts("80"),
	)
	defer stop()

	_, shutdown := tc.SplunkOtelCollectorContainer(
		"otlp_exporter.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			cc.Container = cc.Container.WithBinds("/var/run/docker.sock:/var/run/docker.sock:ro")
			cc.Container = cc.Container.WillWaitForLogs("Discovering for next")
			cc.Container = cc.Container.WithUser(fmt.Sprintf("999:%d", testutils.GetDockerGID(t)))
			return cc
		},
		func(c testutils.Collector) testutils.Collector {
			return c.WithEnv(map[string]string{
				// runner seems to be slow
				"SPLUNK_DISCOVERY_DURATION": "20s",
				// confirm that debug logging doesn't affect runtime
				"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
			}).WithArgs(
				"--discovery",
				"--set", "splunk.discovery.receivers.smartagent/collectd/nginx.config.username=some_user",
				"--set", "splunk.discovery.receivers.smartagent/collectd/nginx.config.password=some_password",
				"--set", `splunk.discovery.extensions.k8s_observer.enabled=false`,
				"--set", `splunk.discovery.extensions.host_observer.enabled=false`,
			)
		},
	)
	defer shutdown()

	expectedResourceMetrics := tc.ResourceMetrics("default.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
