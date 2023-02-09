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

package forwarder

import (
	"context"
	"time"

	"github.com/pkg/errors"
	goliblog "github.com/signalfx/golib/v3/log"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false" singleInstance:"true"`
	// The host:port on which to listen for datapoints.  The listening server
	// accepts datapoints on the same HTTP path that ingest/gateway accepts
	// them (e.g. `/v2/datapoint`, `/v1/trace`).  Requests to other paths will
	// return 404s.
	ListenAddress string `yaml:"listenAddress" default:"127.0.0.1:9080"`
	// HTTP timeout duration for both read and writes. This should be a
	// duration string that is accepted by https://golang.org/pkg/time/#ParseDuration
	ServerTimeout timeutil.Duration `yaml:"serverTimeout" default:"5s"`
	// Whether to emit internal metrics about the HTTP listener
	SendInternalMetrics *bool `yaml:"sendInternalMetrics" default:"false"`
}

// Monitor that accepts and forwards SignalFx data
type Monitor struct {
	Output      types.Output
	cancel      context.CancelFunc
	logger      *utils.ThrottledLogger
	golibLogger goliblog.Logger
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = utils.NewThrottledLogger(logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID}), 30*time.Second)
	m.golibLogger = &utils.LogrusGolibShim{FieldLogger: m.logger.FieldLogger}

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	sink := &outputSink{Output: m.Output}
	listenerMetrics, err := m.startListening(ctx, conf.ListenAddress, conf.ServerTimeout.AsDuration(), sink)
	if err != nil {
		return errors.WithMessage(err, "could not start forwarder listener")
	}

	if *conf.SendInternalMetrics {
		utils.RunOnInterval(ctx, func() {
			m.Output.SendDatapoints(listenerMetrics.Datapoints()...)
		}, time.Duration(conf.IntervalSeconds)*time.Second)
	}

	return nil
}

// Shutdown stops the forwarder and correlation MTSs
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
