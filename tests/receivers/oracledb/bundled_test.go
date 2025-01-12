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

//go:build discovery_integration_oracledb

package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestOracledbDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	dockerSocket := testutils.CreateDockerSocketProxy(t)
	require.NoError(t, dockerSocket.Start())
	t.Cleanup(func() {
		dockerSocket.Stop()
	})

	testutils.AssertAllMetricsReceived(t, "bundled.yaml", "otlp_exporter.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				cc := c.(*testutils.CollectorContainer)
				cc.Container = cc.Container.WillWaitForLogs("Discovering for next")
				return cc
			},
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithEnv(map[string]string{
					"SPLUNK_DISCOVERY_DURATION": "20s",
					// confirm that debug logging doesn't affect runtime
					"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
					"ORACLE_PASSWORD":            "password",
				}).WithArgs(
					"--discovery",
					"--set", "splunk.discovery.receivers.oracledb.config.username=otel",
					"--set", "splunk.discovery.receivers.oracledb.config.password='${ORACLE_PASSWORD}'",
					"--set", "splunk.discovery.receivers.oracledb.config.service=XE",
					"--set", `splunk.discovery.extensions.k8s_observer.enabled=false`,
					"--set", `splunk.discovery.extensions.host_observer.enabled=false`,
					"--set", fmt.Sprintf("splunk.discovery.extensions.docker_observer.config.endpoint=tcp://%s", dockerSocket.ContainerEndpoint),
				)
			},
		},
	)
}
