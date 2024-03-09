package consul

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
	EnhancedMetrics      *bool `yaml:"enhancedMetrics"`
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	CACertificate        string `yaml:"caCertificate"`
	python.CommonConfig  `yaml:",inline"`
	Host                 string `yaml:"host" validate:"required"`
	SignalFxAccessToken  string `yaml:"signalFxAccessToken" neverLog:"true"`
	ACLToken             string `yaml:"aclToken" neverLog:"true"`
	ClientKey            string `yaml:"clientKey"`
	ClientCertificate    string `yaml:"clientCertificate"`
	TelemetryHost        string `yaml:"telemetryHost" default:"0.0.0.0"`
	TelemetryPort        int    `yaml:"telemetryPort" default:"8125"`
	Port                 uint16 `yaml:"port" validate:"required"`
	TelemetryServer      bool   `yaml:"telemetryServer"`
	UseHTTPS             bool   `yaml:"useHTTPS"`
}

// GetExtraMetrics takes into account the EnhancedMetrics flag
func (c *Config) GetExtraMetrics() []string {
	if c.EnhancedMetrics != nil && *c.EnhancedMetrics {
		return []string{"*"}
	}
	return nil
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

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	conf.pyConf = &python.Config{
		MonitorConfig: conf.MonitorConfig,
		Host:          conf.Host,
		Port:          conf.Port,
		ModuleName:    "consul_plugin",
		ModulePaths:   []string{collectd.MakePythonPluginPath(conf.BundleDir, "consul")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath(conf.BundleDir)},
		PluginConfig: map[string]interface{}{
			"ApiHost":           conf.Host,
			"ApiPort":           conf.Port,
			"TelemetryServer":   conf.TelemetryServer,
			"TelemetryHost":     conf.TelemetryHost,
			"TelemetryPort":     conf.TelemetryPort,
			"SfxToken":          conf.SignalFxAccessToken,
			"EnhancedMetrics":   conf.EnhancedMetrics,
			"AclToken":          conf.ACLToken,
			"CaCertificate":     conf.CACertificate,
			"ClientCertificate": conf.ClientCertificate,
			"ClientKey":         conf.ClientKey,
		},
	}

	if conf.UseHTTPS {
		conf.pyConf.PluginConfig["ApiProtocol"] = "https"
	} else {
		conf.pyConf.PluginConfig["ApiProtocol"] = "http"
	}

	return m.PyMonitor.Configure(conf)
}
