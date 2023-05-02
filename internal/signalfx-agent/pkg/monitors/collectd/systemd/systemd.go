//go:build linux
// +build linux

package systemd

import (
	"strings"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
)

const (
	activeState = "ActiveState"
	subState    = "SubState"
	loadState   = "LoadState"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			python.PyMonitor{
				MonitorCore: subproc.New(),
			},
		}
	}, &Config{})
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	pyConf               *python.Config
	// Systemd services to report on
	Services []string `yaml:"services" validate:"required"`
	// Flag for sending metrics about the state of systemd services
	SendActiveState bool `yaml:"sendActiveState"`
	// Flag for sending more detailed metrics about the state of systemd services
	SendSubState bool `yaml:"sendSubState"`
	// Flag for sending metrics about the load state of systemd services
	SendLoadState bool `yaml:"sendLoadState"`
}

// PythonConfig returns the embedded python.Config struct from the interface
func (c *Config) PythonConfig() *python.Config {
	return c.pyConf
}

// GetExtraMetrics returns additional metrics to allow through.
func (c *Config) GetExtraMetrics() []string {
	extraMetrics := make([]string, 0)
	if c.SendActiveState {
		extraMetrics = append(extraMetrics, groupMetricsMap[activeState]...)
	}
	if c.SendSubState {
		extraMetrics = append(extraMetrics, groupMetricsMap[subState]...)
	}
	if c.SendLoadState {
		extraMetrics = append(extraMetrics, groupMetricsMap[loadState]...)
	}
	return extraMetrics
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	python.PyMonitor
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	var services []string
	for _, service := range conf.Services {
		services = append(services, strings.Trim(service, " "))
	}
	serviceStates := []string{subState}
	if conf.SendActiveState {
		serviceStates = append(serviceStates, activeState)
	}
	if conf.SendLoadState {
		serviceStates = append(serviceStates, loadState)
	}
	conf.pyConf = &python.Config{
		MonitorConfig: conf.MonitorConfig,
		ModuleName:    "collectd_systemd",
		ModulePaths:   []string{collectd.MakePythonPluginPath("systemd")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
		PluginConfig: map[string]interface{}{
			"Service":       services,
			"Interval":      conf.IntervalSeconds,
			"Verbose":       false,
			"ServiceStates": serviceStates,
		},
	}
	return m.PyMonitor.Configure(conf)
}
