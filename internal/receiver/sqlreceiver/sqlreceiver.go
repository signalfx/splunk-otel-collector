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

package sqlreceiver

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
)

const typeStr = "sql"

func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		createDefaultConfig,
		component.WithMetricsReceiver(mkCreateReceiver(sql.Open, newDbClient)),
	)
}

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	Driver                                  string  `mapstructure:"driver"`
	DataSource                              string  `mapstructure:"datasource"`
	Queries                                 []Query `mapstructure:"queries"`
}

type Query struct {
	SQL     string   `mapstructure:"sql"`
	Metrics []Metric `mapstructure:"metrics"`
}

type Metric struct {
	MetricName       string   `mapstructure:"metric_name"`
	ValueColumn      string   `mapstructure:"value_column"`
	AttributeColumns []string `mapstructure:"attribute_columns"`
	IsMonotonic      bool     `mapstructure:"is_monotonic"`
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: time.Second,
		},
	}
}

type dbConnectionProvider func(driverName, dataSourceName string) (*sql.DB, error)

type dbClientFactory func(*sql.DB, string, *zap.Logger) dbClient

func mkCreateReceiver(dbConnect dbConnectionProvider, mkClient dbClientFactory) component.CreateMetricsReceiverFunc {
	return func(
		ctx context.Context,
		settings component.ReceiverCreateSettings,
		cfg config.Receiver,
		consumer consumer.Metrics,
	) (component.MetricsReceiver, error) {
		sqlCfg := cfg.(*Config)
		db, err := dbConnect(sqlCfg.Driver, sqlCfg.DataSource)
		if err != nil {
			return nil, fmt.Errorf("failed to open db connection: %w", err)
		}
		var opts []scraperhelper.ScraperControllerOption
		for _, query := range sqlCfg.Queries {
			mp := metricsProvider{
				client:     mkClient(db, query.SQL, settings.TelemetrySettings.Logger),
				metricsCfg: query.Metrics,
			}
			scraper, err := scraperhelper.NewScraper(typeStr, mp.scrape)
			if err != nil {
				return nil, fmt.Errorf("%s: createReceiver: %w", typeStr, err)
			}
			opt := scraperhelper.AddScraper(scraper)
			opts = append(opts, opt)
		}
		return scraperhelper.NewScraperControllerReceiver(
			&sqlCfg.ScraperControllerSettings,
			settings,
			consumer,
			opts...,
		)
	}
}
