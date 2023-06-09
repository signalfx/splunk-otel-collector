package dpfilters

import (
	"github.com/signalfx/golib/v3/datapoint"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// AlwaysMatchFilter is a trivial filter that always matches datapoints
type AlwaysMatchFilter struct{}

// Matches just always returns true
func (m *AlwaysMatchFilter) Matches(*datapoint.Datapoint) bool {
	return true
}

func (m *AlwaysMatchFilter) MatchesMetric(pmetric.Metric) bool {
	return true
}
