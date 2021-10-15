// Copyright Splunk, Inc.
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

package includeconfigsource

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

func TestIncludeConfigSourceLoadConfig(t *testing.T) {
	fileName := path.Join("testdata", "config.yaml")
	v, err := config.NewMapFromFile(fileName)
	require.NoError(t, err)

	factories := map[config.Type]configprovider.Factory{
		typeStr: NewFactory(),
	}

	actualSettings, err := configprovider.Load(context.Background(), v, factories)
	require.NoError(t, err)

	expectedSettings := map[string]configprovider.ConfigSettings{
		"include": &Config{
			Settings: &configprovider.Settings{
				TypeVal: "include",
				NameVal: "include",
			},
		},
		"include/delete_files": &Config{
			Settings: &configprovider.Settings{
				TypeVal: "include",
				NameVal: "include/delete_files",
			},
			DeleteFiles: true,
		},
		"include/watch_files": &Config{
			Settings: &configprovider.Settings{
				TypeVal: "include",
				NameVal: "include/watch_files",
			},
			WatchFiles: true,
		},
	}

	require.Equal(t, expectedSettings, actualSettings)

	params := configprovider.CreateParams{
		Logger: zap.NewNop(),
	}
	cfgSrcs, err := configprovider.Build(context.Background(), actualSettings, params, factories)
	require.NoError(t, err)
	for k := range expectedSettings {
		assert.Contains(t, cfgSrcs, k)
	}
}
