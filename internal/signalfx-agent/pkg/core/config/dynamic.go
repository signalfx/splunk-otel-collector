package config

import (
	"fmt"
	"reflect"
)

// CallConfigure will call the Configure method on an observer or monitor with
// a `conf` object, typed to the correct type.  This allows monitors/observers
// to set the type of the config object to their own config and not have to
// worry about casting or converting.
func CallConfigure(instance, conf interface{}) error {
	instanceVal := reflect.ValueOf(instance)
	_type := instanceVal.Type().PkgPath()

	confVal := reflect.ValueOf(conf)

	method := instanceVal.MethodByName("Configure")
	if !method.IsValid() {
		return fmt.Errorf("no Configure method found for type %s", _type)
	}

	if method.Type().NumIn() != 1 {
		return fmt.Errorf("configure method of %s should take exactly one argument that matches "+
			"the type of the config template provided in the Register function! It has %d arguments.",
			_type, method.Type().NumIn())
	}

	errorIntf := reflect.TypeOf((*error)(nil)).Elem()
	if method.Type().NumOut() != 1 || !method.Type().Out(0).Implements(errorIntf) {
		return fmt.Errorf("configure method for type %s should return an error", _type)
	}

	ret := method.Call([]reflect.Value{confVal})[0]
	if ret.IsNil() {
		return nil
	}
	return ret.Interface().(error)
}
