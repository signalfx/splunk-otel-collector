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

	createClient := func(sql string) dbClient {
		var row map[string]string
		switch sql {
		case queryCPUTimeSQL:
			row = map[string]string{"VALUE": "1"}
		case queryElapsedTimeSQL:
			row = map[string]string{"VALUE": "1"}
		case queryExecutionsTimeSQL:
			row = map[string]string{"VALUE": "1"}
		case queryParseCallsSQL:
			row = map[string]string{"VALUE": "1"}
		case queryPhysicalReadBytesSQL:
			row = map[string]string{"VALUE": "1"}
		case queryPhysicalReadRequestsSQL:
			row = map[string]string{"VALUE": "1"}
		case queryPhysicalWriteBytesSQL:
			row = map[string]string{"VALUE": "1"}
		case queryPhysicalWriteRequestsSQL:
			row = map[string]string{"VALUE": "1"}
		case queryTotalSharableMemSQL:
			row = map[string]string{"VALUE": "1"}
		case queryLongestRunningSQL:
			row = map[string]string{"VALUE": "1"}
		case sessionUsageSQL:
			row = map[string]string{"CPU_USAGE": "45", "PGA_MEMORY": "3455", "PHYSICAL_READS": "12344", "LOGICAL_READS": "345", "HARD_PARSES": "346", "SOFT_PARSES": "7866"}
		case sessionEnqueueDeadlocksSQL:
			row = map[string]string{"VALUE": "1"}
		case sessionExchangeDeadlocksSQL:
			row = map[string]string{"VALUE": "1"}
		case sessionExecuteCountSQL:
			row = map[string]string{"VALUE": "1"}
		case sessionParseCountTotalSQL:
			row = map[string]string{"VALUE": "1"}
		case sessionUserCommitsSQL:
			row = map[string]string{"VALUE": "1"}
		case sessionUserRollbacksSQL:
			row = map[string]string{"VALUE": "1"}
		case sessionCountSQL:
			row = map[string]string{"VALUE": "1"}
		case systemResourceLimitsSQL:
			row = map[string]string{"RESOURCE_NAME": "processes", "CURRENT_UTILIZATION": "3", "MAX_UTILIZATION": "10", "INITIAL_ALLOCATION": "100", "LIMIT_VALUE": "100"}
		case tablespaceUsageSQL:
			row = map[string]string{"TABLESPACE_NAME": "SYS", "BYTES": "1024"}
		case tablespaceMaxSpaceSQL:
			row = map[string]string{"TABLESPACE_NAME": "SYS", "VALUE": "1024"}
		}

		return &fakeDbClient{Responses: [][]metricRow{
			{
				row,
			},
		}}
	}

	scrpr := scraper{
		logger:         zap.NewNop(),
		metricsBuilder: metricsBuilder,
		dbProviderFunc: func() (*sql.DB, error) {
			return nil, nil
		},
		clientProviderFunc: func(db *sql.DB, s string, logger *zap.Logger) dbClient {
			return createClient(s)
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
