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

func TestDisableExcessiveInternalMetrics(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput string
	}{
		{
			name:       "no_prom_receiver",
			input:      "testdata/disable_excessive_internal_metrics/no_prom_receiver.yaml",
			wantOutput: "testdata/disable_excessive_internal_metrics/no_prom_receiver.yaml",
		},
		{
			name:       "no_scrape_configs_prom_receiver",
			input:      "testdata/disable_excessive_internal_metrics/no_scrape_configs.yaml",
			wantOutput: "testdata/disable_excessive_internal_metrics/no_scrape_configs.yaml",
		},
		{
			name:       "different_job",
			input:      "testdata/disable_excessive_internal_metrics/different_job.yaml",
			wantOutput: "testdata/disable_excessive_internal_metrics/different_job.yaml",
		},
		{
			name:       "no_metric_relabel_configs_set",
			input:      "testdata/disable_excessive_internal_metrics/no_metric_relabel_configs_set.yaml",
			wantOutput: "testdata/disable_excessive_internal_metrics/no_metric_relabel_configs_set.yaml",
		},
		{
			name:       "metric_relabel_configs_with_other_actions",
			input:      "testdata/disable_excessive_internal_metrics/metric_relabel_configs_with_other_actions.yaml",
			wantOutput: "testdata/disable_excessive_internal_metrics/metric_relabel_configs_with_other_actions.yaml",
		},
		{
			name:       "metric_relabel_configs_with_batch_drop_action",
			input:      "testdata/disable_excessive_internal_metrics/old_metric_relabel_configs_present_input.yaml",
			wantOutput: "testdata/disable_excessive_internal_metrics/old_metric_relabel_configs_present_output.yaml",
		},
		{
			name:       "all_metric_relabel_configs_are_present",
			input:      "testdata/disable_excessive_internal_metrics/all_metric_relabel_configs_are_present.yaml",
			wantOutput: "testdata/disable_excessive_internal_metrics/all_metric_relabel_configs_are_present.yaml",
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

			err = DisableExcessiveInternalMetrics{}.Convert(context.Background(), cfgMap)
			require.NoError(t, err)

			assert.Equal(t, expectedCfgMap, cfgMap)
		})
	}
}
