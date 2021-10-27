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

package configprovider

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	expcfg "go.opentelemetry.io/collector/config/experimental/config"
	"go.opentelemetry.io/collector/config/experimental/configsource"
)

func TestConfigSourceParser(t *testing.T) {
	ctx := context.Background()

	testFactories := Factories{
		"tstcfgsrc": &mockCfgSrcFactory{},
	}
	tests := []struct {
		factories Factories
		expected  map[string]expcfg.Source
		envvars   map[string]string
		wantErr   error
		name      string
		file      string
	}{
		{
			name:      "basic_config",
			file:      "basic_config",
			factories: testFactories,
			expected: map[string]expcfg.Source{
				"tstcfgsrc": &mockCfgSrcSettings{
					SourceSettings: expcfg.NewSourceSettings(config.NewComponentID("tstcfgsrc")),
					Endpoint:       "some_endpoint",
					Token:          "some_token",
				},
				"tstcfgsrc/named": &mockCfgSrcSettings{
					SourceSettings: expcfg.NewSourceSettings(config.NewComponentIDWithName("tstcfgsrc", "named")),
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
			expected: map[string]expcfg.Source{
				"tstcfgsrc": &mockCfgSrcSettings{
					SourceSettings: expcfg.NewSourceSettings(config.NewComponentID("tstcfgsrc")),
					Endpoint:       "https://env_var_endpoint:8200",
					Token:          "env_var_token",
				},
			},
		},
		{
			name:      "cfgsrc_load_cannot_use_cfgsrc",
			file:      "cfgsrc_load_use_cfgsrc",
			factories: testFactories,
			wantErr:   &errUnknownConfigSource{},
		},
		{
			name:      "bad_name",
			file:      "bad_name",
			factories: testFactories,
			wantErr:   &errInvalidTypeAndNameKey{},
		},
		{
			name: "missing_factory",
			file: "basic_config",
			factories: Factories{
				"not_in_basic_config": &mockCfgSrcFactory{},
			},
			wantErr: &errUnknownType{},
		},
		{
			name:      "unknown_field",
			file:      "unknown_field",
			factories: testFactories,
			wantErr:   &errUnmarshalError{},
		},
		{
			name:      "duplicated_name",
			file:      "duplicated_name",
			factories: testFactories,
			wantErr:   &errDuplicateName{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgFile := path.Join("testdata", tt.file+".yaml")
			v, err := config.NewMapFromFile(cfgFile)
			require.NoError(t, err)

			for key, value := range tt.envvars {
				require.NoError(t, os.Setenv(key, value))
				keyToUnset := key
				defer func() {
					assert.NoError(t, os.Unsetenv(keyToUnset))
				}()
			}

			cfgSrcSettings, err := Load(ctx, v, tt.factories)
			require.IsType(t, tt.wantErr, err)
			assert.Equal(t, tt.expected, cfgSrcSettings)
		})
	}
}

type mockCfgSrcSettings struct {
	expcfg.SourceSettings
	Endpoint string `mapstructure:"endpoint"`
	Token    string `mapstructure:"token"`
}

func (m mockCfgSrcSettings) Validate() error {
	return nil
}

var _ expcfg.Source = (*mockCfgSrcSettings)(nil)

type mockCfgSrcFactory struct {
	ErrOnCreateConfigSource error
}

var _ Factory = (*mockCfgSrcFactory)(nil)

func (m *mockCfgSrcFactory) Type() config.Type {
	return "tstcfgsrc"
}

func (m *mockCfgSrcFactory) CreateDefaultConfig() expcfg.Source {
	return &mockCfgSrcSettings{
		SourceSettings: expcfg.NewSourceSettings(config.NewComponentID("tstcfgsrc")),
		Endpoint:       "default_endpoint",
	}
}

func (m *mockCfgSrcFactory) CreateConfigSource(_ context.Context, _ CreateParams, cfg expcfg.Source) (configsource.ConfigSource, error) {
	if m.ErrOnCreateConfigSource != nil {
		return nil, m.ErrOnCreateConfigSource
	}
	return &testConfigSource{
		ValueMap: map[string]valueEntry{
			cfg.ID().String(): {
				Value: cfg,
			},
		},
	}, nil
}
