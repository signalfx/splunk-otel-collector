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
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestPostgresReceiverProvidesAllMetrics(t *testing.T) {
	tc := newTestcase(t)
	defer tc.printLogsOnFailure()
	defer func() { require.NoError(t, tc.otlp.Shutdown()) }()

	expectedResourceMetrics := tc.resourceMetrics("all.yaml")

	server, client := tc.postgresContainers()
	defer func() {
		require.NoError(t, server.Stop(context.Background()))
		require.NoError(t, client.Stop(context.Background()))
	}()

	collector := tc.splunkOtelCollector("all_metrics_config.yaml")
	defer func() { require.NoError(tc, collector.Shutdown()) }()

	require.NoError(t, tc.otlp.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}

// While overkill for these test purposes, as a repeated pattern these should eventually be moved to testutils.
type testcase struct {
	*testing.T
	logger       *zap.Logger
	observedLogs *observer.ObservedLogs
	otlp         *testutils.OTLPMetricsReceiverSink
}

func newTestcase(t *testing.T) *testcase {
	tc := testcase{T: t}
	var logCore zapcore.Core
	logCore, tc.observedLogs = observer.New(zap.DebugLevel)
	tc.logger = zap.New(logCore)

	var err error
	tc.otlp, err = testutils.NewOTLPMetricsReceiverSink().WithEndpoint("localhost:23456").WithLogger(tc.logger).Build()
	require.NoError(tc, err)
	require.NoError(tc, tc.otlp.Start())
	return &tc
}

func (t *testcase) resourceMetrics(filename string) *testutils.ResourceMetrics {
	expectedResourceMetrics, err := testutils.LoadResourceMetrics(
		path.Join(".", "testdata", "resource_metrics", filename),
	)
	require.NoError(t, err)
	require.NotNil(t, expectedResourceMetrics)
	return expectedResourceMetrics
}

func (t *testcase) postgresContainers() (server, client *testutils.Container) {
	server = testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "server"),
	).WithEnv(map[string]string{
		"POSTGRES_DB":       "test_db",
		"POSTGRES_USER":     "postgres",
		"POSTGRES_PASSWORD": "postgres",
	}).WithExposedPorts(
		"5432:5432",
	).WithName("postgres-server").WithNetworks(
		"postgres",
	).WillWaitForPorts("5432").WillWaitForLogs(
		"database system is ready to accept connections",
	).Build()

	require.NoError(t, server.Start(context.Background()))

	client = testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "client"),
	).WithEnv(map[string]string{
		"POSTGRES_SERVER": "postgres-server",
	}).WithName("postgres-client").WithNetworks(
		"postgres",
	).WillWaitForLogs("Beginning psql requests").Build()

	require.NoError(t, client.Start(context.Background()))

	return server, client
}

func (t *testcase) splunkOtelCollector(configFilename string) *testutils.CollectorProcess {
	collector, err := testutils.NewCollectorProcess().WithConfigPath(
		path.Join(".", "testdata", configFilename),
	).WithLogLevel("debug").WithLogger(t.logger).Build()

	require.NoError(t, err)
	require.NotNil(t, collector)
	require.NoError(t, collector.Start())
	return collector
}

func (t *testcase) printLogsOnFailure() {
	if !t.Failed() {
		return
	}
	fmt.Printf("Logs: \n")
	for _, statement := range t.observedLogs.All() {
		fmt.Printf("%v\n", statement)
	}
}

