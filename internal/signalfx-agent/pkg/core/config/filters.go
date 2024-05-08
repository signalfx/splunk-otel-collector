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
