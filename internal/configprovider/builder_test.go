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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func TestConfigSourceBuild(t *testing.T) {
	ctx := context.Background()
	params := CreateParams{
		Logger:    zap.NewNop(),
		BuildInfo: component.NewDefaultBuildInfo(),
	}

	testFactories := Factories{
		"tstcfgsrc": &mockCfgSrcFactory{},
	}

	tests := []struct {
		configSettings     map[string]Source
		factories          Factories
		expectedCfgSources map[string]ConfigSource
		wantErr            string
		name               string
	}{
		{
			name: "unknown_config_source",
			configSettings: map[string]Source{
				"tstcfgsrc": &mockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.NewIDWithName("unknown_config_source", "tstcfgsrc")),
				},
			},
			factories: testFactories,
			wantErr:   "unknown unknown_config_source config source type for tstcfgsrc",
		},
		{
			name: "creation_error",
			configSettings: map[string]Source{
				"tstcfgsrc": &mockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.NewID("tstcfgsrc")),
				},
			},
			factories: Factories{
				"tstcfgsrc": &mockCfgSrcFactory{
					ErrOnCreateConfigSource: errors.New("forced test error"),
				},
			},
			wantErr: "failed to create config source tstcfgsrc: forced test error",
		},
		{
			name: "factory_return_nil",
			configSettings: map[string]Source{
				"tstcfgsrc": &mockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.NewID("tstcfgsrc")),
				},
			},
			factories: Factories{
				"tstcfgsrc": &mockNilCfgSrcFactory{},
			},
			wantErr: "factory for \"tstcfgsrc\" produced a nil extension",
		},
		{
			name: "base_case",
			configSettings: map[string]Source{
				"tstcfgsrc/named": &mockCfgSrcSettings{
					SourceSettings: NewSourceSettings(component.NewIDWithName("tstcfgsrc", "named")),
					Endpoint:       "some_endpoint",
					Token:          "some_token",
				},
			},
			factories: testFactories,
			expectedCfgSources: map[string]ConfigSource{
				"tstcfgsrc/named": &testConfigSource{
					ValueMap: map[string]valueEntry{
						"tstcfgsrc/named": {
							Value: &mockCfgSrcSettings{
								SourceSettings: NewSourceSettings(component.NewIDWithName("tstcfgsrc", "named")),
								Endpoint:       "some_endpoint",
								Token:          "some_token",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builtCfgSources, err := Build(ctx, tt.configSettings, params, tt.factories)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expectedCfgSources, builtCfgSources)
		})
	}
}

type mockNilCfgSrcFactory struct{}

func (m *mockNilCfgSrcFactory) Type() component.Type {
	return "tstcfgsrc"
}

var _ (Factory) = (*mockNilCfgSrcFactory)(nil)

func (m *mockNilCfgSrcFactory) CreateDefaultConfig() Source {
	return &mockCfgSrcSettings{
		SourceSettings: NewSourceSettings(component.NewID("tstcfgsrc")),
		Endpoint:       "default_endpoint",
	}
}

func (m *mockNilCfgSrcFactory) CreateConfigSource(context.Context, CreateParams, Source) (ConfigSource, error) {
	return nil, nil
}
