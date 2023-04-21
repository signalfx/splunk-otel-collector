package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	// Imports to get sql driver registered
	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
	_ "github.com/snowflakedb/gosnowflake"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Query is used to configure a query statement and the resulting datapoints
type Query struct {
	// A SQL query text that selects one or more rows from a database
	Query string `yaml:"query" validate:"required"`
	// Optional parameters that will replace placeholders in the query string.
	Params []interface{} `yaml:"params"`
	// Metrics that should be generated from the query.
	Metrics []Metric `yaml:"metrics"`
	// A set of [expr] expressions that will be used to convert each row to a
	// set of metrics.  Each of these will be run for each row in the query
	// result set, allowing you to generate multiple datapoints per row.  Each
	// expression should evaluate to a single datapoint or nil.
	DatapointExpressions []string `yaml:"datapointExpressions"`
}

// Metric describes how to derive a metric from the individual rows of a query
// result.
type Metric struct {
	// The name of the metric as it will appear in SignalFx.
	MetricName string `yaml:"metricName" validate:"required"`
	// The column name that holds the datapoint value
	ValueColumn string `yaml:"valueColumn" validate:"required"`
	// The names of the columns that should make up the dimensions of the
	// datapoint.
	DimensionColumns []string `yaml:"dimensionColumns"`
	// Whether the value is a cumulative counters (true) or gauge
	// (false).  If you set this to the wrong value and send in your first
	// datapoint for the metric name with the wrong type, you will have to
	// manually change the type in SignalFx, as it is set in the system based
	// on the first type seen.
	IsCumulative bool `yaml:"isCumulative"`
	// The mapping between dimensions and the columns to be used to attach respective properties
	DimensionPropertyColumns map[string][]string `yaml:"dimensionPropertyColumns"`
}

func (m *Metric) NewDatapoint() *datapoint.Datapoint {
	typ := datapoint.Gauge
	if m.IsCumulative {
		typ = datapoint.Counter
	}
	return datapoint.New(m.MetricName, map[string]string{}, nil, typ, time.Time{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`

	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`

	// Parameters to the connectionString that can be templated into that option using
	// Go template syntax (e.g. `{{.key}}`).
	Params map[string]string `yaml:"params"`

	// The database driver to use, valid values are `postgres`, `mysql`, `sqlserver`,
	// and `snowflake`.
	DBDriver string `yaml:"dbDriver"`
	// A URL or simple option string used to connect to the database.
	// For example, if using PostgreSQL, [see the list of connection string
	// params](https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters).
	ConnectionString string `yaml:"connectionString"`

	// A list of queries to make against the database that are used to generate
	// datapoints.
	Queries []Query `yaml:"queries" validate:"required"`
	// If true, query results will be logged at the info level.
	LogQueries bool `yaml:"logQueries"`
}

// Validate that the config is right
func (c *Config) Validate() error {
	if c.DBDriver != "postgres" && c.DBDriver != "mysql" && c.DBDriver != "sqlserver" && c.DBDriver != "snowflake" {
		return fmt.Errorf("database driver %s is not supported", c.DBDriver)
	}

	if len(c.Queries) == 0 {
		return errors.New("must specify at least one query")
	}

	for i := range c.Queries {
		if len(c.Queries[i].Metrics) == 0 && len(c.Queries[i].DatapointExpressions) == 0 {
			return errors.New("each SQL query must have at least one metric or expression defined on it")
		}
		valueCols := map[string]bool{}
		for _, met := range c.Queries[i].Metrics {
			if seen := valueCols[met.ValueColumn]; seen {
				return fmt.Errorf("sql query metric value column %s is repeated in the same query", met.ValueColumn)
			}
		}
	}
	return nil
}

func (c *Config) renderedDataSource() (string, error) {
	context, err := utils.ConvertToMapViaYAML(c)
	if err != nil {
		return "", err
	}
	for k, v := range c.Params {
		context[k] = v
	}

	rendered, err := utils.RenderSimpleTemplate(c.ConnectionString, context)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(rendered), nil
}

// Monitor for generic SQL queries -> metrics
type Monitor struct {
	Output   types.Output
	database *sql.DB
	cancel   context.CancelFunc
	ctx      context.Context
	logger   logrus.FieldLogger
}

// Configure the monitor and kick off metric gathering
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithField("monitorType", conf.Type).WithField("monitorID", conf.MonitorID)
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// This will "open" a database by verifying that the config is sane but
	// generally won't try and connect to it.  If it does attempt to connect
	// here this should be done within RunOnInterval to handle cases where the
	// database is temporarily down when the monitor starts.
	dataSource, err := conf.renderedDataSource()
	if err != nil {
		return err
	}

	m.database, err = sql.Open(conf.DBDriver, dataSource)
	if err != nil {
		return fmt.Errorf("could not handle %s database config: %v", conf.DBDriver, err)
	}

	for i := range conf.Queries {
		querier, err := newQuerier(&conf.Queries[i], conf.LogQueries, m.logger)
		if err != nil {
			return err
		}

		utils.RunOnInterval(m.ctx, func() {
			if err := querier.doQuery(m.ctx, m.database, m.Output); err != nil {
				querier.logger.WithError(err).Error("Problem running SQL query or converting datapoints")
			}
		}, time.Duration(conf.IntervalSeconds)*time.Second)
	}

	return nil
}

// Shutdown the monitor and close the DB connection
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}

	if m.database != nil {
		m.database.Close()
	}
}
