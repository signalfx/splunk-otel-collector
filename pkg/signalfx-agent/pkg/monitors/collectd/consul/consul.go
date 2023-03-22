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
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	python.CommonConfig  `yaml:",inline"`
	pyConf               *python.Config
	Host                 string `yaml:"host" validate:"required"`
	Port                 uint16 `yaml:"port" validate:"required"`
	// Consul ACL token
	ACLToken string `yaml:"aclToken" neverLog:"true"`
	// Set to `true` to connect to Consul using HTTPS.  You can figure the
	// certificate for the server with the `caCertificate` config option.
	UseHTTPS        bool `yaml:"useHTTPS"`
	TelemetryServer bool `yaml:"telemetryServer"`
	// IP address or DNS to which Consul is configured to send telemetry UDP packets. Relevant only if `telemetryServer` is set to true.
	TelemetryHost string `yaml:"telemetryHost" default:"0.0.0.0"`
	// Port to which Consul is configured to send telemetry UDP packets. Relevant only if `telemetryServer` is set to true.
	TelemetryPort int `yaml:"telemetryPort" default:"8125"`
	// Set to *true* to enable collecting all metrics from Consul's runtime telemetry send via UDP or from the `/agent/metrics` endpoint.
	EnhancedMetrics *bool `yaml:"enhancedMetrics"`
	// If Consul server has HTTPS enabled for the API, specifies the path to the CA's Certificate.
	CACertificate string `yaml:"caCertificate"`
	// If client-side authentication is enabled, specifies the path to the certificate file.
	ClientCertificate string `yaml:"clientCertificate"`
	// If client-side authentication is enabled, specifies the path to the key file.
	ClientKey           string `yaml:"clientKey"`
	SignalFxAccessToken string `yaml:"signalFxAccessToken" neverLog:"true"`
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
		ModulePaths:   []string{collectd.MakePythonPluginPath("consul")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
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
