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
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
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
	logCore, logs := observer.New(zap.DebugLevel)
	logger := zap.New(logCore)

	csPort := testutils.GetAvailablePort(t)
	collector, err := testutils.NewCollectorProcess().
		WithArgs("--config", path.Join(".", "testdata", "envvar_config.yaml")).
		WithLogger(logger).
		WithEnv(map[string]string{
			"SPLUNK_DEBUG_CONFIG_SERVER_PORT": fmt.Sprintf("%d", csPort),
			"OTLP_PROTOCOLS":                  "{ grpc: , http: , }",
		}).
		Build()

	require.NotNil(t, collector)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, collector.Shutdown())
	}()

	err = collector.Start()
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		for _, log := range logs.All() {
			if strings.Contains(log.Message,
				`Set config to [testdata/envvar_config.yaml]`,
			) {
				return true
			}
		}
		return false
	}, 20*time.Second, time.Second)

	require.Eventually(t, func() bool {
		for _, log := range logs.All() {
			// Confirm collector starts and runs successfully
			if strings.Contains(log.Message, "Everything is ready. Begin running and processing data.") {
				return true
			}
		}
		return false
	}, 20*time.Second, time.Second)

	expectedConfig := map[string]any{
		"receivers": map[string]any{
			"otlp": map[string]any{
				"protocols": map[string]any{
					"grpc": nil,
					"http": nil,
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
					"receivers": []any{"otlp"},
					"exporters": []any{"otlp"},
				},
			},
		},
	}
	for _, tc := range []struct {
		expected map[string]any
		endpoint string
	}{
		{expected: expectedConfig, endpoint: "effective"},
	} {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/debug/configz/%s", csPort, tc.endpoint))
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		actual := map[string]any{}
		require.NoError(t, yaml.Unmarshal(body, &actual))

		require.Equal(t, tc.expected, confmap.NewFromStringMap(actual).ToStringMap())
	}

}
