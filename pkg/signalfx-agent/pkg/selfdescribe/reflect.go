package selfdescribe

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Only works if there is an explicit "yaml" struct tag
func getYAMLName(f reflect.StructField) string {
	yamlTag := f.Tag.Get("yaml")
	return strings.SplitN(yamlTag, ",", 2)[0]
}

func isInlinedYAML(f reflect.StructField) bool {
	return strings.Contains(f.Tag.Get("yaml"), ",inline")
}

// Assumes monitors are using the defaults package
func getDefault(f reflect.StructField) interface{} {
	if getRequired(f) {
		return nil
	}
	if f.Tag.Get("noDefault") == strconv.FormatBool(true) {
		return nil
	}

	defTag := f.Tag.Get("default")
	if defTag != "" {
		// These are essentially just noop defaults so don't return them
		if defTag == "{}" || defTag == "[]" {
			return ""
		}
		if strings.HasPrefix(defTag, "{") || strings.HasPrefix(defTag, "[") || defTag == strconv.FormatBool(true) || defTag == strconv.FormatBool(false) {
			var out interface{}
			err := json.Unmarshal([]byte(defTag), &out)
			if err != nil {
				log.WithError(err).Warnf("Could not unmarshal default value `%s` for field %s", defTag, f.Name)
				return defTag
			}
			return out
		}
		if asInt, err := strconv.Atoi(defTag); err == nil {
			return asInt
		}
		return defTag
	}
	if f.Type.Kind() == reflect.Ptr {
		if f.Type.Elem().Kind() == reflect.Bool {
			return "false"
		}
		return nil
	}
	if f.Type.Kind() != reflect.Struct {
		return reflect.Zero(f.Type).Interface()
	}
	return nil
}

// Assumes that monitors are using the validate package to do validation
func getRequired(f reflect.StructField) bool {
	validate := f.Tag.Get("validate")
	for _, v := range strings.Split(validate, ",") {
		if v == "required" {
			return true
		}
	}
	return false
}

// The kind with any pointer removed
func indirectKind(t reflect.Type) reflect.Kind {
	if t == reflect.TypeOf(yaml.MapSlice(nil)) {
		return reflect.Map
	}

	kind := t.Kind()
	if kind == reflect.Ptr {
		return t.Elem().Kind()
	}
	return kind
}

// The type with any pointers removed
func indirectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}
