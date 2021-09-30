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

package vaultconfigsource

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configparser"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

func TestVaultLoadConfig(t *testing.T) {
	fileName := path.Join("testdata", "config.yaml")
	v, err := configparser.NewConfigMapFromFile(fileName)
	require.NoError(t, err)

	factories := map[config.Type]configprovider.Factory{
		typeStr: NewFactory(),
	}

	actualSettings, err := configprovider.Load(context.Background(), v, factories)
	require.NoError(t, err)

	devToken := "dev_token"
	otherToken := "other_token"
	expectedSettings := map[string]configprovider.ConfigSettings{
		"vault": &Config{
			Settings: &configprovider.Settings{
				TypeVal: "vault",
				NameVal: "vault",
			},
			Endpoint:     "http://localhost:8200",
			Path:         "secret/kv",
			PollInterval: 1 * time.Minute,
			Authentication: &Authentication{
				Token: &devToken,
			},
		},
		"vault/poll_interval": &Config{
			Settings: &configprovider.Settings{
				TypeVal: "vault",
				NameVal: "vault/poll_interval",
			},
			Endpoint:     "https://localhost:8200",
			Path:         "other/path/kv",
			PollInterval: 10 * time.Second,
			Authentication: &Authentication{
				Token: &otherToken,
			},
		},
	}

	require.Equal(t, expectedSettings, actualSettings)

	params := configprovider.CreateParams{
		Logger: zap.NewNop(),
	}
	_, err = configprovider.Build(context.Background(), actualSettings, params, factories)
	require.NoError(t, err)
}
