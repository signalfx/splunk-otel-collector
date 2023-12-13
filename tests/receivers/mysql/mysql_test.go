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
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

var mysql = []testutils.Container{testutils.NewContainer().WithContext(
	path.Join(".", "testdata", "server"),
).WithName("mysql").WithExposedPorts("3306:3306").WillWaitForLogs("database system is ready to accept connections")}

// This test ensures the collector can connect to a MySQL DB, and properly get metrics. It's not intended to
// test the receiver itself.
func TestMysqlIntegration(t *testing.T) {
	testutils.AssertAllMetricsReceived(t, "all.yaml", "all_metrics_config.yaml",
		mysql, []testutils.CollectorBuilder{
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithEnv(map[string]string{
					"MYSQLDB_ENDPOINT": "localhost:3306",
					"MYSQLDB_USERNAME": "otelu",
					"MYSQLDB_PASSWORD": "otelp",
				})
			},
		},
	)
}
