package dpfilters

import (
	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"go.opentelemetry.io/collector/pdata/pcommon"
)

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

// MatchesMetricDataPoint returns the opposite of what the original filter would have
// returned.
func (n *NegatedDatapointFilter) MatchesMetricDataPoint(metricName string, dimensions pcommon.Map) bool {
	return !n.DatapointFilter.MatchesMetricDataPoint(metricName, dimensions)
}

// Negate returns the supplied filter negated such Matches returns the
// opposite.
func Negate(f DatapointFilter) DatapointFilter {
	return &NegatedDatapointFilter{
		DatapointFilter: f,
	}
}
