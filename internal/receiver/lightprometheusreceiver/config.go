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
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

func createDefaultConfig() component.Config {
	scs := scraperhelper.NewDefaultScraperControllerSettings(typeStr)
	// set the default collection interval to 30 seconds which is half of the
	// lowest job frequency of 1 minute
	scs.CollectionInterval = time.Second * 30
	return &Config{
		ScraperControllerSettings: scs,
		HTTPClientSettings:        confighttp.NewDefaultHTTPClientSettings(),
		ResourceAttributes: ResourceAttributesConfig{
			ServiceInstanceID: ResourceAttributeConfig{Enabled: true},
			ServiceName:       ResourceAttributeConfig{Enabled: true},
			NetHostName:       ResourceAttributeConfig{Enabled: false},
			NetHostPort:       ResourceAttributeConfig{Enabled: false},
			HTTPScheme:        ResourceAttributeConfig{Enabled: false},
		},
	}
}

// ResourceAttributeConfig provides configuration for a resource attribute.
type ResourceAttributeConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// ResourceAttributesConfig allows users to configure the resource attributes.
type ResourceAttributesConfig struct {
	ServiceName       ResourceAttributeConfig `mapstructure:"service.name"`
	ServiceInstanceID ResourceAttributeConfig `mapstructure:"service.instance.id"`
	NetHostName       ResourceAttributeConfig `mapstructure:"net.host.name"`
	NetHostPort       ResourceAttributeConfig `mapstructure:"net.host.port"`
	HTTPScheme        ResourceAttributeConfig `mapstructure:"http.scheme"`
}

type Config struct {
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	// ResourceAttributes that added to scraped metrics.
	ResourceAttributes ResourceAttributesConfig `mapstructure:"resource_attributes"`
}

func (cfg *Config) Validate() error {
	if cfg.HTTPClientSettings.Endpoint == "" {
		return errors.New(`"endpoint" is required`)
	}
	return nil
}
