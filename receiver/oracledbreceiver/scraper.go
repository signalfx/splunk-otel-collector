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
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
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
	sessionUsageSQL               = "select session_id, cpu as cpu_usage, pga_memory, physical_reads, logical_reads, hard_parses, soft_parses FROM v$sessmetric"
	sessionEnqueueDeadlocksSQL    = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'enqueue deadlocks'"
	sessionExchangeDeadlocksSQL   = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'exchange deadlocks'"
	sessionExecuteCountSQL        = "select ss.SID as session_id, se.value as VALUE  from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'execute count'"
	sessionParseCountTotalSQL     = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'parse count (total)'"
	sessionUserCommitsSQL         = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'user commits'"
	sessionUserRollbacksSQL       = "select ss.SID as session_id, se.value as VALUE from v$session ss, v$sesstat se, v$statname sn where se.STATISTIC# = sn.STATISTIC# and se.SID = ss.SID and NAME = 'user rollbacks'"
	activeSessionTotalSQL         = "select status, type, count(*) as VALUE FROM v$session GROUP BY status, type"
	cachedSessionTotalSQL         = "select type, count(*) as VALUE FROM v$session WHERE status = 'CACHED' GROUP BY type"
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
	scrapeCfg          scraperhelper.ScraperControllerSettings
	startTime          pcommon.Timestamp
	metricsSettings    metadata.MetricsSettings
}

func newScraper(id config.ComponentID, metricsBuilder *metadata.MetricsBuilder, metricsSettings metadata.MetricsSettings, scrapeCfg scraperhelper.ScraperControllerSettings, logger *zap.Logger, providerFunc dbProviderFunc, clientProviderFunc clientProviderFunc) *scraper {
	return &scraper{
		id:                 id,
		metricsBuilder:     metricsBuilder,
		metricsSettings:    metricsSettings,
		scrapeCfg:          scrapeCfg,
		logger:             logger,
		dbProviderFunc:     providerFunc,
		clientProviderFunc: clientProviderFunc,
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
	now := pcommon.NewTimestampFromTime(time.Now())
	if s.metricsSettings.OracledbQueryCPUTime.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, now, queryCPUTimeSQL, s.metricsBuilder.RecordOracledbQueryCPUTimeDataPoint); err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryCPUTimeSQL, err)
		}
	}
	if s.metricsSettings.OracledbQueryElapsedTime.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, now, queryElapsedTimeSQL, s.metricsBuilder.RecordOracledbQueryElapsedTimeDataPoint); err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryElapsedTimeSQL, err)
		}
	}
	if s.metricsSettings.OracledbQueryExecutions.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, now, queryExecutionsTimeSQL, s.metricsBuilder.RecordOracledbQueryExecutionsDataPoint); err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryExecutionsTimeSQL, err)
		}
	}
	if s.metricsSettings.OracledbQueryParseCalls.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, now, queryParseCallsSQL, s.metricsBuilder.RecordOracledbQueryParseCallsDataPoint); err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryParseCallsSQL, err)
		}
	}
	if s.metricsSettings.OracledbQueryPhysicalReadBytes.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, now, queryPhysicalReadBytesSQL, s.metricsBuilder.RecordOracledbQueryPhysicalReadBytesDataPoint); err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryPhysicalReadBytesSQL, err)
		}
	}
	if s.metricsSettings.OracledbQueryPhysicalReadRequests.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, now, queryPhysicalReadRequestsSQL, s.metricsBuilder.RecordOracledbQueryPhysicalReadRequestsDataPoint); err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryPhysicalReadRequestsSQL, err)
		}
	}
	if s.metricsSettings.OracledbQueryPhysicalWriteBytes.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, now, queryPhysicalWriteBytesSQL, s.metricsBuilder.RecordOracledbQueryPhysicalWriteBytesDataPoint); err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryPhysicalWriteBytesSQL, err)
		}
	}
	if s.metricsSettings.OracledbQueryPhysicalWriteRequests.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, now, queryPhysicalWriteRequestsSQL, s.metricsBuilder.RecordOracledbQueryPhysicalWriteRequestsDataPoint); err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryPhysicalWriteRequestsSQL, err)
		}
	}
	if s.metricsSettings.OracledbQueryTotalSharableMem.Enabled {
		if err := s.executeOneQueryWithFullText(ctx, now, queryTotalSharableMemSQL, s.metricsBuilder.RecordOracledbQueryTotalSharableMemDataPoint); err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", queryTotalSharableMemSQL, err)
		}
	}
	runSessionUsage := s.metricsSettings.OracledbSessionCPUUsage.Enabled || s.metricsSettings.OracledbSessionPgaMemory.Enabled ||
		s.metricsSettings.OracledbSessionPhysicalReads.Enabled || s.metricsSettings.OracledbSessionLogicalReads.Enabled || s.metricsSettings.OracledbSessionHardParses.Enabled || s.metricsSettings.OracledbSessionSoftParses.Enabled
	if runSessionUsage {
		client := s.clientProviderFunc(s.db, sessionUsageSQL, s.logger)
		rows, execError := client.metricRows(ctx)
		if execError != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", sessionUsageSQL, execError)
		}

		for _, row := range rows {
			// SELECT session_id, cpu as cpu_usage, pga_memory, physical_reads, logical_reads, hard_parses, soft_parses FROM v$sessmetric
			if s.metricsSettings.OracledbSessionCPUUsage.Enabled {
				value, err := strconv.ParseFloat(row["CPU_USAGE"], 64)
				if err != nil {
					return pmetric.Metrics{}, err
				}
				s.metricsBuilder.RecordOracledbSessionCPUUsageDataPoint(now, value)
			}
			if s.metricsSettings.OracledbSessionPgaMemory.Enabled {
				value, err := strconv.ParseInt(row["PGA_MEMORY"], 10, 64)
				if err != nil {
					return pmetric.Metrics{}, fmt.Errorf("pga_memory value: %s, %w", row["PGA_MEMORY"], err)
				}
				s.metricsBuilder.RecordOracledbSessionPgaMemoryDataPoint(now, value)
			}
			if s.metricsSettings.OracledbSessionPhysicalReads.Enabled {
				value, err := strconv.ParseInt(row["PHYSICAL_READS"], 10, 64)
				if err != nil {
					return pmetric.Metrics{}, fmt.Errorf("physical_reads value: %s, %w", row["PHYSICAL_READS"], err)
				}
				s.metricsBuilder.RecordOracledbSessionPhysicalReadsDataPoint(now, value)
			}
			if s.metricsSettings.OracledbSessionLogicalReads.Enabled {
				value, err := strconv.ParseInt(row["LOGICAL_READS"], 10, 64)
				if err != nil {
					return pmetric.Metrics{}, fmt.Errorf("logical_reads value: %s, %w", row["LOGICAL_READS"], err)
				}
				s.metricsBuilder.RecordOracledbSessionLogicalReadsDataPoint(now, value)
			}
			if s.metricsSettings.OracledbSessionHardParses.Enabled {
				value, err := strconv.ParseInt(row["HARD_PARSES"], 10, 64)
				if err != nil {
					return pmetric.Metrics{}, fmt.Errorf("hard_parses value: %s, %w", row["HARD_PARSES"], err)
				}
				s.metricsBuilder.RecordOracledbSessionHardParsesDataPoint(now, value)
			}
			if s.metricsSettings.OracledbSessionSoftParses.Enabled {
				value, err := strconv.ParseInt(row["SOFT_PARSES"], 10, 64)
				if err != nil {
					return pmetric.Metrics{}, fmt.Errorf("soft_parses value: %s, %w", row["SOFT_PARSES"], err)
				}
				s.metricsBuilder.RecordOracledbSessionSoftParsesDataPoint(now, value)
			}
		}
	}

	if s.metricsSettings.OracledbSessionEnqueueDeadlocks.Enabled {
		if err := s.executeOneQuery(ctx, now, sessionEnqueueDeadlocksSQL, s.metricsBuilder.RecordOracledbSessionEnqueueDeadlocksDataPoint); err != nil {
			return pmetric.Metrics{}, err
		}
	}
	if s.metricsSettings.OracledbSessionExchangeDeadlocks.Enabled {
		if err := s.executeOneQuery(ctx, now, sessionExchangeDeadlocksSQL, s.metricsBuilder.RecordOracledbSessionExchangeDeadlocksDataPoint); err != nil {
			return pmetric.Metrics{}, err
		}
	}
	if s.metricsSettings.OracledbSessionExecuteCount.Enabled {
		if err := s.executeOneQuery(ctx, now, sessionExecuteCountSQL, s.metricsBuilder.RecordOracledbSessionExecuteCountDataPoint); err != nil {
			return pmetric.Metrics{}, err
		}
	}

	if s.metricsSettings.OracledbSessionParseCountTotal.Enabled {
		if err := s.executeOneQuery(ctx, now, sessionParseCountTotalSQL, s.metricsBuilder.RecordOracledbSessionParseCountTotalDataPoint); err != nil {
			return pmetric.Metrics{}, err
		}
	}

	if s.metricsSettings.OracledbSessionUserCommits.Enabled {
		if err := s.executeOneQuery(ctx, now, sessionUserCommitsSQL, s.metricsBuilder.RecordOracledbSessionUserCommitsDataPoint); err != nil {
			return pmetric.Metrics{}, err
		}
	}

	if s.metricsSettings.OracledbSessionUserRollbacks.Enabled {
		if err := s.executeOneQuery(ctx, now, sessionUserRollbacksSQL, s.metricsBuilder.RecordOracledbSessionUserRollbacksDataPoint); err != nil {
			return pmetric.Metrics{}, err
		}
	}

	if s.metricsSettings.OracledbSystemActiveSessionTotal.Enabled {
		client := s.clientProviderFunc(s.db, activeSessionTotalSQL, s.logger)
		rows, err := client.metricRows(ctx)
		if err != nil {
			return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", activeSessionTotalSQL, err)
		}
		for _, row := range rows {
			value, err := strconv.ParseInt(row["VALUE"], 10, 64)
			if err != nil {
				return pmetric.Metrics{}, err
			}
			s.metricsBuilder.RecordOracledbSystemActiveSessionTotalDataPoint(now, value, fmt.Sprintf("%s-%s", row["STATUS"], row["TYPE"]))
		}
	}
	if s.metricsSettings.OracledbSystemCachedSessionTotal.Enabled {
		client := s.clientProviderFunc(s.db, cachedSessionTotalSQL, s.logger)
		rows, err := client.metricRows(ctx)
		if err != nil {
			return pmetric.Metrics{}, err
		}
		for _, row := range rows {
			value, err := strconv.ParseInt(row["VALUE"], 10, 64)
			if err != nil {
				return pmetric.Metrics{}, fmt.Errorf("error executing %s: %w", cachedSessionTotalSQL, err)
			}
			s.metricsBuilder.RecordOracledbSystemCachedSessionTotalDataPoint(now, value, row["TYPE"])
		}
	}

	out := s.metricsBuilder.Emit()
	s.logger.Debug("Done scraping")
	return out, nil
}

func (s *scraper) executeOneQuery(ctx context.Context, now pcommon.Timestamp, query string, recorder func(ts pcommon.Timestamp, val int64)) error {
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
		recorder(now, value)
	}
	return nil
}

func (s *scraper) executeOneQueryWithFullText(ctx context.Context, now pcommon.Timestamp, query string, recorder func(ts pcommon.Timestamp, val int64, oracledbQueryFulltextAttributeValue string)) error {
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
		recorder(now, value, row["SQL_FULLTEXT"])
	}
	return nil
}

func (s *scraper) Shutdown(ctx context.Context) error {
	return s.db.Close()
}
