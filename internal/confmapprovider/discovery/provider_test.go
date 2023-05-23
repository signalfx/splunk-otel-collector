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
)

func TestConfigDProviderHappyPath(t *testing.T) {
	provider, err := New()
	require.NoError(t, err)
	require.NotNil(t, provider)

	assert.Equal(t, "splunk.configd", provider.ConfigDScheme())
	configD := provider.ConfigDProvider()
	assert.Equal(t, "splunk.configd", configD.Scheme())

	configDir := filepath.Join(".", "testdata", "config.d")
	retrieved, err := configD.Retrieve(context.Background(), fmt.Sprintf("%s:%s", configD.Scheme(), configDir), nil)
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
	retrieved, err := configD.Retrieve(context.Background(), fmt.Sprintf("%s:%s", configD.Scheme(), configDir), nil)
	assert.NoError(t, err)
	require.NotNil(t, retrieved)
	conf, err := retrieved.AsRaw()
	assert.NoError(t, err)
	assert.Equal(t, expectedServiceConfig, conf)

	configDir = filepath.Join(".", "testdata", "another-config.d")
	retrieved, err = configD.Retrieve(context.Background(), fmt.Sprintf("%s:%s", configD.Scheme(), configDir), nil)
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
	assert.EqualError(t, err, `uri "not.a.thing:not.a.path" is not supported by splunk.configd provider`)
	assert.Nil(t, retrieved)

	retrieved, err = configD.Retrieve(context.Background(), fmt.Sprintf("%s:not.a.path", discoveryModeScheme), nil)
	assert.EqualError(t, err, `uri "splunk.discovery:not.a.path" is not supported by splunk.configd provider`)
	assert.Nil(t, retrieved)
}
