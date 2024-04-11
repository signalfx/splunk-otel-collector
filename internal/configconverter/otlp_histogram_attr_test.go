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

func TestOTLPHistograms_Enabled(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/otlp_histograms_attr/send_otlp_histrograms_enabled.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)
	pmp := AddOTLPHistogramAttr{}
	assert.NoError(t, pmp.Convert(context.Background(), cfgMap))
	cfgMapExpected, err := confmaptest.LoadConf("testdata/otlp_histograms_attr/send_otlp_histrograms_enabled_expected.yaml")
	require.NoError(t, err)
	assert.Equal(t, cfgMapExpected.ToStringMap(), cfgMap.ToStringMap())
}

func TestOTLPHistograms_EnabledNoTelemetry(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/otlp_histograms_attr/send_otlp_histrograms_enabled_no_telemetry.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)
	pmp := AddOTLPHistogramAttr{}
	assert.NoError(t, pmp.Convert(context.Background(), cfgMap))
	cfgMapExpected, err := confmaptest.LoadConf("testdata/otlp_histograms_attr/send_otlp_histrograms_enabled_no_telemetry_expected.yaml")
	require.NoError(t, err)
	assert.Equal(t, cfgMapExpected.ToStringMap(), cfgMap.ToStringMap())
}

func TestOTLPHistograms_Disabled(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/otlp_histograms_attr/send_otlp_histrograms_disabled.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)
	pmp := AddOTLPHistogramAttr{}
	assert.NoError(t, pmp.Convert(context.Background(), cfgMap))
	cfgMapExpected, err := confmaptest.LoadConf("testdata/otlp_histograms_attr/send_otlp_histrograms_disabled_expected.yaml")
	require.NoError(t, err)
	assert.Equal(t, cfgMapExpected.ToStringMap(), cfgMap.ToStringMap())
}

func TestOTLPHistograms_DisabledExporterNotInUse(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/otlp_histograms_attr/send_otlp_histrograms_disabled_no_exp.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)
	pmp := AddOTLPHistogramAttr{}
	assert.NoError(t, pmp.Convert(context.Background(), cfgMap))
	cfgMapExpected, err := confmaptest.LoadConf("testdata/otlp_histograms_attr/send_otlp_histrograms_disabled_no_exp_expected.yaml")
	require.NoError(t, err)
	assert.Equal(t, cfgMapExpected.ToStringMap(), cfgMap.ToStringMap())
}
