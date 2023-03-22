package config

import (
	"errors"
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
)

// MetricFilter describes a set of subtractive filters applied to datapoints
// right before they are sent.
type MetricFilter struct {
	// A map of dimension key/values to match against.  All key/values must
	// match a datapoint for it to be matched.  The map values can be either a
	// single string or a list of strings.
	Dimensions map[string]interface{} `yaml:"dimensions" default:"{}"`
	// A list of metric names to match against
	MetricNames []string `yaml:"metricNames"`
	// A single metric name to match against
	MetricName string `yaml:"metricName"`
	// (**Only applicable for the top level filters**) Limits this scope of the
	// filter to datapoints from a specific monitor. If specified, any
	// datapoints not from this monitor type will never match against this
	// filter.
	MonitorType string `yaml:"monitorType"`
	// (**Only applicable for the top level filters**) Negates the result of
	// the match so that it matches all datapoints that do NOT match the metric
	// name and dimension values given. This does not negate monitorType, if
	// given.
	Negated bool `yaml:"negated"`
}

// Normalize puts any singular metricName into the metricNames list and also
// returns a normalized dimension set.
func (mf *MetricFilter) Normalize() (map[string][]string, error) {
	if mf.MetricName != "" {
		mf.MetricNames = append(mf.MetricNames, mf.MetricName)
	}

	dimSet := map[string][]string{}
	for k, v := range mf.Dimensions {
		switch s := v.(type) {
		case []interface{}:
			newSet := []string{}
			for _, iv := range s {
				newSet = append(newSet, fmt.Sprintf("%v", iv))
			}
			dimSet[k] = newSet
		case string:
			dimSet[k] = []string{s}
		default:
			return nil, fmt.Errorf("%v should be either a string or string list", v)
		}
	}
	return dimSet, nil
}

// MakeFilter returns an actual filter instance from the config
func (mf *MetricFilter) MakeFilter() (dpfilters.DatapointFilter, error) {
	dimSet, err := mf.Normalize()
	if err != nil {
		return nil, err
	}

	return dpfilters.New(mf.MonitorType, mf.MetricNames, dimSet, mf.Negated)
}

func makeOldFilterSet(excludes []MetricFilter, includes []MetricFilter) (*dpfilters.FilterSet, error) {
	excludeSet := make([]dpfilters.DatapointFilter, 0)
	includeSet := make([]dpfilters.DatapointFilter, 0)
	mtes := make([]MetricFilter, 0, len(excludes))
	mtis := make([]MetricFilter, 0, len(includes))

	for _, mte := range excludes {
		mtes = AddOrMerge(mtes, mte)
	}

	for _, mti := range includes {
		mtis = AddOrMerge(mtis, mti)
	}

	for _, mte := range mtes {
		f, err := mte.MakeFilter()
		if err != nil {
			return nil, err
		}
		excludeSet = append(excludeSet, f)
	}

	for _, mti := range mtis {
		dimSet, err := mti.Normalize()
		if err != nil {
			return nil, err
		}

		f, err := dpfilters.New(mti.MonitorType, mti.MetricNames, dimSet, mti.Negated)
		if err != nil {
			return nil, err
		}

		includeSet = append(includeSet, f)
	}

	return &dpfilters.FilterSet{
		ExcludeFilters: excludeSet,
		IncludeFilters: includeSet,
	}, nil
}

// This should be the preferred filter set creator from now on.  It is much
// simpler to understand.
func makeNewFilterSet(excludes []MetricFilter) (*dpfilters.FilterSet, error) {
	var excludeSet []dpfilters.DatapointFilter
	for _, f := range excludes {
		if f.Negated {
			return nil, errors.New("new filters can't be negated")
		}
		dimSet, err := f.Normalize()
		if err != nil {
			return nil, err
		}

		dpf, err := dpfilters.NewOverridable(f.MetricNames, dimSet)
		if err != nil {
			return nil, err
		}

		excludeSet = append(excludeSet, dpf)
	}
	return &dpfilters.FilterSet{
		ExcludeFilters: excludeSet,
	}, nil
}

// AddOrMerge MetricFilter to list or merge with existing MetricFilter
func AddOrMerge(mtes []MetricFilter, mf2 MetricFilter) []MetricFilter {
	for i, mf1 := range mtes {
		if mf1.ShouldMerge(mf2) {
			mtes[i] = mf1.MergeWith(mf2)
			return mtes
		}
	}
	return append(mtes, mf2)
}

// MergeWith merges mf2's MetricFilter.MetricNames into receiver mf MetricFilter.MetricNames
func (mf *MetricFilter) MergeWith(mf2 MetricFilter) MetricFilter {
	if mf2.MetricName != "" {
		mf2.MetricNames = append(mf2.MetricNames, mf2.MetricName)
	}
	mf.MetricNames = append(mf.MetricNames, mf2.MetricNames...)
	return *mf
}

// ShouldMerge checks if mf2 MetricFilter should be merged into receiver mf MetricFilter
// Filters with same monitorType, negation, and dimensions should be merged
func (mf *MetricFilter) ShouldMerge(mf2 MetricFilter) bool {
	if mf.MonitorType != mf2.MonitorType {
		return false
	}
	if mf.Negated != mf2.Negated {
		return false
	}
	if len(mf.Dimensions) != len(mf2.Dimensions) {
		return false
	}
	// Ensure no differing dimension values
	for k, v := range mf.Dimensions {
		if mf2.Dimensions[k] != v {
			return false
		}
	}
	return true
}
