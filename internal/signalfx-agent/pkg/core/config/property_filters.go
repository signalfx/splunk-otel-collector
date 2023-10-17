package config

// PropertyFilterConfig describes a set of subtractive filters applied to properties
// used to create a PropertyFilter
// Deprecated: This struct is no longer used and doesn't participate in filtering. It will be removed in an upcoming release.
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
