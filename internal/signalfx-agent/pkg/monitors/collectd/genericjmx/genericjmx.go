//go:build linux
// +build linux

// Package genericjmx coordinates the various monitors that rely on the
// GenericJMX Collectd plugin to pull JMX metrics.  All of the GenericJMX
// monitors share the same instance of the coreInstance struct that can be
// gotten by the Instance func in this package.  This ultimately means that all
// GenericJMX config will be written to one file to make it simpler to control
// dependencies.
package genericjmx

import (
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

//go:generate ../../../../scripts/collectd-template-to-go genericjmx.tmpl

// Config has configuration that is specific to GenericJMX. This config should
// be used by a monitors that use the generic JMX collectd plugin.
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`

	// Host to connect to -- JMX must be configured for remote access and
	// accessible from the agent
	Host string `yaml:"host" validate:"required"`
	// JMX connection port (NOT the RMI port) on the application.  This
	// correponds to the `com.sun.management.jmxremote.port` Java property that
	// should be set on the JVM when running the application.
	Port uint16 `yaml:"port" validate:"required"`
	Name string `yaml:"name"`
	// This is how the service type is identified in the SignalFx UI so that
	// you can get built-in content for it.  For custom JMX integrations, it
	// can be set to whatever you like and metrics will get the special
	// property `sf_hostHasService` set to this value.
	ServiceName string `yaml:"serviceName"`
	// The JMX connection string.  This is rendered as a Go template and has
	// access to the other values in this config. NOTE: under normal
	// circumstances it is not advised to set this string directly - setting
	// the host and port as specified above is preferred.
	ServiceURL string `yaml:"serviceURL" default:"service:jmx:rmi:///jndi/rmi://{{.Host}}:{{.Port}}/jmxrmi"`
	// Prefixes the generated plugin instance with prefix.
	// If a second `instancePrefix` is specified in a referenced MBean block,
	// the prefix specified in the Connection block will appear at the
	// beginning of the plugin instance, and the prefix specified in the
	// MBean block will be appended to it
	InstancePrefix string `yaml:"instancePrefix"`
	// Username to authenticate to the server
	Username string `yaml:"username"`
	// User password to authenticate to the server
	Password string `yaml:"password" neverLog:"true"`
	// Takes in key-values pairs of custom dimensions at the connection level.
	CustomDimensions map[string]string `yaml:"customDimensions"`
	// A list of the MBeans defined in `mBeanDefinitions` to actually collect.
	// If not provided, then all defined MBeans will be collected.
	MBeansToCollect []string `yaml:"mBeansToCollect"`
	// A list of the MBeans to omit. This will come handy in cases where only a
	// few MBeans need to omitted from the default list
	MBeansToOmit []string `yaml:"mBeansToOmit"`
	// Specifies how to map JMX MBean values to metrics.  If using a specific
	// service monitor such as cassandra, kafka, or activemq, they come
	// pre-loaded with a set of mappings, and any that you add in this option
	// will be merged with those.  See
	// [collectd GenericJMX](https://collectd.org/documentation/manpages/collectd-java.5.shtml#genericjmx_plugin)
	// for more details.
	MBeanDefinitions MBeanMap `yaml:"mBeanDefinitions"`
}

// JMXMonitorCore should be embedded by all monitors that use the collectd
// GenericJMX plugin.  It has most of the logic they will need.  The individual
// monitors mainly just need to provide their set of default mBean definitions.
type JMXMonitorCore struct {
	collectd.MonitorCore

	DefaultMBeans      MBeanMap
	defaultServiceName string
}

// NewJMXMonitorCore makes a new JMX core as well as the underlying MonitorCore
func NewJMXMonitorCore(defaultMBeans MBeanMap, defaultServiceName string) *JMXMonitorCore {
	mc := &JMXMonitorCore{
		MonitorCore:        *collectd.NewMonitorCore(CollectdTemplate),
		DefaultMBeans:      defaultMBeans,
		defaultServiceName: defaultServiceName,
	}

	mc.MonitorCore.UsesGenericJMX = true
	return mc
}

// Configure configures and runs the plugin in collectd
func (m *JMXMonitorCore) Configure(conf *Config) error {
	conf.MBeanDefinitions = m.DefaultMBeans.MergeWith(conf.MBeanDefinitions)
	if conf.MBeansToCollect == nil {
		conf.MBeansToCollect = conf.MBeanDefinitions.MBeanNames()
	}
	if conf.MBeansToOmit != nil {
		conf.MBeansToCollect = utils.RemoveAllElementsFromStringSlice(conf.MBeansToCollect, conf.MBeansToOmit)
	}
	if conf.ServiceName == "" {
		conf.ServiceName = m.defaultServiceName
	}

	return m.SetConfigurationAndRun(conf)
}
