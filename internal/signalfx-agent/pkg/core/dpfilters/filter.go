// Package dpfilters has logic describing the filtering of unwanted metrics.  Filters
// are configured from the agent configuration file and is intended to be passed
// into each monitor for use if it sends datapoints on its own.
package dpfilters

import (
	"github.com/signalfx/golib/v3/datapoint"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// DatapointFilter can be used to filter out datapoints
type DatapointFilter interface {
	// Matches takes a datapoint and returns whether it is matched by the
	// filter
	Matches(dp *datapoint.Datapoint) bool

	MatchesMetric(m pmetric.Metric) bool
}
