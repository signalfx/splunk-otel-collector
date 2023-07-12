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
