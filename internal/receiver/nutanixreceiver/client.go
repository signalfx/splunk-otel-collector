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
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	clusterConfig "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/config"
	clusterRequests "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/request/clusters"
	diskRequests "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/request/disks"
	storageContainerRequests "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/request/storagecontainers"
	clusterStatsCommon "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/common/v1/stats"
	networkConfig "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/config"
	subnetRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/subnets"
	vmStatsCommon "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/common/v1/stats"
	vmConfig "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/ahv/config"
	vmStatsRequests "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/request/stats"
	vmRequests "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/request/vm"
	volumeStatsCommon "github.com/nutanix/ntnx-api-golang-clients/volumes-go-client/v4/models/common/v1/stats"
	volumeConfig "github.com/nutanix/ntnx-api-golang-clients/volumes-go-client/v4/models/volumes/v4/config"
	volumeGroupRequests "github.com/nutanix/ntnx-api-golang-clients/volumes-go-client/v4/models/volumes/v4/request/volumegroups"
)

type nutanixClient interface {
	serverAddress() string
	serverPort() int64
	listClusters(context.Context) ([]nutanixCluster, error)
	listHosts(context.Context) ([]nutanixHost, error)
	listStorageContainers(context.Context) ([]nutanixStorageContainer, error)
	listVMs(context.Context) ([]nutanixVM, error)
	listVolumeGroups(context.Context) ([]nutanixVolumeGroup, error)
	listDisks(context.Context) ([]nutanixDisk, error)
	listSubnets(context.Context) ([]nutanixSubnet, error)
	getClusterStats(context.Context, nutanixCluster) ([]metricStat, error)
	getHostStats(context.Context, nutanixHost) ([]metricStat, error)
	getStorageContainerStats(context.Context, nutanixStorageContainer) ([]metricStat, error)
	listVMStats(context.Context) (map[string][]metricStat, error)
	getVolumeGroupStats(context.Context, nutanixVolumeGroup) ([]metricStat, error)
	getDiskStats(context.Context, nutanixDisk) ([]metricStat, error)
	collectAdditionalMetrics(context.Context, additionalMetricsRequest) (additionalSnapshot, error)
}

func newPrismClient(cfg *Config) (nutanixClient, error) {
	if cfg.APIVersion == "v2.0" {
		return newPrismElementClient(cfg)
	}
	return newPrismV4Client(cfg)
}

func normalizeEndpoint(endpoint string, port int) (*url.URL, error) {
	endpoint = strings.TrimSpace(endpoint)
	if !strings.Contains(endpoint, "://") {
		endpoint = "https://" + endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		return nil, fmt.Errorf("endpoint %q must include a host", endpoint)
	}
	if u.Port() == "" {
		u.Host = net.JoinHostPort(u.Hostname(), strconv.Itoa(port))
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u, nil
}

func (c *prismClient) serverAddress() string {
	return c.baseURL.Hostname()
}

func (c *prismClient) serverPort() int64 {
	return serverPort(c.baseURL)
}

func serverPort(u *url.URL) int64 {
	port, err := strconv.ParseInt(u.Port(), 10, 64)
	if err != nil {
		return 0
	}
	return port
}

func (c *prismClient) listClusters(ctx context.Context) ([]nutanixCluster, error) {
	clusters, err := listSDKModels(ctx, "clusters", func(ctx context.Context, page, limit int) (any, error) {
		return c.clusters.ListClusters(ctx, &clusterRequests.ListClustersRequest{Page_: &page, Limit_: &limit})
	}, clusterFromV4)
	if err != nil {
		return nil, err
	}
	result := make([]nutanixCluster, 0, len(clusters))
	for _, cluster := range clusters {
		if !isPrismCentralCluster(cluster) {
			result = append(result, cluster)
		}
	}
	return result, nil
}

func (c *prismClient) listHosts(ctx context.Context) ([]nutanixHost, error) {
	return listSDKModels(ctx, "hosts", func(ctx context.Context, page, limit int) (any, error) {
		return c.clusters.ListHosts(ctx, &clusterRequests.ListHostsRequest{Page_: &page, Limit_: &limit})
	}, hostFromV4)
}

func (c *prismClient) listStorageContainers(ctx context.Context) ([]nutanixStorageContainer, error) {
	return listSDKModels(ctx, "storage containers", func(ctx context.Context, page, limit int) (any, error) {
		return c.storageContainers.ListStorageContainers(ctx, &storageContainerRequests.ListStorageContainersRequest{Page_: &page, Limit_: &limit})
	}, storageContainerFromV4)
}

func (c *prismClient) listVMs(ctx context.Context) ([]nutanixVM, error) {
	return listSDKModels(ctx, "virtual machines", func(ctx context.Context, page, limit int) (any, error) {
		return c.virtualMachines.ListVms(ctx, &vmRequests.ListVmsRequest{Page_: &page, Limit_: &limit})
	}, vmFromV4)
}

func (c *prismClient) listVolumeGroups(ctx context.Context) ([]nutanixVolumeGroup, error) {
	return listSDKModels(ctx, "volume groups", func(ctx context.Context, page, limit int) (any, error) {
		return c.volumeGroups.ListVolumeGroups(ctx, &volumeGroupRequests.ListVolumeGroupsRequest{Page_: &page, Limit_: &limit})
	}, volumeGroupFromV4)
}

func (c *prismClient) listDisks(ctx context.Context) ([]nutanixDisk, error) {
	return listSDKModels(ctx, "disks", func(ctx context.Context, page, limit int) (any, error) {
		return c.disks.ListDisks(ctx, &diskRequests.ListDisksRequest{Page_: &page, Limit_: &limit})
	}, diskFromV4)
}

func (c *prismClient) listSubnets(ctx context.Context) ([]nutanixSubnet, error) {
	return listSDKModels(ctx, "subnets", func(ctx context.Context, page, limit int) (any, error) {
		return c.subnets.ListSubnets(ctx, &subnetRequests.ListSubnetsRequest{Page_: &page, Limit_: &limit})
	}, subnetFromV4)
}

func (c *prismClient) getClusterStats(ctx context.Context, cluster nutanixCluster) ([]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if cluster.ID == "" {
		return nil, nil
	}
	start, end, sampling := c.statQuery()
	statType := clusterStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	response, err := c.clusters.GetClusterStats(ctx, &clusterRequests.GetClusterStatsRequest{
		ExtId:             &cluster.ID,
		StartTime_:        &start,
		EndTime_:          &end,
		SamplingInterval_: &sampling,
		StatType_:         &statType,
	})
	return statsFromSDKResponse(response, err)
}

func (c *prismClient) getHostStats(ctx context.Context, host nutanixHost) ([]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if host.ClusterID == "" || host.ID == "" {
		return nil, nil
	}
	start, end, sampling := c.statQuery()
	statType := clusterStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	response, err := c.clusters.GetHostStats(ctx, &clusterRequests.GetHostStatsRequest{
		ClusterExtId:      &host.ClusterID,
		ExtId:             &host.ID,
		StartTime_:        &start,
		EndTime_:          &end,
		SamplingInterval_: &sampling,
		StatType_:         &statType,
	})
	return statsFromSDKResponse(response, err)
}

func (c *prismClient) getStorageContainerStats(ctx context.Context, container nutanixStorageContainer) ([]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if container.ID == "" {
		return nil, nil
	}
	start, end, sampling := c.statQuery()
	statType := clusterStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	response, err := c.storageContainers.GetStorageContainerStats(ctx, &storageContainerRequests.GetStorageContainerStatsRequest{
		ExtId:             &container.ID,
		StartTime_:        &start,
		EndTime_:          &end,
		SamplingInterval_: &sampling,
		StatType_:         &statType,
	})
	return statsFromSDKResponse(response, err)
}

func (c *prismClient) listVMStats(ctx context.Context) (map[string][]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	start, end, sampling := c.statQuery()
	statType := vmStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	result := map[string][]metricStat{}
	page := 0
	processed := 0

	for {
		limit := v4PageSize
		response, err := c.virtualMachineStats.ListVmStats(ctx, &vmStatsRequests.ListVmStatsRequest{
			StartTime_:        &start,
			EndTime_:          &end,
			SamplingInterval_: &sampling,
			StatType_:         &statType,
			Page_:             &page,
			Limit_:            &limit,
		})
		if err != nil {
			return nil, err
		}
		responseMap, err := sdkResponseMap(response)
		if err != nil {
			return nil, fmt.Errorf("decode virtual machine stats response: %w", err)
		}
		items := responseDataMaps(responseMap)
		for _, item := range items {
			vmID := firstString(item, "extId", "ext_id", "id")
			stats := mapList(item, "stats")
			if vmID == "" || len(stats) == 0 {
				continue
			}
			result[vmID] = statsFromMap(stats[len(stats)-1])
		}
		processed += len(items)
		total, hasTotal := responseTotal(responseMap)
		if len(items) == 0 || (hasTotal && processed >= total) || (!hasTotal && len(items) < v4PageSize) {
			break
		}
		page++
	}

	return result, nil
}

func (c *prismClient) getVolumeGroupStats(ctx context.Context, volumeGroup nutanixVolumeGroup) ([]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if volumeGroup.ID == "" {
		return nil, nil
	}
	start, end, sampling := c.statQuery()
	statType := volumeStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	response, err := c.volumeGroups.GetVolumeGroupStats(ctx, &volumeGroupRequests.GetVolumeGroupStatsRequest{
		ExtId:             &volumeGroup.ID,
		StartTime_:        &start,
		EndTime_:          &end,
		SamplingInterval_: &sampling,
		StatType_:         &statType,
	})
	return statsFromSDKResponse(response, err)
}

func (c *prismClient) getDiskStats(ctx context.Context, disk nutanixDisk) ([]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if disk.ID == "" {
		return nil, nil
	}
	start, end, sampling := c.statQuery()
	statType := clusterStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	response, err := c.disks.GetDiskStats(ctx, &diskRequests.GetDiskStatsRequest{
		ExtId:             &disk.ID,
		StartTime_:        &start,
		EndTime_:          &end,
		SamplingInterval_: &sampling,
		StatType_:         &statType,
	})
	return statsFromSDKResponse(response, err)
}

type sdkPageFetcher func(context.Context, int, int) (any, error)

func listSDKModels[T, R any](ctx context.Context, name string, fetch sdkPageFetcher, convert func(T) R) ([]R, error) {
	entities, err := listSDKEntities(ctx, name, fetch)
	if err != nil {
		return nil, err
	}
	result := make([]R, 0, len(entities))
	for i, entity := range entities {
		data, err := json.Marshal(entity)
		if err != nil {
			return nil, fmt.Errorf("encode %s response item %d: %w", name, i, err)
		}
		var model T
		if err := json.Unmarshal(data, &model); err != nil {
			return nil, fmt.Errorf("decode %s response item %d: %w", name, i, err)
		}
		result = append(result, convert(model))
	}
	return result, nil
}

func listSDKEntities(ctx context.Context, name string, fetch sdkPageFetcher) ([]map[string]any, error) {
	var result []map[string]any
	for page := 0; ; page++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		response, err := fetch(ctx, page, v4PageSize)
		if err != nil {
			return nil, err
		}
		responseMap, err := sdkResponseMap(response)
		if err != nil {
			return nil, fmt.Errorf("decode %s response: %w", name, err)
		}
		entities := responseDataMaps(responseMap)
		result = append(result, entities...)
		total, hasTotal := responseTotal(responseMap)
		if len(entities) == 0 || (hasTotal && len(result) >= total) || len(entities) < v4PageSize {
			return result, nil
		}
	}
}

func sdkResponseMap(response any) (map[string]any, error) {
	data, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func statsFromSDKResponse(response any, callErr error) ([]metricStat, error) {
	if callErr != nil {
		return nil, callErr
	}
	responseMap, err := sdkResponseMap(response)
	if err != nil {
		return nil, err
	}
	items := responseDataMaps(responseMap)
	if len(items) == 0 {
		return nil, nil
	}
	return statsFromMap(items[len(items)-1]), nil
}

func (c *prismClient) statQuery() (time.Time, time.Time, int) {
	interval := c.interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	end := time.Now()
	start := end.Add(-interval)
	sampling := int(math.Max(1, interval.Seconds()))
	return start, end, sampling
}

func clusterFromV4(cluster clusterConfig.Cluster) nutanixCluster {
	result := nutanixCluster{
		ID:   stringValue(cluster.ExtId),
		Name: stringValue(cluster.Name),
	}
	if cluster.Config != nil {
		result.Functions = make([]string, 0, len(cluster.Config.ClusterFunction))
		for _, function := range cluster.Config.ClusterFunction {
			result.Functions = append(result.Functions, function.GetName())
		}
	}
	return result
}

func hostFromV4(host clusterConfig.Host) nutanixHost {
	result := nutanixHost{
		ID:   stringValue(host.ExtId),
		Name: stringValue(host.HostName),
	}
	if host.Cluster != nil {
		result.ClusterID = stringValue(host.Cluster.Uuid)
		result.ClusterName = stringValue(host.Cluster.Name)
	}
	return result
}

func storageContainerFromV4(container clusterConfig.StorageContainer) nutanixStorageContainer {
	id := stringValue(container.ExtId)
	if id == "" {
		id = stringValue(container.ContainerExtId)
	}
	return nutanixStorageContainer{
		ID:                id,
		Name:              stringValue(container.Name),
		ClusterID:         stringValue(container.ClusterExtId),
		ClusterName:       stringValue(container.ClusterName),
		Encrypted:         container.IsEncrypted,
		ReplicationFactor: intValue(container.ReplicationFactor),
	}
}

func vmFromV4(vm vmConfig.Vm) nutanixVM {
	result := nutanixVM{
		ID:                stringValue(vm.ExtId),
		Name:              stringValue(vm.Name),
		NumSockets:        intValue(vm.NumSockets),
		NumCoresPerSocket: intValue(vm.NumCoresPerSocket),
		MemoryBytes:       int64Value(vm.MemorySizeBytes),
		NICCount:          len(vm.Nics),
		HasGPU:            len(vm.Gpus) > 0,
	}
	if vm.Cluster != nil {
		result.ClusterID = stringValue(vm.Cluster.ExtId)
	}
	if vm.Host != nil {
		result.HostID = stringValue(vm.Host.ExtId)
	}
	if vm.PowerState != nil {
		result.PowerState = normalizeEnumName(vm.PowerState.GetName())
	}
	if vm.BootConfig != nil {
		switch vm.BootConfig.GetValue().(type) {
		case *vmConfig.LegacyBoot, vmConfig.LegacyBoot:
			result.BootType = "legacy"
		case *vmConfig.UefiBoot, vmConfig.UefiBoot:
			result.BootType = "uefi"
		}
	}
	if vm.ProtectionType != nil {
		result.ProtectionType = normalizeEnumName(vm.ProtectionType.GetName())
	}
	if vm.ProtectionPolicyState != nil && vm.ProtectionPolicyState.Policy != nil {
		result.ProtectionPolicyID = stringValue(vm.ProtectionPolicyState.Policy.ExtId)
	}
	if vm.GuestTools != nil {
		result.GuestTools = nutanixGuestTools{
			Installed:          vm.GuestTools.IsInstalled,
			Enabled:            vm.GuestTools.IsEnabled,
			Reachable:          vm.GuestTools.IsReachable,
			VSSSnapshotCapable: vm.GuestTools.IsVssSnapshotCapable,
		}
	}
	for _, disk := range vm.Disks {
		if disk.DiskAddress == nil || disk.DiskAddress.BusType == nil {
			result.DiskBuses = append(result.DiskBuses, "")
			continue
		}
		result.DiskBuses = append(result.DiskBuses, normalizeEnumName(disk.DiskAddress.BusType.GetName()))
	}
	return result
}

func volumeGroupFromV4(volumeGroup volumeConfig.VolumeGroup) nutanixVolumeGroup {
	result := nutanixVolumeGroup{
		ID:        stringValue(volumeGroup.ExtId),
		Name:      stringValue(volumeGroup.Name),
		ClusterID: stringValue(volumeGroup.ClusterReference),
	}
	if volumeGroup.SharingStatus != nil {
		result.SharingStatus = normalizeEnumName(volumeGroup.SharingStatus.GetName())
	}
	return result
}

func diskFromV4(disk clusterConfig.Disk) nutanixDisk {
	result := nutanixDisk{
		ID:          stringValue(disk.ExtId),
		Serial:      stringValue(disk.SerialNumber),
		ClusterID:   stringValue(disk.ClusterExtId),
		ClusterName: stringValue(disk.ClusterName),
		HostID:      stringValue(disk.NodeExtId),
		HostName:    stringValue(disk.HostName),
	}
	if disk.StorageTier != nil {
		result.StorageTier = normalizeEnumName(disk.StorageTier.GetName())
	}
	return result
}

func subnetFromV4(subnet networkConfig.Subnet) nutanixSubnet {
	result := nutanixSubnet{
		ID:                 stringValue(subnet.ExtId),
		Name:               stringValue(subnet.Name),
		AdvancedNetworking: subnet.IsAdvancedNetworking,
		External:           subnet.IsExternal,
	}
	if subnet.ClusterReference != nil {
		result.ClusterIDs = append(result.ClusterIDs, *subnet.ClusterReference)
	}
	result.ClusterIDs = append(result.ClusterIDs, subnet.ClusterReferenceList...)
	if subnet.SubnetType != nil {
		result.SubnetType = normalizeEnumName(subnet.SubnetType.GetName())
	}
	return result
}
