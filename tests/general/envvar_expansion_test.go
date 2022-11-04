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
	"testing"
	"time"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func TestExpandedDollarSignsViaStandardEnvVar(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	tc.SkipIfNotContainer()

	_, shutdown := tc.SplunkOtelCollectorWithEnv(
		"envvar_labels.yaml",
		map[string]string{"AN_ENVVAR": "an-envvar-value"},
	)

	defer shutdown()

	expectedResourceMetrics := tc.ResourceMetrics("envvar_labels.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}

func TestExpandedDollarSignsViaEnvConfigSource(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	tc.SkipIfNotContainer()

	_, shutdown := tc.SplunkOtelCollectorWithEnv(
		"env_config_source_labels.yaml",
		map[string]string{"AN_ENVVAR": "an-envvar-value"},
	)

	defer shutdown()

	expectedResourceMetrics := tc.ResourceMetrics("env_config_source_labels.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}

func TestIncompatibleExpandedDollarSignsViaEnvConfigSource(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

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
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}

func TestExpandedYamlViaEnvConfigSource(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	tc.SkipIfNotContainer()

	_, shutdown := tc.SplunkOtelCollectorWithEnv(
		"yaml_from_env.yaml",
		map[string]string{"YAML": "[{action: update, include: .*, match_type: regexp, operations: [{action: add_label, new_label: yaml-from-env, new_value: value-from-env}]}]"},
	)

	defer shutdown()
	expectedResourceMetrics := tc.ResourceMetrics("yaml_from_env.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}

func TestCollectorProcessWithEnvVarConfig(t *testing.T) {

	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	tc.SkipIfNotContainer()
	c, shutdown := tc.SplunkOtelCollectorWithEnv(
		"envvar_config.yaml",
		map[string]string{
			"OTLP_PROTOCOLS": "{ grpc: , http: , }",
		},
	)

	cc := c.(*testutils.CollectorContainer)

	defer shutdown()

	expectedConfig := map[string]any{
		"receivers": map[string]any{
			"otlp": map[string]any{
				"protocols": map[string]any{
					"grpc": nil,
					"http": nil,
				},
			},
			"hostmetrics": map[string]any{
				"scrapers": map[string]any{
					"cpu":    nil,
					"memory": nil,
				},
			},
		},
		"exporters": map[string]any{
			"otlp": map[string]any{
				"endpoint": "localhost:23456",
				"tls": map[string]any{
					"insecure": true,
				},
			},
		},
		"service": map[string]any{
			"pipelines": map[string]any{
				"metrics": map[string]any{
					"receivers": []any{"hostmetrics"},
					"exporters": []any{"otlp"},
				},
			},
		},
	}

	actual := cc.EffectiveConfig(t, 55554)

	require.Equal(t, expectedConfig, confmap.NewFromStringMap(actual).ToStringMap())

}
