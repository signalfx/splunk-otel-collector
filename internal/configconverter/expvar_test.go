// Copyright Splunk, Inc.
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

package configconverter

import (
	"context"
	"encoding/json"
	"expvar"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"
)

func TestGetExpvarConverter(t *testing.T) {
	converter1 := GetExpvarConverter()
	converter2 := GetExpvarConverter()

	// Verify both calls return the same instance (singleton)
	assert.Same(t, converter1, converter2)

	// Verify the converter is properly initialized
	assert.NotNil(t, converter1)
	assert.NotNil(t, converter1.initial)
	assert.NotNil(t, converter1.effective)
	assert.Empty(t, converter1.initial)
	assert.Empty(t, converter1.effective)

	// Verify expvar variables are published
	initialVar := expvar.Get("splunk.config.initial")
	effectiveVar := expvar.Get("splunk.config.effective")
	assert.NotNil(t, initialVar)
	assert.NotNil(t, effectiveVar)
}

func TestExpvarConverter_OnRetrieve(t *testing.T) {
	converter := GetExpvarConverter()

	// Test retrieving configuration for different schemes
	testData := map[string]any{
		"key1": "value1",
		"key2": map[string]any{
			"nested": "value2",
		},
	}

	converter.OnRetrieve("file", testData)

	// Verify data is stored in initial map
	assert.Contains(t, converter.initial, "file")
	assert.Equal(t, testData, converter.initial["file"])

	// Test multiple schemes
	envData := map[string]any{
		"env_key": "env_value",
	}
	converter.OnRetrieve("env", envData)

	assert.Contains(t, converter.initial, "file")
	assert.Contains(t, converter.initial, "env")
	assert.Equal(t, testData, converter.initial["file"])
	assert.Equal(t, envData, converter.initial["env"])
}

func TestExpvarConverter_Convert(t *testing.T) {
	converter := GetExpvarConverter()

	// Test converting configuration
	configData := map[string]any{
		"receivers": map[string]any{
			"otlp": map[string]any{
				"protocols": map[string]any{
					"grpc": nil,
					"http": nil,
				},
			},
		},
		"exporters": map[string]any{
			"debug": nil,
		},
		"service": map[string]any{
			"pipelines": map[string]any{
				"traces": map[string]any{
					"receivers": []any{"otlp"},
					"exporters": []any{"debug"},
				},
			},
		},
	}

	conf := confmap.NewFromStringMap(configData)
	err := converter.Convert(context.Background(), conf)

	require.NoError(t, err)
	assert.Equal(t, configData, converter.effective)
}

func TestExpvarConverter_Convert_EmptyConfig(t *testing.T) {
	converter := GetExpvarConverter()

	conf := confmap.NewFromStringMap(map[string]any{})
	err := converter.Convert(context.Background(), conf)

	require.NoError(t, err)
	assert.Empty(t, converter.effective)
}

func TestExpvarConverter_OnNew(t *testing.T) {
	converter := GetExpvarConverter()

	converter.initial = map[string]any{}
	converter.effective = map[string]any{}

	// OnNew should be a no-op
	converter.OnNew()

	// State should remain unchanged
	assert.Empty(t, converter.initial)
	assert.Empty(t, converter.effective)
}

func TestExpvarConverter_OnShutdown(t *testing.T) {
	converter := GetExpvarConverter()

	// Add some data first
	converter.OnRetrieve("test", map[string]any{"key": "value"})
	conf := confmap.NewFromStringMap(map[string]any{"test": "config"})
	err := converter.Convert(context.Background(), conf)
	require.NoError(t, err)

	// OnShutdown should be a no-op
	converter.OnShutdown()

	// State should remain unchanged
	assert.NotEmpty(t, converter.initial)
	assert.NotEmpty(t, converter.effective)
}

func TestExpvarConverter_SingletonPersistence(t *testing.T) {
	// First access creates the singleton
	converter1 := GetExpvarConverter()
	converter1.OnRetrieve("test", map[string]any{"initial": "data"})

	conf := confmap.NewFromStringMap(map[string]any{"effective": "data"})
	err := converter1.Convert(context.Background(), conf)
	require.NoError(t, err)

	// Second access should return the same instance with preserved state
	converter2 := GetExpvarConverter()
	assert.Same(t, converter1, converter2)
	assert.Contains(t, converter2.initial, "test")
	assert.Contains(t, converter2.effective, "effective")
}

func TestExpvarConverter_SimpleRedaction(t *testing.T) {
	// Test that simpleRedact correctly redacts sensitive information
	// uses all expected keys from the redactKeys map
	anchorsToBeRedacted := []string{
		"access",
		"api_key",
		"apikey",
		"auth",
		"credential",
		"creds",
		"login",
		"password",
		"pwd",
		"token",
		"user",
		"X-SF-Token",
	}

	unredacted := make(map[string]any)
	for _, anchor := range anchorsToBeRedacted {
		unredacted[anchor] = "unredacted_value"
	}
	unredacted["unrelated_key"] = "safe_value"

	converter := GetExpvarConverter()
	converter.initial = unredacted
	converter.effective = unredacted

	redacted := simpleRedact(unredacted)

	initialConfigJSONStr := expvar.Get("splunk.config.initial")
	var initialConfigYAMLStr string
	require.NoError(t, json.Unmarshal([]byte(initialConfigJSONStr.String()), &initialConfigYAMLStr))
	var initialConfigMap map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(initialConfigYAMLStr), &initialConfigMap))
	assert.Equal(t, redacted, initialConfigMap)

	effectiveConfigJSONStr := expvar.Get("splunk.config.effective")
	var effectiveConfigYAMLStr string
	require.NoError(t, json.Unmarshal([]byte(effectiveConfigJSONStr.String()), &effectiveConfigYAMLStr))
	var effectiveConfigMap map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(effectiveConfigYAMLStr), &effectiveConfigMap))
	assert.Equal(t, redacted, effectiveConfigMap)
}
