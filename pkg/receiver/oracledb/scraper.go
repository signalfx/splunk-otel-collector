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
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/scrapererror"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/receiver/oracledbreceiver/internal/metadata"
)

const (
	sessionUsageSQL             = "select sum(cpu) as cpu_usage, sum(pga_memory) as pga_memory, sum(physical_reads) as physical_reads, sum(logical_reads) as logical_reads, sum(hard_parses) as hard_parses, sum(soft_parses) as soft_parses FROM v$sessmetric"
	sessionEnqueueDeadlocksSQL  = "select sum(se.value) as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and NAME = 'enqueue deadlocks'"
	sessionExchangeDeadlocksSQL = "select sum(se.value) as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and NAME = 'exchange deadlocks'"
	sessionExecuteCountSQL      = "select sum(se.value) as VALUE  from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and NAME = 'execute count'"
	sessionParseCountTotalSQL   = "select sum(se.value) as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and NAME = 'parse count (total)'"
	sessionUserCommitsSQL       = "select sum(se.value) as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and NAME = 'user commits'"
	sessionUserRollbacksSQL     = "select sum(se.value) as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and NAME = 'user rollbacks'"
	sessionCountSQL             = "select status, type, count(*) as VALUE FROM v$session GROUP BY status, type"
	systemResourceLimitsSQL     = "select RESOURCE_NAME, CURRENT_UTILIZATION, MAX_UTILIZATION, CASE WHEN TRIM(INITIAL_ALLOCATION) LIKE 'UNLIMITED' THEN '-1' ELSE TRIM(INITIAL_ALLOCATION) END as INITIAL_ALLOCATION, CASE WHEN TRIM(LIMIT_VALUE) LIKE 'UNLIMITED' THEN '-1' ELSE TRIM(LIMIT_VALUE) END as LIMIT_VALUE from v$resource_limit"
	tablespaceUsageSQL          = "select TABLESPACE_NAME, BYTES from DBA_DATA_FILES"
	tablespaceMaxSpaceSQL       = "select TABLESPACE_NAME, (BLOCK_SIZE*MAX_EXTENTS) AS VALUE FROM DBA_TABLESPACES"
)

type dbProviderFunc func() (*sql.DB, error)

type clientProviderFunc func(*sql.DB, string, *zap.Logger) dbClient

type scraper struct {
	sessionUserCommitsClient       dbClient
	sessionExecuteCountClient      dbClient
	sessionUsageRunner             dbClient
	sessionExchangeDeadlocksClient dbClient
	sessionEnqueueDeadlocksClient  dbClient
	tablespaceMaxSpaceClient       dbClient
	tablespaceUsageClient          dbClient
	systemResourceLimitsClient     dbClient
	sessionCountClient             dbClient
	sessionUserRollbacksClient     dbClient
	sessionParseCountTotalClient   dbClient
	db                             *sql.DB
	clientProviderFunc             clientProviderFunc
	metricsBuilder                 *metadata.MetricsBuilder
	dbProviderFunc                 dbProviderFunc
	logger                         *zap.Logger
	id                             config.ComponentID
	instanceName                   string
	scrapeCfg                      scraperhelper.ScraperControllerSettings
	startTime                      pcommon.Timestamp
	metricsSettings                metadata.MetricsSettings
}

func newScraper(id config.ComponentID, metricsBuilder *metadata.MetricsBuilder, metricsSettings metadata.MetricsSettings, scrapeCfg scraperhelper.ScraperControllerSettings, logger *zap.Logger, providerFunc dbProviderFunc, clientProviderFunc clientProviderFunc, instanceName string) *scraper {
	return &scraper{
		id:                 id,
		metricsBuilder:     metricsBuilder,
		metricsSettings:    metricsSettings,
		scrapeCfg:          scrapeCfg,
		logger:             logger,
		dbProviderFunc:     providerFunc,
		clientProviderFunc: clientProviderFunc,
		instanceName:       instanceName,
	}
}

var _ scraperhelper.Scraper = (*scraper)(nil)

func (s *scraper) ID() config.ComponentID {
	return s.id
}

func (s *scraper) Start(context.Context, component.Host) error {
	s.startTime = pcommon.NewTimestampFromTime(time.Now())
	var err error
	s.db, err = s.dbProviderFunc()
	if err != nil {
		return fmt.Errorf("failed to open db connection: %w", err)
	}
	s.sessionUsageRunner = s.clientProviderFunc(s.db, sessionUsageSQL, s.logger)
	s.sessionEnqueueDeadlocksClient = s.clientProviderFunc(s.db, sessionEnqueueDeadlocksSQL, s.logger)
	s.sessionExchangeDeadlocksClient = s.clientProviderFunc(s.db, sessionExchangeDeadlocksSQL, s.logger)
	s.sessionExecuteCountClient = s.clientProviderFunc(s.db, sessionExecuteCountSQL, s.logger)
	s.sessionParseCountTotalClient = s.clientProviderFunc(s.db, sessionParseCountTotalSQL, s.logger)
	s.sessionUserCommitsClient = s.clientProviderFunc(s.db, sessionUserCommitsSQL, s.logger)
	s.sessionUserRollbacksClient = s.clientProviderFunc(s.db, sessionUserRollbacksSQL, s.logger)
	s.sessionCountClient = s.clientProviderFunc(s.db, sessionCountSQL, s.logger)
	s.systemResourceLimitsClient = s.clientProviderFunc(s.db, systemResourceLimitsSQL, s.logger)
	s.tablespaceUsageClient = s.clientProviderFunc(s.db, tablespaceUsageSQL, s.logger)
	s.tablespaceMaxSpaceClient = s.clientProviderFunc(s.db, tablespaceMaxSpaceSQL, s.logger)
	return nil
}

func (s *scraper) Scrape(ctx context.Context) (pmetric.Metrics, error) {
	s.logger.Debug("Begin scrape")

	var scrapeErrors []error

	runSessionUsage := s.metricsSettings.OracledbSessionCPUUsage.Enabled || s.metricsSettings.OracledbSessionPgaMemory.Enabled ||
		s.metricsSettings.OracledbSessionPhysicalReads.Enabled || s.metricsSettings.OracledbSessionLogicalReads.Enabled || s.metricsSettings.OracledbSessionHardParses.Enabled || s.metricsSettings.OracledbSessionSoftParses.Enabled
	if runSessionUsage {
		rows, execError := s.sessionUsageRunner.metricRows(ctx)
		if execError != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", sessionUsageSQL, execError))
		}

		for _, row := range rows {
			// SELECT session_id, cpu as cpu_usage, pga_memory, physical_reads, logical_reads, hard_parses, soft_parses FROM v$sessmetric
			if s.metricsSettings.OracledbSessionCPUUsage.Enabled {
				value, err := strconv.ParseFloat(row["CPU_USAGE"], 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("value: %q, %s, %w", row["CPU_USAGE"], sessionUsageSQL, err))
				}
				s.metricsBuilder.RecordOracledbSessionCPUUsageDataPoint(pcommon.NewTimestampFromTime(time.Now()), value)
			}
			if s.metricsSettings.OracledbSessionPgaMemory.Enabled {
				value, err := strconv.ParseInt(row["PGA_MEMORY"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("pga_memory value: %q, %w", row["PGA_MEMORY"], err))
				}
				s.metricsBuilder.RecordOracledbSessionPgaMemoryDataPoint(pcommon.NewTimestampFromTime(time.Now()), value)
			}
			if s.metricsSettings.OracledbSessionPhysicalReads.Enabled {
				value, err := strconv.ParseInt(row["PHYSICAL_READS"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("physical_reads value: %q, %w", row["PHYSICAL_READS"], err))
				}
				s.metricsBuilder.RecordOracledbSessionPhysicalReadsDataPoint(pcommon.NewTimestampFromTime(time.Now()), value)
			}
			if s.metricsSettings.OracledbSessionLogicalReads.Enabled {
				value, err := strconv.ParseInt(row["LOGICAL_READS"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("logical_reads value: %q, %w", row["LOGICAL_READS"], err))
				}
				s.metricsBuilder.RecordOracledbSessionLogicalReadsDataPoint(pcommon.NewTimestampFromTime(time.Now()), value)
			}
			if s.metricsSettings.OracledbSessionHardParses.Enabled {
				value, err := strconv.ParseInt(row["HARD_PARSES"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("hard_parses value: %q, %w", row["HARD_PARSES"], err))
				}
				s.metricsBuilder.RecordOracledbSessionHardParsesDataPoint(pcommon.NewTimestampFromTime(time.Now()), value)
			}
			if s.metricsSettings.OracledbSessionSoftParses.Enabled {
				value, err := strconv.ParseInt(row["SOFT_PARSES"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("soft_parses value: %q, %w", row["SOFT_PARSES"], err))
				}
				s.metricsBuilder.RecordOracledbSessionSoftParsesDataPoint(pcommon.NewTimestampFromTime(time.Now()), value)
			}
		}
	}

	if s.metricsSettings.OracledbSessionEnqueueDeadlocks.Enabled {
		if err := s.executeOneQuery(ctx, s.sessionEnqueueDeadlocksClient, sessionEnqueueDeadlocksSQL, s.metricsBuilder.RecordOracledbSessionEnqueueDeadlocksDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}
	if s.metricsSettings.OracledbSessionExchangeDeadlocks.Enabled {
		if err := s.executeOneQuery(ctx, s.sessionExchangeDeadlocksClient, sessionExchangeDeadlocksSQL, s.metricsBuilder.RecordOracledbSessionExchangeDeadlocksDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}
	if s.metricsSettings.OracledbSessionExecuteCount.Enabled {
		if err := s.executeOneQuery(ctx, s.sessionExecuteCountClient, sessionExecuteCountSQL, s.metricsBuilder.RecordOracledbSessionExecuteCountDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	if s.metricsSettings.OracledbSessionParseCountTotal.Enabled {
		if err := s.executeOneQuery(ctx, s.sessionParseCountTotalClient, sessionParseCountTotalSQL, s.metricsBuilder.RecordOracledbSessionParseCountTotalDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	if s.metricsSettings.OracledbSessionUserCommits.Enabled {
		if err := s.executeOneQuery(ctx, s.sessionUserCommitsClient, sessionUserCommitsSQL, s.metricsBuilder.RecordOracledbSessionUserCommitsDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	if s.metricsSettings.OracledbSessionUserRollbacks.Enabled {
		if err := s.executeOneQuery(ctx, s.sessionUserRollbacksClient, sessionUserRollbacksSQL, s.metricsBuilder.RecordOracledbSessionUserRollbacksDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	if s.metricsSettings.OracledbSystemSessionCount.Enabled {
		rows, err := s.sessionCountClient.metricRows(ctx)
		if err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", sessionCountSQL, err))
		}
		for _, row := range rows {
			value, err := strconv.ParseInt(row["VALUE"], 10, 64)
			if err != nil {
				scrapeErrors = append(scrapeErrors, fmt.Errorf("value: %q: %q, %w", row["VALUE"], sessionCountSQL, err))
			}
			s.metricsBuilder.RecordOracledbSystemSessionCountDataPoint(pcommon.NewTimestampFromTime(time.Now()), value, row["TYPE"], row["STATUS"])
		}
	}

	if s.metricsSettings.OracledbSystemResourceLimits.Enabled {
		rows, err := s.systemResourceLimitsClient.metricRows(ctx)
		if err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", systemResourceLimitsSQL, err))
		}
		for _, row := range rows {
			resourceName := row["RESOURCE_NAME"]

			currentUtilization, err := strconv.ParseInt(row["CURRENT_UTILIZATION"], 10, 64)
			if err != nil {
				scrapeErrors = append(scrapeErrors, fmt.Errorf("current utilization for %q: %q, %s, %w", resourceName, row["CURRENT_UTILIZATION"], systemResourceLimitsSQL, err))
			} else {
				s.metricsBuilder.RecordOracledbSystemResourceLimitsDataPoint(pcommon.NewTimestampFromTime(time.Now()), currentUtilization, resourceName, "current_utilization")
			}
			maxUtilization, err := strconv.ParseInt(row["MAX_UTILIZATION"], 10, 64)
			if err != nil {
				scrapeErrors = append(scrapeErrors, fmt.Errorf("max utilization for %q: %q, %s, %w", resourceName, row["MAX_UTILIZATION"], systemResourceLimitsSQL, err))
			} else {
				s.metricsBuilder.RecordOracledbSystemResourceLimitsDataPoint(pcommon.NewTimestampFromTime(time.Now()), maxUtilization, resourceName, "max")
			}

			initialAllocation, err := strconv.ParseInt(strings.Trim(row["INITIAL_ALLOCATION"], " "), 10, 64)
			if err != nil {
				scrapeErrors = append(scrapeErrors, fmt.Errorf("initial allocation for %q: %q, %s, %w", resourceName, row["INITIAL_ALLOCATION"], systemResourceLimitsSQL, err))
			} else {
				s.metricsBuilder.RecordOracledbSystemResourceLimitsDataPoint(pcommon.NewTimestampFromTime(time.Now()), initialAllocation, resourceName, "initial")
			}

			limitValue, err := strconv.ParseInt(strings.Trim(row["LIMIT_VALUE"], " "), 10, 64)
			if err != nil {
				scrapeErrors = append(scrapeErrors, fmt.Errorf("limit value for %q: %q, %s, %w", resourceName, row["LIMIT_VALUE"], systemResourceLimitsSQL, err))
			} else {
				s.metricsBuilder.RecordOracledbSystemResourceLimitsDataPoint(pcommon.NewTimestampFromTime(time.Now()), limitValue, resourceName, "current")
			}
		}
	}
	if s.metricsSettings.OracledbTablespaceSize.Enabled {
		rows, err := s.tablespaceUsageClient.metricRows(ctx)
		if err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", tablespaceUsageSQL, err))
		} else {
			now := pcommon.NewTimestampFromTime(time.Now())
			for _, row := range rows {
				tablespaceName := row["TABLESPACE_NAME"]
				value, err := strconv.ParseInt(row["BYTES"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("bytes for %q: %q, %s, %w", tablespaceName, row["BYTES"], tablespaceUsageSQL, err))
				} else {
					s.metricsBuilder.RecordOracledbTablespaceSizeDataPoint(now, value, tablespaceName)
				}
			}
		}
	}
	if s.metricsSettings.OracledbTablespaceMaxSize.Enabled {
		rows, err := s.tablespaceMaxSpaceClient.metricRows(ctx)
		if err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", tablespaceMaxSpaceSQL, err))
		} else {
			now := pcommon.NewTimestampFromTime(time.Now())
			for _, row := range rows {
				tablespaceName := row["TABLESPACE_NAME"]
				var value int64
				ok := true
				if row["VALUE"] == "" {
					value = 0
				} else {
					value, err = strconv.ParseInt(row["VALUE"], 10, 64)
					if err != nil {
						ok = false
						scrapeErrors = append(scrapeErrors, fmt.Errorf("value for %q: %q, %s, %w", tablespaceName, row["VALUE"], tablespaceMaxSpaceSQL, err))
					}
				}
				if ok {
					s.metricsBuilder.RecordOracledbTablespaceMaxSizeDataPoint(now, value, tablespaceName)
				}
			}
		}
	}

	out := s.metricsBuilder.Emit(metadata.WithDbOracleInstanceName(s.instanceName))
	s.logger.Debug("Done scraping")
	if len(scrapeErrors) > 0 {
		return out, scrapererror.NewPartialScrapeError(multierr.Combine(scrapeErrors...), len(scrapeErrors))
	}
	return out, nil
}

func (s *scraper) executeOneQuery(ctx context.Context, client dbClient, query string, recorder func(ts pcommon.Timestamp, val int64)) error {
	rows, err := client.metricRows(ctx)
	if err != nil {
		return fmt.Errorf("error executing %s: %w", query, err)
	}
	for _, row := range rows {
		value, err := strconv.ParseInt(row["VALUE"], 10, 64)
		if err != nil {
			return fmt.Errorf("value: %s, %s, %w", row["VALUE"], query, err)
		}
		recorder(pcommon.NewTimestampFromTime(time.Now()), value)
	}
	return nil
}

func (s *scraper) Shutdown(ctx context.Context) error {
	return s.db.Close()
}
