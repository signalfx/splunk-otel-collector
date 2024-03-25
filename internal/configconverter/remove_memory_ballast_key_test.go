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
	pmp := RemoveMemoryBallastKey{}
	conf := confmap.NewFromStringMap(map[string]interface{}{"foo": "bar"})
	assert.NoError(t, pmp.Convert(context.Background(), conf))
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, conf.ToStringMap())
}

func TestRemoveMemoryBallastConverter(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/memory_ballast.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)
	pmp := RemoveMemoryBallastKey{}
	assert.NoError(t, pmp.Convert(context.Background(), cfgMap))
	assert.Equal(t, map[string]interface{}{"service": map[string]interface{}{"extensions": []interface{}{"health_check", "http_forwarder", "zpages", "smartagent"}}}, cfgMap.ToStringMap())
}
