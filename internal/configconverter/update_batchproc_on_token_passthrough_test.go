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

func TestUpdateBatchProcOnIncludeMetadata(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/include_metadata_on_sapm_token_passthrough/agent_config_w_include_metadata.yaml")
	require.NotNil(t, cfgMap)
	require.NoError(t, err)

	expectedCfgMap, err := confmaptest.LoadConf("testdata/include_metadata_on_sapm_token_passthrough/expected_agent_config_w_include_metadata.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = UpdateBatchProcOnTokenPassthrough(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.Equal(t, expectedCfgMap, cfgMap)
}

func TestNoIncludeMetadata(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/include_metadata_on_sapm_token_passthrough/agent_config_wo_include_metadata.yaml")
	require.NotNil(t, cfgMap)
	require.NoError(t, err)

	expectedCfgMap, err := confmaptest.LoadConf("testdata/include_metadata_on_sapm_token_passthrough/agent_config_wo_include_metadata.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = UpdateBatchProcOnTokenPassthrough(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.Equal(t, expectedCfgMap, cfgMap)
}

func TestIncludeMetadataDisabled(t *testing.T) {
	cfgMap, err := confmaptest.LoadConf("testdata/include_metadata_on_sapm_token_passthrough/agent_config_w_include_metadata_disabled.yaml")
	require.NotNil(t, cfgMap)
	require.NoError(t, err)

	expectedCfgMap, err := confmaptest.LoadConf("testdata/include_metadata_on_sapm_token_passthrough/agent_config_w_include_metadata_disabled.yaml")
	require.NoError(t, err)
	require.NotNil(t, cfgMap)

	err = UpdateBatchProcOnTokenPassthrough(context.Background(), cfgMap)
	require.NoError(t, err)

	assert.Equal(t, expectedCfgMap, cfgMap)
}
