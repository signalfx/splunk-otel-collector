// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentreceiver

import (
	"fmt"
	"reflect"
)

// setStructFieldWithExplicitType sets the first occurrence of a (potentially embedded) struct field w/ name fieldName and first occurring fieldType
// to desired value, if available.  Returns true if successfully set, false otherwise.
// Error contains if field undetected or if issues occur with reflect usage.
func setStructFieldWithExplicitType(strukt any, fieldName string, value any, fieldTypes ...reflect.Type) (bool, error) {
	var err error
	for _, fieldType := range fieldTypes {
		var set bool
		set, err = setStructField(strukt, fieldName, value, fieldType, false)
		if set {
			return true, err
		}
	}
	return false, err
}

// setStructFieldIfZeroValue same as SetStructField() but only if first occurrence of the field is of zero value.
func setStructFieldIfZeroValue(strukt any, fieldName string, value any) (bool, error) {
	fieldType := reflect.TypeOf(value)
	return setStructField(strukt, fieldName, value, fieldType, true)
}

// getSettableStructFieldValue finds the first occurrence of a valid, settable field value from a potentially embedded struct
// with a fieldName of desired type.
// Based on https://github.com/signalfx/signalfx-agent/blob/731c2a0b5ff5ac324130453b02dd9cb7c912c0d5/pkg/utils/reflection.go#L36
func getSettableStructFieldValue(strukt any, fieldName string, fieldType reflect.Type) (*reflect.Value, error) {
	reflectedStruct := reflect.Indirect(reflect.ValueOf(strukt))
	if reflectedStruct.IsValid() && reflectedStruct.Type().Kind() == reflect.Struct {
		fieldValue := reflectedStruct.FieldByName(fieldName)
		if !fieldValue.IsValid() || !fieldValue.CanSet() || fieldValue.Type() != fieldType {
			valueCandidates := make([]reflect.Value, 0)
			for i := 0; i < reflectedStruct.Type().NumField(); i++ {
				field := reflectedStruct.Type().Field(i)
				if field.Type.Kind() == reflect.Struct && field.Anonymous && reflectedStruct.Field(i).CanSet() {
					candidate, _ := getSettableStructFieldValue(reflectedStruct.Field(i).Interface(), fieldName, fieldType)
					if candidate != nil {
						valueCandidates = append(valueCandidates, *candidate)
					}
				}
			}
			for _, value := range valueCandidates {
				if value.IsValid() {
					return &value, nil
				}
			}
			return nil, nil
		}
		return &fieldValue, nil
	}
	return nil, fmt.Errorf("invalid struct instance: %#v (CanAddr(): %t)", strukt, reflectedStruct.CanAddr())
}

func setStructField(strukt any, fieldName string, value any, fieldType reflect.Type, onlyIfZero bool) (bool, error) {
	valuePtr, err := getSettableStructFieldValue(strukt, fieldName, fieldType)
	if err != nil {
		return false, err
	}

	if valuePtr != nil {
		fieldValue := *valuePtr
		if !fieldValue.IsValid() {
			return false, fmt.Errorf("canot set invalid field value: %v", fieldValue)
		}
		if !(onlyIfZero && !fieldValue.IsZero()) {
			fieldValue.Set(reflect.ValueOf(value))
			return true, nil
		}
		return false, nil
	}
	return false, fmt.Errorf("no field %s of type %s detected", fieldName, fieldType)
}
