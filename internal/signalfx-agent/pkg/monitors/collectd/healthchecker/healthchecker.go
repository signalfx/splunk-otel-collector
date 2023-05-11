package healthchecker

import (
	"errors"
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"

	"github.com/signalfx/signalfx-agent/pkg/core/config"

	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
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
	python.CommonConfig  `yaml:",inline"`
	pyConf               *python.Config
	Host                 string `yaml:"host" validate:"required"`
	Port                 uint16 `yaml:"port" validate:"required"`
	Name                 string `yaml:"name"`
	// The HTTP path that contains a JSON document to verify
	Path string `yaml:"path" default:"/"`
	// If `jsonKey` and `jsonVal` are given, the given endpoint will be
	// interpreted as a JSON document and will be expected to contain the given
	// key and value for the service to be considered healthy.
	JSONKey string `yaml:"jsonKey"`
	// This can be either a string or numeric type
	JSONVal interface{} `yaml:"jsonVal"`
	// If true, the endpoint will be connected to on HTTPS instead of plain
	// HTTP.  It is invalid to specify this if `tcpCheck` is true.
	UseHTTPS bool `yaml:"useHTTPS"`
	// If true, and `useHTTPS` is true, the server's SSL/TLS cert will not be
	// verified.
	SkipSecurity bool `yaml:"skipSecurity"`
	// If true, the plugin will verify that it can connect to the given
	// host/port value. JSON checking is not supported.
	TCPCheck bool `yaml:"tcpCheck"`
}

// PythonConfig returns the embedded python.Config struct from the interface
func (c *Config) PythonConfig() *python.Config {
	c.pyConf.CommonConfig = c.CommonConfig
	return c.pyConf
}

// Validate the given config
func (c *Config) Validate() error {
	if c.TCPCheck && (c.SkipSecurity || c.UseHTTPS) {
		return errors.New("neither skipSecurity nor useHTTPS should be set when tcpCheck is true")
	}
	if c.TCPCheck && (c.JSONKey != "" || c.JSONVal != nil) {
		return errors.New("cannot do JSON value check with tcpCheck set to true")
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
		ModuleName:    "health_checker",
		ModulePaths:   []string{collectd.MakePythonPluginPath("health_checker")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
		PluginConfig: map[string]interface{}{
			"Instance": conf.Name,
			"JSONKey":  conf.JSONKey,
			"JSONVal":  conf.JSONVal,
		},
	}

	if conf.TCPCheck {
		conf.pyConf.PluginConfig["URL"] = conf.Host
		conf.pyConf.PluginConfig["TCP"] = conf.Port
	} else {
		protocol := "http"
		if conf.UseHTTPS {
			protocol = "https"
		}
		conf.pyConf.PluginConfig["URL"] = fmt.Sprintf("%s://%s:%d%s", protocol, conf.Host, conf.Port, conf.Path)
		conf.pyConf.PluginConfig["SkipSecurity"] = conf.SkipSecurity
	}

	if conf.Name == "" {
		conf.pyConf.PluginConfig["Instance"] = fmt.Sprintf("%s-%d", conf.Host, conf.Port)
	}

	return m.PyMonitor.Configure(conf)
}
