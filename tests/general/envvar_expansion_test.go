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
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestExpandedDollarSignsViaEnvConfigSource(t *testing.T) {
	testutils.RunMetricsCollectionTest(t, "env_config_source_labels.yaml", "env_config_source_labels_expected.yaml",
		testutils.WithCollectorEnvVars(map[string]string{"AN_ENVVAR": "an-envvar-value"}),
		testutils.WithCompareMetricsOptions(
			pmetrictest.IgnoreScopeVersion(),
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
			pmetrictest.IgnoreScopeMetricsOrder(),
			pmetrictest.IgnoreResourceMetricsOrder(),
			pmetrictest.IgnoreMetricsOrder(),
			pmetrictest.IgnoreMetricValues(),
		),
	)
}

func TestExpandedYamlViaEnvConfigSource(t *testing.T) {
	testutils.RunMetricsCollectionTest(t, "yaml_from_env.yaml", "yaml_from_env_expected.yaml",
		testutils.WithCollectorEnvVars(map[string]string{
			"YAML": "[{action: update, include: .*, match_type: regexp, operations: [{action: add_label, " +
				"new_label: yaml-from-env, new_value: value-from-env}]}]",
		}),
		testutils.WithCompareMetricsOptions(
			pmetrictest.IgnoreScopeVersion(),
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreStartTimestamp(),
			pmetrictest.IgnoreMetricDataPointsOrder(),
			pmetrictest.IgnoreScopeMetricsOrder(),
			pmetrictest.IgnoreResourceMetricsOrder(),
			pmetrictest.IgnoreMetricsOrder(),
			pmetrictest.IgnoreMetricValues(),
		),
	)
}

func TestEnvConfigSource(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	configServerPort := testutils.GetAvailablePort(t)
	cc, shutdown := tc.SplunkOtelCollectorContainer(
		"envvar_config.yaml",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(
				map[string]string{
					"OVERRIDDEN_DEFAULT":              "{ grpc: , http: , }",
					"SPLUNK_DEBUG_CONFIG_SERVER_PORT": fmt.Sprintf("%d", configServerPort),
				},
			)
		},
	)
	defer shutdown()

	expectedInitial := map[string]any{
		"file": map[string]any{
			"config_sources": map[string]any{
				"env": map[string]any{
					"defaults": map[string]any{
						"OVERRIDDEN_DEFAULT": "{ http: , }",
						"USED_DEFAULT":       "localhost:23456",
					},
				},
			},
			"receivers": map[string]any{
				"otlp": map[string]any{
					"protocols": "${env:OVERRIDDEN_DEFAULT}",
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
					"endpoint": "${env:USED_DEFAULT}",
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
		},
	}

	require.Equal(t, expectedInitial, cc.InitialConfig(t, configServerPort))

	expectedEffective := map[string]any{
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

	require.Equal(t, expectedEffective, cc.EffectiveConfig(t, configServerPort))
}
