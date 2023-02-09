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

package diskio

import (
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
	log "github.com/sirupsen/logrus"
)

var logger = log.WithFields(log.Fields{"monitorType": monitorType})

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"false"`
	// The devices to include/exclude. This is an [overridable
	// set](https://docs.splunk.com/observability/gdi/smart-agent/smart-agent-resources.html#filtering-data-using-the-smart-agent).
	Disks []string `yaml:"disks" default:"[\"*\", \"!/^loop[0-9]+$/\", \"!/^dm-[0-9]+$/\"]"`
	// (Windows Only) The frequency that wildcards in counter paths should
	// be expanded and how often to refresh counters from configuration.
	// This is expressed as a duration.
	CountersRefreshInterval timeutil.Duration `yaml:"counterRefreshInterval" default:"60s"`
	// (Windows Only) Print out the configurations that match available
	// performance counters.  This used for debugging.
	PrintValid bool `yaml:"printValid"`
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
