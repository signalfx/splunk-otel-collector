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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/metadata"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

const typeStr = "databricks"

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(newReceiverFactory(newDatabricksClient), component.StabilityLevelAlpha),
	)
}

type dbClientFactory func(baseURL string, tok string, httpClient *http.Client, logger *zap.Logger) databricksClientIntf

func newReceiverFactory(dbClientFactory dbClientFactory) receiver.CreateMetricsFunc {
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
		dbsvc := newDatabricksService(dbClientFactory(dbcfg.Endpoint, dbcfg.Token, httpClient, settings.Logger), dbcfg.MaxResults)
		ssvc := newSparkService(
			settings.Logger,
			dbsvc,
			httpClient,
			dbcfg.SparkAPIEndpoint,
			dbcfg.SparkUIPort,
			dbcfg.OrgID,
			dbcfg.Token,
			newSparkUnmarshaler,
		)
		dbScraper := scraper{
			logger:      settings.Logger,
			rmp:         newRunMetricsProvider(dbsvc),
			dbmp:        dbMetricsProvider{dbsvc: dbsvc},
			builder:     metadata.NewMetricsBuilder(dbcfg.Metrics, settings.BuildInfo),
			resourceOpt: metadata.WithDatabricksInstanceName(dbcfg.InstanceName),
			scmb:        sparkCoreMetricsBuilder{ssvc: ssvc},
			semb:        sparkMetricsBuilder{ssvc: ssvc},
		}
		collectorScraper, err := scraperhelper.NewScraper(typeStr, dbScraper.scrape)
		if err != nil {
			return nil, fmt.Errorf("%s: createReceiverFunc closure: %w", typeStr, err)
		}
		return scraperhelper.NewScraperControllerReceiver(
			&dbcfg.ScraperControllerSettings,
			settings,
			consumer,
			scraperhelper.AddScraper(collectorScraper),
		)
	}
}

func newSparkUnmarshaler(logger *zap.Logger, httpClient *http.Client, sparkProxyURL string, orgID string, port int, token string, clusterID string) spark.Unmarshaler {
	return spark.NewUnmarshaler(httpClient, dbSparkProxyURL(sparkProxyURL, orgID, clusterID, port), token)
}

func dbSparkProxyURL(sparkProxyURL string, orgID string, clusterID string, port int) string {
	return fmt.Sprintf("%s/driver-proxy-api/o/%s/%s/%d", sparkProxyURL, orgID, clusterID, port)
}
