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
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const typeStr = "databricks"

func NewFactory() component.ReceiverFactory {
	return receiverhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		receiverhelper.WithMetrics(createReceiverFunc(newAPIClient)),
	)
}

type Config struct {
	InstanceName                            string `mapstructure:"instance_name"`
	BaseURL                                 string `mapstructure:"base_url"`
	Token                                   string
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			ReceiverSettings:   config.NewReceiverSettings(config.NewComponentID(typeStr)),
			CollectionInterval: time.Second * 30,
		},
	}
}

func createReceiverFunc(createAPIClient func(baseURL string, tok string) databricksAPI) func(
	_ context.Context,
	settings component.ReceiverCreateSettings,
	cfg config.Receiver,
	consumer consumer.Metrics,
) (component.MetricsReceiver, error) {
	return func(
		_ context.Context,
		settings component.ReceiverCreateSettings,
		cfg config.Receiver,
		consumer consumer.Metrics,
	) (component.MetricsReceiver, error) {
		dbcfg := cfg.(*Config)
		p := newPaginator(createAPIClient(dbcfg.BaseURL, dbcfg.Token))
		s := scraper{
			instanceName: dbcfg.InstanceName,
			rmp:          newRunMetricsProvider(p),
			jmp:          newMetricsProvider(p),
		}
		scrpr, err := scraperhelper.NewScraper(typeStr, s.scrape)
		if err != nil {
			return nil, fmt.Errorf("%s: createMetricsReceiver(): %w", typeStr, err)
		}
		return scraperhelper.NewScraperControllerReceiver(
			&dbcfg.ScraperControllerSettings,
			settings,
			consumer,
			scraperhelper.AddScraper(scrpr),
		)
	}
}
