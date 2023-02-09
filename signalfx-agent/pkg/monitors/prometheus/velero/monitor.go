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

package velero

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{Monitor: prometheusexporter.Monitor{}}
	}, &Config{})
}

// Config for this monitor
type Config struct {
	prometheusexporter.Config `yaml:",inline"`
}

// Monitor for Velero
type Monitor struct {
	prometheusexporter.Monitor
}

// Configure the underlying Prometheus exporter monitor
func (m *Monitor) Configure(conf *Config) error {
	return m.Monitor.Configure(&conf.Config)
}
