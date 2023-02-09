// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package solr

import (
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
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
	// Cluster name of this solr cluster.
	Cluster string `yaml:"cluster"`
	// EnhancedMetrics boolean to indicate whether stats from /metrics are needed
	EnhancedMetrics *bool `yaml:"enhancedMetrics" default:"false"`
	// IncludeMetrics metric names from the /admin/metrics endpoint to include (valid when EnhancedMetrics is "false")
	IncludeMetrics []string `yaml:"includeMetrics"`
	// ExcludeMetrics metric names from the /admin/metrics endpoint to exclude (valid when EnhancedMetrics is "true")
	ExcludeMetrics []string `yaml:"excludeMetrics"`
}

func (c *Config) GetExtraMetrics() []string {
	if c.EnhancedMetrics != nil && *c.EnhancedMetrics {
		return []string{"*"}
	}
	return c.IncludeMetrics
}

var _ config.ExtraMetrics = &Config{}

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
		ModuleName:    "solr_collectd",
		ModulePaths:   []string{collectd.MakePythonPluginPath("solr")},
		TypesDBPaths:  []string{collectd.DefaultTypesDBPath()},
		PluginConfig: map[string]interface{}{
			"Host":            conf.Host,
			"Port":            conf.Port,
			"Cluster":         conf.Cluster,
			"EnhancedMetrics": conf.EnhancedMetrics,
			"IncludeMetric": map[string]interface{}{
				"#flatten": true,
				"values":   conf.IncludeMetrics,
			},
			"ExcludeMetric": map[string]interface{}{
				"#flatten": true,
				"values":   conf.ExcludeMetrics,
			},
			"Interval": conf.IntervalSeconds,
		},
	}

	return m.PyMonitor.Configure(conf)
}
