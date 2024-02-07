// Copyright  Splunk, Inc.
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestDisableKubeletUtilizationMetrics(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput string
	}{
		{
			name:       "no_kubeletstats_receiver",
			input:      "testdata/disable_kubelet_utilization_metrics/no_kubeletstats_receiver.yaml",
			wantOutput: "testdata/disable_kubelet_utilization_metrics/no_kubeletstats_receiver.yaml",
		},
		{
			name:       "no_signalfx_exporter",
			input:      "testdata/disable_kubelet_utilization_metrics/no_signalfx_exporter.yaml",
			wantOutput: "testdata/disable_kubelet_utilization_metrics/no_signalfx_exporter.yaml",
		},
		{
			name:       "disable_all_metrics",
			input:      "testdata/disable_kubelet_utilization_metrics/disable_all_metrics_input.yaml",
			wantOutput: "testdata/disable_kubelet_utilization_metrics/disable_all_metrics_output.yaml",
		},
		{
			name:       "do_not_change_enabled_metrics",
			input:      "testdata/disable_kubelet_utilization_metrics/do_not_change_enabled_metrics_input.yaml",
			wantOutput: "testdata/disable_kubelet_utilization_metrics/do_not_change_enabled_metrics_output.yaml",
		},
		{
			name:       "all_metrics_included_in_signalfx_exporter",
			input:      "testdata/disable_kubelet_utilization_metrics/all_metrics_included_in_signalfx_exporter.yaml",
			wantOutput: "testdata/disable_kubelet_utilization_metrics/all_metrics_included_in_signalfx_exporter.yaml",
		},
		{
			name:       "partially_excluded_in_signalfx_exporter",
			input:      "testdata/disable_kubelet_utilization_metrics/partially_excluded_in_signalfx_exporter_input.yaml",
			wantOutput: "testdata/disable_kubelet_utilization_metrics/partially_excluded_in_signalfx_exporter_output.yaml",
		},
		{
			name:       "partially_included_in_signalfx_exporter",
			input:      "testdata/disable_kubelet_utilization_metrics/partially_included_in_signalfx_exporter_input.yaml",
			wantOutput: "testdata/disable_kubelet_utilization_metrics/partially_included_in_signalfx_exporter_output.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedCfgMap, err := confmaptest.LoadConf(tt.wantOutput)
			require.NoError(t, err)
			require.NotNil(t, expectedCfgMap)

			cfgMap, err := confmaptest.LoadConf(tt.input)
			require.NoError(t, err)
			require.NotNil(t, cfgMap)

			err = DisableKubeletUtilizationMetrics{}.Convert(context.Background(), cfgMap)
			require.NoError(t, err)

			assert.Equal(t, expectedCfgMap, cfgMap)
		})
	}
}
