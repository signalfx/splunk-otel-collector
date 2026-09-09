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
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	v4PageSize        = 100
	v4RequestAttempts = 3
	v4StatsWorkers    = 10
)

type v4RESTClient struct {
	baseURL    *url.URL
	httpClient *http.Client
	username   string
	password   string
	interval   time.Duration
}

type v4EntityEndpoint struct {
	entityType string
	listPath   string
	statsPath  string
}

func newV4RESTClient(baseURL *url.URL, cfg *Config) *v4RESTClient {
	return &v4RESTClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.TLS.InsecureSkipVerify}, //nolint:gosec // configured by the user
			},
			Timeout: cfg.ControllerConfig.Timeout,
		},
		username: cfg.Username,
		password: string(cfg.Password),
		interval: cfg.ControllerConfig.CollectionInterval,
	}
}

func (c *prismClient) collectAdditionalMetrics(ctx context.Context, request additionalMetricsRequest) (additionalSnapshot, error) {
	var result additionalSnapshot
	collectors := []struct {
		collect func(context.Context) additionalSnapshot
		enabled bool
	}{
		{collect: c.restClient.collectNetworking, enabled: request.Networking},
		{collect: c.restClient.collectPrismCentral, enabled: request.PrismCentral},
		{collect: func(ctx context.Context) additionalSnapshot {
			return c.restClient.collectDataProtection(ctx, request.VMs)
		}, enabled: request.DataProtection},
		{collect: c.restClient.collectMicrosegmentation, enabled: request.Microsegmentation},
		{collect: c.restClient.collectFiles, enabled: request.Files},
		{collect: c.restClient.collectObjects, enabled: request.Objects},
	}
	for _, collector := range collectors {
		if !collector.enabled {
			continue
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}
		partial := collector.collect(ctx)
		result.Metrics = append(result.Metrics, partial.Metrics...)
		result.Errors = append(result.Errors, partial.Errors...)
		if err := ctx.Err(); err != nil {
			return result, err
		}
	}
	return result, nil
}

func (c *v4RESTClient) collectNetworking(ctx context.Context) additionalSnapshot {
	endpoints := []v4EntityEndpoint{
		{entityType: "bgp_session", listPath: "/api/networking/v4.2/config/bgp-sessions"},
		{entityType: "gateway", listPath: "/api/networking/v4.2/config/gateways"},
		{entityType: "layer2_stretch", listPath: "/api/networking/v4.2/config/layer2-stretches", statsPath: "/api/networking/v4.2/stats/layer2-stretches/%s"},
		{entityType: "network_controller", listPath: "/api/networking/v4.2/config/controllers"},
		{entityType: "routing_policy", listPath: "/api/networking/v4.2/config/routing-policies"},
		{entityType: "traffic_mirror", listPath: "/api/networking/v4.2/config/traffic-mirrors", statsPath: "/api/networking/v4.2/stats/traffic-mirrors/%s"},
		{entityType: "uplink_bond", listPath: "/api/networking/v4.2/config/uplink-bonds"},
		{entityType: "virtual_switch", listPath: "/api/networking/v4.2/config/virtual-switches"},
		{entityType: "vpn_connection", listPath: "/api/networking/v4.2/config/vpn-connections", statsPath: "/api/networking/v4.2/stats/vpn-connections/%s"},
	}

	var result additionalSnapshot
	for _, endpoint := range endpoints {
		if endpoint.statsPath == "" {
			count, err := c.count(ctx, endpoint.listPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("list networking %s: %w", endpoint.entityType, err))
				continue
			}
			result.Metrics = append(result.Metrics, inventoryCountMetric("networking", endpoint.entityType, "", "", count))
			continue
		}

		entities, err := c.listAll(ctx, endpoint.listPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("list networking %s: %w", endpoint.entityType, err))
			continue
		}
		result.Metrics = append(result.Metrics, inventoryCountMetric("networking", endpoint.entityType, "", "", len(entities)))
		stats, errs := c.collectEntityStats(ctx, "networking", endpoint.entityType, entities, func(entity map[string]any) string {
			return fmt.Sprintf(endpoint.statsPath, url.PathEscape(firstString(entity, "extId", "ext_id", "id")))
		})
		result.Metrics = append(result.Metrics, stats...)
		result.Errors = append(result.Errors, errs...)
	}

	vpcs, err := c.listAll(ctx, "/api/networking/v4.2/config/vpcs")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list networking vpc: %w", err))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("networking", "vpc", "", "", len(vpcs)))
		var externalSubnets []map[string]any
		for _, vpc := range vpcs {
			vpcID := firstString(vpc, "extId", "ext_id", "id")
			vpcName := firstString(vpc, "name")
			for _, subnet := range mapList(vpc, "externalSubnets", "external_subnets") {
				subnet["vpcExtId"] = vpcID
				subnet["vpcName"] = vpcName
				externalSubnets = append(externalSubnets, subnet)
			}
		}
		stats, errs := c.collectEntityStats(ctx, "networking", "vpc_external_subnet", externalSubnets, func(entity map[string]any) string {
			return fmt.Sprintf(
				"/api/networking/v4.2/stats/vpc/%s/external-subnets/%s",
				url.PathEscape(firstString(entity, "vpcExtId")),
				url.PathEscape(firstString(entity, "subnetReference", "subnet_reference", "extId")),
			)
		})
		result.Metrics = append(result.Metrics, stats...)
		result.Errors = append(result.Errors, errs...)
	}
	return result
}

func (c *v4RESTClient) collectPrismCentral(ctx context.Context) additionalSnapshot {
	var result additionalSnapshot

	categories, err := c.listAll(ctx, "/api/prism/v4.2/config/categories")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list categories: %w", err))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("prism", "category", "", "", len(categories)))
		for _, categoryType := range []string{"system", "user", "internal"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("prism", "category", "type", categoryType, countMapsByString(categories, "type", categoryType)))
		}
		keys := map[string]struct{}{}
		for _, category := range categories {
			if key := firstString(category, "key"); key != "" {
				keys[key] = struct{}{}
			}
		}
		result.Metrics = append(result.Metrics, inventoryCountMetric("prism", "category_key", "", "", len(keys)))
	}

	result.appendCount(ctx, c, "prism", "task", "", "", "/api/prism/v4.2/config/tasks", "")
	for _, status := range []string{"queued", "running", "canceling", "succeeded", "failed", "canceled", "suspended"} {
		filter := fmt.Sprintf("status eq Prism.Config.TaskStatus'%s'", strings.ToUpper(status))
		result.appendCount(ctx, c, "prism", "task", "status", status, "/api/prism/v4.2/config/tasks", filter)
	}

	alertsPath := "/api/monitoring/v4.2/serviceability/alerts"
	result.appendCount(ctx, c, "monitoring", "alert", "", "", alertsPath, "")
	for _, resolved := range []bool{true, false} {
		state := strconv.FormatBool(resolved)
		result.appendCount(ctx, c, "monitoring", "alert", "resolved", state, alertsPath, "isResolved eq "+state)
	}
	for _, acknowledged := range []bool{true, false} {
		state := strconv.FormatBool(acknowledged)
		result.appendCount(ctx, c, "monitoring", "alert", "acknowledged", state, alertsPath, "isAcknowledged eq "+state)
	}
	for _, severity := range []string{"info", "warning", "critical"} {
		severityFilter := fmt.Sprintf("severity eq Monitoring.Common.Severity'%s'", strings.ToUpper(severity))
		result.appendCount(ctx, c, "monitoring", "alert", "severity", severity, alertsPath, severityFilter)
		result.appendCount(ctx, c, "monitoring", "alert", "unresolved_severity", severity, alertsPath, "isResolved eq false and "+severityFilter)
	}
	return result
}

func (c *v4RESTClient) collectDataProtection(ctx context.Context, vms []nutanixVM) additionalSnapshot {
	var result additionalSnapshot
	policies, err := c.listAll(ctx, "/api/datapolicies/v4.2/config/protection-policies")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list protection policies: %w", err))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protection_policy", "", "", len(policies)))
		counts := countProtectionPolicySchedules(policies)
		result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protection_policy_schedule", "", "", counts["total"]))
		for _, consistency := range []string{"crash_consistent", "application_consistent"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protection_policy_schedule", "consistency", consistency, counts[consistency]))
		}
		for _, rpo := range []string{"sync", "nearsync", "async"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protection_policy_schedule", "rpo", rpo, counts[rpo]))
		}
		protectedVMsByRPO := countProtectedVMsByRPO(policies, vms)
		for _, rpo := range []string{"sync", "nearsync", "async"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protected_vm", "rpo", rpo, protectedVMsByRPO[rpo]))
		}
	}

	recoveryPoints, err := c.count(ctx, "/api/dataprotection/v4.2/config/recovery-points")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list recovery points: %w", err))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "recovery_point", "", "", recoveryPoints))
	}
	return result
}

func (c *v4RESTClient) collectMicrosegmentation(ctx context.Context) additionalSnapshot {
	var result additionalSnapshot
	policies, err := c.listAll(ctx, "/api/microseg/v4.2/config/policies")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list microsegmentation policies: %w", err))
	} else {
		result.Metrics = append(result.Metrics,
			inventoryCountMetric("microseg", "network_security_policy", "", "", len(policies)),
			inventoryCountMetric("microseg", "network_security_policy", "scope", "vlan", countMicrosegPoliciesByScope(policies, "vlan", "all_vlan")),
			inventoryCountMetric("microseg", "network_security_policy", "scope", "vpc", countMicrosegPoliciesByScope(policies, "vpc", "all_vpc", "vpc_list")),
		)
		for _, state := range []string{"save", "monitor", "enforce"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("microseg", "network_security_policy", "state", state, countMapsByString(policies, "state", state)))
		}
		for _, policyType := range []string{"quarantine", "isolation", "application"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("microseg", "network_security_policy", "type", policyType, countMapsByString(policies, "type", policyType)))
		}
	}
	for _, endpoint := range []v4EntityEndpoint{
		{entityType: "address_group", listPath: "/api/microseg/v4.2/config/address-groups"},
		{entityType: "service_group", listPath: "/api/microseg/v4.2/config/service-groups"},
	} {
		count, countErr := c.count(ctx, endpoint.listPath)
		if countErr != nil {
			result.Errors = append(result.Errors, fmt.Errorf("list microsegmentation %s: %w", endpoint.entityType, countErr))
			continue
		}
		result.Metrics = append(result.Metrics, inventoryCountMetric("microseg", endpoint.entityType, "", "", count))
	}
	return result
}

func (c *v4RESTClient) collectFiles(ctx context.Context) additionalSnapshot {
	var result additionalSnapshot
	fileServers, err := c.listAll(ctx, "/api/files/v4.0/config/file-servers")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list file servers: %w", err))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("files", "file_server", "", "", len(fileServers)))
		stats, errs := c.collectEntityStats(ctx, "files", "file_server", fileServers, func(entity map[string]any) string {
			return "/api/files/v4.0/stats/file-servers/" + url.PathEscape(firstString(entity, "extId", "ext_id", "id"))
		})
		result.Metrics = append(result.Metrics, stats...)
		result.Errors = append(result.Errors, errs...)
	}

	if count, countErr := c.count(ctx, "/api/files/v4.0/config/unified-namespaces"); countErr != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list unified namespaces: %w", countErr))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("files", "unified_namespace", "", "", count))
	}

	for _, fileServer := range fileServers {
		fileServerID := firstString(fileServer, "extId", "ext_id", "id")
		if fileServerID == "" {
			continue
		}
		escapedFileServerID := url.PathEscape(fileServerID)
		for _, nested := range []v4EntityEndpoint{
			{entityType: "antivirus_server", listPath: "/api/files/v4.0/config/file-servers/" + escapedFileServerID + "/anti-virus-servers", statsPath: "/api/files/v4.0/stats/file-servers/" + escapedFileServerID + "/anti-virus-servers/%s"},
			{entityType: "mount_target", listPath: "/api/files/v4.0/config/file-servers/" + escapedFileServerID + "/mount-targets", statsPath: "/api/files/v4.0/stats/file-servers/" + escapedFileServerID + "/mount-targets/%s"},
		} {
			entities, listErr := c.listAll(ctx, nested.listPath)
			if listErr != nil {
				result.Errors = append(result.Errors, fmt.Errorf("list files %s for %s: %w", nested.entityType, fileServerID, listErr))
				continue
			}
			result.Metrics = append(result.Metrics, inventoryCountMetric("files", nested.entityType, "file_server", fileServerID, len(entities)))
			nestedStats, nestedErrs := c.collectEntityStats(ctx, "files", nested.entityType, entities, func(entity map[string]any) string {
				return fmt.Sprintf(nested.statsPath, url.PathEscape(firstString(entity, "extId", "ext_id", "id")))
			})
			result.Metrics = append(result.Metrics, nestedStats...)
			result.Errors = append(result.Errors, nestedErrs...)
		}
	}
	return result
}

func (c *v4RESTClient) collectObjects(ctx context.Context) additionalSnapshot {
	var result additionalSnapshot
	objectStores, err := c.listAll(ctx, "/api/objects/v4.1/config/object-stores")
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list object stores: %w", err))
		return result
	}
	result.Metrics = append(result.Metrics, inventoryCountMetric("objects", "object_store", "", "", len(objectStores)))
	stats, errs := c.collectEntityStats(ctx, "objects", "object_store", objectStores, func(entity map[string]any) string {
		return "/api/objects/v4.1/stats/object-stores/" + url.PathEscape(firstString(entity, "extId", "ext_id", "id"))
	})
	result.Metrics = append(result.Metrics, stats...)
	result.Errors = append(result.Errors, errs...)
	return result
}

func (c *v4RESTClient) collectEntityStats(
	ctx context.Context,
	domain string,
	entityType string,
	entities []map[string]any,
	path func(map[string]any) string,
) ([]additionalMetric, []error) {
	metrics := make([]additionalMetric, len(entities))
	errorsByIndex := make([]error, len(entities))
	sem := make(chan struct{}, v4StatsWorkers)
	var wg sync.WaitGroup
	for i := range entities {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				errorsByIndex[index] = ctx.Err()
				return
			}
			entity := entities[index]
			entityID := firstString(entity, "extId", "ext_id", "id", "subnetReference", "subnet_reference")
			if entityID == "" {
				return
			}
			stats, err := c.stats(ctx, path(entity))
			if err != nil {
				errorsByIndex[index] = fmt.Errorf("get %s %s stats for %s: %w", domain, entityType, entityID, err)
				return
			}
			metrics[index] = additionalMetric{
				Name:        "nutanix." + domain + ".entity.stat",
				Description: "Latest Nutanix " + strings.ReplaceAll(domain, "_", " ") + " entity statistic",
				Attributes: map[string]string{
					"nutanix.entity.type": entityType,
					"nutanix.entity.id":   entityID,
					"nutanix.entity.name": firstString(entity, "name", "vpcName"),
				},
				Stats: stats,
			}
		}(i)
	}
	wg.Wait()
	return compactAdditionalMetrics(metrics), compactErrors(errorsByIndex)
}

func (c *v4RESTClient) stats(ctx context.Context, path string) ([]metricStat, error) {
	interval := c.interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	samplingInterval := int(math.Max(1, interval.Seconds()))
	includeStatType := true
	if strings.HasPrefix(path, "/api/files/") {
		if interval < 5*time.Minute {
			interval = 5 * time.Minute
		}
		samplingInterval = 300
		includeStatType = false
	} else if strings.HasPrefix(path, "/api/objects/") && samplingInterval < 120 {
		samplingInterval = 120
	}
	end := time.Now()
	start := end.Add(-interval)
	query := url.Values{
		"$startTime":        []string{start.Format(time.RFC3339Nano)},
		"$endTime":          []string{end.Format(time.RFC3339Nano)},
		"$samplingInterval": []string{strconv.Itoa(samplingInterval)},
	}
	if includeStatType {
		query.Set("$statType", "LAST")
	}
	response, err := c.get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	entities := responseDataMaps(response)
	if len(entities) == 0 {
		return nil, nil
	}
	return statsFromMap(entities[len(entities)-1]), nil
}

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

func (c *v4RESTClient) count(ctx context.Context, path string) (int, error) {
	return c.countWithFilter(ctx, path, "")
}

func (c *v4RESTClient) countWithFilter(ctx context.Context, path, filter string) (int, error) {
	query := url.Values{"$page": []string{"0"}, "$limit": []string{"1"}}
	if filter != "" {
		query.Set("$filter", filter)
	}
	response, err := c.get(ctx, path, query)
	if err != nil {
		return 0, err
	}
	if total, ok := responseTotal(response); ok {
		return total, nil
	}
	return len(responseDataMaps(response)), nil
}

func (s *additionalSnapshot) appendCount(ctx context.Context, client *v4RESTClient, domain, entityType, stateType, state, path, filter string) {
	count, err := client.countWithFilter(ctx, path, filter)
	if err != nil {
		s.Errors = append(s.Errors, fmt.Errorf("count %s %s: %w", domain, entityType, err))
		return
	}
	s.Metrics = append(s.Metrics, inventoryCountMetric(domain, entityType, stateType, state, count))
}

func (c *v4RESTClient) listAll(ctx context.Context, path string) ([]map[string]any, error) {
	var result []map[string]any
	for page := 0; ; page++ {
		response, err := c.get(ctx, path, url.Values{
			"$page":  []string{strconv.Itoa(page)},
			"$limit": []string{strconv.Itoa(v4PageSize)},
		})
		if err != nil {
			return nil, err
		}
		entities := responseDataMaps(response)
		result = append(result, entities...)
		total, hasTotal := responseTotal(response)
		if len(entities) == 0 || (hasTotal && len(result) >= total) || len(entities) < v4PageSize {
			return result, nil
		}
	}
}

func (c *v4RESTClient) get(ctx context.Context, path string, query url.Values) (map[string]any, error) {
	var lastErr error
	for attempt := 0; attempt < v4RequestAttempts; attempt++ {
		requestURL := *c.baseURL
		requestURL.Path = path
		requestURL.RawQuery = query.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), http.NoBody)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(c.username, c.password)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt+1 == v4RequestAttempts {
				break
			}
			if waitErr := waitForRetry(ctx, attempt, ""); waitErr != nil {
				return nil, waitErr
			}
			continue
		}

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			decoder := json.NewDecoder(io.LimitReader(resp.Body, 32*1024*1024))
			decoder.UseNumber()
			var response map[string]any
			decodeErr := decoder.Decode(&response)
			resp.Body.Close()
			if decodeErr != nil {
				return nil, fmt.Errorf("decode %s response: %w", path, decodeErr)
			}
			return response, nil
		}

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
		resp.Body.Close()
		lastErr = fmt.Errorf("request %s returned %s: %s", path, resp.Status, strings.TrimSpace(string(body)))
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < http.StatusInternalServerError {
			return nil, lastErr
		}
		if attempt+1 == v4RequestAttempts {
			break
		}
		if waitErr := waitForRetry(ctx, attempt, resp.Header.Get("Retry-After")); waitErr != nil {
			return nil, waitErr
		}
	}
	return nil, lastErr
}

func waitForRetry(ctx context.Context, attempt int, retryAfter string) error {
	delay := time.Duration(1<<attempt) * 250 * time.Millisecond
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
		delay = time.Duration(seconds) * time.Second
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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
