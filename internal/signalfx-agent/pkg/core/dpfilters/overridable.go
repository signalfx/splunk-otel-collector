package dpfilters

import (
	"errors"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
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

func (f *overridableDatapointFilter) MatchesMetric(m pmetric.Metric) bool {
	var attributes pcommon.Map
	switch m.Type() {
	case pmetric.MetricTypeGauge:
		attributes = m.Gauge().DataPoints().At(0).Attributes()
	case pmetric.MetricTypeSummary:
		attributes = m.Summary().DataPoints().At(0).Attributes()
	case pmetric.MetricTypeSum:
		attributes = m.Sum().DataPoints().At(0).Attributes()
	case pmetric.MetricTypeHistogram:
		attributes = m.Histogram().DataPoints().At(0).Attributes()
	case pmetric.MetricTypeExponentialHistogram:
		attributes = m.ExponentialHistogram().DataPoints().At(0).Attributes()
	case pmetric.MetricTypeEmpty:
		attributes = pcommon.NewMap()
	}
	return (f.metricFilter == nil || f.metricFilter.Matches(m.Name())) &&
		(f.dimFilter == nil || f.dimFilter.MatchesAny(attributes.AsRaw()))
}
