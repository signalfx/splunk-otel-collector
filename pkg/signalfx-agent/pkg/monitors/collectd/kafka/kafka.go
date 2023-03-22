//go:build linux
// +build linux

package kafka

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/genericjmx"
	yaml "gopkg.in/yaml.v2"
)

var serviceName = "kafka"

// Monitor is the main type that represents the monitor
type Monitor struct {
	*genericjmx.JMXMonitorCore
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	genericjmx.Config `yaml:",inline"`
	// Cluster name to which the broker belongs
	ClusterName string `yaml:"clusterName" validate:"required"`
}

// Configure configures the kafka monitor and instantiates the generic jmx
// monitor
func (m *Monitor) Configure(conf *Config) error {
	m.Output.AddExtraDimension("cluster", conf.ClusterName)
	return m.JMXMonitorCore.Configure(&conf.Config)
}

func init() {
	var defaultMBeans genericjmx.MBeanMap
	err := yaml.Unmarshal([]byte(defaultMBeanYAML), &defaultMBeans)
	if err != nil {
		panic("YAML for GenericJMX MBeans is invalid: " + err.Error())
	}
	defaultMBeans = defaultMBeans.MergeWith(genericjmx.DefaultMBeans)

	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			genericjmx.NewJMXMonitorCore(defaultMBeans, serviceName),
		}
	}, &Config{})
}
