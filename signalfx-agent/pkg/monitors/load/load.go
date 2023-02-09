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

package load

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/load"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	if runtime.GOOS != "windows" {
		monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
	}
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"true" acceptsEndpoints:"false"`
	PerCPU               bool `yaml:"perCPU" default:"false"`
}

// Monitor for load
type Monitor struct {
	Output types.Output
	cancel func()
	logger logrus.FieldLogger
}

// Configure is the main function of the monitor, it will report host metadata
// on a varied interval
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		avgLoad, err := load.Avg()
		if err != nil {
			m.logger.WithError(err).Error("Failed to get load statistics")
			return
		}

		divisor := 1.0
		if conf.PerCPU {
			divisor = float64(runtime.NumCPU())
		}

		m.Output.SendDatapoints([]*datapoint.Datapoint{
			datapoint.New(loadLongterm, nil, datapoint.NewFloatValue(avgLoad.Load15/divisor), datapoint.Gauge, time.Time{}),
			datapoint.New(loadMidterm, nil, datapoint.NewFloatValue(avgLoad.Load5/divisor), datapoint.Gauge, time.Time{}),
			datapoint.New(loadShortterm, nil, datapoint.NewFloatValue(avgLoad.Load1/divisor), datapoint.Gauge, time.Time{}),
		}...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
