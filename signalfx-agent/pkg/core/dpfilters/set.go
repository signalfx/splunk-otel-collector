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

import (
	"github.com/signalfx/golib/v3/datapoint"
)

// FilterSet is a collection of datapont filters, any one of which must match
// for a datapoint to be matched.
type FilterSet struct {
	ExcludeFilters []DatapointFilter
	// IncludeFilters are optional and serve as a top-priority list of matchers
	// that will cause a datapoint to always be sent
	IncludeFilters []DatapointFilter
}

var _ DatapointFilter = &FilterSet{}

// Matches sends a datapoint through each of the filters in the set and returns
// true if at least one of them matches the datapoint.
func (fs *FilterSet) Matches(dp *datapoint.Datapoint) bool {
	for _, ex := range fs.ExcludeFilters {
		if ex.Matches(dp) {
			// If we match an exclusionary filter, run through each inclusion
			// filter and see if anything includes the metrics.
			for _, incl := range fs.IncludeFilters {
				if incl.Matches(dp) {
					return false
				}
			}
			return true
		}
	}
	return false
}
