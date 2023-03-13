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
