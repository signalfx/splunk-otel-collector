package dpfilters

import (
	"github.com/signalfx/golib/v3/datapoint"
	"go.opentelemetry.io/collector/pdata/pmetric"
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

// MatchesMetric returns the opposite of what the original filter would have
// returned.
func (n *NegatedDatapointFilter) MatchesMetric(m pmetric.Metric) bool {
	return !n.DatapointFilter.MatchesMetric(m)
}

// Negate returns the supplied filter negated such Matches returns the
// opposite.
func Negate(f DatapointFilter) DatapointFilter {
	return &NegatedDatapointFilter{
		DatapointFilter: f,
	}
}
