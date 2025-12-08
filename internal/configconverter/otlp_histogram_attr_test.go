// Copyright The OpenTelemetry Authors
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

// Taken from https://github.com/open-telemetry/opentelemetry-collector/blob/v0.66.0/confmap/converter/overwritepropertiesconverter/properties_test.go
// to prevent breaking changes.
package configconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestOTLPHistogramsAttrs(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput string
	}{
		{
			name:       "enabled",
			input:      "testdata/otlp_histograms_attr/send_otlp_histograms_enabled.yaml",
			wantOutput: "testdata/otlp_histograms_attr/send_otlp_histograms_enabled_expected.yaml",
		},
		{
			name:       "enabled_no_telemetry",
			input:      "testdata/otlp_histograms_attr/send_otlp_histograms_enabled_no_telemetry.yaml",
			wantOutput: "testdata/otlp_histograms_attr/send_otlp_histograms_enabled_no_telemetry_expected.yaml",
		},
		{
			name:       "disabled",
			input:      "testdata/otlp_histograms_attr/send_otlp_histograms_disabled.yaml",
			wantOutput: "testdata/otlp_histograms_attr/send_otlp_histograms_disabled_expected.yaml",
		},
		{
			name:       "disabled_not_in_use_exporter",
			input:      "testdata/otlp_histograms_attr/send_otlp_histograms_disabled_no_exp.yaml",
			wantOutput: "testdata/otlp_histograms_attr/send_otlp_histograms_disabled_no_exp_expected.yaml",
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

			err = AddOTLPHistogramAttr(context.Background(), cfgMap)
			require.NoError(t, err)

			assert.Equal(t, expectedCfgMap.ToStringMap(), cfgMap.ToStringMap())
		})
	}
}
