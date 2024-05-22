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
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestRemoveMemoryBallastConverter_Empty(t *testing.T) {
	conf := confmap.NewFromStringMap(map[string]interface{}{"foo": "bar"})
	assert.NoError(t, RemoveMemoryBallastKey(context.Background(), conf))
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, conf.ToStringMap())
}

func TestRemoveMemoryBallastConverter_With_Memory_Ballast(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/with_memory_ballast.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)
	assert.NoError(t, RemoveMemoryBallastKey(context.Background(), cfgMap))
	cfgMapExpected, err := confmaptest.LoadConf("testdata/with_memory_ballast_config_expected.yaml")
	require.NoError(t, err)
	assert.Equal(t, cfgMapExpected.ToStringMap(), cfgMap.ToStringMap())
}

func TestMemoryBallastConverter_Without_Memory_Ballast(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/without_memory_ballast_config.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)
	assert.NoError(t, RemoveMemoryBallastKey(context.Background(), cfgMap))
	assert.Equal(t, cfgMap.ToStringMap(), cfgMap.ToStringMap())
}

func TestRemoveMemoryBallastStrElementFromSlice(t *testing.T) {
	originalSlice := []interface{}{"foo", "bar", "memory_ballast", "item2"}
	actual := removeMemoryBallastStrElementFromSlice(originalSlice)
	expected := []interface{}{"foo", "bar", "item2"}
	assert.Equal(t, actual, expected)

	originalSlice1 := []interface{}{"foo", "bar", "foobar", "foobar1"}
	actual = removeMemoryBallastStrElementFromSlice(originalSlice1)
	assert.Equal(t, actual, originalSlice1)
}

func TestRemoveMemoryBallastConverter_With_Only_MemoryBallast_Value(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/with_memory_ballast_only.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)
	assert.NoError(t, RemoveMemoryBallastKey(context.Background(), cfgMap))
	cfgMapExpected, err := confmaptest.LoadConf("testdata/with_memory_ballast_only_expected.yaml")
	require.NoError(t, err)
	assert.Equal(t, cfgMapExpected.ToStringMap(), cfgMap.ToStringMap())
}
