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

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestJMXReceiverProvidesAllJVMMetrics(t *testing.T) {
	// JMX metrics are missing when running on arm so the test only checks a subset
	// of what would be received otherwise.
	expectedMetrics := "all.yaml"
	if testutils.CollectorImageIsForArm(t) {
		t.Skip("apparent metric gathering issue on qemu")
	}

	containers := []testutils.Container{
		testutils.NewContainer().WithContext(
			filepath.Join("..", "smartagent", "collectd-cassandra", "testdata", "server"),
		).WithExposedPorts("7199:7199").WithName("jmx").WillWaitForPorts("7199"),
	}

	testutils.AssertAllMetricsReceived(
		t, expectedMetrics, "all_metrics_config.yaml", containers,
		[]testutils.CollectorBuilder{
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithEnv(map[string]string{
					"OTEL_TRACES_EXPORTER": "none",
					"OTEL_LOGS_EXPORTER":   "none",
				})
			},
		})
}
