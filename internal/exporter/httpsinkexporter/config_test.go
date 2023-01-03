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

package httpsinkexporter

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {

	configs, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, configs)

	e0cm, err := configs.Sub("httpsink")
	require.NoError(t, err)
	e0 := createDefaultConfig()
	require.NoError(t, component.UnmarshalConfig(e0cm, e0))

	assert.Equal(t, NewFactory().CreateDefaultConfig(), e0)

	e1cm, err := configs.Sub("httpsink/2")
	require.NoError(t, err)
	e1 := createDefaultConfig()
	require.NoError(t, component.UnmarshalConfig(e1cm, e1))

	assert.Equal(t,
		&Config{
			Endpoint: "localhost:3333",
		}, e1)
}
