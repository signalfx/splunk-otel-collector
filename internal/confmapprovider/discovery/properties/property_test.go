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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropertyEBNF(t *testing.T) {
	require.Equal(t, `Property = "splunk" <dot> "discovery" <dot> ("receivers" | "extensions") <dot> ComponentID <dot> (("config" <dot>) | "enabled") (<string> | <dot> | <forwardslash>)* .
ComponentID = ~(<forwardslash> | (<dot> (?= ("config" | "enabled"))))+ (<forwardslash> ~(<dot> (?= ("config" | "enabled")))+*)? .`, parser.String())
}

func TestWordifyHappyPath(t *testing.T) {
	cBytes, err := os.ReadFile(filepath.Join(".", "testdata", "utf8-corpus.txt"))
	require.NoError(t, err)
	corpus := string(cBytes)
	nonWordRE := regexp.MustCompile(`[^\w]+`)
	require.True(t, nonWordRE.MatchString(corpus))

	wordifiedCorpus := wordify(corpus)
	require.False(t, nonWordRE.MatchString(wordifiedCorpus))

	unwordifiedCorpus, err := unwordify(wordifiedCorpus)
	require.NoError(t, err)
	require.Equal(t, corpus, unwordifiedCorpus)
}

func TestUnwordifyInvalidHexInEscapedForm(t *testing.T) {
	notHex := "_xnothex__xgggg "
	out, err := unwordify(notHex)
	require.NoError(t, err)
	require.Equal(t, notHex, out)

	notHex = "_xfffff__xa_"
	out, err = unwordify(notHex)
	assert.Empty(t, out)
	require.EqualError(
		t, err,
		`failed parsing env var hex-encoded content: "fffff": encoding/hex: odd length hex string; "a": encoding/hex: odd length hex string`,
	)
}

func TestValidProperties(t *testing.T) {
	for _, tt := range []struct {
		expected *Property
		key      string
		val      string
	}{
		{key: "splunk.discovery.receivers.receivertype.config.key", val: "val",
			expected: &Property{
				ComponentType: "receivers",
				Component:     ComponentID{Type: "receivertype"},
				Type:          "config",
				Key:           "key",
				Val:           "val",
				stringMap: map[string]any{
					"receivers": map[string]any{
						"receivertype": map[string]any{
							"config": map[string]any{
								"key": "val",
							},
						},
					},
				},
				Input: "splunk.discovery.receivers.receivertype.config.key",
			},
		},
		{key: "splunk.discovery.extensions.extension-type/extensionname.config.key", val: "val",
			expected: &Property{
				ComponentType: "extensions",
				Component:     ComponentID{Type: "extension-type", Name: "extensionname"},
				Type:          "config",
				Key:           "key",
				Val:           "val",
				stringMap: map[string]any{
					"extensions": map[string]any{
						"extension-type/extensionname": map[string]any{
							"config": map[string]any{
								"key": "val",
							},
						},
					},
				},
				Input: "splunk.discovery.extensions.extension-type/extensionname.config.key",
			},
		},
		{key: "splunk.discovery.receivers.receivertype/.config.key", val: "val",
			expected: &Property{
				ComponentType: "receivers",
				Component:     ComponentID{Type: "receivertype"},
				Type:          "config",
				Key:           "key",
				Val:           "val",
				stringMap: map[string]any{
					"receivers": map[string]any{
						"receivertype": map[string]any{
							"config": map[string]any{
								"key": "val",
							},
						},
					},
				},
				Input: "splunk.discovery.receivers.receivertype/.config.key",
			},
		},
		{key: "splunk.discovery.receivers.receiver_type/config.config.one::two::three", val: "val",
			expected: &Property{
				ComponentType: "receivers",
				Component:     ComponentID{Type: "receiver_type", Name: "config"},
				Type:          "config",
				Key:           "one::two::three",
				Val:           "val",
				stringMap: map[string]any{
					"receivers": map[string]any{
						"receiver_type/config": map[string]any{
							"config": map[string]any{
								"one": map[string]any{"two": map[string]any{"three": "val"}},
							},
						},
					},
				},
				Input: "splunk.discovery.receivers.receiver_type/config.config.one::two::three",
			},
		},
		{key: "splunk.discovery.receivers.receiver.type////.config.one::config", val: "val",
			expected: &Property{
				ComponentType: "receivers",
				Component:     ComponentID{Type: "receiver.type", Name: "///"},
				Type:          "config",
				Key:           "one::config",
				Val:           "val",
				stringMap: map[string]any{
					"receivers": map[string]any{
						"receiver.type////": map[string]any{
							"config": map[string]any{
								"one": map[string]any{"config": "val"}},
						},
					},
				},
				Input: "splunk.discovery.receivers.receiver.type////.config.one::config",
			},
		},
		{key: "splunk.discovery.extensions.extension--0-1-with-config-in-type-_x64__x86_🙈🙉🙊4:000x0;;0;;0;;-___-----type/e/x/t/e%ns<i>o<=n=>nam/e-with-config.config.o::n::e.config", val: "val",
			expected: &Property{
				ComponentType: "extensions",
				Component:     ComponentID{Type: "extension--0-1-with-config-in-type-_x64__x86_🙈🙉🙊4:000x0;;0;;0;;-___-----type", Name: "e/x/t/e%ns<i>o<=n=>nam/e-with-config"},
				Type:          "config",
				Key:           "o::n::e.config",
				Val:           "val",
				stringMap: map[string]any{
					"extensions": map[string]any{
						"extension--0-1-with-config-in-type-_x64__x86_🙈🙉🙊4:000x0;;0;;0;;-___-----type/e/x/t/e%ns<i>o<=n=>nam/e-with-config": map[string]any{
							"config": map[string]any{
								"o": map[string]any{"n": map[string]any{"e.config": "val"}}},
						},
					},
				},
				Input: "splunk.discovery.extensions.extension--0-1-with-config-in-type-_x64__x86_🙈🙉🙊4:000x0;;0;;0;;-___-----type/e/x/t/e%ns<i>o<=n=>nam/e-with-config.config.o::n::e.config",
			},
		},
		{key: "splunk.discovery.receivers.receiver.type////.enabled", val: "false",
			expected: &Property{
				stringMap: map[string]any{
					"receivers": map[string]any{
						"receiver.type////": map[string]any{
							"enabled": "false",
						},
					},
				},
				ComponentType: "receivers",
				Component:     ComponentID{Type: "receiver.type", Name: "///"},
				Type:          "enabled",
				Key:           "",
				Val:           "false",
				Input:         "splunk.discovery.receivers.receiver.type////.enabled",
			},
		},
		{key: "splunk.discovery.receivers.receiver.type////.enabled", val: "T",
			expected: &Property{
				stringMap: map[string]any{
					"receivers": map[string]any{
						"receiver.type////": map[string]any{
							"enabled": "true",
						},
					},
				},
				ComponentType: "receivers",
				Component:     ComponentID{Type: "receiver.type", Name: "///"},
				Type:          "enabled",
				Key:           "",
				Val:           "true",
				Input:         "splunk.discovery.receivers.receiver.type////.enabled",
			},
		},
	} {
		t.Run(fmt.Sprintf("%s=%s", tt.key, tt.val), func(t *testing.T) {
			p, err := NewProperty(tt.key, tt.val)
			require.NoError(t, err)
			if tt.expected == nil {
				require.Nil(t, p)
				return
			}
			require.NotNil(t, p)
			require.Equal(t, *tt.expected, *p)

			// confirm env var rendering and equivalence
			envVarP, ok, err := NewPropertyFromEnvVar(p.ToEnvVar(), tt.val)
			require.True(t, ok)
			require.NoError(t, err)
			require.Equal(t, p, envVarP)
		})
	}
}

func TestInvalidProperties(t *testing.T) {
	for _, tt := range []struct {
		property, expectedError string
	}{
		{property: "splunk.discovery.invalid", expectedError: "invalid property \"splunk.discovery.invalid\" (parsing error): splunk.discovery:1:18: unexpected token \"invalid\" (expected (\"receivers\" | \"extensions\") <dot> ComponentID <dot> ((\"config\" <dot>) | \"enabled\") (<string> | <dot> | <forwardslash>)*)"},
		{property: "splunk.discovery.extensions.config.one.two", expectedError: "invalid property \"splunk.discovery.extensions.config.one.two\" (parsing error): splunk.discovery:1:43: unexpected token \"<EOF>\" (expected <dot> ((\"config\" <dot>) | \"enabled\") (<string> | <dot> | <forwardslash>)*)"},
		{property: "splunk.discovery.receivers.type/name.config", expectedError: "invalid property \"splunk.discovery.receivers.type/name.config\" (parsing error): splunk.discovery:1:44: unexpected token \"<EOF>\" (expected <dot>)"},
	} {
		t.Run(tt.property, func(t *testing.T) {
			p, err := NewProperty(tt.property, "val")
			require.Error(t, err)
			require.EqualError(t, err, tt.expectedError)
			require.Nil(t, p)
		})
	}
}
