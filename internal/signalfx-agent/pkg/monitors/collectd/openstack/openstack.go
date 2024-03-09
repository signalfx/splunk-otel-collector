package openstack

import (
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"gopkg.in/yaml.v2"

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
	NovaListServersSearchOpts map[string]string `yaml:"novaListServersSearchOpts"`
	pyConf                    *python.Config
	QueryHypervisorMetrics    *bool `yaml:"queryHypervisorMetrics" default:"true"`
	QueryServerMetrics        *bool `yaml:"queryServerMetrics" default:"true"`
	config.MonitorConfig      `yaml:",inline" acceptsEndpoints:"false"`
	RegionName                string `yaml:"regionName"`
	ProjectName               string `yaml:"projectName"`
	ProjectDomainID           string `yaml:"projectDomainID"`
	Password                  string `yaml:"password" validate:"required"`
	UserDomainID              string `yaml:"userDomainID"`
	Username                  string `yaml:"username" validate:"required"`
	AuthURL                   string `yaml:"authURL" validate:"required"`
	python.CommonConfig       `yaml:",inline"`
	HTTPTimeout               float64 `yaml:"httpTimeout"`
	RequestBatchSize          int     `yaml:"requestBatchSize" default:"5"`
	SkipVerify                bool    `yaml:"skipVerify"`
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
	novaListServersSearchOpts := "{}"
	if len(conf.NovaListServersSearchOpts) > 0 {
		marshaled, err := yaml.Marshal(conf.NovaListServersSearchOpts)
		if err != nil {
			return fmt.Errorf("failed to parse novaListServersSearchOpts: %w", err)
		}
		novaListServersSearchOpts = string(marshaled)
	}
	conf.pyConf = &python.Config{
		ModuleName:    "openstack_metrics",
		ModulePaths:   []string{collectd.MakePythonPluginPath(conf.BundleDir, "openstack")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath(conf.MonitorConfig.BundleDir)},
		MonitorConfig: conf.MonitorConfig,
		PluginConfig: map[string]interface{}{
			"AuthURL":                   conf.AuthURL,
			"Username":                  conf.Username,
			"Password":                  conf.Password,
			"ProjectName":               conf.ProjectName,
			"ProjectDomainId":           conf.ProjectDomainID,
			"RegionName":                conf.RegionName,
			"UserDomainId":              conf.UserDomainID,
			"SSLVerify":                 !conf.SkipVerify,
			"Interval":                  conf.IntervalSeconds,
			"HTTPTimeout":               conf.HTTPTimeout,
			"RequestBatchSize":          conf.RequestBatchSize,
			"QueryServerMetrics":        conf.QueryServerMetrics,
			"QueryHypervisorMetrics":    conf.QueryHypervisorMetrics,
			"NovaListServersSearchOpts": novaListServersSearchOpts,
		},
	}

	return m.PyMonitor.Configure(conf)
}
