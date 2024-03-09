package types

import "fmt"

// Dimension represents a SignalFx dimension object and its associated
// properties and tags.
type Dimension struct {
	Properties        map[string]string
	Tags              map[string]bool
	Name              string
	Value             string
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
