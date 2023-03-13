//go:build linux
// +build linux

package apache

//go:generate ../../../../scripts/collectd-template-to-go apache.tmpl

import (
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

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`

	// The hostname of the Apache server
	Host string `yaml:"host" validate:"required"`
	// The port number of the Apache server
	Port uint16 `yaml:"port" validate:"required"`
	// This will be sent as the `plugin_instance` dimension and can be any name
	// you like.
	Name string `yaml:"name"`
	// You can specify a username and password to do basic HTTP auth

	// The URL, either a final URL or a Go template that will be populated with
	// the host and port values.
	URL      string `yaml:"url" default:"http://{{.Host}}:{{.Port}}/mod_status?auto"`
	Username string `yaml:"username"`
	Password string `yaml:"password" neverLog:"true"`
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
}

// Configure configures and runs the plugin in collectd
func (am *Monitor) Configure(conf *Config) error {
	return am.SetConfigurationAndRun(conf)
}
