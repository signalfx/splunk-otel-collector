// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configsource

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
)

func TestConfigSourceParser(t *testing.T) {
	ctx := context.Background()

	testFactories := Factories{
		component.MustNewType("tstcfgsrc"): &MockCfgSrcFactory{},
	}
	tests := []struct {
		factories        Factories
		expectedSettings map[string]Settings
		expectedConf     map[string]any
		envvars          map[string]string
		wantErr          string
		name             string
		file             string
	}{
		{
			name:      "basic_config",
			file:      "basic_config",
			factories: testFactories,
			expectedSettings: map[string]Settings{
				"tstcfgsrc": &MockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.MustNewID("tstcfgsrc")),
					Endpoint:       "some_endpoint",
					Token:          "some_token",
				},
				"tstcfgsrc/named": &MockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.MustNewIDWithName("tstcfgsrc", "named")),
					Endpoint:       "default_endpoint",
				},
			},
		},
		{
			name:      "env_var_on_load",
			file:      "env_var_on_load",
			factories: testFactories,
			envvars: map[string]string{
				"ENV_VAR_ENDPOINT": "env_var_endpoint",
				"ENV_VAR_TOKEN":    "env_var_token",
			},
			expectedSettings: map[string]Settings{
				"tstcfgsrc": &MockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.MustNewID("tstcfgsrc")),
					Endpoint:       "https://env_var_endpoint:8200",
					Token:          "env_var_token",
				},
			},
			expectedConf: map[string]any{"ignored_by_parser": map[string]any{"some_field": "$ENV_VAR_TOKEN"}},
		},
		{
			name:      "cfgsrc_load_cannot_use_cfgsrc",
			file:      "cfgsrc_load_use_cfgsrc",
			factories: testFactories,
			wantErr:   "config source \"tstcfgsrc\" not found",
		},
		{
			name:      "bad_name",
			file:      "bad_name",
			factories: testFactories,
			wantErr:   "invalid config_sources type and name key \"tstcfgsrc/\"",
		},
		{
			name: "missing_factory",
			file: "basic_config",
			factories: Factories{
				component.MustNewType("not_in_basic_config"): &MockCfgSrcFactory{},
			},
			wantErr: "unknown config_sources type \"tstcfgsrc\"",
		},
		{
			name:      "unknown_field",
			file:      "unknown_field",
			factories: testFactories,
			wantErr:   "error reading config_sources configuration for \"tstcfgsrc\"",
		},
		{
			name:      "duplicated_name",
			file:      "duplicated_name",
			factories: testFactories,
			wantErr:   "duplicate config_sources name tstcfgsrc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgFile := path.Join("testdata", tt.file+".yaml")
			v, err := confmaptest.LoadConf(cfgFile)
			require.NoError(t, err)

			for key, value := range tt.envvars {
				require.NoError(t, os.Setenv(key, value))
				keyToUnset := key
				defer func() {
					assert.NoError(t, os.Unsetenv(keyToUnset))
				}()
			}

			cfgSrcSettings, splitConf, err := SettingsFromConf(ctx, v, tt.factories, map[string]confmap.Provider{
				"env": envprovider.NewFactory().Create(confmap.ProviderSettings{}),
			})
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, splitConf)
			} else {
				require.NoError(t, err)
				expectedConf := tt.expectedConf
				if expectedConf == nil {
					expectedConf = map[string]any{}
				}
				require.Equal(t, expectedConf, splitConf.ToStringMap())
			}
			assert.Equal(t, tt.expectedSettings, cfgSrcSettings)
		})
	}
}
