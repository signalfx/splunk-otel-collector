package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/signalfx/signalfx-agent/pkg/utils"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// ToString converts a config struct to a pseudo-yaml text outut.  If a struct
// field has the 'neverLog' tag, its value will be replaced by asterisks, or
// completely omitted if the tag value is 'omit'.
func ToString(conf interface{}) string {
	if conf == nil {
		return ""
	}

	var out string
	confValue := reflect.Indirect(reflect.ValueOf(conf))
	if !confValue.IsValid() {
		return ""
	}

	isSlice := confValue.Type().Kind() == reflect.Slice
	if isSlice {
		out := ""
		for j := 0; j < confValue.Len(); j++ {
			s := utils.IndentLines(ToString(confValue.Index(j).Interface()), 2)
			if len(s) > 0 {
				out += strings.Trim("-"+s[1:], "\n") + "\n"
			}
		}
		return strings.Trim(out, "\n")
	}

	if !utils.IsStructOrPointerToStruct(confValue.Type()) {
		yamlBytes, err := yaml.Marshal(conf)
		if err != nil {
			log.WithError(err).Error("Could not marshal yaml for diagnostic text conversion")
			return ""
		}
		asYaml := string(yamlBytes)
		// Output empty maps as blank instead of a pair of braces
		if confValue.Type().Kind() == reflect.Map && strings.HasPrefix(asYaml, "{}") {
			return ""
		}

		return strings.TrimSuffix(asYaml, "\n")
	}

	confStruct := confValue.Type()

	for i := 0; i < confStruct.NumField(); i++ {
		field := confStruct.Field(i)

		// PkgPath is empty only for exported fields, so it it's non-empty the
		// field is private
		if field.PkgPath != "" {
			continue
		}

		fieldName := utils.YAMLNameOfField(field)

		if fieldName == "" && !field.Anonymous {
			continue
		}

		neverLogVal, neverLogPresent := field.Tag.Lookup("neverLog")
		var val string
		if neverLogPresent {
			if neverLogVal == "omit" {
				continue
			}
			if v, ok := confValue.Field(i).Interface().(string); ok && v == "" {
				val = ""
			} else {
				val = "***************"
			}
		} else {
			val = ToString(confValue.Field(i).Interface())
		}

		// Flatten embedded struct's representation
		if field.Anonymous {
			out += val
			continue
		}

		separator := " "
		if len(val) > 0 && (strings.Contains(val, "\n") || field.Type.Kind() == reflect.Map) {
			separator = "\n"
			val = utils.IndentLines(val, 2)
		}

		out += fmt.Sprintf("%s:%s%s\n", fieldName, separator, strings.TrimSuffix(val, "\n"))
	}
	return out
}
