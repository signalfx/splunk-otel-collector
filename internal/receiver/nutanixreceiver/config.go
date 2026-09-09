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
	typeStr           = "nutanix"
	defaultPort       = 9440
	defaultAPIVersion = "v4"
)

type TLSConfig struct {
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

type MetricCategoryConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type MetricsConfig struct {
	Clusters          MetricCategoryConfig `mapstructure:"clusters"`
	DataProtection    MetricCategoryConfig `mapstructure:"data_protection"`
	Disks             MetricCategoryConfig `mapstructure:"disks"`
	Files             MetricCategoryConfig `mapstructure:"files"`
	Hosts             MetricCategoryConfig `mapstructure:"hosts"`
	Microsegmentation MetricCategoryConfig `mapstructure:"microsegmentation"`
	Networking        MetricCategoryConfig `mapstructure:"networking"`
	Objects           MetricCategoryConfig `mapstructure:"objects"`
	PrismCentral      MetricCategoryConfig `mapstructure:"prism_central"`
	StorageContainers MetricCategoryConfig `mapstructure:"storage_containers"`
	VMs               MetricCategoryConfig `mapstructure:"vms"`
	VolumeGroups      MetricCategoryConfig `mapstructure:"volume_groups"`
}

type Config struct {
	Endpoint   string              `mapstructure:"endpoint"`
	APIVersion string              `mapstructure:"api_version"`
	Username   string              `mapstructure:"username"`
	Password   configopaque.String `mapstructure:"password"`

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
		APIVersion:       defaultAPIVersion,
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
	if cfg.APIVersion != "v4" && cfg.APIVersion != "v2.0" {
		return errors.New(`"api_version" must be either "v4" for Prism Central or "v2.0" for Prism Element`)
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
		!cfg.Metrics.DataProtection.Enabled &&
		!cfg.Metrics.Disks.Enabled &&
		!cfg.Metrics.Files.Enabled &&
		!cfg.Metrics.Hosts.Enabled &&
		!cfg.Metrics.Microsegmentation.Enabled &&
		!cfg.Metrics.Networking.Enabled &&
		!cfg.Metrics.Objects.Enabled &&
		!cfg.Metrics.PrismCentral.Enabled &&
		!cfg.Metrics.StorageContainers.Enabled &&
		!cfg.Metrics.VMs.Enabled &&
		!cfg.Metrics.VolumeGroups.Enabled {
		return errors.New("at least one metric category must be enabled")
	}
	if cfg.APIVersion == "v2.0" && (cfg.Metrics.DataProtection.Enabled ||
		cfg.Metrics.Disks.Enabled ||
		cfg.Metrics.Files.Enabled ||
		cfg.Metrics.Microsegmentation.Enabled ||
		cfg.Metrics.Networking.Enabled ||
		cfg.Metrics.Objects.Enabled ||
		cfg.Metrics.PrismCentral.Enabled) {
		return errors.New("disks, networking, prism_central, data_protection, microsegmentation, files, and objects metrics require api_version v4")
	}
	return nil
}
