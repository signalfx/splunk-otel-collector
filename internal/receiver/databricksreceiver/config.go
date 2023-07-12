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

package databricksreceiver

import (
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

const maxMaxResults = 25

func createDefaultConfig() component.Config {
	scs := scraperhelper.NewDefaultScraperControllerSettings(typeStr)
	// set the default collection interval to 30 seconds which is half of the
	// lowest job frequency of 1 minute
	scs.CollectionInterval = time.Second * 30
	return &Config{
		MaxResults:                maxMaxResults, // 25 is the max the API supports
		ScraperControllerSettings: scs,
		MetricsBuilderConfig:      metadata.DefaultMetricsBuilderConfig(),
	}
}

type Config struct {
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
	InstanceName                            string `mapstructure:"instance_name"`
	Token                                   string `mapstructure:"token"`
	SparkOrgID                              string `mapstructure:"spark_org_id"`
	SparkEndpoint                           string `mapstructure:"spark_endpoint"`
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	MaxResults                              int                           `mapstructure:"max_results"`
	SparkUIPort                             int                           `mapstructure:"spark_ui_port"`
	MetricsBuilderConfig                    metadata.MetricsBuilderConfig `mapstructure:"squash,"`
}

func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("`endpoint` is empty")
	}
	if c.InstanceName == "" {
		return errors.New("`instance_name` is empty")
	}
	if c.Token == "" {
		return errors.New("`token` is empty")
	}
	if c.MaxResults < 0 || c.MaxResults > maxMaxResults {
		return fmt.Errorf("max_results must be between 0 and %d, inclusive", maxMaxResults)
	}
	return nil
}
