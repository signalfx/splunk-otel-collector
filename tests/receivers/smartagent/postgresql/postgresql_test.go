// Copyright 2021 Splunk, Inc.
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

package tests

import (
	"path"
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestPostgresReceiverProvidesAllMetrics(t *testing.T) {
	server := testutils.NewContainer().WithContext(path.Join(".", "testdata", "server")).WithEnv(
		map[string]string{"POSTGRES_DB": "test_db", "POSTGRES_USER": "postgres", "POSTGRES_PASSWORD": "postgres"},
	).WithExposedPorts("5432:5432").WithName("postgres-server").WithNetworks(
		"postgres",
	).WillWaitForPorts("5432").WillWaitForLogs("database system is ready to accept connections")

	client := testutils.NewContainer().WithContext(path.Join(".", "testdata", "client")).WithEnv(
		map[string]string{"POSTGRES_SERVER": "postgres-server"},
	).WithName("postgres-client").WithNetworks("postgres").WillWaitForLogs("Beginning psql requests")
	containers := []testutils.Container{server, client}

	testutils.AssertAllMetricsReceived(t, "all.yaml", "all_metrics_config.yaml", containers)
}
