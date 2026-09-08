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
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// prismElementClient uses the cluster-local Prism Element v2.0 REST API.
// Prism Element does not expose the Prism Central v4 API namespaces.
type prismElementClient struct {
	baseURL    *url.URL
	httpClient *http.Client
	username   string
	password   string
	cluster    nutanixCluster
}

func newPrismElementClient(cfg *Config) (*prismElementClient, error) {
	baseURL, err := normalizeEndpoint(cfg.Endpoint, cfg.Port)
	if err != nil {
		return nil, err
	}

	return &prismElementClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.TLS.InsecureSkipVerify}, //nolint:gosec // configured by the user
			},
			Timeout: cfg.ControllerConfig.Timeout,
		},
		username: cfg.Username,
		password: string(cfg.Password),
	}, nil
}

func (c *prismElementClient) serverAddress() string {
	return c.baseURL.Hostname()
}

func (c *prismElementClient) serverPort() int64 {
	return serverPort(c.baseURL)
}

func (c *prismElementClient) request(ctx context.Context, path string, query url.Values, result any) error {
	requestURL := *c.baseURL
	requestURL.Path = "/api/nutanix/v2.0/" + strings.Trim(path, "/")
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s failed: %w", requestURL.Path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
		return fmt.Errorf("request %s returned %s: %s", requestURL.Path, resp.Status, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode %s response: %w", requestURL.Path, err)
	}
	return nil
}

func (c *prismElementClient) ensureCluster(ctx context.Context) error {
	if c.cluster.ID != "" || c.cluster.Name != "" {
		return nil
	}
	_, err := c.listClusters(ctx)
	return err
}

func (c *prismElementClient) listClusters(ctx context.Context) ([]nutanixCluster, error) {
	var response map[string]any
	if err := c.request(ctx, "cluster", nil, &response); err != nil {
		return nil, err
	}

	entity := response
	if entities := entityList(response); len(entities) > 0 {
		entity = entities[0]
	} else if nested, ok := response["cluster"].(map[string]any); ok {
		entity = nested
	}

	c.cluster = nutanixCluster{
		ID:    firstString(entity, "id", "uuid", "cluster_uuid", "clusterUuid"),
		Name:  firstString(entity, "name", "cluster_name", "clusterName"),
		Stats: legacyStats(entity),
	}
	return []nutanixCluster{c.cluster}, nil
}

func (c *prismElementClient) listHosts(ctx context.Context) ([]nutanixHost, error) {
	if err := c.ensureCluster(ctx); err != nil {
		return nil, err
	}
	var response map[string]any
	if err := c.request(ctx, "hosts", nil, &response); err != nil {
		return nil, err
	}

	entities := entityList(response)
	result := make([]nutanixHost, 0, len(entities))
	for _, entity := range entities {
		result = append(result, nutanixHost{
			ID:          firstString(entity, "id", "uuid"),
			Name:        firstString(entity, "name", "host_name", "hostName"),
			ClusterID:   valueOr(firstString(entity, "cluster_uuid", "clusterUuid"), c.cluster.ID),
			ClusterName: valueOr(firstString(entity, "cluster_name", "clusterName"), c.cluster.Name),
			Stats:       legacyStats(entity),
		})
	}
	return result, nil
}

func (c *prismElementClient) listStorageContainers(ctx context.Context) ([]nutanixStorageContainer, error) {
	if err := c.ensureCluster(ctx); err != nil {
		return nil, err
	}
	var response map[string]any
	if err := c.request(ctx, "storage_containers", nil, &response); err != nil {
		return nil, err
	}

	entities := entityList(response)
	result := make([]nutanixStorageContainer, 0, len(entities))
	for _, entity := range entities {
		result = append(result, nutanixStorageContainer{
			ID:          firstString(entity, "id", "uuid"),
			Name:        firstString(entity, "name", "container_name", "containerName"),
			ClusterID:   valueOr(firstString(entity, "cluster_uuid", "clusterUuid"), c.cluster.ID),
			ClusterName: valueOr(firstString(entity, "cluster_name", "clusterName"), c.cluster.Name),
			Stats:       legacyStats(entity),
		})
	}
	return result, nil
}

func (c *prismElementClient) listVMs(ctx context.Context) ([]nutanixVM, error) {
	if err := c.ensureCluster(ctx); err != nil {
		return nil, err
	}
	var response map[string]any
	if err := c.request(ctx, "vms", nil, &response); err != nil {
		return nil, err
	}

	entities := entityList(response)
	result := make([]nutanixVM, 0, len(entities))
	for _, entity := range entities {
		result = append(result, vmFromElement(entity, c.cluster))
	}
	return result, nil
}

func (c *prismElementClient) listVolumeGroups(ctx context.Context) ([]nutanixVolumeGroup, error) {
	if err := c.ensureCluster(ctx); err != nil {
		return nil, err
	}
	var response map[string]any
	if err := c.request(ctx, "volume_groups", nil, &response); err != nil {
		return nil, err
	}

	entities := entityList(response)
	result := make([]nutanixVolumeGroup, 0, len(entities))
	for _, entity := range entities {
		result = append(result, nutanixVolumeGroup{
			ID:        firstString(entity, "id", "uuid"),
			Name:      firstString(entity, "name", "volume_group_name", "volumeGroupName"),
			ClusterID: valueOr(firstString(entity, "cluster_uuid", "clusterUuid"), c.cluster.ID),
			Stats:     legacyStats(entity),
		})
	}
	return result, nil
}

func (c *prismElementClient) getClusterStats(context.Context, nutanixCluster) ([]metricStat, error) {
	return c.cluster.Stats, nil
}

func (c *prismElementClient) getHostStats(_ context.Context, host nutanixHost) ([]metricStat, error) {
	return host.Stats, nil
}

func (c *prismElementClient) getStorageContainerStats(_ context.Context, container nutanixStorageContainer) ([]metricStat, error) {
	return container.Stats, nil
}

func (c *prismElementClient) listVMStats(ctx context.Context) (map[string][]metricStat, error) {
	vms, err := c.listVMs(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]metricStat, len(vms))
	for _, vm := range vms {
		result[vm.ID] = vm.Stats
	}
	return result, nil
}

func (c *prismElementClient) getVolumeGroupStats(_ context.Context, volumeGroup nutanixVolumeGroup) ([]metricStat, error) {
	return volumeGroup.Stats, nil
}

func vmFromElement(entity map[string]any, cluster nutanixCluster) nutanixVM {
	vm := nutanixVM{
		ID:                firstString(entity, "id", "uuid"),
		Name:              firstString(entity, "name", "vm_name", "vmName"),
		ClusterID:         valueOr(firstString(entity, "cluster_uuid", "clusterUuid"), cluster.ID),
		HostID:            firstString(entity, "host_uuid", "hostUuid"),
		PowerState:        strings.ToLower(firstString(entity, "power_state", "powerState")),
		MemoryBytes:       int64(firstNumber(entity, "memory_mb", "memoryMb")) * 1024 * 1024,
		NumSockets:        int(firstNumber(entity, "num_sockets", "numSockets")),
		NumCoresPerSocket: int(firstNumber(entity, "num_cores_per_socket", "numCoresPerSocket", "num_cores_per_vcpu")),
		NICCount:          countList(entity, "nic_list", "nics", "nicList"),
		Stats:             legacyStats(entity),
	}
	if vm.PowerState == "" {
		vm.PowerState = "unknown"
	}
	if vm.NumSockets == 0 {
		vm.NumSockets = 1
		vm.NumCoresPerSocket = int(firstNumber(entity, "num_vcpus", "numVcpus"))
	}
	vm.DiskBuses = diskBuses(entity)
	return vm
}

func entityList(response map[string]any) []map[string]any {
	values, ok := response["entities"].([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if entity, ok := value.(map[string]any); ok {
			result = append(result, entity)
		}
	}
	return result
}

func firstString(entity map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := entity[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func firstNumber(entity map[string]any, keys ...string) float64 {
	for _, key := range keys {
		switch value := entity[key].(type) {
		case float64:
			return value
		case json.Number:
			parsed, _ := value.Float64()
			return parsed
		case string:
			parsed, _ := strconv.ParseFloat(value, 64)
			return parsed
		}
	}
	return 0
}

func valueOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func countList(entity map[string]any, keys ...string) int {
	for _, key := range keys {
		if values, ok := entity[key].([]any); ok {
			return len(values)
		}
	}
	return 0
}

func diskBuses(entity map[string]any) []string {
	for _, key := range []string{"vm_disks", "disks", "disk_list", "vmdisk_list"} {
		values, ok := entity[key].([]any)
		if !ok {
			continue
		}
		buses := make([]string, 0, len(values))
		for _, value := range values {
			if disk, ok := value.(map[string]any); ok {
				buses = append(buses, strings.ToLower(firstString(disk, "bus", "device_bus", "deviceBus")))
			}
		}
		return buses
	}
	return nil
}

func legacyStats(entity map[string]any) []metricStat {
	var result []metricStat
	for _, key := range []string{"stats", "usage_stats", "usageStats"} {
		if value, ok := entity[key]; ok {
			appendNumericStats(&result, key, value)
		}
	}
	return result
}

func appendNumericStats(result *[]metricStat, prefix string, value any) {
	switch value := value.(type) {
	case map[string]any:
		for key, nested := range value {
			appendNumericStats(result, prefix+"_"+key, nested)
		}
	case []any:
		if len(value) > 0 {
			appendNumericStats(result, prefix, value[len(value)-1])
		}
	case float64:
		*result = append(*result, metricStat{Name: prefix, Value: value})
	case json.Number:
		if parsed, err := value.Float64(); err == nil {
			*result = append(*result, metricStat{Name: prefix, Value: parsed})
		}
	case bool:
		if value {
			*result = append(*result, metricStat{Name: prefix, Value: 1})
		} else {
			*result = append(*result, metricStat{Name: prefix, Value: 0})
		}
	case string:
		if strings.EqualFold(value, "on") {
			*result = append(*result, metricStat{Name: prefix, Value: 1})
		} else if strings.EqualFold(value, "off") {
			*result = append(*result, metricStat{Name: prefix, Value: 0})
		} else if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			*result = append(*result, metricStat{Name: prefix, Value: parsed})
		}
	}
}

var _ nutanixClient = (*prismElementClient)(nil)
