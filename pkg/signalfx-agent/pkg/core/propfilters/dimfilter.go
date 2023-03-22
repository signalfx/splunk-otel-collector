// Package propfilters has logic describing the filtering of unwanted properties.  Filters
// are configured from the agent configuration file and is intended to be passed
// into each monitor for use if it sends properties on its own.
package propfilters

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

// DimensionFilter is designed to filter SignalFx property objects.  It
// can filter based on property names, property values, dimension names,
// and dimension values. It also supports both static, globbed, and regex
// patterns for filter values.
// All properties matched will be filtered.
type DimensionFilter interface {
	// Filters out properties from Dimension object
	FilterDimension(dim *types.Dimension) *types.Dimension
	MatchesDimension(name string, value string) bool
	FilterProperties(properties map[string]string) map[string]string
}

// basicDimensionFilter is an implementation of DimensionFilter
type basicDimensionFilter struct {
	propertyNameFilter   filter.StringFilter
	propertyValueFilter  filter.StringFilter
	dimensionNameFilter  filter.StringFilter
	dimensionValueFilter filter.StringFilter
}

// New returns a new filter with the given configuration
func New(propertyNames []string, propertyValues []string, dimensionNames []string,
	dimensionValues []string) (DimensionFilter, error) {

	var propertyNameFilter filter.StringFilter
	if len(propertyNames) > 0 {
		var err error
		propertyNameFilter, err = filter.NewBasicStringFilter(propertyNames)
		if err != nil {
			return nil, err
		}
	}

	var propertyValueFilter filter.StringFilter
	if len(propertyValues) > 0 {
		var err error
		propertyValueFilter, err = filter.NewBasicStringFilter(propertyValues)
		if err != nil {
			return nil, err
		}
	}

	var dimensionNameFilter filter.StringFilter
	if len(dimensionNames) > 0 {
		var err error
		dimensionNameFilter, err = filter.NewBasicStringFilter(dimensionNames)
		if err != nil {
			return nil, err
		}
	}

	var dimensionValueFilter filter.StringFilter
	if len(dimensionValues) > 0 {
		var err error
		dimensionValueFilter, err = filter.NewBasicStringFilter(dimensionValues)
		if err != nil {
			return nil, err
		}
	}

	return &basicDimensionFilter{
		propertyNameFilter:   propertyNameFilter,
		propertyValueFilter:  propertyValueFilter,
		dimensionNameFilter:  dimensionNameFilter,
		dimensionValueFilter: dimensionValueFilter,
	}, nil
}

// Filter applies the filter to the given Dimension and returns a new
// filtered Dimensions
func (f *basicDimensionFilter) FilterDimension(dim *types.Dimension) *types.Dimension {
	if dim == nil {
		return nil
	}
	var filteredProperties map[string]string

	if f.MatchesDimension(dim.Name, dim.Value) {
		filteredProperties = f.FilterProperties(dim.Properties)
	} else {
		filteredProperties = dim.Properties
	}

	// If the filtering has removed all properties, then don't consider this
	// dimension at all.
	if len(filteredProperties) == 0 && len(dim.Tags) == 0 {
		return nil
	}

	return &types.Dimension{
		Name:       dim.Name,
		Value:      dim.Value,
		Properties: filteredProperties,
		Tags:       dim.Tags,
	}
}

// FilterProperties uses the propertyNameFilter and propertyValueFilter given to
// filter out properties in a map if either the name or value matches
func (f *basicDimensionFilter) FilterProperties(properties map[string]string) map[string]string {
	filteredProperties := make(map[string]string, len(properties))
	for propName, propValue := range properties {
		if (!f.propertyNameFilter.Matches(propName)) ||
			(!f.propertyValueFilter.Matches(propValue)) {
			filteredProperties[propName] = propValue
		}
	}

	return filteredProperties
}

// MatchesDimension checks both dimensionNameFilter and dimensionValueFilter
// and if both match, returns true
func (f *basicDimensionFilter) MatchesDimension(name string, value string) bool {
	return f.dimensionNameFilter.Matches(name) && f.dimensionValueFilter.Matches(value)
}
