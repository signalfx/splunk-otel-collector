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

package zookeeperconfigsource

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	expcfg "go.opentelemetry.io/collector/config/experimental/config"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

func TestZookeeperLoadConfig(t *testing.T) {
	fileName := path.Join("testdata", "config.yaml")
	v, err := config.NewMapFromFile(fileName)
	require.NoError(t, err)

	factories := map[config.Type]configprovider.Factory{
		typeStr: NewFactory(),
	}

	actualSettings, err := configprovider.Load(context.Background(), v, factories)
	require.NoError(t, err)

	expectedSettings := map[string]expcfg.Source{
		"zookeeper": &Config{
			SourceSettings: expcfg.NewSourceSettings(config.NewComponentID(typeStr)),
			Endpoints:      []string{"http://localhost:1234"},
			Timeout:        time.Second * 10,
		},
		"zookeeper/timeout": &Config{
			SourceSettings: expcfg.NewSourceSettings(config.NewComponentIDWithName(typeStr, "timeout")),
			Endpoints:      []string{"https://localhost:3010"},
			Timeout:        time.Second * 8,
		},
	}

	require.Equal(t, expectedSettings, actualSettings)

	params := configprovider.CreateParams{
		Logger: zap.NewNop(),
	}
	_, err = configprovider.Build(context.Background(), actualSettings, params, factories)
	require.NoError(t, err)
}
