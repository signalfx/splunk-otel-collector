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
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false"`
	python.CommonConfig  `yaml:",inline"`
	pyConf               *python.Config
	// Keystone authentication URL/endpoint for the OpenStack cloud
	AuthURL string `yaml:"authURL" validate:"required"`
	// Username to authenticate with keystone identity
	Username string `yaml:"username" validate:"required"`
	// Password to authenticate with keystone identity
	Password string `yaml:"password" validate:"required"`
	// Specify the name of Project to be monitored (**default**:"demo")
	ProjectName string `yaml:"projectName"`
	// The project domain (**default**:"default")
	ProjectDomainID string `yaml:"projectDomainID"`
	// The region name for URL discovery, defaults to the first region if multiple regions are available.
	RegionName string `yaml:"regionName"`
	// The user domain id (**default**:"default")
	UserDomainID string `yaml:"userDomainID"`
	// Skip SSL certificate validation
	SkipVerify bool `yaml:"skipVerify"`
	// The HTTP client timeout in seconds for all requests
	HTTPTimeout float64 `yaml:"httpTimeout"`
	// The maximum number of concurrent requests for each metric class
	RequestBatchSize int `yaml:"requestBatchSize" default:"5"`
	// Whether to query server metrics (useful to disable for TripleO Undercloud)
	QueryServerMetrics *bool `yaml:"queryServerMetrics" default:"true"`
	// Whether to query hypervisor metrics (useful to disable for TripleO Undercloud)
	QueryHypervisorMetrics *bool `yaml:"queryHypervisorMetrics" default:"true"`
	// Optional search_opts mapping for collectd-openstack Nova client servers.list(search_opts=novaListServerSearchOpts).
	// For more information see https://docs.openstack.org/api-ref/compute/#list-servers.
	NovaListServersSearchOpts map[string]string `yaml:"novaListServersSearchOpts"`
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
		marshalled, err := yaml.Marshal(conf.NovaListServersSearchOpts)
		if err != nil {
			return fmt.Errorf("failed to parse novaListServersSearchOpts: %w", err)
		}
		novaListServersSearchOpts = string(marshalled)
	}
	conf.pyConf = &python.Config{
		ModuleName:    "openstack_metrics",
		ModulePaths:   []string{collectd.MakePythonPluginPath("openstack")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
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
