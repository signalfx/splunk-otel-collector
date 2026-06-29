// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package nutanixreceiver

import (
	"reflect"
	"strings"
)

type metricStat struct {
	Name  string
	Value float64
}

type nutanixCluster struct {
	ID    string
	Name  string
	Stats []metricStat
}

type nutanixHost struct {
	ID          string
	Name        string
	ClusterID   string
	ClusterName string
	Stats       []metricStat
}

type nutanixStorageContainer struct {
	ID          string
	Name        string
	ClusterID   string
	ClusterName string
	Stats       []metricStat
}

type nutanixVM struct {
	ID                string
	Name              string
	ClusterID         string
	HostID            string
	PowerState        string
	DiskBuses         []string
	Stats             []metricStat
	MemoryBytes       int64
	NumSockets        int
	NumCoresPerSocket int
	NICCount          int
}

type nutanixVolumeGroup struct {
	ID        string
	Name      string
	ClusterID string
	Stats     []metricStat
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func cloneAttrs(attrs map[string]string) map[string]string {
	if len(attrs) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(attrs))
	for k, v := range attrs {
		cloned[k] = v
	}
	return cloned
}

func sanitizeAttributeValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ".", "_")
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, "/", "_")
	return value
}

func normalizeEnumName(value string) string {
	value = strings.TrimPrefix(value, "$")
	return strings.ToLower(value)
}

func statsFromStruct(value any) []metricStat {
	rv := reflect.ValueOf(value)
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()
	stats := make([]metricStat, 0)
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		name := jsonFieldName(field)
		if name == "" || skipStatField(name) {
			continue
		}

		fieldValue := rv.Field(i)
		if value, ok := latestValueFromSeries(fieldValue); ok {
			stats = append(stats, metricStat{Name: name, Value: value})
			continue
		}
		if value, ok := numberFromValue(fieldValue); ok {
			stats = append(stats, metricStat{Name: name, Value: value})
		}
	}
	return stats
}

func jsonFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	if tag == "" {
		return lowerFirst(field.Name)
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

func lowerFirst(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToLower(value[:1]) + value[1:]
}

func skipStatField(name string) bool {
	switch name {
	case "$objectType", "$reserved", "$unknownFields", "extId", "tenantId", "links",
		"timestamp", "cluster", "containerExtId", "volumeGroupExtId", "stats":
		return true
	default:
		return false
	}
}

func latestValueFromSeries(value reflect.Value) (float64, bool) {
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return 0, false
		}
		value = value.Elem()
	}
	if !value.IsValid() || value.Kind() != reflect.Slice {
		return 0, false
	}

	for i := value.Len() - 1; i >= 0; i-- {
		item := value.Index(i)
		for item.IsValid() && item.Kind() == reflect.Pointer {
			if item.IsNil() {
				return 0, false
			}
			item = item.Elem()
		}
		if !item.IsValid() || item.Kind() != reflect.Struct {
			continue
		}
		valueField := item.FieldByName("Value")
		if value, ok := numberFromValue(valueField); ok {
			return value, true
		}
	}
	return 0, false
}

func numberFromValue(value reflect.Value) (float64, bool) {
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return 0, false
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return 0, false
	}

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint()), true
	case reflect.Float32, reflect.Float64:
		return value.Float(), true
	default:
		return 0, false
	}
}
