package marathon

import (
	"errors"

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
	// Make this single instance since we can't add dimensions
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true" singleInstance:"true"`
	python.CommonConfig  `yaml:",inline"`
	pyConf               *python.Config
	Host                 string `yaml:"host" validate:"required"`
	Port                 uint16 `yaml:"port" validate:"required"`
	// Username used to authenticate with Marathon.
	Username string `yaml:"username"`
	// Password used to authenticate with Marathon.
	Password string `yaml:"password" neverLog:"true"`
	// Set to either `http` or `https`.
	Scheme string `yaml:"scheme" default:"http"`
	// The dcos authentication URL which the plugin uses to get authentication
	// tokens from. Set scheme to "https" if operating DC/OS in strict mode and
	// dcosAuthURL to "https://leader.mesos/acs/api/v1/auth/login"
	// (which is the default DNS entry provided by DC/OS)
	DCOSAuthURL string `yaml:"dcosAuthURL"`
}

// PythonConfig returns the embedded python.Config struct from the interface
func (c *Config) PythonConfig() *python.Config {
	c.pyConf.CommonConfig = c.CommonConfig
	return c.pyConf
}

// Validate config issues
func (c *Config) Validate() error {
	if c.DCOSAuthURL != "" && c.Scheme != "https" {
		return errors.New("scheme must be set to https when using a DCOSAuthURL")
	}
	return nil
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	python.PyMonitor
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	conf.pyConf = &python.Config{
		MonitorConfig: conf.MonitorConfig,
		Host:          conf.Host,
		Port:          conf.Port,
		ModuleName:    "marathon",
		ModulePaths:   []string{collectd.MakePythonPluginPath("marathon")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
		PluginConfig: map[string]interface{}{
			"verbose": false,
		},
	}

	// marathon's configuration is different, all configurations are
	// packed into an array of values for a given host
	host := []interface{}{conf.Scheme, conf.Host, conf.Port}
	if conf.Username != "" {
		host = append(host, conf.Username)
	}
	if conf.Password != "" {
		host = append(host, conf.Password)
	}
	if conf.DCOSAuthURL != "" {
		host = append(host, conf.DCOSAuthURL)
	}
	conf.pyConf.PluginConfig["host"] = host

	return m.PyMonitor.Configure(conf)
}
