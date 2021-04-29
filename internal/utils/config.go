// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"reflect"
	"strings"
)

// RespectYamlTagsInAllSettings recursively walks through all tags and fixes them.
// Viper is case insensitive and doesn't preserve a record of actual yaml map key
// cases from the provided config, which is a problem when unmarshalling custom
// agent monitor configs.  Here we use a map of lowercase to supported case tag key
// names and update the keys where applicable.
func RespectYamlTagsInAllSettings(s reflect.Type, settings map[string]interface{}) {
	yamlTags := yamlTagsFromStruct(s)
	recursivelyCapitalizeConfigKeys(settings, yamlTags)
}

// yamlTagsFromStruct Walks through a custom monitor config struct type,
// creating a map of lowercase to supported yaml struct tag name cases.
func yamlTagsFromStruct(s reflect.Type) map[string]string {
	yamlTags := map[string]string{}
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		tag := field.Tag
		yamlTag := strings.Split(tag.Get("yaml"), ",")[0]
		lowerTag := strings.ToLower(yamlTag)
		if yamlTag != lowerTag {
			yamlTags[lowerTag] = yamlTag
		}

		fieldType := field.Type
		switch fieldType.Kind() {
		case reflect.Struct:
			otherFields := yamlTagsFromStruct(fieldType)
			for k, v := range otherFields {
				yamlTags[k] = v
			}
		case reflect.Ptr:
			fieldTypeElem := fieldType.Elem()
			if fieldTypeElem.Kind() == reflect.Struct {
				otherFields := yamlTagsFromStruct(fieldTypeElem)
				for k, v := range otherFields {
					yamlTags[k] = v
				}
			}
		}
	}

	return yamlTags
}

func recursivelyCapitalizeConfigKeys(settings map[string]interface{}, yamlTags map[string]string) {
	for key, val := range settings {
		updatedKey := yamlTags[key]
		if updatedKey != "" {
			delete(settings, key)
			settings[updatedKey] = val
			if m, ok := val.(map[string]interface{}); ok {
				recursivelyCapitalizeConfigKeys(m, yamlTags)
			}
		}
	}
}
