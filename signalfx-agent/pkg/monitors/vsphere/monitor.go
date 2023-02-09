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

package vsphere

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

type Monitor struct {
	Output types.Output
	cancel func()
	logger logrus.FieldLogger
}

func init() {
	monitors.Register(
		&monitorMetadata,
		func() interface{} { return &Monitor{} },
		&model.Config{},
	)
}

func (m *Monitor) Configure(conf *model.Config) error {
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})
	r := newRunner(ctx, m.logger, conf, m)
	// 20 seconds is the fixed, real-time metrics interval for vsphere/esxi
	utils.RunOnInterval(ctx, r.run, model.RealtimeMetricsInterval*time.Second)
	return nil
}

func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
