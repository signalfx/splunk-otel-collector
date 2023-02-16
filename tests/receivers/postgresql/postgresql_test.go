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
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

var postgresqldb = []testutils.Container{testutils.NewContainer().WithContext(
	path.Join(".", "testdata", "server"),
).WithName("postgresqldb").WithExposedPorts("15432:5432").WillWaitForLogs("database system is ready to accept connections")}

// This test ensures the collector can connect to a PostgreSQL DB, and properly get metrics. It's not intended to
// test the receiver itself.
func TestPostgresqlDBIntegration(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, stop := tc.Containers(postgresqldb...)
	defer stop()

	_, shutdown := tc.SplunkOtelCollector(
		"all_metrics_config.yaml",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(map[string]string{
				"POSTGRESQLDB_ENDPOINT": "localhost:15432",
				"POSTGRESQLDB_USERNAME": "otelu",
				"POSTGRESQLDB_PASSWORD": "otelp",
			})
		},
	)
	defer shutdown()

	expectedResourceMetrics := tc.ResourceMetrics("all.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
