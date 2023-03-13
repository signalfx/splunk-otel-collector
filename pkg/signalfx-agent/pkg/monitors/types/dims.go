package types

import "fmt"

// Dimension represents a SignalFx dimension object and its associated
// properties and tags.
type Dimension struct {
	// Name of the dimension
	Name string
	// Value of the dimension
	Value string
	// Properties to be set on the dimension.  If the value of any key is blank
	// string and MergeIntoExisting is true, that property will be deleted from
	// the dimension.
	Properties map[string]string
	// Tags to apply to the dimension value.  The maps keys are the tag names
	// and the value should normally be set to `true`.  If the value of the map
	// entry for a particular tag is set to false and `MergeIntoExisting` is
	// true, that tag will be deleted on the backend.
	Tags map[string]bool
	// Whether to do a query of existing dimension properties/tags and merge
	// the given values into those before updating, or whether to entirely
	// replace the set of properties/tags.
	MergeIntoExisting bool
}

func (d *Dimension) String() string {
	return fmt.Sprintf("{name: %q; value: %q; props: %v; tags: %v; mergeIntoExisting: %v}", d.Name, d.Value, d.Properties, d.Tags, d.MergeIntoExisting)
}

// DimensionKey is what uniquely identifies a dimension, its name and value
// together.
type DimensionKey struct {
	Name  string
	Value string
}

func (dk DimensionKey) String() string {
	return fmt.Sprintf("[%s/%s]", dk.Name, dk.Value)
}

func (d *Dimension) Key() DimensionKey {
	return DimensionKey{
		Name:  d.Name,
		Value: d.Value,
	}
}

// Copy creates a copy of the given Dimension object
func (d *Dimension) Copy() *Dimension {
	clonedProperties := make(map[string]string)
	for k, v := range d.Properties {
		clonedProperties[k] = v
	}

	clonedTags := make(map[string]bool)
	for k, v := range d.Tags {
		clonedTags[k] = v
	}

	return &Dimension{
		Name:              d.Name,
		Value:             d.Value,
		Properties:        clonedProperties,
		Tags:              clonedTags,
		MergeIntoExisting: d.MergeIntoExisting,
	}
}
