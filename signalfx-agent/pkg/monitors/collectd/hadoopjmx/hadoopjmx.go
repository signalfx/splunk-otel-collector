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

//go:build linux
// +build linux

package hadoopjmx

import (
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/genericjmx"
	yaml "gopkg.in/yaml.v2"
)

var serviceName = "hadoop"

type nodeType string

const (
	nameNode        nodeType = "nameNode"
	resourceManager nodeType = "resourceManager"
	nodeManager     nodeType = "nodeManager"
	dataNode        nodeType = "dataNode"
)

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	genericjmx.Config `yaml:",inline"`
	// Hadoop Node Type
	NodeType nodeType `yaml:"nodeType" validate:"required"`
}

// Validate that the nodeType is one of our defined constants
func (c *Config) Validate() error {
	switch c.NodeType {
	case nameNode, dataNode, resourceManager, nodeManager:
		return nil
	default:
		return fmt.Errorf("required configuration nodeType '%s' is invalid", c.NodeType)
	}
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	*genericjmx.JMXMonitorCore
}

// Configure configures the hadoopjmx monitor and instantiates the generic jmx
// monitor
func (m *Monitor) Configure(conf *Config) error {
	// create the mbean map with the appropriate mbeans for the given node type
	var newMBeans genericjmx.MBeanMap
	switch conf.NodeType {
	case nameNode:
		newMBeans = genericjmx.DefaultMBeans.MergeWith(loadMBeans(defaultNameNodeMBeanYAML))
	case dataNode:
		newMBeans = genericjmx.DefaultMBeans.MergeWith(loadMBeans(defaultDataNodeMBeanYAML))
	case resourceManager:
		newMBeans = genericjmx.DefaultMBeans.MergeWith(loadMBeans(defaultResourceManagerMBeanYAML))
	case nodeManager:
		newMBeans = genericjmx.DefaultMBeans.MergeWith(loadMBeans(defaultNodeManagerMBeanYAML))
	}

	m.JMXMonitorCore.DefaultMBeans = newMBeans

	// invoke the JMXMonitorCore configuration callback
	return m.JMXMonitorCore.Configure(&conf.Config)
}

// loadMBeans validates the mbean yaml and unmarshals the mbean yaml to an MBeanMap
func loadMBeans(mBeanYaml string) genericjmx.MBeanMap {
	var mbeans genericjmx.MBeanMap

	if err := yaml.Unmarshal([]byte(mBeanYaml), &mbeans); err != nil {
		panic("YAML for GenericJMX MBeans is invalid: " + err.Error())
	}

	return mbeans
}

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			genericjmx.NewJMXMonitorCore(genericjmx.DefaultMBeans, serviceName),
		}
	}, &Config{})
}
