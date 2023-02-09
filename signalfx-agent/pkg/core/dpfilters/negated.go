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

package dpfilters

import "github.com/signalfx/golib/v3/datapoint"

// NegatedDatapointFilter is a datapoint filter whose Matches method is made
// opposite
type NegatedDatapointFilter struct {
	DatapointFilter
}

// Matches returns the opposite of what the original filter would have
// returned.
func (n *NegatedDatapointFilter) Matches(dp *datapoint.Datapoint) bool {
	return !n.DatapointFilter.Matches(dp)
}

// Negate returns the supplied filter negated such Matches returns the
// opposite.
func Negate(f DatapointFilter) DatapointFilter {
	return &NegatedDatapointFilter{
		DatapointFilter: f,
	}
}
