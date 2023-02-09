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

	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

type runner struct {
	ctx                   context.Context
	log                   logrus.FieldLogger
	conf                  *model.Config
	vsm                   *vSphereMonitor
	vsphereReloadInterval int // seconds
}

func newRunner(ctx context.Context, log logrus.FieldLogger, conf *model.Config, monitor *Monitor) runner {
	vsphereReloadInterval := int(conf.InventoryRefreshInterval.AsDuration().Seconds())
	vsm := newVsphereMonitor(conf, log, monitor.Output.SendDatapoints)
	return runner{
		ctx:                   ctx,
		log:                   log,
		conf:                  conf,
		vsphereReloadInterval: vsphereReloadInterval,
		vsm:                   vsm,
	}
}

// Called periodically. This is the entry point to the vSphere monitor.
func (r *runner) run() {
	err := r.vsm.firstTimeSetup(r.ctx)
	if err != nil {
		r.log.WithError(err).Error("firstTimeSetup failed")
		return
	}
	r.vsm.generateDatapoints()
	if r.vsm.isTimeForVSphereInfoReload(r.vsphereReloadInterval) {
		r.vsm.reloadVSphereInfo()
	}
}
