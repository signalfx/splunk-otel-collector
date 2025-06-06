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

package lightprometheusreceiver

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

func TestValidConfig(t *testing.T) {
	configs, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))
	require.NoError(t, err)
	require.NotNil(t, configs)

	cm, err := configs.Sub("lightprometheus")
	require.NoError(t, err)

	cfg := createDefaultConfig().(*Config)
	err = cm.Unmarshal(&cfg)
	require.NoError(t, err)
	require.NoError(t, cfg.Validate())

	expectedCfg := &Config{
		ControllerConfig: scraperhelper.ControllerConfig{
			CollectionInterval: 10 * time.Second,
			InitialDelay:       time.Second,
		},
		ClientConfig: confighttp.NewDefaultClientConfig(),
		ResourceAttributes: ResourceAttributesConfig{
			ServiceInstanceID: ResourceAttributeConfig{Enabled: false},
			ServiceName:       ResourceAttributeConfig{Enabled: false},
			ServerAddress:     ResourceAttributeConfig{Enabled: true},
			ServerPort:        ResourceAttributeConfig{Enabled: false},
			URLScheme:         ResourceAttributeConfig{Enabled: false},
		},
	}
	expectedCfg.ClientConfig.Endpoint = "http://localhost:9090/metrics"
	require.Equal(t, expectedCfg, cfg)
}

func TestInvalidConfig(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	require.ErrorContains(t, cfg.Validate(), "endpoint")
}
