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

package propfilters

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// FilterSet is a collection of dimension filters, any one of which must match
// for a dimension property to be matched.
type FilterSet struct {
	Filters []DimensionFilter
}

// FilterDimension sends a *types.Dimension through each of the filters in the set
// and filters properties. All original properties will be returned if no filter matches
// , or a subset of the original if some are filtered, or nil if all are filtered.
func (fs *FilterSet) FilterDimension(dim *types.Dimension) *types.Dimension {
	filteredDim := &(*dim)
	for _, f := range fs.Filters {
		filteredDim = f.FilterDimension(filteredDim)
	}
	return filteredDim
}
