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
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

const (
	typeStr     = "nutanix"
	defaultPort = 9440
)

type TLSConfig struct {
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

type MetricCategoryConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type MetricsConfig struct {
	Clusters          MetricCategoryConfig `mapstructure:"clusters"`
	Hosts             MetricCategoryConfig `mapstructure:"hosts"`
	StorageContainers MetricCategoryConfig `mapstructure:"storage_containers"`
	VMs               MetricCategoryConfig `mapstructure:"vms"`
	VolumeGroups      MetricCategoryConfig `mapstructure:"volume_groups"`
}

type Config struct {
	Endpoint string              `mapstructure:"endpoint"`
	Username string              `mapstructure:"username"`
	Password configopaque.String `mapstructure:"password"`

	scraperhelper.ControllerConfig `mapstructure:",squash"`
	Metrics                        MetricsConfig `mapstructure:"metrics"`
	TLS                            TLSConfig     `mapstructure:"tls"`
	Port                           int           `mapstructure:"port"`
}

func createDefaultConfig() component.Config {
	scs := scraperhelper.NewDefaultControllerConfig()
	scs.CollectionInterval = 30 * time.Second
	scs.Timeout = 30 * time.Second

	return &Config{
		ControllerConfig: scs,
		Port:             defaultPort,
		Metrics: MetricsConfig{
			Clusters:          MetricCategoryConfig{Enabled: true},
			Hosts:             MetricCategoryConfig{Enabled: true},
			StorageContainers: MetricCategoryConfig{Enabled: true},
			VMs:               MetricCategoryConfig{Enabled: true},
			VolumeGroups:      MetricCategoryConfig{Enabled: true},
		},
	}
}

func (cfg *Config) Validate() error {
	if err := cfg.ControllerConfig.Validate(); err != nil {
		return err
	}
	if cfg.Endpoint == "" {
		return errors.New(`"endpoint" is required`)
	}
	if cfg.Username == "" {
		return errors.New(`"username" is required`)
	}
	if cfg.Password == "" {
		return errors.New(`"password" is required`)
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return errors.New(`"port" must be between 1 and 65535`)
	}
	if !cfg.Metrics.Clusters.Enabled &&
		!cfg.Metrics.Hosts.Enabled &&
		!cfg.Metrics.StorageContainers.Enabled &&
		!cfg.Metrics.VMs.Enabled &&
		!cfg.Metrics.VolumeGroups.Enabled {
		return errors.New("at least one metric category must be enabled")
	}
	return nil
}
