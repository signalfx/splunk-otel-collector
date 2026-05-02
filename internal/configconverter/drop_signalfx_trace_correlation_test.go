package configconverter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/featuregate"
)

func TestDropSignalFxTracesExporterIfFeatureGateEnabled(t *testing.T) {
	tests := []struct {
		name               string
		featureGateEnabled bool
		input              string
		expected           string
	}{
		{
			name:               "feature_gate_enabled",
			featureGateEnabled: true,
			input:              "testdata/drop_signalfx_exporter_traces_pipelines/input_config.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/expected_config.yaml",
		},
		{
			name:               "feature_gate_disabled",
			featureGateEnabled: false,
			input:              "testdata/drop_signalfx_exporter_traces_pipelines/input_config.yaml",
			expected:           "testdata/drop_signalfx_exporter_traces_pipelines/input_config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, featuregate.GlobalRegistry().Set(dropTraceCorrelationPipelineFeatureGateID, tt.featureGateEnabled))
			t.Cleanup(func() {
				_ = featuregate.GlobalRegistry().Set(dropTraceCorrelationPipelineFeatureGateID, false)
			})

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
