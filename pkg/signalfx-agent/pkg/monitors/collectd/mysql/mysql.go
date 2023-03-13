//go:build linux
// +build linux

package mysql

//go:generate ../../../../scripts/collectd-template-to-go mysql.tmpl

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

// Database configures a particular MySQL database
type Database struct {
	Name     string `yaml:"name" validate:"required"`
	Username string `yaml:"username"`
	Password string `yaml:"password" neverLog:"true"`
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`

	Host string `yaml:"host" validate:"required"`
	Port uint16 `yaml:"port" validate:"required"`
	Name string `yaml:"name"`
	// A list of databases along with optional authentication credentials.
	Databases []Database `yaml:"databases" validate:"required"`
	// These credentials serve as defaults for all databases if not overridden
	Username string `yaml:"username"`
	Password string `yaml:"password" neverLog:"true"`
	// A SignalFx extension to the plugin that allows us to disable the normal
	// behavior of the MySQL collectd plugin where the `host` dimension is set
	// to the hostname of the MySQL database server.  When `false` (the
	// recommended and default setting), the globally configured `hostname`
	// config is used instead.
	ReportHost  bool `yaml:"reportHost"`
	InnodbStats bool `yaml:"innodbStats"`
}

// Validate will check the config for correctness.
func (c *Config) Validate() error {
	if len(c.Databases) == 0 {
		return errors.New("you must specify at least one database for MySQL")
	}

	for _, db := range c.Databases {
		if db.Username == "" && c.Username == "" {
			return errors.New("username is required for MySQL monitoring")
		}
	}
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
