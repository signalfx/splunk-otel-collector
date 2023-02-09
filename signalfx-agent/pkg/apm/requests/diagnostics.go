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

package requests

import (
	"sync/atomic"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
)

// InternalMetrics returns datapoints that describe the current state of the
// dimension update client
func (rs *ReqSender) InternalMetrics() []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		sfxclient.CumulativeP("sfxagent.dim_updates_started", map[string]string{"client": rs.clientName}, &rs.TotalRequestsStarted),
		sfxclient.CumulativeP("sfxagent.dim_updates_completed", map[string]string{"client": rs.clientName}, &rs.TotalRequestsCompleted),
		sfxclient.CumulativeP("sfxagent.dim_updates_failed", map[string]string{"client": rs.clientName}, &rs.TotalRequestsFailed),
		sfxclient.Gauge("sfxagent.dim_request_senders", map[string]string{"client": rs.clientName}, atomic.LoadInt64(&rs.RunningWorkers)),
	}
}
