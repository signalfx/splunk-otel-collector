package config

import (
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/signalfx-agent/pkg/core/propfilters"
)

// PropertyFilterConfig describes a set of subtractive filters applied to properties
// used to create a PropertyFilter
type PropertyFilterConfig struct {
	// A single property name to match
	PropertyName *string `yaml:"propertyName" default:"*"`
	// A property value to match
	PropertyValue *string `yaml:"propertyValue" default:"*"`
	// A dimension name to match
	DimensionName *string `yaml:"dimensionName" default:"*"`
	// A dimension value to match
	DimensionValue *string `yaml:"dimensionValue" default:"*"`
}

// MakePropertyFilter returns an actual filter instance from the config
func (pfc *PropertyFilterConfig) MakePropertyFilter() (propfilters.DimensionFilter, error) {
	pfc.PropertyName = setDefault(pfc.PropertyName)
	pfc.PropertyValue = setDefault(pfc.PropertyValue)
	pfc.DimensionName = setDefault(pfc.DimensionName)
	pfc.DimensionValue = setDefault(pfc.DimensionValue)

	propertyNames := []string{*pfc.PropertyName}
	propertyValues := []string{*pfc.PropertyValue}
	dimensionNames := []string{*pfc.DimensionName}
	dimensionValues := []string{*pfc.DimensionValue}
	return propfilters.New(propertyNames, propertyValues,
		dimensionNames, dimensionValues)
}

func setDefault(s *string) *string {
	if s == nil {
		return pointer.String("*")
	}
	return s
}

func makePropertyFilterSet(conf []PropertyFilterConfig) (*propfilters.FilterSet, error) {
	fs := make([]propfilters.DimensionFilter, 0)
	for _, pte := range conf {
		f, err := pte.MakePropertyFilter()
		if err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}

	return &propfilters.FilterSet{
		Filters: fs,
	}, nil
}
