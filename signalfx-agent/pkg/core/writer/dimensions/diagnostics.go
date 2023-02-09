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

package dimensions

import (
	"sync/atomic"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
)

// InternalMetrics returns datapoints that describe the current state of the
// dimension update client
func (dc *DimensionClient) InternalMetrics() []*datapoint.Datapoint {
	dps := []*datapoint.Datapoint{
		sfxclient.Gauge("sfxagent.dim_updates_currently_delayed", nil, atomic.LoadInt64(&dc.DimensionsCurrentlyDelayed)),
		sfxclient.CumulativeP("sfxagent.dim_updates_dropped", nil, &dc.TotalDimensionsDropped),
		sfxclient.CumulativeP("sfxagent.dim_updates_invalid", nil, &dc.TotalInvalidDimensions),
		sfxclient.CumulativeP("sfxagent.dim_updates_flappy_total", nil, &dc.TotalFlappyUpdates),
		sfxclient.CumulativeP("sfxagent.dim_updates_duplicates", nil, &dc.TotalDuplicates),
		// All 4xx HTTP responses that are not retried except 404 (which is retried)
		sfxclient.CumulativeP("sfxagent.dim_updates_client_errors", nil, &dc.TotalClientError4xxResponses),
		sfxclient.CumulativeP("sfxagent.dim_updates_retries", nil, &dc.TotalRetriedUpdates),
		sfxclient.Cumulative("sfxagent.dim_updates_deduplicator_size", nil, int64(dc.deduplicator.history.Len())),
	}
	return append(dps, dc.requestSender.InternalMetrics()...)
}
