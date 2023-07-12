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

package properties

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestConf(t *testing.T) {
	for _, path := range []string{
		filepath.Join(".", "testdata", "valid-mapping.yaml"),
		filepath.Join(".", "testdata", "valid-set.yaml"),
		filepath.Join(".", "testdata", "valid-mix.yaml"),
	} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			conf, err := confmaptest.LoadConf(path)
			require.NoError(t, err)
			loaded, warning, fatal := LoadConf(conf.ToStringMap())
			require.NoError(t, warning)
			require.NoError(t, fatal)
			require.Equal(t, map[string]any{
				"extensions": map[string]any{
					"docker_observer": map[string]any{
						"enabled": "false",
					},
					"host_observer/with_a_name": map[string]any{
						"config": map[string]any{
							"refresh_interval": "1h",
						},
					},
				},
				"receivers": map[string]any{
					"a_receiver": map[string]any{
						"config": map[string]any{
							"some_field": "some_value",
						},
					},
					"another_receiver/with-name": map[string]any{
						"config": map[string]any{
							"parent": map[string]any{
								"child_one": map[string]any{
									"another_field": "another_value",
									"child_two": map[string]any{
										"another_field":     "another_value",
										"yet_another_field": "yet_another_value",
									},
								},
							},
						},
					},
				},
			}, loaded.ToStringMap())
		})
	}
}

func TestInvalidPropertiesIgnoredWhenLoadingConf(t *testing.T) {
	for _, fixture := range []struct {
		expectedError string
		path          string
	}{
		{
			expectedError: `unknown property "splunk.discovery.receivers.a_receiver.unknown"; unknown property "splunk.discovery.receivers.another_receiver/with-name.another_unknown"`,
			path:          filepath.Join(".", "testdata", "invalid-mapping-unknown-subfield.yaml"),
		},
		{
			expectedError: `unknown discovery property mapping "splunk.discovery.123"; unknown discovery property mapping "splunk.discovery.processors"`,
			path:          filepath.Join(".", "testdata", "invalid-mapping-unknown-component.yaml"),
		},
		{
			expectedError: `invalid property "splunk.discovery.123" (parsing error): splunk.discovery:1:18: unexpected token "123" (expected ("receivers" | "extensions") <dot> ComponentID <dot> (("config" <dot>) | "enabled") (<string> | <dot> | <forwardslash>)*); invalid property "splunk.discovery.processors.a_processor.config.some_field" (parsing error): splunk.discovery:1:18: unexpected token "processors" (expected ("receivers" | "extensions") <dot> ComponentID <dot> (("config" <dot>) | "enabled") (<string> | <dot> | <forwardslash>)*); invalid property "splunk.discovery.receivers.a_receiver.unknown.some_field" (parsing error): splunk.discovery:1:57: unexpected token "<EOF>" (expected <dot> (("config" <dot>) | "enabled") (<string> | <dot> | <forwardslash>)*); invalid property "splunk.discovery.receivers.another_receiver/with-name.another_unknown.some_other_field" (parsing error): splunk.discovery:1:87: unexpected token "<EOF>" (expected <dot> (("config" <dot>) | "enabled") (<string> | <dot> | <forwardslash>)*)`,
			path:          filepath.Join(".", "testdata", "invalid-set.yaml"),
		},
	} {
		t.Run(filepath.Base(fixture.path), func(t *testing.T) {
			conf, err := confmaptest.LoadConf(fixture.path)
			require.NoError(t, err)
			require.NotNil(t, conf)
			loaded, warning, fatal := LoadConf(conf.ToStringMap())
			require.NoError(t, fatal)
			require.EqualError(t, warning, fixture.expectedError)
			require.Equal(t, map[string]any{
				"receivers": map[string]any{
					"another_receiver/with-name": map[string]any{
						"config": map[string]any{
							"parent": map[string]any{
								"child_one": map[string]any{
									"another_field": "another_value",
									"child_two": map[string]any{
										"another_field":     "another_value",
										"yet_another_field": "yet_another_value",
									},
								},
							},
						},
					},
				},
			}, loaded.ToStringMap())
		})
	}
}
