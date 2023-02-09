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

package correlations

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
)

// InternalMetrics returns datapoints that describe the current state of the
// dimension update client
func (cc *Client) InternalMetrics() []*datapoint.Datapoint {
	dps := []*datapoint.Datapoint{
		sfxclient.CumulativeP("sfxagent.correlation_updates_invalid", nil, &cc.TotalInvalidDimensions),
		// All 4xx HTTP responses that are not retried except 404 (which is retried)
		sfxclient.CumulativeP("sfxagent.correlation_updates_client_errors", nil, &cc.TotalClientError4xxResponses),
		sfxclient.CumulativeP("sfxagent.correlation_updates_retries", nil, &cc.TotalRetriedUpdates),
	}
	return append(dps, cc.requestSender.InternalMetrics()...)
}
