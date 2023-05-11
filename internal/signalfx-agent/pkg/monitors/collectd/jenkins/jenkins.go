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
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	python.CommonConfig  `yaml:",inline"`
	pyConf               *python.Config
	Host                 string `yaml:"host" validate:"required"`
	Port                 uint16 `yaml:"port" validate:"required"`
	Path                 string `yaml:"path"`
	// Key required for collecting metrics.  The access key located at
	// `Manage Jenkins > Configure System > Metrics > ADD.`
	// If empty, click `Generate`.
	MetricsKey string `yaml:"metricsKey" validate:"required"`
	// Whether to enable enhanced metrics
	EnhancedMetrics *bool `yaml:"enhancedMetrics"`
	// Set to *true* to exclude job metrics retrieved from `/api/json` endpoint
	ExcludeJobMetrics *bool `yaml:"excludeJobMetrics"`
	// Used to enable individual enhanced metrics when `enhancedMetrics` is
	// false
	IncludeMetrics []string `yaml:"includeMetrics"`
	// User with security access to jenkins
	Username string `yaml:"username"`
	// API Token of the user
	APIToken string `yaml:"apiToken" neverLog:"true"`
	// Whether to enable HTTPS.
	UseHTTPS bool `yaml:"useHTTPS"`
	// Path to the keyfile
	SSLKeyFile string `yaml:"sslKeyFile"`
	// Path to the certificate
	SSLCertificate string `yaml:"sslCertificate"`
	// Path to the ca file
	SSLCACerts string `yaml:"sslCACerts"`
	// Skip SSL certificate validation
	SkipVerify bool `yaml:"skipVerify"`
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
		ModulePaths:   []string{collectd.MakePythonPluginPath("jenkins")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
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
