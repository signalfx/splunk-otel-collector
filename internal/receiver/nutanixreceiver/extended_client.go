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
	attrs := map[string]string{}
	if stateType != "" {
		attribute := inventoryMetricPrefix(domain, entityType) + "." + stateType
		if domain == "files" && stateType == "file_server" {
			attribute = "nutanix.files.file_server.id"
		}
		attrs[attribute] = state
	}
	return additionalMetric{
		Name:        inventoryMetricPrefix(domain, entityType) + ".count",
		Description: inventoryMetricDescription(entityType),
		Unit:        inventoryMetricUnit(entityType),
		Attributes:  attrs,
		Value:       float64(count),
	}
}

func inventoryMetricDescription(entityType string) string {
	switch entityType {
	case "address_group":
		return "Number of Nutanix microsegmentation address groups."
	case "alert":
		return "Number of Nutanix monitoring alerts."
	case "antivirus_server":
		return "Number of Nutanix Files antivirus servers."
	case "bgp_session":
		return "Number of Nutanix BGP sessions."
	case "category":
		return "Number of Nutanix Prism Central categories."
	case "category_key":
		return "Number of unique category keys in Nutanix Prism Central."
	case "gateway":
		return "Number of Nutanix gateways."
	case "layer2_stretch":
		return "Number of Nutanix Layer 2 stretches."
	case "mount_target":
		return "Number of Nutanix Files mount targets."
	case "network_controller":
		return "Number of Nutanix network controllers."
	case "network_security_policy":
		return "Number of Nutanix network security policies."
	case "object_store":
		return "Number of Nutanix Objects object stores."
	case "protected_vm":
		return "Number of Nutanix virtual machines protected by a data protection policy."
	case "protection_policy":
		return "Number of Nutanix data protection policies."
	case "protection_policy_schedule":
		return "Number of schedules configured for Nutanix data protection policies."
	case "recovery_point":
		return "Number of Nutanix data protection recovery points."
	case "routing_policy":
		return "Number of Nutanix routing policies."
	case "service_group":
		return "Number of Nutanix microsegmentation service groups."
	case "task":
		return "Number of Nutanix Prism Central tasks."
	case "traffic_mirror":
		return "Number of Nutanix traffic mirrors."
	case "unified_namespace":
		return "Number of Nutanix Files unified namespaces."
	case "uplink_bond":
		return "Number of Nutanix uplink bonds."
	case "virtual_switch":
		return "Number of Nutanix virtual switches."
	case "vpc":
		return "Number of Nutanix virtual private clouds."
	case "vpn_connection":
		return "Number of Nutanix VPN connections."
	default:
		return "Number of Nutanix " + strings.ReplaceAll(entityType, "_", " ") + "."
	}
}

func inventoryMetricPrefix(domain, entityType string) string {
	if domain == "networking" && entityType == "vpc_external_subnet" {
		entityType = "vpc.external_subnet"
	}
	return "nutanix." + domain + "." + entityType
}

func inventoryMetricUnit(entityType string) string {
	switch entityType {
	case "address_group", "service_group":
		return "{group}"
	case "alert":
		return "{alert}"
	case "antivirus_server", "file_server":
		return "{server}"
	case "bgp_session":
		return "{session}"
	case "category":
		return "{category}"
	case "category_key":
		return "{key}"
	case "gateway":
		return "{gateway}"
	case "layer2_stretch":
		return "{stretch}"
	case "mount_target":
		return "{target}"
	case "network_controller":
		return "{controller}"
	case "network_security_policy", "protection_policy", "routing_policy":
		return "{policy}"
	case "object_store":
		return "{store}"
	case "protected_vm":
		return "{vm}"
	case "recovery_point":
		return "{recovery_point}"
	case "protection_policy_schedule":
		return "{schedule}"
	case "task":
		return "{task}"
	case "traffic_mirror":
		return "{mirror}"
	case "unified_namespace":
		return "{namespace}"
	case "uplink_bond":
		return "{bond}"
	case "virtual_switch":
		return "{switch}"
	case "vpc":
		return "{vpc}"
	case "vpn_connection":
		return "{connection}"
	default:
		return "{" + entityType + "}"
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
