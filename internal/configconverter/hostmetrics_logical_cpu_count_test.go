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

package configconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func TestIncludeHostMetricsLogicalCPUCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "adds_signalfx_include_for_explicitly_enabled_metric",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    include_metrics:
      - metric_name: system.cpu.logical.count
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "adds_signalfx_include_when_default_translation_rules_explicitly_enabled",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    disable_default_translation_rules: false
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    disable_default_translation_rules: false
    include_metrics:
      - metric_name: system.cpu.logical.count
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "appends_to_existing_signalfx_include_metrics",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    include_metrics:
      - metric_name: cpu.idle
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    include_metrics:
      - metric_name: cpu.idle
      - metric_name: system.cpu.logical.count
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "named_signalfx_and_hostmetrics_receivers",
			input: `receivers:
  host_metrics/main:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
  hostmetrics/legacy:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx/main:
  signalfx/disabled:
    disable_default_translation_rules: true
service:
  pipelines:
    metrics/main:
      receivers: [host_metrics/main]
      exporters: [signalfx/main]
    metrics/disabled:
      receivers: [hostmetrics/legacy]
      exporters: [signalfx/disabled]
`,
			expected: `receivers:
  host_metrics/main:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
  hostmetrics/legacy:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx/main:
    include_metrics:
      - metric_name: system.cpu.logical.count
  signalfx/disabled:
    disable_default_translation_rules: true
service:
  pipelines:
    metrics/main:
      receivers: [host_metrics/main]
      exporters: [signalfx/main]
    metrics/disabled:
      receivers: [hostmetrics/legacy]
      exporters: [signalfx/disabled]
`,
		},
		{
			name: "no_change_when_metric_explicitly_disabled",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: false
exporters:
  signalfx:
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: false
exporters:
  signalfx:
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "no_change_when_metric_not_explicitly_enabled",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
exporters:
  signalfx:
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
exporters:
  signalfx:
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "no_change_when_metric_already_in_include_metrics",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    include_metrics:
      - metric_name: system.cpu.logical.count
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    include_metrics:
      - metric_name: system.cpu.logical.count
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "no_change_when_metric_already_in_include_metric_names",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    include_metrics:
      - metric_names: [cpu.idle, system.cpu.logical.count]
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    include_metrics:
      - metric_names: [cpu.idle, system.cpu.logical.count]
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "no_change_when_metric_already_in_exclude_metrics",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    exclude_metrics:
      - metric_name: system.cpu.logical.count
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    exclude_metrics:
      - metric_name: system.cpu.logical.count
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "no_change_without_signalfx_exporter",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  otlp:
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [otlp]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  otlp:
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [otlp]
`,
		},
		{
			name: "no_change_when_signalfx_default_translation_rules_disabled",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    disable_default_translation_rules: true
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
    disable_default_translation_rules: true
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "no_change_without_cpu_scraper",
			input: `receivers:
  host_metrics:
    scrapers:
      memory:
exporters:
  signalfx:
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      memory:
exporters:
  signalfx:
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
		{
			name: "no_change_for_logs_pipeline",
			input: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
service:
  pipelines:
    logs/signalfx:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
			expected: `receivers:
  host_metrics:
    scrapers:
      cpu:
        metrics:
          system.cpu.logical.count:
            enabled: true
exporters:
  signalfx:
service:
  pipelines:
    logs/signalfx:
      receivers: [host_metrics]
      exporters: [signalfx]
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := confFromYaml(t, tt.input)
			expected := confFromYaml(t, tt.expected)

			require.NoError(t, IncludeHostMetricsLogicalCPUCount(context.Background(), in))
			require.Equal(t, expected.ToStringMap(), in.ToStringMap())
		})
	}
}

func TestIncludeHostMetricsLogicalCPUCountNilConfig(t *testing.T) {
	require.NoError(t, IncludeHostMetricsLogicalCPUCount(context.Background(), (*confmap.Conf)(nil)))
}
