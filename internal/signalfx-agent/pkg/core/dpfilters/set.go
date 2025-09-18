package dpfilters

import (
	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// FilterSet is a collection of datapont filters, any one of which must match
// for a datapoint to be matched.
type FilterSet struct {
	ExcludeFilters []DatapointFilter
}

var _ DatapointFilter = &FilterSet{}

// Matches sends a datapoint through each of the filters in the set and returns
// true if at least one of them matches the datapoint.
func (fs *FilterSet) Matches(dp *datapoint.Datapoint) bool {
	for _, ex := range fs.ExcludeFilters {
		if ex.Matches(dp) {
			return true
		}
	}
	return false
}

// MatchesMetricDataPoint sends a datapoint through each of the filters in the set and returns
// true if at least one of them matches the datapoint.
func (fs *FilterSet) MatchesMetricDataPoint(metricName string, dimensions pcommon.Map) bool {
	for _, ex := range fs.ExcludeFilters {
		if ex.MatchesMetricDataPoint(metricName, dimensions) {
			return true
		}
	}
	return false
}
