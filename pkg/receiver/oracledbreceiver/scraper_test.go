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
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/receiver/oracledbreceiver/internal/metadata"
)

func TestScraper_ErrorOnStart(t *testing.T) {
	scrpr := scraper{
		dbProviderFunc: func() (*sql.DB, error) {
			return nil, errors.New("oops")
		},
	}
	err := scrpr.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
}

func TestScraper_Scrape(t *testing.T) {
	metricsBuilder := metadata.NewMetricsBuilder(metadata.DefaultMetricsSettings(), component.NewDefaultBuildInfo())
	fakeClient := &fakeDbClient{
		Responses: [][]metricRow{
			{
				{"value": "1", "sql_fulltext": "foo"}, // cpu time
			},
			{
				{"value": "2", "sql_fulltext": "foo"}, // elapsed time
			},
			{
				{"value": "3", "sql_fulltext": "foo"}, // executions time
			},
			{
				{"value": "4", "sql_fulltext": "foo"}, // parse calls
			},
			{
				{"value": "5", "sql_fulltext": "foo"}, // read bytes
			},
			{
				{"value": "6", "sql_fulltext": "foo"}, // read requests
			},
			{
				{"value": "7", "sql_fulltext": "foo"}, // write bytes
			},
			{
				{"value": "8", "sql_fulltext": "foo"}, // write requests
			},
			{
				{"value": "9", "sql_fulltext": "foo"}, // total sharable mem
			},
			{
				{"cpu_usage": "45", "pga_memory": "3455", "physical_reads": "12344", "logical_reads": "345", "hard_parses": "346", "soft_parses": "7866"}, // session
			},
			{
				{"value": "10"}, // session enqueue deadlocks
			},
			{
				{"value": "11"}, // session exchange deadlocks
			},
			{
				{"value": "12"}, // session executions count
			},
			{
				{"value": "13"}, // session parse count
			},
			{
				{"value": "14"}, // session user commits
			},
			{
				{"value": "15"}, // session user rollbacks
			},
			{
				{"value": "16"}, // active sessions
			},
			{
				{"value": "17"}, // cached sessions
			},
		},
	}

	scrpr := scraper{
		logger:         zap.NewNop(),
		metricsBuilder: metricsBuilder,
		dbProviderFunc: func() (*sql.DB, error) {
			return nil, nil
		},
		clientProviderFunc: func(db *sql.DB, s string, logger *zap.Logger) dbClient {
			return fakeClient
		},
		id:              config.ComponentID{},
		metricsSettings: metadata.DefaultMetricsSettings(),
	}
	err := scrpr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	m, err := scrpr.Scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 23, m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	captured := m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	assert.Equal(t, int64(1), captured.At(0).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, "oracledb.query.cpu_time", captured.At(0).Name())
	assert.Equal(t, int64(2), captured.At(1).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(3), captured.At(2).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(4), captured.At(3).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(5), captured.At(4).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(6), captured.At(5).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(7), captured.At(6).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(8), captured.At(7).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(9), captured.At(8).Gauge().DataPoints().At(0).IntVal())
	assert.Equal(t, "oracledb.query.total_sharable_mem", captured.At(8).Name())
	assert.Equal(t, float64(45), captured.At(9).Gauge().DataPoints().At(0).DoubleVal())
	assert.Equal(t, "oracledb.session.cpu_usage", captured.At(9).Name())
	assert.Equal(t, "oracledb.session.enqueue_deadlocks", captured.At(10).Name())
	assert.Equal(t, int64(10), captured.At(10).Gauge().DataPoints().At(0).IntVal())
	assert.Equal(t, "oracledb.session.exchange_deadlocks", captured.At(11).Name())
	assert.Equal(t, int64(11), captured.At(11).Gauge().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(12), captured.At(12).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(346), captured.At(13).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(345), captured.At(14).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(13), captured.At(15).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(3455), captured.At(16).Gauge().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(12344), captured.At(17).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(7866), captured.At(18).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(14), captured.At(19).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(15), captured.At(20).Sum().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(16), captured.At(21).Gauge().DataPoints().At(0).IntVal())
	assert.Equal(t, int64(17), captured.At(22).Gauge().DataPoints().At(0).IntVal())

}
