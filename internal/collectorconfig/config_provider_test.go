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

package collectorconfig

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmapprovider"
)

func TestParseConfigSource_ConfigFile(t *testing.T) {
	_ = os.Setenv("CFG_SRC", "my_cfg_src_val")
	_ = os.Setenv("LEGACY", "my_legacy_val")
	pp := NewConfigMapProvider(component.BuildInfo{}, false, "testdata/config.yaml", "", nil)
	assertProviderOK(t, pp)
}

func TestParseConfigSource_InMemory(t *testing.T) {
	cfgYaml, err := os.ReadFile("testdata/config.yaml")
	require.NoError(t, err)
	pp := NewConfigMapProvider(component.BuildInfo{}, false, "", string(cfgYaml), nil)
	assertProviderOK(t, pp)
}

func assertProviderOK(t *testing.T, provider configmapprovider.Provider) {
	ctx := context.Background()
	retrieved, err := provider.Retrieve(ctx, nil)
	require.NoError(t, err)
	cfgMap, err := retrieved.Get(ctx)
	require.NoError(t, err)
	v := cfgMap.Get("config_source_env_key")
	assert.Equal(t, "my_cfg_src_val", v)
	v = cfgMap.Get("legacy_env_key")
	assert.Equal(t, "my_legacy_val", v)
}
