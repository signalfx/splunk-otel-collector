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

package nutanixreceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

func TestValidConfig(t *testing.T) {
	configs, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	cm, err := configs.Sub(typeStr)
	require.NoError(t, err)

	cfg := createDefaultConfig().(*Config)
	require.NoError(t, cm.Unmarshal(&cfg))
	require.NoError(t, cfg.Validate())

	require.Equal(t, &Config{
		ControllerConfig: scraperhelper.ControllerConfig{
			CollectionInterval: 10 * time.Second,
			InitialDelay:       time.Second,
			Timeout:            5 * time.Second,
		},
		Endpoint: "https://prism.example.com:9440",
		Port:     defaultPort,
		Username: "readonly",
		Password: configopaque.String("secret"),
		TLS: TLSConfig{
			InsecureSkipVerify: true,
		},
		Metrics: MetricsConfig{
			Clusters:          MetricCategoryConfig{Enabled: true},
			Hosts:             MetricCategoryConfig{Enabled: true},
			StorageContainers: MetricCategoryConfig{Enabled: true},
			VMs:               MetricCategoryConfig{Enabled: true},
			VolumeGroups:      MetricCategoryConfig{Enabled: true},
		},
	}, cfg)
}

func TestInvalidConfig(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	require.ErrorContains(t, cfg.Validate(), "endpoint")

	cfg.Endpoint = "prism.example.com"
	require.ErrorContains(t, cfg.Validate(), "username")

	cfg.Username = "readonly"
	require.ErrorContains(t, cfg.Validate(), "password")

	cfg.Password = "secret"
	cfg.Port = 0
	require.ErrorContains(t, cfg.Validate(), "port")

	cfg.Port = defaultPort
	cfg.ControllerConfig.Timeout = -time.Second
	require.ErrorContains(t, cfg.Validate(), "timeout")

	cfg.ControllerConfig.Timeout = time.Second
	cfg.Metrics = MetricsConfig{}
	require.ErrorContains(t, cfg.Validate(), "at least one")
}
