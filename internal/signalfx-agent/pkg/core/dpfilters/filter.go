// Package dpfilters has logic describing the filtering of unwanted metrics.  Filters
// are configured from the agent configuration file and is intended to be passed
// into each monitor for use if it sends datapoints on its own.
package dpfilters

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/core/common/dpmeta"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

// DatapointFilter can be used to filter out datapoints
type DatapointFilter interface {
	// Matches takes a datapoint and returns whether it is matched by the
	// filter
	Matches(*datapoint.Datapoint) bool
}

// BasicDatapointFilter is designed to filter SignalFx datapoint objects.  It
// can filter based on the monitor type, dimensions, or the metric name.  It
// supports both static, globbed, and regex patterns for filter values. If
// dimensions are specified, they must all match for the datapoint to match. If
// multiple metric names are given, only one must match for the datapoint to
// match the filter since datapoints can only have one metric name.
type basicDatapointFilter struct {
	monitorType  string
	dimFilter    filter.StringMapFilter
	metricFilter filter.StringFilter
	negated      bool
}

// New returns a new filter with the given configuration
func New(monitorType string, metricNames []string, dimensions map[string][]string, negated bool) (DatapointFilter, error) {
	var dimFilter filter.StringMapFilter
	if len(dimensions) > 0 {
		var err error
		dimFilter, err = filter.NewStringMapFilter(dimensions)
		if err != nil {
			return nil, err
		}
	}

	var metricFilter filter.StringFilter
	if len(metricNames) > 0 {
		var err error
		metricFilter, err = filter.NewBasicStringFilter(metricNames)
		if err != nil {
			return nil, err
		}
	}

	return &basicDatapointFilter{
		monitorType:  monitorType,
		metricFilter: metricFilter,
		dimFilter:    dimFilter,
		negated:      negated,
	}, nil
}

// Matches tests a datapoint to see whether it is excluded by this filter.  In
// order to match on monitor type, the datapoint should have the "monitorType"
// key set in it's Meta field.
func (f *basicDatapointFilter) Matches(dp *datapoint.Datapoint) bool {
	if dpMonitorType, ok := dp.Meta[dpmeta.MonitorTypeMeta].(string); ok {
		if f.monitorType != "" && dpMonitorType != f.monitorType {
			return false
		}
	} else {
		// If we have a monitorType on the filter but none on the datapoint, it
		// can never match.
		if f.monitorType != "" {
			return false
		}
	}

	matched := (f.metricFilter == nil || f.metricFilter.Matches(dp.Metric)) &&
		(f.dimFilter == nil || f.dimFilter.Matches(dp.Dimensions))

	if f.negated {
		return !matched
	}
	return matched
}
