package jenkins

import (
	"strconv"

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
	ExcludeJobMetrics    *bool `yaml:"excludeJobMetrics"`
	pyConf               *python.Config
	EnhancedMetrics      *bool `yaml:"enhancedMetrics"`
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	MetricsKey           string `yaml:"metricsKey" validate:"required"`
	Path                 string `yaml:"path"`
	SSLCertificate       string `yaml:"sslCertificate"`
	Host                 string `yaml:"host" validate:"required"`
	python.CommonConfig  `yaml:",inline"`
	Username             string   `yaml:"username"`
	APIToken             string   `yaml:"apiToken" neverLog:"true"`
	SSLCACerts           string   `yaml:"sslCACerts"`
	SSLKeyFile           string   `yaml:"sslKeyFile"`
	IncludeMetrics       []string `yaml:"includeMetrics"`
	Port                 uint16   `yaml:"port" validate:"required"`
	UseHTTPS             bool     `yaml:"useHTTPS"`
	SkipVerify           bool     `yaml:"skipVerify"`
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
		ModuleName:    "jenkins",
		ModulePaths:   []string{collectd.MakePythonPluginPath(conf.BundleDir, "jenkins")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath(conf.BundleDir)},
		PluginConfig: map[string]interface{}{
			"Host":                conf.Host,
			"Port":                conf.Port,
			"Path":                conf.Path,
			"Interval":            conf.IntervalSeconds,
			"MetricsKey":          conf.MetricsKey,
			"EnhancedMetrics":     conf.EnhancedMetrics,
			"ExcludeJobMetrics":   conf.ExcludeJobMetrics,
			"Username":            conf.Username,
			"APIToken":            conf.APIToken,
			"ssl_keyfile":         conf.SSLKeyFile,
			"ssl_certificate":     conf.SSLCertificate,
			"ssl_ca_certs":        conf.SSLCACerts,
			"ssl_enabled":         strconv.FormatBool(conf.UseHTTPS),
			"ssl_cert_validation": strconv.FormatBool(!conf.SkipVerify),
			"IncludeMetric": map[string]interface{}{
				"#flatten": true,
				"values":   conf.IncludeMetrics,
			},
		},
	}

	return m.PyMonitor.Configure(conf)
}
