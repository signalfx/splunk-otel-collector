package utils

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// ConvertToMapViaYAML takes a struct and converts it to map[string]interface{}
// by marshalling it to yaml and back to a map.  This will return nil if the
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

var yamlLineNumberRE = regexp.MustCompile(`line (\d+): `)

// ParseLineNumberFromYAMLError takes an error message nested in yaml.TypeError
// and returns a line number if indicated in the error message.  This is pretty
// hacky but is the only way to actually get at the line number in the standard
// yaml package.
func ParseLineNumberFromYAMLError(e string) (int, bool) {
	match := yamlLineNumberRE.FindStringSubmatch(e)
	if len(match) > 0 {
		asInt, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, false
		}
		return asInt, true
	}
	return 0, false
}

// YAMLErrorWithContext will wrap an given YAML error that contains a clear
// message about where the error occurred in parsing the YAML.
func YAMLErrorWithContext(content []byte, err error) error {
	var out string

	typeErr, ok := err.(*yaml.TypeError)
	var errs []string
	if ok {
		errs = typeErr.Errors
	} else {
		errs = []string{err.Error()}
	}
	// Provide some context about where the parse error occurred
	for _, e := range errs {
		line, valid := ParseLineNumberFromYAMLError(e)
		if !valid {
			return err
		}
		context := string(content)
		lines := strings.Split(context, "\n")
		context = strings.Join(lines[int(math.Max(float64(line-5), 0)):line], "\n")
		context += "\n^^^^^^^\n"
		context += strings.Join(lines[line:int(math.Min(float64(line+5), float64(len(lines))))], "\n")
		out += fmt.Sprintf(
			"Could not unmarshal config file:\n\n%s\n\n%s\n",
			context,
			yamlLineNumberRE.ReplaceAllString(err.Error(), ""))
	}
	return errors.New(out)
}
