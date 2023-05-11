//go:build linux
// +build linux

package disk

//go:generate ../../../../scripts/collectd-template-to-go disk.tmpl

import (
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			MonitorCore: *collectd.NewMonitorCore(CollectdTemplate),
		}
	}, &Config{})
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"true"`

	// Which devices to include/exclude
	Disks []string `yaml:"disks" default:"[\"/^loop[0-9]+$/\", \"/^dm-[0-9]+$/\"]"`

	// If true, the disks selected by `disks` will be excluded and all others
	// included.
	IgnoreSelected bool `yaml:"ignoreSelected" default:"true"`
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	return m.SetConfigurationAndRun(conf)
}
