//go:build linux
// +build linux

package memcached

//go:generate ../../../../scripts/collectd-template-to-go memcached.tmpl

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

	Host       string `yaml:"host" validate:"required"`
	Port       uint16 `yaml:"port" validate:"required"`
	Name       string `yaml:"name"`
	ReportHost bool   `yaml:"reportHost"`
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
}

// Configure configures and runs the plugin in collectd
func (mm *Monitor) Configure(conf *Config) error {
	return mm.SetConfigurationAndRun(conf)
}
