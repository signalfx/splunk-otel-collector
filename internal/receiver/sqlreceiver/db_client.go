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
	"strconv"
	"time"

	// register db drivers
	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/snowflakedb/gosnowflake"
	"go.uber.org/zap"
)

type dbClient interface {
	metricRows(ctx context.Context) ([]metricRow, error)
}

type dbSQLClient struct {
	db     *sql.DB
	logger *zap.Logger
	sql    string
}

func newDbClient(db *sql.DB, sql string, logger *zap.Logger) dbClient {
	return dbSQLClient{
		db:     db,
		sql:    sql,
		logger: logger,
	}
}

type metricRow map[string]string

func (cl dbSQLClient) metricRows(ctx context.Context) ([]metricRow, error) {
	sqlRows, err := cl.db.QueryContext(ctx, cl.sql)
	if err != nil {
		return nil, err
	}
	var out []metricRow
	rr := reusableRow{
		attrs: map[string]func() string{},
	}
	types, err := sqlRows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	for _, columnType := range types {
		colName := columnType.Name()
		switch columnType.ScanType().Name() {
		case "int8", "int16", "int32", "int64":
			var v int64
			rr.attrs[colName] = func() string {
				return strconv.FormatInt(v, 10)
			}
			rr.scanDest = append(rr.scanDest, &v)
		case "string":
			var v string
			rr.attrs[colName] = func() string {
				return v
			}
			rr.scanDest = append(rr.scanDest, &v)
		case "Time":
			var v time.Time
			rr.attrs[colName] = v.String
			rr.scanDest = append(rr.scanDest, &v)
		default:
			cl.logger.Error("column type not supported", zap.String("type", columnType.ScanType().Name()))
		}
	}
	for sqlRows.Next() {
		err = sqlRows.Scan(rr.scanDest...)
		if err != nil {
			return nil, err
		}
		out = append(out, rr.toMetricRow())
	}
	return out, nil
}

type reusableRow struct {
	attrs    map[string]func() string
	scanDest []any
}

func (r reusableRow) toMetricRow() metricRow {
	out := metricRow{}
	for k, f := range r.attrs {
		out[k] = f()
	}
	return out
}
