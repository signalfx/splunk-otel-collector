//go:build linux
// +build linux

package postgresql

//go:generate ../../../../scripts/collectd-template-to-go postgresql.tmpl

import (
	"errors"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			*collectd.NewMonitorCore(CollectdTemplate),
		}
	}, &Config{})
}

// Database configures a particular PostgreSQL database
type Database struct {
	// The name of the database
	Name string `yaml:"name" validate:"required"`
	// Username used to access the database
	Username string `yaml:"username"`
	// Password used to access the database
	Password string `yaml:"password" neverLog:"true"`
	// Interval to query the database in seconds
	Interval int `yaml:"interval"`
	// Skip expired values in query output
	ExpireDelay int `yaml:"expireDelay"`
	// Specify whether to use an ssl connection with PostgreSQL.
	// (prefer(default), disable, allow, require)
	SSLMode string `yaml:"sslMode"`
	// Specify the Kerberos service name used to authenticate with kerberos 5 or
	// GSSAPI
	KRBSrvName string `yaml:"krbSrvName"`
	// Queries used to generate metrics. These will override the default set.
	// If no queries are specified, the default set will be used
	// [`custom_deadlocks`, `backends`, `transactions`, `queries`, `queries_by_table`,
	// `query_plans`, `table_states`, `query_plans_by_table`, `table_states_by_table`,
	// `disk_io`, `disk_io_by_table`, `disk_usage`]
	Queries []string `yaml:"queries"`
}

// Result maps values from a query to a metric
type Result struct {
	// Type defines a metric type
	Type string `yaml:"type" validate:"required"`
	// Specifies columns in the SQL result to use as the metric
	// value.  The number of columns must match the expected number of values
	// for the metric type.
	ValuesFrom []string `yaml:"valuesFrom" validate:"required"`
	// A prefix for the type instance
	InstancePrefix string `yaml:"instancePrefix"`
	// Specifies columns in the SQL result to uses for the type
	// instance.  Multiple columns are joined with a hyphen "-".
	InstancesFrom []string `yaml:"instancesFrom"`
}

// Query adds a new query for retrieving metrics
type Query struct {
	// Name used to refer to the query in the database block
	Name string `yaml:"name" validate:"required"`
	// Statement is a SQL statement to execute
	Statement string `yaml:"statement" validate:"required"`
	// Result blocks that define mappings of SQL query results to
	// metrics
	Results []Result `yaml:"results" validate:"required"`
	// Parameters used to fill in $1,$2,$... tokens in the SQL
	// statement.  Acceptable values are hostname, database, instance, username,
	// interval
	Params []string `yaml:"params"`
	// Specifies the column that should be used to populate
	// plugin instance
	PluginInstanceFrom string `yaml:"pluginInstanceFrom"`
	// The minimum version of PostgreSQL that the query is
	// compatible with.  The version must be specified as a two decimal digit.
	// Ex. 7.2.3 -> 70203
	MinVersion int `yaml:"minVersion"`
	// The maximum version of PostgreSQL that the query is
	// compatible with.  The version must be specified as a two decimal digit.
	// Ex. 7.2.3 -> 70203
	MaxVersion int `yaml:"maxVersion"`
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	Host                 string `yaml:"host" validate:"required"`
	Port                 uint16 `yaml:"port" validate:"required"`
	// A list of databases along with optional authentication credentials.
	Databases []Database `yaml:"databases" validate:"required"`
	// PostgreSQL queries and metric mappings
	Queries []Query `yaml:"queries"`
	// A username that serves as a default for all databases if not overridden
	Username string `yaml:"username"`
	// A password that serves as a default for all databases if not overridden
	Password string `yaml:"password" neverLog:"true"`
	// A SignalFx extension to the plugin that allows us to disable the normal
	// behavior of the PostgreSQL collectd plugin where the `host` dimension is set
	// to the hostname of the PostgreSQL database server.  When `false` (the
	// recommended and default setting), the globally configured `hostname`
	// config is used instead.
	ReportHost bool `yaml:"reportHost"`
}

// Validate will check the config for correctness.
func (c *Config) Validate() error {
	if len(c.Databases) == 0 {
		return errors.New("you must specify at least one database for PostgreSQL")
	}

	for _, db := range c.Databases {
		if db.Username == "" && c.Username == "" {
			return errors.New("username is required for PostgreSQL monitoring")
		}
	}
	return nil
}

func (c *Config) GetExtraMetrics() []string {
	return nil
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
}

// Configure configures and runs the plugin in collectd
func (am *Monitor) Configure(conf *Config) error {
	return am.SetConfigurationAndRun(conf)
}
