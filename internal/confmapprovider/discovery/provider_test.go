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

package discovery

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/internal/settings"
)

func TestConfigDProviderHappyPath(t *testing.T) {
	provider, err := New()
	require.NoError(t, err)
	require.NotNil(t, provider)

	assert.Equal(t, "splunk.configd", provider.ConfigDScheme())
	configD := provider.ConfigDProvider()
	assert.Equal(t, "splunk.configd", configD.Scheme())

	configDir := filepath.Join(".", "testdata", "config.d")
	retrieved, err := configD.Retrieve(context.Background(), fmt.Sprintf("%s:false%c%s", configD.Scheme(), rune(30), configDir), nil)
	assert.NoError(t, err)
	require.NotNil(t, retrieved)

	conf, err := retrieved.AsRaw()
	assert.NoError(t, err)
	assert.Equal(t, expectedServiceConfig, conf)

	assert.NoError(t, configD.Shutdown(context.Background()))
}

func TestConfigDProviderDifferentConfigDirs(t *testing.T) {
	provider, err := New()
	require.NoError(t, err)
	require.NotNil(t, provider)

	configD := provider.ConfigDProvider()
	configDir := filepath.Join(".", "testdata", "config.d")
	retrieved, err := configD.Retrieve(context.Background(), fmt.Sprintf("%s:false%c%s", configD.Scheme(), rune(30), configDir), nil)
	assert.NoError(t, err)
	require.NotNil(t, retrieved)
	conf, err := retrieved.AsRaw()
	assert.NoError(t, err)
	assert.Equal(t, expectedServiceConfig, conf)

	configDir = filepath.Join(".", "testdata", "another-config.d")
	retrieved, err = configD.Retrieve(context.Background(), fmt.Sprintf("%s:false%c%s", configD.Scheme(), rune(30), configDir), nil)
	assert.NoError(t, err)
	require.NotNil(t, retrieved)
	conf, err = retrieved.AsRaw()
	assert.NoError(t, err)
	anotherExpectedServiceConfig := map[string]any{
		"exporters": map[string]any{
			"signalfx": map[string]any{
				"api_url":    "http://0.0.0.0/different-api",
				"ingest_url": "http://0.0.0.0/different-ingest",
			},
		},
		"extensions": map[string]any{},
		"processors": map[string]any{},
		"receivers":  map[string]any{},
		"service":    map[string]any{},
	}
	assert.Equal(t, anotherExpectedServiceConfig, conf)
}

func TestConfigDProviderInvalidURIs(t *testing.T) {
	provider, err := New()
	require.NoError(t, err)
	require.NotNil(t, provider)

	configD := provider.ConfigDProvider()
	require.NotNil(t, configD)
	retrieved, err := configD.Retrieve(context.Background(), "not.a.thing:not.a.path", nil)
	assert.EqualError(t, err, `uri failed validation: uri "not.a.thing:not.a.path" is not supported by splunk.configd provider`)
	assert.Nil(t, retrieved)

	retrieved, err = configD.Retrieve(context.Background(), fmt.Sprintf("%s:not.a.path", settings.DiscoveryModeScheme), nil)
	assert.EqualError(t, err, `uri failed validation: uri "splunk.discovery:not.a.path" is not supported by splunk.configd provider`)
	assert.Nil(t, retrieved)
}

func TestConfigDirAndDryRun(t *testing.T) {
	for _, test := range []struct {
		expectedConfigDir, name, uri, scheme, expectedError string
		expectedDryRun                                      bool
	}{
		{
			name:              "dry run configd",
			uri:               fmt.Sprintf("splunk.configd:true%csome.config.dir", rune(30)),
			scheme:            "splunk.configd",
			expectedError:     "",
			expectedDryRun:    true,
			expectedConfigDir: "some.config.dir",
		},
		{
			name:              "dry run discovery mode",
			uri:               fmt.Sprintf("splunk.discovery:true%csome.config.dir", rune(30)),
			scheme:            "splunk.discovery",
			expectedError:     "",
			expectedDryRun:    true,
			expectedConfigDir: "some.config.dir",
		},
		{
			name:              "no dry run configd",
			uri:               fmt.Sprintf("splunk.configd:false%csome.config.dir", rune(30)),
			scheme:            "splunk.configd",
			expectedError:     "",
			expectedDryRun:    false,
			expectedConfigDir: "some.config.dir",
		},
		{
			name:              "no dry run discovery mode",
			uri:               fmt.Sprintf("splunk.discovery:false%csome.config.dir", rune(30)),
			scheme:            "splunk.discovery",
			expectedError:     "",
			expectedDryRun:    false,
			expectedConfigDir: "some.config.dir",
		},
		{
			name:              "invalid dry run configd",
			uri:               fmt.Sprintf("splunk.configd:notabool%csome.config.dir", rune(30)),
			scheme:            "splunk.configd",
			expectedError:     fmt.Sprintf(`invalid dry run arg "notabool" from %q`, fmt.Sprintf("splunk.configd:notabool%csome.config.dir", rune(30))),
			expectedDryRun:    false,
			expectedConfigDir: "",
		},
		{
			name:              "invalid dry run discovery mode",
			uri:               fmt.Sprintf("splunk.discovery:notabool%csome.config.dir", rune(30)),
			scheme:            "splunk.discovery",
			expectedError:     fmt.Sprintf(`invalid dry run arg "notabool" from %q`, fmt.Sprintf("splunk.discovery:notabool%csome.config.dir", rune(30))),
			expectedDryRun:    false,
			expectedConfigDir: "",
		},
		{
			name:              "config.d missing dryRun uri",
			uri:               "splunk.configd:invalid.uri",
			scheme:            "splunk.configd",
			expectedError:     `invalid uri missing record separator: "splunk.configd:invalid.uri"`,
			expectedDryRun:    false,
			expectedConfigDir: "",
		},
		{
			name:              "discovery missing dryRun uri",
			uri:               "splunk.discovery:invalid.uri",
			scheme:            "splunk.discovery",
			expectedError:     `invalid uri missing record separator: "splunk.discovery:invalid.uri"`,
			expectedDryRun:    false,
			expectedConfigDir: "",
		},
		{
			name:              "invalid scheme",
			uri:               "some.uri",
			scheme:            "not.a.valid.scheme",
			expectedError:     `uri "some.uri" is not supported by not.a.valid.scheme provider`,
			expectedDryRun:    false,
			expectedConfigDir: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			configDir, dryRun, err := configDirAndDryRun(test.uri, test.scheme)
			require.Equal(t, test.expectedConfigDir, configDir)
			require.Equal(t, test.expectedDryRun, dryRun)
			if test.expectedError == "" {
				require.Nil(t, err)
			} else {
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}
