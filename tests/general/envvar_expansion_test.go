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

package tests

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestExpandedDollarSignsViaStandardEnvVar(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPMetricsReceiverSink()

	image := os.Getenv("SPLUNK_OTEL_COLLECTOR_IMAGE")
	if strings.TrimSpace(image) == "" {
		t.Skipf("skipping container-only test")
	}

	hostMetricsEnvVars := map[string]string{
		"OTLP_ENDPOINT":  tc.OTLPEndpoint,
		"SPLUNK_TEST_ID": tc.ID,
		"AN_ENVVAR":      "an-envvar-value"}

	collector, err := testutils.NewCollectorContainer().
		WithImage(image).
		WithEnv(hostMetricsEnvVars).
		WithConfigPath(path.Join(".", "testdata", "envvar_labels.yaml")).
		WithLogger(tc.Logger).
		Build()

	require.NoError(t, err)
	require.NotNil(t, collector)
	require.NoError(t, collector.Start())
	defer func() { require.NoError(t, collector.Shutdown()) }()

	expectedResourceMetrics := tc.ResourceMetrics("envvar_labels.yaml")
	require.NoError(t, tc.OTLPMetricsReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}


func TestExpandedDollarSignsViaEnvConfigSource(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPMetricsReceiverSink()

	image := os.Getenv("SPLUNK_OTEL_COLLECTOR_IMAGE")
	if strings.TrimSpace(image) == "" {
		t.Skipf("skipping container-only test")
	}

	hostMetricsEnvVars := map[string]string{
		"OTLP_ENDPOINT":  tc.OTLPEndpoint,
		"SPLUNK_TEST_ID": tc.ID,
		"AN_ENVVAR":      "an-envvar-value"}

	collector, err := testutils.NewCollectorContainer().
		WithImage(image).
		WithEnv(hostMetricsEnvVars).
		WithConfigPath(path.Join(".", "testdata", "env_config_source_labels.yaml")).
		WithLogger(tc.Logger).
		Build()

	require.NoError(t, err)
	require.NotNil(t, collector)
	require.NoError(t, collector.Start())
	defer func() { require.NoError(t, collector.Shutdown()) }()

	expectedResourceMetrics := tc.ResourceMetrics("env_config_source_labels.yaml")
	require.NoError(t, tc.OTLPMetricsReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
