// Copyright Splunk, Inc.
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

package splunksecretconfigsource

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

func TestSplunkSecretLoadConfig(t *testing.T) {
	fileName := path.Join("testdata", "config.yaml")
	v, err := confmaptest.LoadConf(fileName)
	require.NoError(t, err)

	factories := map[component.Type]configsource.Factory{
		component.MustNewType(typeStr): NewFactory(),
	}

	actualSettings, splitConf, err := configsource.SettingsFromConf(context.Background(), v, factories, nil)
	require.NoError(t, err)
	require.NotNil(t, splitConf)

	expectedSettings := map[string]configsource.Settings{
		"splunk_secret": &Config{
			SourceSettings: configsource.NewSourceSettings(component.MustNewID(typeStr)),
			Endpoint:       "https://localhost:8089",
			SessionKey:     "some_session_key",
			App:            defaultApp,
			User:           defaultUser,
			Timeout:        defaultTimeout,
		},
		"splunk_secret/insecure": &Config{
			SourceSettings:     configsource.NewSourceSettings(component.MustNewIDWithName(typeStr, "insecure")),
			Endpoint:           "https://localhost:8089",
			SessionKey:         "some_session_key",
			App:                defaultApp,
			User:               defaultUser,
			InsecureSkipVerify: true,
			Timeout:            5 * time.Second,
		},
	}
	require.Equal(t, expectedSettings, actualSettings)
	require.Empty(t, splitConf.ToStringMap())

	_, err = configsource.BuildConfigSources(context.Background(), actualSettings, zap.NewNop(), factories)
	require.NoError(t, err)
}
