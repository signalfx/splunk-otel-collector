package zookeeper

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"

	"github.com/signalfx/signalfx-agent/pkg/core/config"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"
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
	pyConf               *python.Config
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	python.CommonConfig  `yaml:",inline"`
	Host                 string `yaml:"host" validate:"required"`
	Name                 string `yaml:"name"`
	Port                 uint16 `yaml:"port" validate:"required"`
}

// PythonConfig returns the embedded python.Config struct from the interface
func (c *Config) PythonConfig() *python.Config {
	c.pyConf.CommonConfig = c.CommonConfig
	return c.pyConf
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	python.PyMonitor
}

// Configure configures and runs the plugin in python
func (rm *Monitor) Configure(conf *Config) error {
	conf.pyConf = &python.Config{
		MonitorConfig: conf.MonitorConfig,
		ModuleName:    "zk-collectd",
		ModulePaths:   []string{collectd.MakePythonPluginPath(conf.BundleDir, "zookeeper")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath(conf.BundleDir)},
		Host:          conf.Host,
		Port:          conf.Port,
		PluginConfig: map[string]interface{}{
			"Hosts":    conf.Host,
			"Port":     conf.Port,
			"Instance": conf.Name,
		},
	}

	return rm.PyMonitor.Configure(conf)
}
