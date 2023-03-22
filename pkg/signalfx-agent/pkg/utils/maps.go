// Package utils hold miscellaneous utility functions
package utils

import (
	"fmt"
	"sort"

	"github.com/iancoleman/strcase"
)

// DuplicateInterfaceMapKeysAsCamelCase takes a map[string]interface{} and camel cases the keys
func DuplicateInterfaceMapKeysAsCamelCase(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range m {
		out[k] = v
		out[strcase.ToLowerCamel(k)] = v
	}
	return out
}

// MergeStringMaps merges n maps with a later map's keys overriding earlier maps
func MergeStringMaps(maps ...map[string]string) map[string]string {
	ret := map[string]string{}

	for _, m := range maps {
		for k, v := range m {
			ret[k] = v
		}
	}

	return ret
}

// RemoveEmptyMapValues will strip a map of any key/value pairs for which the
// value is the empty string.
func RemoveEmptyMapValues(m map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range m {
		if v != "" {
			out[k] = v
		}
	}
	return out
}

// StringMapToInterfaceMap converts a map[string]string to a map[string]interface{}.
func StringMapToInterfaceMap(m map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// MergeInterfaceMaps merges any number of map[string]interface{} with a later
// map's keys overriding earlier maps.  Nil values do not override earlier
// values.
func MergeInterfaceMaps(maps ...map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{}

	for i := range maps {
		for k, v := range maps[i] {
			if ret[k] == nil || v != nil {
				ret[k] = v
			}
		}
	}

	return ret
}

// CloneStringMap makes a shallow copy of a map[string]string
func CloneStringMap(m map[string]string) map[string]string {
	m2 := make(map[string]string, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

// CloneInterfaceMap makes a shallow copy of a map[string]interface{}
func CloneInterfaceMap(m map[string]interface{}) map[string]interface{} {
	m2 := make(map[string]interface{}, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

// CloneFullInterfaceMap makes a shallow copy of a map[interface{}]interface{}
func CloneFullInterfaceMap(m map[interface{}]interface{}) map[interface{}]interface{} {
	m2 := make(map[interface{}]interface{}, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

// CloneAndFilterStringMapWithFunc clones a string map and only includes
// key/value pairs for which the filter function returns true
func CloneAndFilterStringMapWithFunc(in map[string]string, filter func(string, string) bool) (out map[string]string) {
	out = make(map[string]string, len(in))
	for k, v := range in {
		if filter(k, v) {
			out[k] = v
		}
	}
	return out
}

// CloneAndExcludeStringMapByKey clones a string map excluding the specified keys
func CloneAndExcludeStringMapByKey(in map[string]string, exclude map[string]bool) map[string]string {
	if len(exclude) == 0 {
		return CloneStringMap(in)
	}
	var out = make(map[string]string, len(in))
	for k, v := range in {
		if exclude[k] {
			out[k] = v
		}
	}
	return out
}

// InterfaceMapToStringMap converts a map[interface{}]interface{} to a
// map[string]string.  Keys and values will be converted with fmt.Sprintf so
// the original key/values don't have to be strings.
func InterfaceMapToStringMap(m map[interface{}]interface{}) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
	}
	return out
}

// SortMapKeys returns a slice of all of the keys of a map sorted
// alphabetically ascending.
func SortMapKeys(m map[string]interface{}) []string {
	if len(m) == 0 {
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// StringInterfaceMapToAllInterfaceMap converts a map[string]interface{} to a
// map[interface{}]interface{}
func StringInterfaceMapToAllInterfaceMap(in map[string]interface{}) map[interface{}]interface{} {
	out := make(map[interface{}]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// FormatStringMapCompact formats a string map as a string in a compact form
func FormatStringMapCompact(in map[string]string) string {
	out := "{"

	for k, v := range in {
		out += k + ": " + v + ", "
	}

	if len(in) > 0 {
		// Strip last comma
		out = out[:len(out)-2]
	}

	return out + "}"
}

func StringInterfaceMapToStringMap(in map[string]interface{}) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		if strVal, ok := v.(string); ok {
			out[k] = strVal
		} else if stringer, ok := v.(fmt.Stringer); ok {
			out[k] = stringer.String()
		} else {
			out[k] = fmt.Sprintf("%v", v)
		}
	}
	return out
}
