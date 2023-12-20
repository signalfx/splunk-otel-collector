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
	"path"
	"runtime"
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestMysqlDockerObserver(t *testing.T) {
	port := "3306"
	mysql := []testutils.Container{testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "server"),
	).WithName("mysql").WithExposedPorts(port + ":" + port).WillWaitForPorts(port).WillWaitForLogs("")}
	testutils.SkipIfNotContainerTest(t)
	if runtime.GOOS == "darwin" {
		t.Skip("unable to share sockets between mac and d4m vm: https://github.com/docker/for-mac/issues/483#issuecomment-758836836")
	}

	testutils.AssertAllMetricsReceived(t, "bundled.yaml", "all_metrics_config.yaml",
		mysql, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				cc := c.(*testutils.CollectorContainer)
				cc.Container = cc.Container.WithBinds("/var/run/docker.sock:/var/run/docker.sock:ro")
				cc.Container = cc.Container.WillWaitForLogs("Discovering for next")
				cc.Container = cc.Container.WithUser(fmt.Sprintf("999:%d", testutils.GetDockerGID(t)))
				return cc
			},
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithEnv(map[string]string{
					"SPLUNK_DISCOVERY_DURATION":  "10s",
					"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
					"MYSQLDB_ENDPOINT":           "127.0.0.1:" + port,
					"MYSQLDB_USERNAME":           "root",
					"MYSQLDB_PASSWORD":           "testuser",
				}).WithArgs(
					"--discovery",
					"--set", `splunk.discovery.extensions.k8s_observer.enabled=false`,
					"--set", `splunk.discovery.extensions.host_observer.enabled=false`,
				)
			},
		},
	)
}
