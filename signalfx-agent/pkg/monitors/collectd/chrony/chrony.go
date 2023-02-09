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

package chrony

//go:generate ../../../../scripts/collectd-template-to-go chrony.tmpl

import (
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			MonitorCore: *collectd.NewMonitorCore(CollectdTemplate),
		}
	}, &Config{})
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	// This cannot be multi instance until there is a way to differentiate them
	// in collectd
	config.MonitorConfig `yaml:",inline" singleInstance:"true"`

	// The hostname of the chronyd instance
	Host string `yaml:"host" validate:"required" default:"localhost"`
	// The UDP port number of the chronyd instance.  Defaults to 323 in
	// collectd if unspecified.
	Port *uint16 `yaml:"port"`
	// How long to wait for a response from chronyd before considering it down.
	// Defaults to 2 seconds in the collectd plugin if not specified
	Timeout *uint `yaml:"timeout"`
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	return m.SetConfigurationAndRun(conf)
}
