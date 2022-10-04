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
	"go.opentelemetry.io/collector/receiver/scrapererror"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/receiver/oracledbreceiver/internal/metadata"
)

const (
	queryCPUTimeSQL               = "select * from (select s.sql_id as sql_id, s.SQL_FULLTEXT, s.CPU_TIME as VALUE FROM v$sqlstats s  where to_char(LAST_ACTIVE_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.EXECUTIONS DESC )"
	queryElapsedTimeSQL           = "select * from (select s.sql_id as sql_id, s.SQL_FULLTEXT, s.ELAPSED_TIME as VALUE FROM v$sqlstats s  where to_char(LAST_ACTIVE_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.ELAPSED_TIME DESC )"
	queryExecutionsTimeSQL        = "select * from (select s.sql_id as sql_id, s.SQL_FULLTEXT, s.EXECUTIONS as VALUE FROM v$sqlstats s  where to_char(LAST_ACTIVE_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.ELAPSED_TIME DESC )"
	queryParseCallsSQL            = "select * from (select s.sql_id as sql_id, s.SQL_FULLTEXT, s.PARSE_CALLS as VALUE FROM v$sqlstats s  where to_char(LAST_ACTIVE_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.EXECUTIONS DESC )"
	queryPhysicalReadBytesSQL     = "select * from (select s.sql_id as sql_id, s.SQL_FULLTEXT, s.PHYSICAL_READ_BYTES as VALUE FROM v$sqlstats s  where to_char(LAST_ACTIVE_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.EXECUTIONS DESC )"
	queryPhysicalReadRequestsSQL  = "select * from (select s.sql_id as sql_id, s.SQL_FULLTEXT, s.PHYSICAL_READ_REQUESTS as VALUE FROM v$sqlstats s  where to_char(LAST_ACTIVE_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.EXECUTIONS DESC )"
	queryPhysicalWriteBytesSQL    = "select * from (select s.sql_id as sql_id, s.SQL_FULLTEXT, s.PHYSICAL_WRITE_BYTES as VALUE FROM v$sqlstats s  where to_char(LAST_ACTIVE_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.ELAPSED_TIME DESC )"
	queryPhysicalWriteRequestsSQL = "select * from (select s.sql_id as sql_id, s.SQL_FULLTEXT, s.PHYSICAL_WRITE_REQUESTS as VALUE FROM v$sqlstats s  where to_char(LAST_ACTIVE_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.ELAPSED_TIME DESC )"
	queryTotalSharableMemSQL      = "select * from (select s.sql_id as sql_id, s.SQL_FULLTEXT, s.TOTAL_SHARABLE_MEM as VALUE FROM v$sqlstats s  where to_char(LAST_ACTIVE_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.ELAPSED_TIME DESC )"
	queryLongestRunningSQL        = "select s.SQL_ID, s.START_TIME, s.ELAPSED_SECONDS as VALUE FROM v$session_longops s where to_char(START_TIME , 'mm/dd/yyyy HH:MI') between to_char(sysdate - 30/(24*60), 'mm/dd/yyyy HH:MI') and to_char(sysdate, 'mm/dd/yyyy HH:MI') ORDER BY s.ELAPSED_SECONDS DESC"
	sessionUsageSQL               = "select session_id, cpu as cpu_usage, pga_memory, physical_reads, logical_reads, hard_parses, soft_parses FROM v$sessmetric"
	sessionEnqueueDeadlocksSQL    = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'enqueue deadlocks'"
	sessionExchangeDeadlocksSQL   = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'exchange deadlocks'"
	sessionExecuteCountSQL        = "select ss.SID as session_id, se.value as VALUE  from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'execute count'"
	sessionParseCountTotalSQL     = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'parse count (total)'"
	sessionUserCommitsSQL         = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'user commits'"
	sessionUserRollbacksSQL       = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'user rollbacks'"
	sessionCountSQL               = "select status, type, count(*) as VALUE FROM v$session GROUP BY status, type"
	systemResourceLimitsSQL       = "select RESOURCE_NAME, CURRENT_UTILIZATION, MAX_UTILIZATION, INITIAL_ALLOCATION, LIMIT_VALUE from v$resource_limit"
	tablespaceUsageSQL            = "select TABLESPACE_NAME, BYTES from DBA_DATA_FILES"
	tablespaceMaxSpaceSQL         = "select TABLESPACE_NAME, (BLOCK_SIZE*MAX_EXTENTS) AS VALUE FROM DBA_TABLESPACES"
)

type dbProviderFunc func() (*sql.DB, error)

type clientProviderFunc func(*sql.DB, string, *zap.Logger) dbClient

type scraper struct {
	logger             *zap.Logger
	metricsBuilder     *metadata.MetricsBuilder
	dbProviderFunc     dbProviderFunc
	clientProviderFunc clientProviderFunc
	db                 *sql.DB
	id                 config.ComponentID
	instanceName       string
	scrapeCfg          scraperhelper.ScraperControllerSettings
	startTime          pcommon.Timestamp
	metricsSettings    metadata.MetricsSettings
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
	return nil
}

func (s *scraper) Scrape(ctx context.Context) (pmetric.Metrics, error) {
	s.logger.Debug("Begin scrape")

	var scrapeErrors []error

	if s.metricsSettings.OracledbQueryCPUTime.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, queryCPUTimeSQL, s.metricsBuilder.RecordOracledbQueryCPUTimeDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", queryCPUTimeSQL, err))
		}
	}
	if s.metricsSettings.OracledbQueryLongRunning.Enabled {
		client := s.clientProviderFunc(s.db, queryLongestRunningSQL, s.logger)
		rows, err := client.metricRows(ctx)
		if err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryLongestRunningSQL, err)
		}
		for _, row := range rows {
			value, err := strconv.ParseInt(row["VALUE"], 10, 64)
			if err != nil {
				scrapeErrors = append(scrapeErrors, fmt.Errorf("value: %q, %s, %w", row["VALUE"], queryLongestRunningSQL, err))
			} else {
				s.metricsBuilder.RecordOracledbQueryLongRunningDataPoint(pcommon.NewTimestampFromTime(time.Now()), value, row["SQL_ID"])
			}
		}
	}
	if s.metricsSettings.OracledbQueryElapsedTime.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, queryElapsedTimeSQL, s.metricsBuilder.RecordOracledbQueryElapsedTimeDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", queryElapsedTimeSQL, err))
		}
	}
	if s.metricsSettings.OracledbQueryExecutions.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, queryExecutionsTimeSQL, s.metricsBuilder.RecordOracledbQueryExecutionsDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", queryExecutionsTimeSQL, err))
		}
	}
	if s.metricsSettings.OracledbQueryParseCalls.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, queryParseCallsSQL, s.metricsBuilder.RecordOracledbQueryParseCallsDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", queryParseCallsSQL, err))
		}
	}
	if s.metricsSettings.OracledbQueryPhysicalReadBytes.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, queryPhysicalReadBytesSQL, s.metricsBuilder.RecordOracledbQueryPhysicalReadBytesDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", queryPhysicalReadBytesSQL, err))
		}
	}
	if s.metricsSettings.OracledbQueryPhysicalReadRequests.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, queryPhysicalReadRequestsSQL, s.metricsBuilder.RecordOracledbQueryPhysicalReadRequestsDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", queryPhysicalReadRequestsSQL, err))
		}
	}
	if s.metricsSettings.OracledbQueryPhysicalWriteBytes.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, queryPhysicalWriteBytesSQL, s.metricsBuilder.RecordOracledbQueryPhysicalWriteBytesDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", queryPhysicalWriteBytesSQL, err))
		}
	}
	if s.metricsSettings.OracledbQueryPhysicalWriteRequests.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, queryPhysicalWriteRequestsSQL, s.metricsBuilder.RecordOracledbQueryPhysicalWriteRequestsDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", queryPhysicalWriteRequestsSQL, err))
		}
	}
	if s.metricsSettings.OracledbQueryTotalSharableMem.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, queryTotalSharableMemSQL, s.metricsBuilder.RecordOracledbQueryTotalSharableMemDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, fmt.Errorf("error executing %s: %w", queryTotalSharableMemSQL, err))
		}
	}
	runSessionUsage := s.metricsSettings.OracledbSessionCPUUsage.Enabled || s.metricsSettings.OracledbSessionPgaMemory.Enabled ||
		s.metricsSettings.OracledbSessionPhysicalReads.Enabled || s.metricsSettings.OracledbSessionLogicalReads.Enabled || s.metricsSettings.OracledbSessionHardParses.Enabled || s.metricsSettings.OracledbSessionSoftParses.Enabled
	if runSessionUsage {
		client := s.clientProviderFunc(s.db, sessionUsageSQL, s.logger)
		rows, execError := client.metricRows(ctx)
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
				s.metricsBuilder.RecordOracledbSessionCPUUsageDataPoint(pcommon.NewTimestampFromTime(time.Now()), value, row["SESSION_ID"])
			}
			if s.metricsSettings.OracledbSessionPgaMemory.Enabled {
				value, err := strconv.ParseInt(row["PGA_MEMORY"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("pga_memory value: %q, %w", row["PGA_MEMORY"], err))
				}
				s.metricsBuilder.RecordOracledbSessionPgaMemoryDataPoint(pcommon.NewTimestampFromTime(time.Now()), value, row["SESSION_ID"])
			}
			if s.metricsSettings.OracledbSessionPhysicalReads.Enabled {
				value, err := strconv.ParseInt(row["PHYSICAL_READS"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("physical_reads value: %q, %w", row["PHYSICAL_READS"], err))
				}
				s.metricsBuilder.RecordOracledbSessionPhysicalReadsDataPoint(pcommon.NewTimestampFromTime(time.Now()), value, row["SESSION_ID"])
			}
			if s.metricsSettings.OracledbSessionLogicalReads.Enabled {
				value, err := strconv.ParseInt(row["LOGICAL_READS"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("logical_reads value: %q, %w", row["LOGICAL_READS"], err))
				}
				s.metricsBuilder.RecordOracledbSessionLogicalReadsDataPoint(pcommon.NewTimestampFromTime(time.Now()), value, row["SESSION_ID"])
			}
			if s.metricsSettings.OracledbSessionHardParses.Enabled {
				value, err := strconv.ParseInt(row["HARD_PARSES"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("hard_parses value: %q, %w", row["HARD_PARSES"], err))
				}
				s.metricsBuilder.RecordOracledbSessionHardParsesDataPoint(pcommon.NewTimestampFromTime(time.Now()), value, row["SESSION_ID"])
			}
			if s.metricsSettings.OracledbSessionSoftParses.Enabled {
				value, err := strconv.ParseInt(row["SOFT_PARSES"], 10, 64)
				if err != nil {
					scrapeErrors = append(scrapeErrors, fmt.Errorf("soft_parses value: %q, %w", row["SOFT_PARSES"], err))
				}
				s.metricsBuilder.RecordOracledbSessionSoftParsesDataPoint(pcommon.NewTimestampFromTime(time.Now()), value, row["SESSION_ID"])
			}
		}
	}

	if s.metricsSettings.OracledbSessionEnqueueDeadlocks.Enabled {
		if err := s.executeOneQueryWithSessionID(ctx, sessionEnqueueDeadlocksSQL, s.metricsBuilder.RecordOracledbSessionEnqueueDeadlocksDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}
	if s.metricsSettings.OracledbSessionExchangeDeadlocks.Enabled {
		if err := s.executeOneQueryWithSessionID(ctx, sessionExchangeDeadlocksSQL, s.metricsBuilder.RecordOracledbSessionExchangeDeadlocksDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}
	if s.metricsSettings.OracledbSessionExecuteCount.Enabled {
		if err := s.executeOneQueryWithSessionID(ctx, sessionExecuteCountSQL, s.metricsBuilder.RecordOracledbSessionExecuteCountDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	if s.metricsSettings.OracledbSessionParseCountTotal.Enabled {
		if err := s.executeOneQueryWithSessionID(ctx, sessionParseCountTotalSQL, s.metricsBuilder.RecordOracledbSessionParseCountTotalDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	if s.metricsSettings.OracledbSessionUserCommits.Enabled {
		if err := s.executeOneQueryWithSessionID(ctx, sessionUserCommitsSQL, s.metricsBuilder.RecordOracledbSessionUserCommitsDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	if s.metricsSettings.OracledbSessionUserRollbacks.Enabled {
		if err := s.executeOneQueryWithSessionID(ctx, sessionUserRollbacksSQL, s.metricsBuilder.RecordOracledbSessionUserRollbacksDataPoint); err != nil {
			scrapeErrors = append(scrapeErrors, err)
		}
	}

	if s.metricsSettings.OracledbSystemSessionCount.Enabled {
		client := s.clientProviderFunc(s.db, sessionCountSQL, s.logger)
		rows, err := client.metricRows(ctx)
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
		client := s.clientProviderFunc(s.db, systemResourceLimitsSQL, s.logger)
		rows, err := client.metricRows(ctx)
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
			ok := true
			var initialAllocation int64
			if strings.Contains(row["INITIAL_ALLOCATION"], "UNLIMITED") {
				initialAllocation = 0x7FF0000000000000 // max int64
			} else {
				initialAllocation, err = strconv.ParseInt(strings.Trim(row["INITIAL_ALLOCATION"], " "), 10, 64)
				if err != nil {
					ok = false
					scrapeErrors = append(scrapeErrors, fmt.Errorf("initial allocation for %q: %q, %s, %w", resourceName, row["INITIAL_ALLOCATION"], systemResourceLimitsSQL, err))
				}
			}
			if ok {
				s.metricsBuilder.RecordOracledbSystemResourceLimitsDataPoint(pcommon.NewTimestampFromTime(time.Now()), initialAllocation, resourceName, "initial")
			}

			var limitValue int64
			ok = true
			if strings.Contains(row["LIMIT_VALUE"], "UNLIMITED") {
				limitValue = 0x7FF0000000000000 // max int64
			} else {
				limitValue, err = strconv.ParseInt(strings.Trim(row["LIMIT_VALUE"], " "), 10, 64)
				if err != nil {
					ok = false
					scrapeErrors = append(scrapeErrors, fmt.Errorf("limit value for %q: %q, %s, %w", resourceName, row["LIMIT_VALUE"], systemResourceLimitsSQL, err))
				}
			}
			if ok {
				s.metricsBuilder.RecordOracledbSystemResourceLimitsDataPoint(pcommon.NewTimestampFromTime(time.Now()), limitValue, resourceName, "current")
			}
		}
	}
	if s.metricsSettings.OracledbTablespaceSize.Enabled {
		client := s.clientProviderFunc(s.db, tablespaceUsageSQL, s.logger)
		rows, err := client.metricRows(ctx)
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
		client := s.clientProviderFunc(s.db, tablespaceMaxSpaceSQL, s.logger)
		rows, err := client.metricRows(ctx)
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

	out := s.metricsBuilder.Emit(metadata.WithOracledbInstanceName(s.instanceName))
	s.logger.Debug("Done scraping")
	if len(scrapeErrors) > 0 {
		return out, scrapererror.NewPartialScrapeError(multierr.Combine(scrapeErrors...), len(scrapeErrors))
	}
	return out, nil
}

func (s *scraper) executeOneQueryWithSessionID(ctx context.Context, query string, recorder func(ts pcommon.Timestamp, val int64, sessionID string)) error {
	client := s.clientProviderFunc(s.db, query, s.logger)
	rows, err := client.metricRows(ctx)
	if err != nil {
		return fmt.Errorf("error executing %s: %w", query, err)
	}
	for _, row := range rows {
		value, err := strconv.ParseInt(row["VALUE"], 10, 64)
		if err != nil {
			return fmt.Errorf("value: %s, %s, %w", row["VALUE"], query, err)
		}
		recorder(pcommon.NewTimestampFromTime(time.Now()), value, row["SESSION_ID"])
	}
	return nil
}

func (s *scraper) executeOneQueryWithFullText(ctx context.Context, query string, recorder func(ts pcommon.Timestamp, val int64, oracledbQueryIDAttributeValue string, oracledbQueryFulltextAttributeValue string)) error {
	client := s.clientProviderFunc(s.db, query, s.logger)
	rows, err := client.metricRows(ctx)
	if err != nil {
		return fmt.Errorf("error executing %s: %w", query, err)
	}
	for _, row := range rows {
		value, err := strconv.ParseInt(row["VALUE"], 10, 64)
		if err != nil {
			return fmt.Errorf("value: %s, %s, %w", row["VALUE"], query, err)
		}
		fullText := row["SQL_FULLTEXT"]
		if len(fullText) > 256 {
			fullText = fullText[:253] + "..."
		}
		recorder(pcommon.NewTimestampFromTime(time.Now()), value, row["SQL_ID"], fullText)
	}
	return nil
}

func (s *scraper) Shutdown(ctx context.Context) error {
	return s.db.Close()
}
