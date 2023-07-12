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

package properties

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvVarPropertyEBNF(t *testing.T) {
	require.Equal(t, `EnvVarProperty = "SPLUNK" <underscore> "DISCOVERY" <underscore> ("RECEIVERS" | "EXTENSIONS") <underscore> EnvVarComponentID <underscore> ("CONFIG" | "ENABLED") (<underscore> (<string> | <underscore>)+)* .
EnvVarComponentID = ~(<underscore> (?= ("CONFIG" | "ENABLED")))+ (<underscore> "x2f" <underscore> (~(?= <underscore> (?= ("CONFIG" | "ENABLED")))+ | ""))? .`, envVarParser.String())
}

func TestValidEnvVarProperties(t *testing.T) {
	for _, tt := range []struct {
		expected *Property
		envVar   string
		val      string
	}{
		{envVar: "SPLUNK_DISCOVERY_RECEIVERS_receiver_x2d_type_x2f__CONFIG_one",
			val: "val",
			expected: &Property{
				stringMap: map[string]any{
					"receivers": map[string]any{
						"receiver-type": map[string]any{
							"config": map[string]any{
								"one": "val"},
						},
					},
				},
				ComponentType: "receivers",
				Component:     ComponentID{Type: "receiver-type"},
				Type:          "config",
				Key:           "one",
				Val:           "val",
				Input:         "splunk.discovery.receivers.receiver-type/.config.one",
			},
		},
		{envVar: "SPLUNK_DISCOVERY_EXTENSIONS_extension_x2e_type_x2f_extension____name_CONFIG_one_x3a__x3a_two",
			val: "a.val",
			expected: &Property{
				stringMap: map[string]any{
					"extensions": map[string]any{
						"extension.type/extension____name": map[string]any{
							"config": map[string]any{
								"one": map[string]any{
									"two": "a.val",
								},
							},
						},
					},
				},
				ComponentType: "extensions",
				Component:     ComponentID{Type: "extension.type", Name: "extension____name"},
				Type:          "config",
				Key:           "one::two",
				Val:           "a.val",
				Input:         "splunk.discovery.extensions.extension.type/extension____name.config.one::two",
			},
		},
		{envVar: "SPLUNK_DISCOVERY_EXTENSIONS_extension_x2e_type_x2f_extension____name_ENABLED",
			val: "False",
			expected: &Property{
				stringMap: map[string]any{
					"extensions": map[string]any{
						"extension.type/extension____name": map[string]any{
							"enabled": "false",
						},
					},
				},
				ComponentType: "extensions",
				Component:     ComponentID{Type: "extension.type", Name: "extension____name"},
				Type:          "enabled",
				Val:           "false",
				Input:         "splunk.discovery.extensions.extension.type/extension____name.enabled",
			},
		},
		{envVar: "SPLUNK_DISCOVERY_RECEIVERS_receiver_x2d_type_x2f__ENABLED",
			val: "true",
			expected: &Property{
				stringMap: map[string]any{
					"receivers": map[string]any{
						"receiver-type": map[string]any{
							"enabled": "true",
						},
					},
				},
				ComponentType: "receivers",
				Component:     ComponentID{Type: "receiver-type"},
				Type:          "enabled",
				Val:           "true",
				Input:         "splunk.discovery.receivers.receiver-type/.enabled",
			},
		},
	} {
		t.Run(tt.envVar, func(t *testing.T) {
			p, ok, err := NewPropertyFromEnvVar(tt.envVar, tt.val)
			require.True(t, ok)
			require.NoError(t, err)
			require.NotNil(t, p)
			require.Equal(t, tt.expected, p)
		})
	}
}

func TestInvalidEnvVarProperties(t *testing.T) {
	for _, tt := range []struct {
		envVar, expectedError string
	}{
		{envVar: "SPLUNK_DISCOVERY_NOTVALIDCOMPONENT_TYPE_CONFIG_ONE", expectedError: "invalid env var property (parsing error): invalid property env var (parsing error): SPLUNK_DISCOVERY:1:18: unexpected token \"NOTVALIDCOMPONENT\" (expected (\"RECEIVERS\" | \"EXTENSIONS\") <underscore> EnvVarComponentID <underscore> (\"CONFIG\" | \"ENABLED\") (<underscore> (<string> | <underscore>)+)*)"},
		{envVar: "SPLUNK_DISCOVERY_RECEIVERS_TYPE_NOTCONFIG_ONE", expectedError: "invalid env var property (parsing error): invalid property env var (parsing error): SPLUNK_DISCOVERY:1:46: unexpected token \"<EOF>\" (expected <underscore> (\"CONFIG\" | \"ENABLED\") (<underscore> (<string> | <underscore>)+)*)"},
		{envVar: "SPLUNK_DISCOVERY_EXTENSIONS_TYPE_x2f_NAME_CONFIG_", expectedError: "invalid env var property (parsing error): invalid property env var (parsing error): SPLUNK_DISCOVERY:1:50: sub-expression (<string> | <underscore>)+ must match at least once"},
	} {
		t.Run(tt.envVar, func(t *testing.T) {
			p, ok, err := NewPropertyFromEnvVar(tt.envVar, "val")
			require.True(t, ok)
			require.Error(t, err)
			require.EqualError(t, err, tt.expectedError)
			require.Nil(t, p)
		})
	}
}
