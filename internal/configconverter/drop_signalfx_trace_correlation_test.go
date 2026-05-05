// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/featuregate"
)

func TestDropSignalFxTracesExporterIfFeatureGateEnabled(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expected           string
		featureGateEnabled bool
	}{
		{
			name:               "feature_gate_enabled",
			input:              "testdata/drop_signalfx_exporter_traces_pipelines/input_config.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/expected_config.yaml",
			featureGateEnabled: true,
		},
		{
			name:               "feature_gate_disabled",
			input:              "testdata/drop_signalfx_exporter_traces_pipelines/input_config.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/input_config.yaml",
			featureGateEnabled: false,
		},
		{
			name:               "agent_config_feature_gate_enabled",
			input:              "../../cmd/otelcol/config/collector/agent_config.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/agent_config_expected.yaml",
			featureGateEnabled: true,
		},
		{
			name:               "agent_config_feature_gate_disabled",
			input:              "../../cmd/otelcol/config/collector/agent_config.yaml",
			expected:           "../../cmd/otelcol/config/collector/agent_config.yaml",
			featureGateEnabled: false,
		},
		{
			name:               "ecs_ec2_config_feature_gate_enabled",
			input:              "../../cmd/otelcol/config/collector/ecs_ec2_config.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/ecs_ec2_config_expected.yaml",
			featureGateEnabled: true,
		},
		{
			name:               "ecs_ec2_config_feature_gate_disabled",
			input:              "../../cmd/otelcol/config/collector/ecs_ec2_config.yaml",
			expected:           "../../cmd/otelcol/config/collector/ecs_ec2_config.yaml",
			featureGateEnabled: false,
		},
		{
			name:               "fargate_config_feature_gate_enabled",
			input:              "../../cmd/otelcol/config/collector/fargate_config.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/fargate_config_expected.yaml",
			featureGateEnabled: true,
		},
		{
			name:               "fargate_config_feature_gate_disabled",
			input:              "../../cmd/otelcol/config/collector/fargate_config.yaml",
			expected:           "../../cmd/otelcol/config/collector/fargate_config.yaml",
			featureGateEnabled: false,
		},
		{
			name:               "gateway_config_feature_gate_enabled",
			input:              "../../cmd/otelcol/config/collector/gateway_config.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/gateway_config_expected.yaml",
			featureGateEnabled: true,
		},
		{
			name:               "gateway_config_feature_gate_disabled",
			input:              "../../cmd/otelcol/config/collector/gateway_config.yaml",
			expected:           "../../cmd/otelcol/config/collector/gateway_config.yaml",
			featureGateEnabled: false,
		},
		{
			name:               "upstream_agent_config_feature_gate_enabled",
			input:              "../../cmd/otelcol/config/collector/upstream_agent_config.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/upstream_agent_config_expected.yaml",
			featureGateEnabled: true,
		},
		{
			name:               "upstream_agent_config_feature_gate_disabled",
			input:              "../../cmd/otelcol/config/collector/upstream_agent_config.yaml",
			expected:           "../../cmd/otelcol/config/collector/upstream_agent_config.yaml",
			featureGateEnabled: false,
		},
		{
			name:               "full_config_linux_feature_gate_enabled",
			input:              "../../cmd/otelcol/config/collector/full_config_linux.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/full_config_linux_expected.yaml",
			featureGateEnabled: true,
		},
		{
			name:               "full_config_linux_feature_gate_disabled",
			input:              "../../cmd/otelcol/config/collector/full_config_linux.yaml",
			expected:           "../../cmd/otelcol/config/collector/full_config_linux.yaml",
			featureGateEnabled: false,
		},
		{
			name:               "logs_config_linux_feature_gate_enabled",
			input:              "../../cmd/otelcol/config/collector/logs_config_linux.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/logs_config_linux_expected.yaml",
			featureGateEnabled: true,
		},
		{
			name:               "logs_config_linux_feature_gate_disabled",
			input:              "../../cmd/otelcol/config/collector/logs_config_linux.yaml",
			expected:           "../../cmd/otelcol/config/collector/logs_config_linux.yaml",
			featureGateEnabled: false,
		},
		{
			name:               "otlp_config_linux_feature_gate_enabled",
			input:              "../../cmd/otelcol/config/collector/otlp_config_linux.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/otlp_config_linux_expected.yaml",
			featureGateEnabled: true,
		},
		{
			name:               "otlp_config_linux_feature_gate_disabled",
			input:              "../../cmd/otelcol/config/collector/otlp_config_linux.yaml",
			expected:           "../../cmd/otelcol/config/collector/otlp_config_linux.yaml",
			featureGateEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, featuregate.GlobalRegistry().Set(dropTraceCorrelationPipelineFeatureGateID, tt.featureGateEnabled))

			cfgMap, err := confmaptest.LoadConf(tt.input)
			require.NoError(t, err)
			require.NotNil(t, cfgMap)

			expectedCfgMap, err := confmaptest.LoadConf(tt.expected)
			require.NoError(t, err)
			require.NotNil(t, expectedCfgMap)

			require.NoError(t, DropSignalFxTracesExporterIfFeatureGateEnabled(t.Context(), cfgMap))

			require.Equal(t, expectedCfgMap, cfgMap)
		})
	}
}

func TestIsTracePipeline(t *testing.T) {
	tests := []struct {
		name     string
		pipeline string
		expected bool
	}{
		{
			name:     "traces_pipeline",
			pipeline: "traces",
			expected: true,
		},
		{
			name:     "alternate_traces_pipeline",
			pipeline: "traces/alternate",
			expected: true,
		},
		{
			name:     "metrics_pipeline",
			pipeline: "metrics",
			expected: false,
		},
		{
			name:     "internal_metrics_pipeline",
			pipeline: "metrics/internal",
			expected: false,
		},
		{
			name:     "logs_pipeline",
			pipeline: "logs",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, isTracePipeline(tt.pipeline))
		})
	}
}

func TestIsSignalFxExporter(t *testing.T) {
	tests := []struct {
		name     string
		exporter string
		expected bool
	}{
		{
			name:     "default_signalfx_exporter",
			exporter: "signalfx",
			expected: true,
		},
		{
			name:     "alternate_signalfx_exporter",
			exporter: "signalfx/alternate",
			expected: true,
		},
		{
			name:     "otlp_http_exporter",
			exporter: "otlp_http",
			expected: false,
		},
		{
			name:     "splunk_hec_exporter",
			exporter: "splunk_hec",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, isSignalFxExporter(tt.exporter))
		})
	}
}

func TestNoOpConfigs(t *testing.T) {
	tests := []struct {
		config        *confmap.Conf
		name          string
		errorExpected bool
	}{
		{
			name: "nil",
		},
		{
			name:          "empty",
			config:        confmap.New(),
			errorExpected: false,
		},
		{
			name: "invalid_service",
			config: confmap.NewFromStringMap(map[string]any{
				"service": "not-a-map",
			}),
			errorExpected: true,
		},
		{
			name: "invalid_pipelines",
			config: confmap.NewFromStringMap(map[string]any{
				"service": map[string]any{
					"pipelines": "not-a-map",
				},
			}),
			errorExpected: true,
		},
		{
			name: "no_service",
			config: confmap.NewFromStringMap(map[string]any{
				"receivers": map[string]any{
					"otlp": map[string]any{},
				},
			}),
			errorExpected: false,
		},
		{
			name: "no_traces_pipeline",
			config: confmap.NewFromStringMap(map[string]any{
				"service": map[string]any{
					"pipelines": map[string]any{
						"metrics": map[string]any{
							"exporters": []any{"signalfx"},
						},
					},
				},
			}),
			errorExpected: false,
		},
		{
			name: "no_exporters",
			config: confmap.NewFromStringMap(map[string]any{
				"service": map[string]any{
					"pipelines": map[string]any{
						"traces": map[string]any{
							"receivers": []any{"otlp"},
						},
					},
				},
			}),
			errorExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, featuregate.GlobalRegistry().Set(dropTraceCorrelationPipelineFeatureGateID, true))

			var original *confmap.Conf
			config := tt.config
			if config != nil {
				original = confmap.NewFromStringMap(config.ToStringMap())
				config = confmap.NewFromStringMap(config.ToStringMap())
			}

			err := DropSignalFxTracesExporterIfFeatureGateEnabled(t.Context(), config)
			if tt.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, original, config)
		})
	}
}
