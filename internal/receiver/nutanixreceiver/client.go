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
	"fmt"
	"math"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	v4converged "github.com/nutanix-cloud-native/prism-go-client/converged/v4"
	prismv4 "github.com/nutanix-cloud-native/prism-go-client/v4"
	clusterConfig "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/config"
	clusterStats "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/stats"
	clusterCommonStats "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/common/v1/stats"
	vmAPI "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/api"
	vmClient "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/client"
	vmCommonStats "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/common/v1/stats"
	vmConfig "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/ahv/config"
	vmStats "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/ahv/stats"
	volumeCommonStats "github.com/nutanix/ntnx-api-golang-clients/volumes-go-client/v4/models/common/v1/stats"
	volumeConfig "github.com/nutanix/ntnx-api-golang-clients/volumes-go-client/v4/models/volumes/v4/config"
	volumeStats "github.com/nutanix/ntnx-api-golang-clients/volumes-go-client/v4/models/volumes/v4/stats"
)

type nutanixClient interface {
	serverAddress() string
	serverPort() int64
	listClusters(context.Context) ([]nutanixCluster, error)
	listHosts(context.Context) ([]nutanixHost, error)
	listStorageContainers(context.Context) ([]nutanixStorageContainer, error)
	listVMs(context.Context) ([]nutanixVM, error)
	listVolumeGroups(context.Context) ([]nutanixVolumeGroup, error)
	getClusterStats(context.Context, nutanixCluster) ([]metricStat, error)
	getHostStats(context.Context, nutanixHost) ([]metricStat, error)
	getStorageContainerStats(context.Context, nutanixStorageContainer) ([]metricStat, error)
	listVMStats(context.Context) (map[string][]metricStat, error)
	getVolumeGroupStats(context.Context, nutanixVolumeGroup) ([]metricStat, error)
}

type prismClient struct {
	baseURL      *url.URL
	v4Client     *prismv4.Client
	services     *v4converged.Client
	vmStatsAPI   *vmAPI.StatsApi
	statInterval time.Duration
}

func newPrismClient(cfg *Config) (*prismClient, error) {
	baseURL, err := normalizeEndpoint(cfg.Endpoint, cfg.Port)
	if err != nil {
		return nil, err
	}

	credentials := prismgoclient.Credentials{
		Endpoint: baseURL.Host,
		Port:     baseURL.Port(),
		URL:      baseURL.String(),
		Username: cfg.Username,
		Password: string(cfg.Password),
		Insecure: cfg.TLS.InsecureSkipVerify,
	}

	v4Client, err := prismv4.NewV4Client(credentials)
	if err != nil {
		return nil, err
	}

	return &prismClient{
		baseURL:      baseURL,
		v4Client:     v4Client,
		services:     v4converged.NewClientFromV4SDKClient(v4Client),
		vmStatsAPI:   vmAPI.NewStatsApi(newVMAPIClient(baseURL, cfg, credentials)),
		statInterval: cfg.ControllerConfig.CollectionInterval,
	}, nil
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

func newVMAPIClient(baseURL *url.URL, cfg *Config, credentials prismgoclient.Credentials) *vmClient.ApiClient {
	apiClient := vmClient.NewApiClient()
	apiClient.Host = baseURL.Hostname()
	apiClient.Port = int(serverPort(baseURL))
	apiClient.VerifySSL = !cfg.TLS.InsecureSkipVerify
	apiClient.ReadTimeout = cfg.ControllerConfig.Timeout
	apiClient.ConnectTimeout = cfg.ControllerConfig.Timeout
	apiClient.SetUserName(credentials.Username)
	apiClient.SetPassword(credentials.Password)
	return apiClient
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
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	clusters, err := c.services.Clusters.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]nutanixCluster, 0, len(clusters))
	for _, cluster := range clusters {
		result = append(result, clusterFromV4(cluster))
	}
	return result, nil
}

func (c *prismClient) listHosts(ctx context.Context) ([]nutanixHost, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	hosts, err := c.services.Clusters.ListAllHosts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]nutanixHost, 0, len(hosts))
	for _, host := range hosts {
		result = append(result, hostFromV4(host))
	}
	return result, nil
}

func (c *prismClient) listStorageContainers(ctx context.Context) ([]nutanixStorageContainer, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	containers, err := c.services.StorageContainers.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]nutanixStorageContainer, 0, len(containers))
	for _, container := range containers {
		result = append(result, storageContainerFromV4(container))
	}
	return result, nil
}

func (c *prismClient) listVMs(ctx context.Context) ([]nutanixVM, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	vms, err := c.services.VMs.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]nutanixVM, 0, len(vms))
	for _, vm := range vms {
		result = append(result, vmFromV4(vm))
	}
	return result, nil
}

func (c *prismClient) listVolumeGroups(ctx context.Context) ([]nutanixVolumeGroup, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	volumeGroups, err := c.services.VolumeGroups.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]nutanixVolumeGroup, 0, len(volumeGroups))
	for _, volumeGroup := range volumeGroups {
		result = append(result, volumeGroupFromV4(volumeGroup))
	}
	return result, nil
}

func (c *prismClient) getClusterStats(ctx context.Context, cluster nutanixCluster) ([]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if cluster.ID == "" {
		return nil, nil
	}
	start, end, sampling := c.statQuery()
	statType := clusterCommonStats.DOWNSAMPLINGOPERATOR_LAST
	stats, err := v4converged.CallAPI[*clusterStats.ClusterStatsApiResponse, clusterStats.ClusterStats](
		c.v4Client.ClustersApiInstance.GetClusterStats(&cluster.ID, start, end, sampling, &statType, nil),
	)
	if err != nil {
		return nil, err
	}
	return statsFromStruct(stats), nil
}

func (c *prismClient) getHostStats(ctx context.Context, host nutanixHost) ([]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if host.ClusterID == "" || host.ID == "" {
		return nil, nil
	}
	start, end, sampling := c.statQuery()
	statType := clusterCommonStats.DOWNSAMPLINGOPERATOR_LAST
	stats, err := v4converged.CallAPI[*clusterStats.HostStatsApiResponse, clusterStats.HostStats](
		c.v4Client.ClustersApiInstance.GetHostStats(&host.ClusterID, &host.ID, start, end, sampling, &statType, nil),
	)
	if err != nil {
		return nil, err
	}
	return statsFromStruct(stats), nil
}

func (c *prismClient) getStorageContainerStats(ctx context.Context, container nutanixStorageContainer) ([]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if container.ID == "" {
		return nil, nil
	}
	start, end, sampling := c.statQuery()
	statType := clusterCommonStats.DOWNSAMPLINGOPERATOR_LAST
	stats, err := v4converged.CallAPI[*clusterStats.GetStorageContainerStatsApiResponse, clusterStats.StorageContainerStats](
		c.v4Client.StorageContainerAPI.GetStorageContainerStats(&container.ID, start, end, sampling, &statType),
	)
	if err != nil {
		return nil, err
	}
	return statsFromStruct(stats), nil
}

func (c *prismClient) listVMStats(ctx context.Context) (map[string][]metricStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	start, end, sampling := c.statQuery()
	statType := vmCommonStats.DOWNSAMPLINGOPERATOR_LAST
	result := map[string][]metricStat{}
	page := 0

	for {
		response, err := c.vmStatsAPI.ListVmStats(start, end, sampling, &statType, &page, nil, nil, nil, nil)
		items, total, err := v4converged.CallListAPI[*vmStats.ListVmStatsApiResponse, vmStats.VmStats](response, err)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item.ExtId == nil || len(item.Stats) == 0 {
				continue
			}
			result[*item.ExtId] = statsFromStruct(item.Stats[len(item.Stats)-1])
		}
		if len(result) >= total || len(items) == 0 {
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
	statType := volumeCommonStats.DOWNSAMPLINGOPERATOR_LAST
	stats, err := v4converged.CallAPI[*volumeStats.GetVolumeGroupStatsApiResponse, volumeStats.VolumeGroupStats](
		c.v4Client.VolumeGroupsApiInstance.GetVolumeGroupStats(&volumeGroup.ID, start, end, sampling, &statType, nil),
	)
	if err != nil {
		return nil, err
	}
	return statsFromStruct(stats), nil
}

func (c *prismClient) statQuery() (*time.Time, *time.Time, *int) {
	interval := c.statInterval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	end := time.Now()
	start := end.Add(-interval)
	sampling := int(math.Max(1, interval.Seconds()))
	return &start, &end, &sampling
}

func clusterFromV4(cluster clusterConfig.Cluster) nutanixCluster {
	return nutanixCluster{
		ID:   stringValue(cluster.ExtId),
		Name: stringValue(cluster.Name),
	}
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
		ID:          id,
		Name:        stringValue(container.Name),
		ClusterID:   stringValue(container.ClusterExtId),
		ClusterName: stringValue(container.ClusterName),
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
	return nutanixVolumeGroup{
		ID:        stringValue(volumeGroup.ExtId),
		Name:      stringValue(volumeGroup.Name),
		ClusterID: stringValue(volumeGroup.ClusterReference),
	}
}
