package utils

import (
	"reflect"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// ConvertToMapViaYAML takes a struct and converts it to map[string]interface{}
// by marshaling it to yaml and back to a map.  This will return nil if the
// conversion was not successful.
func ConvertToMapViaYAML(obj interface{}) (map[string]interface{}, error) {
	str, err := yaml.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var newMap map[string]interface{}
	if err := yaml.Unmarshal(str, &newMap); err != nil {
		return nil, err
	}

	return newMap, nil
}

// YAMLNameOfField returns the YAML key that is used for the given struct
// field.  It does this by actually serializing the field and parsing the
// output string.  If the field has no key (e.g. if the `yaml:"-"` tag is set,
// this will return an empty string.
func YAMLNameOfField(field reflect.StructField) string {
	if strings.HasPrefix(field.Tag.Get("yaml"), ",inline") {
		return ""
	}
	tmp := reflect.New(reflect.StructOf([]reflect.StructField{field})).Elem()
	asYaml, _ := yaml.Marshal(tmp.Interface())
	parts := strings.SplitN(string(asYaml), ":", 2)
	if parts[0] == string(asYaml) {
		return ""
	}
	return parts[0]
}
