//go:build linux
// +build linux

// Package netinterface wraps the "interface" collectd plugin for gather
// network interface metrics.  It is called netinterface because "interface" is
// a keyword in golang.
package netinterface

//go:generate ../../../../scripts/collectd-template-to-go interface.tmpl

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
	// List of interface names to exclude from monitoring
	ExcludedInterfaces []string `yaml:"excludedInterfaces" default:"[\"/^lo\\\\d*$/\", \"/^docker.*/\", \"/^t(un|ap)\\\\d*$/\", \"/^veth.*$/\"]"`
	// List of all the interfaces you want to monitor, all others will be
	// ignored.  If you set both included and excludedInterfaces, only
	// includedInterfaces will be honored.
	IncludedInterfaces []string `yaml:"includedInterfaces"`
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	newConf := *conf
	// Get rid of default excluded if includedInterfaces is explicitly
	// provided.
	if len(newConf.IncludedInterfaces) > 0 {
		newConf.ExcludedInterfaces = nil
	}
	return m.SetConfigurationAndRun(&newConf)
}
