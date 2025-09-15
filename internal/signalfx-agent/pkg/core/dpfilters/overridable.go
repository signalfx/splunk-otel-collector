package dpfilters

import (
	"errors"

	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

type overridableDatapointFilter struct {
	dimFilter    filter.StringMapFilter
	metricFilter filter.StringFilter
}

// NewOverridable returns a new overridable filter with the given configuration
func NewOverridable(metricNames []string, dimensions map[string][]string) (DatapointFilter, error) {
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
		metricFilter, err = filter.NewOverridableStringFilter(metricNames)
		if err != nil {
			return nil, err
		}
	}

	if metricFilter == nil && dimFilter == nil {
		return nil, errors.New("metric filter must have at least one metric or dimension defined on it")
	}

	return &overridableDatapointFilter{
		metricFilter: metricFilter,
		dimFilter:    dimFilter,
	}, nil
}

// Matches tests a datapoint to see whether it is excluded by this filter.
func (f *overridableDatapointFilter) Matches(dp *datapoint.Datapoint) bool {
	return (f.metricFilter == nil || f.metricFilter.Matches(dp.Metric)) &&
		(f.dimFilter == nil || f.dimFilter.Matches(dp.Dimensions))
}

func (f *overridableDatapointFilter) MatchesMetricDataPoint(metricName string, dimensions pcommon.Map) bool {
	if f.metricFilter != nil && !f.metricFilter.Matches(metricName) {
		return false
	}
	if f.dimFilter != nil {
		return f.dimFilter.MatchesMap(dimensions)
	}
	return true
}
