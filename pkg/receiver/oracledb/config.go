// Copyright Splunk, Inc.
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

package oracledbreceiver // import "github.com/signalfx/splunk-otel-collector/receiver/oracledbreceiver"

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/signalfx/splunk-otel-collector/receiver/oracledbreceiver/internal/metadata"
)

type Config struct {
	config.ReceiverSettings                 `mapstructure:",squash"`
	DataSource                              string `mapstructure:"datasource"`
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	MetricsSettings                         metadata.MetricsSettings `mapstructure:"metrics"`
}

func (c Config) Validate() error {
	if c.DataSource == "" {
		return errors.New("'datasource' cannot be empty")
	}
	if _, err := url.Parse(c.DataSource); err != nil {
		return fmt.Errorf("'datasource' is invalid: %w", err)
	}
	return nil
}

func CreateDefaultConfig() component.ReceiverConfig {
	return &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			ReceiverSettings:   config.NewReceiverSettings(component.NewID(typeStr)),
			CollectionInterval: 10 * time.Second,
		},
		MetricsSettings: metadata.DefaultMetricsSettings(),
	}
}
