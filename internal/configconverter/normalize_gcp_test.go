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

package configconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestNormalizeGcp(t *testing.T) {
	expectedCfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config_post_migration.yaml")
	require.NotNil(t, expectedCfgMap)
	require.NoError(t, err)

	cfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = NormalizeGcp{}.Convert(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.Equal(t, expectedCfgMap, cfgMap)
}

func TestNormalizeGcpMany(t *testing.T) {
	expectedCfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config_post_migration.yaml")
	require.NotNil(t, expectedCfgMap)
	require.NoError(t, err)

	cfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config_many.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = NormalizeGcp{}.Convert(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.Equal(t, expectedCfgMap, cfgMap)
}

func TestNormalizeGcpSame(t *testing.T) {
	expectedCfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config_post_migration.yaml")
	require.NotNil(t, expectedCfgMap)
	require.NoError(t, err)

	cfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config_post_migration.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = NormalizeGcp{}.Convert(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.Equal(t, expectedCfgMap, cfgMap)
}

func TestNormalizeGcpNoop(t *testing.T) {
	expectedCfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config_no_op.yaml")
	require.NotNil(t, expectedCfgMap)
	require.NoError(t, err)

	cfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config_no_op.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = NormalizeGcp{}.Convert(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.Equal(t, expectedCfgMap, cfgMap)
}

func TestNormalizeGcpSubresources(t *testing.T) {
	expectedCfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config_subresources_post_migration.yaml")
	require.NotNil(t, expectedCfgMap)
	require.NoError(t, err)

	cfgMap, err := confmaptest.LoadConf("testdata/normalize_gcp/upstream_agent_config_subresources.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = NormalizeGcp{}.Convert(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.Equal(t, expectedCfgMap, cfgMap)
}
