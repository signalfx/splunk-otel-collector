// Package dpfilters has logic describing the filtering of unwanted metrics.  Filters
// are configured from the agent configuration file and is intended to be passed
// into each monitor for use if it sends datapoints on its own.
package dpfilters

import (
	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// DatapointFilter can be used to filter out datapoints
type DatapointFilter interface {
	// Matches takes a datapoint and returns whether it is matched by the
	// filter
	Matches(dp *datapoint.Datapoint) bool

	MatchesMetricDataPoint(metricName string, dimensions pcommon.Map) bool
}
