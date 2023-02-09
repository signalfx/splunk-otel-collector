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

// Package cadvisor contains a monitor that pulls cadvisor stats either
// directly from cadvisor or from the kubelet /stats endpoint that exposes
// cadvisor.
package cadvisor

import (
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/cadvisor/converter"
)

// Monitor pulls metrics from a cAdvisor-compatible endpoint
type Monitor struct {
	monConfig *config.MonitorConfig
	stop      chan bool
	logger    log.FieldLogger
}

// Configure and start/restart cadvisor plugin
func (m *Monitor) Configure(
	monConfig *config.MonitorConfig,
	sendDPs func(...*datapoint.Datapoint),
	statProvider converter.InfoProvider,
	hasPodEphemeralStorageStatsGroupEnabled bool) error {

	m.monConfig = monConfig
	if m.logger == nil {
		m.logger = log.WithFields(log.Fields{"monitorType": monConfig.Type, "monitorID": monConfig.MonitorID})
	}

	collector := converter.NewCadvisorCollector(statProvider, sendDPs, monConfig.ExtraDimensions, m.logger)

	m.stop = m.monitorNode(collector, hasPodEphemeralStorageStatsGroupEnabled)

	return nil
}

func (m *Monitor) monitorNode(collector *converter.CadvisorCollector, hasPodEphemeralStorageStatsGroupEnabled bool) (stop chan bool) {
	ticker := time.NewTicker(time.Duration(m.monConfig.IntervalSeconds) * time.Second)
	stop = make(chan bool, 1)

	go func() {
		collector.Collect(hasPodEphemeralStorageStatsGroupEnabled)
		for {
			select {
			case <-stop:
				m.logger.Info("Stopping cAdvisor collection")
				ticker.Stop()
				return
			case <-ticker.C:
				collector.Collect(hasPodEphemeralStorageStatsGroupEnabled)
			}
		}
	}()

	return stop
}

// Shutdown cadvisor plugin
func (m *Monitor) Shutdown() {
	// tell cadvisor to stop
	if m.stop != nil {
		close(m.stop)
	}
}
