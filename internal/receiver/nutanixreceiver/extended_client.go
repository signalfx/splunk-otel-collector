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
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

const (
	v4PageSize        = 100
	v4RequestAttempts = 3
	v4StatsWorkers    = 10
)

func statsFromMap(entity map[string]any) []metricStat {
	var result []metricStat
	for key, value := range entity {
		if skipRawField(key) {
			continue
		}
		appendNumericStats(&result, key, value)
	}
	return result
}

func inventoryCountMetric(domain, entityType, stateType, state string, count int) additionalMetric {
	attrs := map[string]string{"nutanix.entity.type": entityType}
	if stateType != "" {
		attrs["nutanix.entity.state_type"] = stateType
		attrs["nutanix.entity.state"] = state
	}
	return additionalMetric{
		Name:        "nutanix." + domain + ".entity.count",
		Description: "Number of Nutanix " + strings.ReplaceAll(domain, "_", " ") + " entities",
		Unit:        "{entity}",
		Attributes:  attrs,
		Value:       float64(count),
	}
}

func responseDataMaps(response map[string]any) []map[string]any {
	data, ok := response["data"]
	if !ok {
		return nil
	}
	switch data := data.(type) {
	case []any:
		result := make([]map[string]any, 0, len(data))
		for _, item := range data {
			if entity, ok := item.(map[string]any); ok {
				result = append(result, entity)
			}
		}
		return result
	case map[string]any:
		return []map[string]any{data}
	default:
		return nil
	}
}

func responseTotal(response map[string]any) (int, bool) {
	metadata, ok := response["metadata"].(map[string]any)
	if !ok {
		return 0, false
	}
	for _, key := range []string{"totalAvailableResults", "total_available_results"} {
		if value, exists := metadata[key]; exists {
			return int(numberValue(value)), true
		}
	}
	return 0, false
}

func numberValue(value any) float64 {
	switch value := value.(type) {
	case json.Number:
		result, _ := value.Float64()
		return result
	case float64:
		return value
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case string:
		result, _ := strconv.ParseFloat(value, 64)
		return result
	default:
		return 0
	}
}

func mapList(entity map[string]any, keys ...string) []map[string]any {
	for _, key := range keys {
		items, ok := entity[key].([]any)
		if !ok {
			continue
		}
		result := make([]map[string]any, 0, len(items))
		for _, item := range items {
			if value, ok := item.(map[string]any); ok {
				result = append(result, value)
			}
		}
		return result
	}
	return nil
}

func countMapsByString(entities []map[string]any, key, expected string) int {
	return countMaps(entities, func(entity map[string]any) bool {
		return stringFieldEquals(entity, key, expected)
	})
}

func countMicrosegPoliciesByScope(entities []map[string]any, expected ...string) int {
	values := make(map[string]struct{}, len(expected))
	for _, value := range expected {
		values[normalizeEnumName(value)] = struct{}{}
	}
	return countMaps(entities, func(entity map[string]any) bool {
		_, ok := values[normalizeEnumName(firstString(entity, "scope"))]
		return ok
	})
}

func countMaps(entities []map[string]any, matches func(map[string]any) bool) int {
	count := 0
	for _, entity := range entities {
		if matches(entity) {
			count++
		}
	}
	return count
}

func stringFieldEquals(entity map[string]any, key, expected string) bool {
	return normalizeEnumName(firstString(entity, key)) == normalizeEnumName(expected)
}

func countProtectionPolicySchedules(policies []map[string]any) map[string]int {
	counts := map[string]int{}
	for _, policy := range policies {
		policyCounts := map[string]int{}
		for _, replication := range mapList(policy, "replicationConfigurations", "replication_configurations") {
			schedule, _ := replication["schedule"].(map[string]any)
			if len(schedule) == 0 {
				continue
			}
			policyCounts["total"]++
			consistency := normalizeEnumName(firstString(schedule, "recoveryPointType", "recovery_point_type"))
			if consistency != "" {
				policyCounts[consistency]++
			}
			rpoValue := firstValue(schedule, "recoveryPointObjectiveTimeSeconds", "recovery_point_objective_time_seconds")
			if rpoValue == nil {
				continue
			}
			if rpo := rpoClass(numberValue(rpoValue)); rpo != "" {
				policyCounts[rpo]++
			}
		}
		for key, value := range policyCounts {
			counts[key] += int(math.Ceil(float64(value) / 2))
		}
	}
	return counts
}

func countProtectedVMsByRPO(policies []map[string]any, vms []nutanixVM) map[string]int {
	policyRPO := make(map[string]string, len(policies))
	for _, policy := range policies {
		policyID := firstString(policy, "extId", "ext_id", "id")
		if policyID == "" {
			continue
		}
		for _, replication := range mapList(policy, "replicationConfigurations", "replication_configurations") {
			schedule, _ := replication["schedule"].(map[string]any)
			rpoValue := firstValue(schedule, "recoveryPointObjectiveTimeSeconds", "recovery_point_objective_time_seconds")
			if rpoValue == nil {
				continue
			}
			policyRPO[policyID] = rpoClass(numberValue(rpoValue))
			break
		}
	}

	counts := map[string]int{}
	for i := range vms {
		if rpo := policyRPO[vms[i].ProtectionPolicyID]; rpo != "" {
			counts[rpo]++
		}
	}
	return counts
}

func rpoClass(seconds float64) string {
	switch {
	case seconds == 0:
		return "sync"
	case seconds > 0 && seconds <= 900:
		return "nearsync"
	case seconds > 900:
		return "async"
	default:
		return ""
	}
}

func firstValue(entity map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := entity[key]; ok {
			return value
		}
	}
	return nil
}

func skipRawField(name string) bool {
	switch name {
	case "$objectType", "$reserved", "$unknownFields", "extId", "tenantId", "links", "timestamp", "metadata":
		return true
	default:
		return false
	}
}

func compactAdditionalMetrics(metrics []additionalMetric) []additionalMetric {
	result := make([]additionalMetric, 0, len(metrics))
	for _, metric := range metrics {
		if metric.Name != "" && (len(metric.Stats) > 0 || metric.Unit != "") {
			result = append(result, metric)
		}
	}
	return result
}

func compactErrors(errs []error) []error {
	result := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			result = append(result, err)
		}
	}
	return result
}
