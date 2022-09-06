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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestExpandedDollarSignsViaStandardEnvVar(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPMetricsReceiverSink()

	tc.SkipIfNotContainer()

	_, shutdown := tc.SplunkOtelCollectorWithEnv(
		"envvar_labels.yaml",
		map[string]string{"AN_ENVVAR": "an-envvar-value"},
	)

	defer shutdown()

	expectedResourceMetrics := tc.ResourceMetrics("envvar_labels.yaml")
	require.NoError(t, tc.OTLPMetricsReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}

func TestExpandedDollarSignsViaEnvConfigSource(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPMetricsReceiverSink()

	tc.SkipIfNotContainer()

	_, shutdown := tc.SplunkOtelCollectorWithEnv(
		"env_config_source_labels.yaml",
		map[string]string{"AN_ENVVAR": "an-envvar-value"},
	)

	defer shutdown()

	expectedResourceMetrics := tc.ResourceMetrics("env_config_source_labels.yaml")
	require.NoError(t, tc.OTLPMetricsReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}

func TestIncompatibleExpandedDollarSignsViaEnvConfigSource(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPMetricsReceiverSink()

	tc.SkipIfNotContainer()

	_, shutdown := tc.SplunkOtelCollectorWithEnv(
		"env_config_source_labels.yaml",
		map[string]string{
			"SPLUNK_DOUBLE_DOLLAR_CONFIG_SOURCE_COMPATIBLE": "false",
			"AN_ENVVAR": "an-envvar-value",
		},
	)

	defer shutdown()

	expectedResourceMetrics := tc.ResourceMetrics("incompat_env_config_source_labels.yaml")
	require.NoError(t, tc.OTLPMetricsReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}

func TestExpandedYamlViaEnvConfigSource(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPMetricsReceiverSink()

	tc.SkipIfNotContainer()

	_, shutdown := tc.SplunkOtelCollectorWithEnv(
		"yaml_from_env.yaml",
		map[string]string{"YAML": "[{action: update, include: .*, match_type: regexp, operations: [{action: add_label, new_label: yaml-from-env, new_value: value-from-env}]}]"},
	)

	defer shutdown()

	expectedResourceMetrics := tc.ResourceMetrics("yaml_from_env.yaml")
	require.NoError(t, tc.OTLPMetricsReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
