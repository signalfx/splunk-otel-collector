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

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/databricks"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
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
		dbrcfg := cfg.(*Config)
		httpClient, err := dbrcfg.ToClient(nil, settings.TelemetrySettings)
		if err != nil {
			return nil, fmt.Errorf("newReceiverFactory failed to create client from config: %w", err)
		}
		dbrsvc := databricks.NewService(databricks.NewRawClient(dbrcfg.Token, dbrcfg.Endpoint, httpClient, settings.Logger), dbrcfg.MaxResults)
		ssvc := spark.NewService(settings.Logger, httpClient, dbrcfg.Token, dbrcfg.SparkEndpoint, dbrcfg.SparkOrgID, dbrcfg.SparkUIPort)
		dbrScraper := scraper{
			dbrInstanceName: dbrcfg.InstanceName,
			logger:          settings.Logger,
			rmp:             databricks.NewRunMetricsProvider(dbrsvc),
			dbrmp:           databricks.MetricsProvider{Svc: dbrsvc},
			metricsBuilder:  metadata.NewMetricsBuilder(dbrcfg.MetricsBuilderConfig, settings),
			scmb:            spark.ClusterMetricsBuilder{Ssvc: ssvc},
			semb: spark.ExtraMetricsBuilder{
				Ssvc:   ssvc,
				Logger: settings.Logger,
			},
			dbrsvc: dbrsvc,
		}
		collectorScraper, err := scraperhelper.NewScraper(typeStr, dbrScraper.scrape)
		if err != nil {
			return nil, fmt.Errorf("newReceiverFactory failed to create scraper: %w", err)
		}
		return scraperhelper.NewScraperControllerReceiver(
			&dbrcfg.ScraperControllerSettings,
			settings,
			consumer,
			scraperhelper.AddScraper(collectorScraper),
		)
	}
}
