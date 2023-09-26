package utils

import "reflect"

// GetStructFieldNames returns a slice with the names of all of the fields in
// the struct `s`.  This will panic if `s` is not a struct.
func GetStructFieldNames(s interface{}) []string {
	v := reflect.Indirect(reflect.ValueOf(s))
	out := []string{}

	for i := 0; i < v.Type().NumField(); i++ {
		out = append(out, v.Type().Field(i).Name)
	}

	return out
}

func FindFirstFieldOfType(st interface{}, typ reflect.Type) reflect.Value {
	val := reflect.Indirect(reflect.ValueOf(st))
	for i := 0; i < val.NumField(); i++ {
		if val.Field(i).Type() == typ {
			return val.Field(i)
		}
	}
	return reflect.ValueOf(nil)
}
