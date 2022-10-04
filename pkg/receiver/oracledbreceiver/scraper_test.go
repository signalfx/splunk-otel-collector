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
	"go.opentelemetry.io/collector/receiver/scrapererror"
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

var queryResponses = map[string]map[string]string{
	queryCPUTimeSQL:               {"VALUE": "1"},
	queryElapsedTimeSQL:           {"VALUE": "1"},
	queryExecutionsTimeSQL:        {"VALUE": "1"},
	queryParseCallsSQL:            {"VALUE": "1"},
	queryPhysicalReadBytesSQL:     {"VALUE": "1"},
	queryPhysicalReadRequestsSQL:  {"VALUE": "1"},
	queryPhysicalWriteBytesSQL:    {"VALUE": "1"},
	queryPhysicalWriteRequestsSQL: {"VALUE": "1"},
	queryTotalSharableMemSQL:      {"VALUE": "1"},
	queryLongestRunningSQL:        {"VALUE": "1"},
	sessionUsageSQL:               {"CPU_USAGE": "45", "PGA_MEMORY": "3455", "PHYSICAL_READS": "12344", "LOGICAL_READS": "345", "HARD_PARSES": "346", "SOFT_PARSES": "7866"},
	sessionEnqueueDeadlocksSQL:    {"VALUE": "1"},
	sessionExchangeDeadlocksSQL:   {"VALUE": "1"},
	sessionExecuteCountSQL:        {"VALUE": "1"},
	sessionParseCountTotalSQL:     {"VALUE": "1"},
	sessionUserCommitsSQL:         {"VALUE": "1"},
	sessionUserRollbacksSQL:       {"VALUE": "1"},
	sessionCountSQL:               {"VALUE": "1"},
	systemResourceLimitsSQL:       {"RESOURCE_NAME": "processes", "CURRENT_UTILIZATION": "3", "MAX_UTILIZATION": "10", "INITIAL_ALLOCATION": "100", "LIMIT_VALUE": "100"},
	tablespaceUsageSQL:            {"TABLESPACE_NAME": "SYS", "BYTES": "1024"},
	tablespaceMaxSpaceSQL:         {"TABLESPACE_NAME": "SYS", "VALUE": "1024"},
}

func TestScraper_Scrape(t *testing.T) {
	metricsBuilder := metadata.NewMetricsBuilder(metadata.DefaultMetricsSettings(), component.NewDefaultBuildInfo())

	scrpr := scraper{
		logger:         zap.NewNop(),
		metricsBuilder: metricsBuilder,
		dbProviderFunc: func() (*sql.DB, error) {
			return nil, nil
		},
		clientProviderFunc: func(db *sql.DB, s string, logger *zap.Logger) dbClient {
			return &fakeDbClient{Responses: [][]metricRow{
				{
					queryResponses[s],
				},
			}}
		},
		id:              config.ComponentID{},
		metricsSettings: metadata.DefaultMetricsSettings(),
	}
	err := scrpr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	m, err := scrpr.Scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 26, m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
}

func TestPartial_InvalidScrape(t *testing.T) {
	metricsBuilder := metadata.NewMetricsBuilder(metadata.DefaultMetricsSettings(), component.NewDefaultBuildInfo())

	scrpr := scraper{
		logger:         zap.NewNop(),
		metricsBuilder: metricsBuilder,
		dbProviderFunc: func() (*sql.DB, error) {
			return nil, nil
		},
		clientProviderFunc: func(db *sql.DB, s string, logger *zap.Logger) dbClient {
			if s == tablespaceUsageSQL {
				return &fakeDbClient{Responses: [][]metricRow{
					{
						{},
					},
				}}
			}
			return &fakeDbClient{Responses: [][]metricRow{
				{
					queryResponses[s],
				},
			}}
		},
		id:              config.ComponentID{},
		metricsSettings: metadata.DefaultMetricsSettings(),
	}
	err := scrpr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	_, err = scrpr.Scrape(context.Background())
	require.Error(t, err)
	require.True(t, scrapererror.IsPartialScrapeError(err))
	require.EqualError(t, err, `bytes for "": "", select TABLESPACE_NAME, BYTES from DBA_DATA_FILES, strconv.ParseInt: parsing "": invalid syntax`)
}
