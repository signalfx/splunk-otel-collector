package selfdescribe

type structMetadata struct {
	Name    string          `json:"name"`
	Doc     string          `json:"doc"`
	Package string          `json:"package"`
	Fields  []fieldMetadata `json:"fields"`
}

type fieldMetadata struct {
	YAMLName    string      `json:"yamlName"`
	Doc         string      `json:"doc"`
	Default     interface{} `json:"default"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
	ElementKind string      `json:"elementKind"`
	// Element is the metadata for the element type of a slice or the value
	// type of a map if they are structs.
	ElementStruct *structMetadata `json:"elementStruct,omitempty"`
}
