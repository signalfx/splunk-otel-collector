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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
)

const typeStr = "databricks"

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(newReceiverFactory(), component.StabilityLevelDevelopment),
	)
}

func newReceiverFactory() receiver.CreateMetricsFunc {
	return func(
		_ context.Context,
		settings receiver.CreateSettings,
		cfg component.Config,
		consumer consumer.Metrics,
	) (receiver.Metrics, error) {
		dbcfg := cfg.(*Config)
		httpClient, err := dbcfg.ToClient(nil, settings.TelemetrySettings)
		if err != nil {
			return nil, fmt.Errorf("newReceiverFactory: failed to create client from config: %w", err)
		}
		dbClient := newDatabricksRawClient(dbcfg.Token, dbcfg.Endpoint, httpClient, settings.Logger)
		dbsvc := newDatabricksService(dbClient, dbcfg.MaxResults)
		ssvc := newSparkService(settings.Logger, dbsvc, httpClient, dbcfg.Token, dbcfg.SparkAPIURL, dbcfg.SparkOrgID, dbcfg.SparkUIPort)
		dbScraper := scraper{
			logger:      settings.Logger,
			rmp:         newRunMetricsProvider(dbsvc),
			dbmp:        dbMetricsProvider{dbsvc: dbsvc},
			builder:     metadata.NewMetricsBuilder(dbcfg.Metrics, settings.BuildInfo),
			resourceOpt: metadata.WithDatabricksInstanceName(dbcfg.InstanceName),
			scmb:        sparkClusterMetricsBuilder{ssvc: ssvc},
			semb: sparkExtraMetricsBuilder{
				ssvc:   ssvc,
				logger: settings.Logger,
			},
			dbsvc: dbsvc,
		}
		collectorScraper, err := scraperhelper.NewScraper(typeStr, dbScraper.scrape)
		if err != nil {
			return nil, fmt.Errorf("newReceiverFactory: failed to create scraper: %w", err)
		}
		return scraperhelper.NewScraperControllerReceiver(
			&dbcfg.ScraperControllerSettings,
			settings,
			consumer,
			scraperhelper.AddScraper(collectorScraper),
		)
	}
}
