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

package statsd

//go:generate ../../../../scripts/collectd-template-to-go statsd.tmpl

import (
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			*collectd.NewMonitorCore(CollectdTemplate),
		}
	}, &Config{})
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"true"`

	// The host/address on which to bind the UDP listener that accepts statsd
	// datagrams
	ListenAddress string `yaml:"listenAddress" default:"localhost"`
	// The port on which to listen for statsd messages
	ListenPort      uint16  `yaml:"listenPort" default:"8125"`
	DeleteSets      bool    `yaml:"deleteSets"`
	DeleteCounters  bool    `yaml:"deleteCounters"`
	DeleteTimers    bool    `yaml:"deleteTimers"`
	DeleteGauges    bool    `yaml:"deleteGauges"`
	TimerPercentile float64 `yaml:"timerPercentile"`
	TimerUpper      bool    `yaml:"timerUpper"`
	TimerCount      bool    `yaml:"timerCount"`
	TimerSum        bool    `yaml:"timerSum"`
	TimerLower      bool    `yaml:"timerLower"`
	CounterSum      bool    `yaml:"counterSum"`
}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
}

// Configure configures and runs the plugin in collectd
func (am *Monitor) Configure(conf *Config) error {
	return am.SetConfigurationAndRun(conf)
}
