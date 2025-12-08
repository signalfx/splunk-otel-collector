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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestOracledbDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	dockerSocketProxy, err := testutils.CreateDockerSocketProxy(t)
	require.NoError(t, err)

	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer("otlp_exporter.yaml", func(c testutils.Collector) testutils.Collector {
		cc := c.(*testutils.CollectorContainer)
		cc.Container = cc.Container.WillWaitForLogs("Everything is ready")
		return cc
	},
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(map[string]string{
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
				"--set", fmt.Sprintf("splunk.discovery.extensions.docker_observer.config.endpoint=tcp://%s", dockerSocketProxy.ContainerEndpoint),
			)
		})
	defer shutdown()

	metricNames := []string{
		"oracledb.cpu_time",
		"oracledb.enqueue_deadlocks",
		"oracledb.exchange_deadlocks",
		"oracledb.executions",
		"oracledb.hard_parses",
		"oracledb.logical_reads",
		"oracledb.parse_calls",
		"oracledb.pga_memory",
		"oracledb.physical_reads",
		"oracledb.user_commits",
		"oracledb.user_rollbacks",
		"oracledb.sessions.usage",
		"oracledb.processes.usage",
		"oracledb.processes.limit",
		"oracledb.sessions.usage",
		"oracledb.sessions.limit",
		"oracledb.enqueue_locks.usage",
		"oracledb.enqueue_locks.limit",
		"oracledb.enqueue_resources.usage",
		"oracledb.enqueue_resources.limit",
		"oracledb.transactions.usage",
		"oracledb.transactions.limit",
		"oracledb.dml_locks.usage",
		"oracledb.dml_locks.limit",
		"oracledb.tablespace_size.limit",
		"oracledb.tablespace_size.usage",
	}

	missingMetrics := make(map[string]any, len(metricNames))
	for _, m := range metricNames {
		missingMetrics[m] = struct{}{}
	}

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		for i := 0; i < len(tc.OTLPReceiverSink.AllMetrics()); i++ {
			m := tc.OTLPReceiverSink.AllMetrics()[i]
			for j := 0; j < m.ResourceMetrics().Len(); j++ {
				rm := m.ResourceMetrics().At(j)
				for k := 0; k < rm.ScopeMetrics().Len(); k++ {
					sm := rm.ScopeMetrics().At(k)
					for l := 0; l < sm.Metrics().Len(); l++ {
						delete(missingMetrics, sm.Metrics().At(l).Name())
					}
				}
			}
		}
		msg := "Missing metrics:\n"
		for k := range missingMetrics {
			msg += fmt.Sprintf("- %q\n", k)
		}
		assert.Len(tt, missingMetrics, 0, msg)
	}, 1*time.Minute, 1*time.Second)
}
