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
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
)

const typeStr = "databricks"

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createReceiverFunc(newAPIClient), component.StabilityLevelAlpha),
	)
}

type Config struct {
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
	InstanceName                            string `mapstructure:"instance_name"`
	Token                                   string
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	MaxResults                              int `mapstructure:"max_results"`
}

func createDefaultConfig() component.Config {
	scs := scraperhelper.NewDefaultScraperControllerSettings(typeStr)
	// we set the default collection interval to 30 seconds which is half of the
	// lowest job frequency of 1 minute
	scs.CollectionInterval = time.Second * 30
	return &Config{
		MaxResults:                25, // 25 is the max the API supports
		ScraperControllerSettings: scs,
	}
}

func createReceiverFunc(createAPIClient func(baseURL string, tok string, httpClient *http.Client, logger *zap.Logger) apiClientInterface) func(
	_ context.Context,
	settings receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	return func(
		_ context.Context,
		settings receiver.CreateSettings,
		cfg component.Config,
		consumer consumer.Metrics,
	) (receiver.Metrics, error) {
		dbcfg := cfg.(*Config)
		httpClient, err := dbcfg.ToClient(nil, settings.TelemetrySettings)
		if err != nil {
			return nil, fmt.Errorf("%s: createReceiverFunc closure: %w", typeStr, err)
		}
		c := newDatabricksClient(createAPIClient(dbcfg.Endpoint, dbcfg.Token, httpClient, settings.Logger), dbcfg.MaxResults)
		s := scraper{
			instanceName: dbcfg.InstanceName,
			rmp:          newRunMetricsProvider(c),
			mp:           newMetricsProvider(c),
		}
		scrpr, err := scraperhelper.NewScraper(typeStr, s.scrape)
		if err != nil {
			return nil, fmt.Errorf("%s: createReceiverFunc closure: %w", typeStr, err)
		}
		return scraperhelper.NewScraperControllerReceiver(
			&dbcfg.ScraperControllerSettings,
			settings,
			consumer,
			scraperhelper.AddScraper(scrpr),
		)
	}
}
