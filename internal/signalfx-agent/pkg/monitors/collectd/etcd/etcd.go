package etcd

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
	pyConf               *python.Config
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	python.CommonConfig  `yaml:",inline"`
	Host                 string `yaml:"host" validate:"required"`
	ClusterName          string `yaml:"clusterName" validate:"required"`
	SSLKeyFile           string `yaml:"sslKeyFile"`
	SSLCertificate       string `yaml:"sslCertificate"`
	SSLCACerts           string `yaml:"sslCACerts"`
	Port                 uint16 `yaml:"port" validate:"required"`
	SkipSSLValidation    bool   `yaml:"skipSSLValidation"`
	EnhancedMetrics      bool   `yaml:"enhancedMetrics"`
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
	m.Logger().Warn("The ectd-collectd plugin is deprecated. Please use the etcd monitor instead. See https://docs.splunk.com/observability/en/gdi/monitors-databases/etcd.html for more information. This plugin will be removed in a future release.")
	conf.pyConf = &python.Config{
		MonitorConfig: conf.MonitorConfig,
		Host:          conf.Host,
		Port:          conf.Port,
		ModuleName:    "etcd_plugin",
		ModulePaths:   []string{collectd.MakePythonPluginPath(conf.BundleDir, "etcd")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath(conf.BundleDir)},
		PluginConfig: map[string]interface{}{
			"Host":     conf.Host,
			"Port":     conf.Port,
			"Interval": conf.IntervalSeconds,
			"Cluster":  conf.ClusterName,
			// Format as a string because collectd passes through bools as strings whereas
			// we pass them through as bools so the logic currently used in collectd-etcd
			// does not work correctly with bools. Maybe subproc should be changed to
			// behave the same as collectd?
			"ssl_cert_validation": strconv.FormatBool(!conf.SkipSSLValidation),
			"EnhancedMetrics":     conf.EnhancedMetrics,
			"ssl_keyfile":         conf.SSLKeyFile,
			"ssl_certificate":     conf.SSLCertificate,
			"ssl_ca_certs":        conf.SSLCACerts,
		},
	}

	return m.PyMonitor.Configure(conf)
}

// GetExtraMetrics returns additional metrics that should be allowed through.
func (c *Config) GetExtraMetrics() []string {
	if c.EnhancedMetrics {
		return monitorMetadata.NonDefaultMetrics()
	}
	return nil
}
