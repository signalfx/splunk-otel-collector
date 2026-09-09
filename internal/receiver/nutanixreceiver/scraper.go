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
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/nutanixreceiver/internal/metadata"
)

type scraper struct {
	cfg       *Config
	client    nutanixClient
	logger    *zap.Logger
	startTime pcommon.Timestamp
}

type prismSnapshot struct {
	clusters          []nutanixCluster
	disks             []nutanixDisk
	hosts             []nutanixHost
	subnets           []nutanixSubnet
	storageContainers []nutanixStorageContainer
	vms               []nutanixVM
	volumeGroups      []nutanixVolumeGroup
	additional        []additionalMetric
}

func newScraper(settings receiver.Settings, cfg *Config) *scraper {
	return &scraper{cfg: cfg, logger: settings.Logger}
}

func (s *scraper) start(context.Context, component.Host) error {
	client, err := newPrismClient(s.cfg)
	if err != nil {
		return err
	}
	s.client = client
	s.startTime = pcommon.NewTimestampFromTime(time.Now())
	return nil
}

func (s *scraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	if s.client == nil {
		return pmetric.NewMetrics(), errors.New("nutanix client is not initialized")
	}

	snapshot, err := s.fetchSnapshot(ctx)
	if err != nil {
		return pmetric.NewMetrics(), err
	}

	builder := newMetricBuilder(s.client, s.startTime, s.cfg.APIVersion)
	builder.addSnapshot(snapshot, s.cfg)
	return builder.metrics, nil
}

func (s *scraper) fetchSnapshot(ctx context.Context) (prismSnapshot, error) {
	var snapshot prismSnapshot
	var err error

	if s.cfg.Metrics.Clusters.Enabled || s.cfg.Metrics.Disks.Enabled || s.cfg.Metrics.Networking.Enabled {
		snapshot.clusters, err = s.client.listClusters(ctx)
		if err != nil {
			if s.cfg.APIVersion == "v4" && isNotFoundError(err) {
				return snapshot, fmt.Errorf("v4 cluster API is unavailable on the configured endpoint; configure api_version v2.0 when endpoint is Prism Element: %w", err)
			}
			return snapshot, fmt.Errorf("failed to list clusters: %w", err)
		}
		snapshot.clusters = filterManagedClusters(snapshot.clusters)
		if s.cfg.Metrics.Clusters.Enabled {
			stats, statsErrors, statsErr := fetchStatsConcurrently(ctx, "cluster", snapshot.clusters, func(cluster nutanixCluster) string { return cluster.ID }, s.client.getClusterStats)
			if statsErr != nil {
				return snapshot, statsErr
			}
			for i := range snapshot.clusters {
				snapshot.clusters[i].Stats = stats[i]
			}
			for _, statsError := range statsErrors {
				s.warnPartialFailure(statsError)
			}
		}
	}

	if s.cfg.Metrics.Disks.Enabled {
		snapshot.disks, err = s.client.listDisks(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list disks: %w", err)
		}
		stats, statsErrors, statsErr := fetchStatsConcurrently(ctx, "disk", snapshot.disks, func(disk nutanixDisk) string { return disk.ID }, s.client.getDiskStats)
		if statsErr != nil {
			return snapshot, statsErr
		}
		for i := range snapshot.disks {
			snapshot.disks[i].Stats = stats[i]
		}
		for _, statsError := range statsErrors {
			s.warnPartialFailure(statsError)
		}
	}

	if s.cfg.Metrics.Networking.Enabled {
		snapshot.subnets, err = s.client.listSubnets(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list subnets: %w", err)
		}
	}

	if s.cfg.Metrics.Hosts.Enabled || s.cfg.Metrics.Clusters.Enabled {
		snapshot.hosts, err = s.client.listHosts(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list hosts: %w", err)
		}
		if s.cfg.Metrics.Hosts.Enabled {
			stats, statsErrors, statsErr := fetchStatsConcurrently(ctx, "host", snapshot.hosts, func(host nutanixHost) string { return host.ID }, s.client.getHostStats)
			if statsErr != nil {
				return snapshot, statsErr
			}
			for i := range snapshot.hosts {
				snapshot.hosts[i].Stats = stats[i]
			}
			for _, statsError := range statsErrors {
				s.warnPartialFailure(statsError)
			}
		}
	}

	if s.cfg.Metrics.StorageContainers.Enabled {
		snapshot.storageContainers, err = s.client.listStorageContainers(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list storage containers: %w", err)
		}
		stats, statsErrors, statsErr := fetchStatsConcurrently(ctx, "storage container", snapshot.storageContainers, func(container nutanixStorageContainer) string { return container.ID }, s.client.getStorageContainerStats)
		if statsErr != nil {
			return snapshot, statsErr
		}
		for i := range snapshot.storageContainers {
			snapshot.storageContainers[i].Stats = stats[i]
		}
		for _, statsError := range statsErrors {
			s.warnPartialFailure(statsError)
		}
	}

	if s.cfg.Metrics.VMs.Enabled || s.cfg.Metrics.Clusters.Enabled || s.cfg.Metrics.Hosts.Enabled || s.cfg.Metrics.DataProtection.Enabled {
		snapshot.vms, err = s.client.listVMs(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list vms: %w", err)
		}
		if s.cfg.Metrics.VMs.Enabled && s.cfg.APIVersion == "v4" {
			statsByVM, statsErr := s.client.listVMStats(ctx)
			if statsErr != nil {
				if ctx.Err() != nil {
					return snapshot, ctx.Err()
				}
				if !isSkippableStatsError(statsErr) {
					s.warnPartialFailure(fmt.Errorf("failed to list vm stats: %w", statsErr))
				}
				statsByVM = map[string][]metricStat{}
			}
			for i := range snapshot.vms {
				snapshot.vms[i].Stats = statsByVM[snapshot.vms[i].ID]
			}
		}
	}

	if s.cfg.Metrics.VolumeGroups.Enabled || s.cfg.Metrics.Clusters.Enabled {
		snapshot.volumeGroups, err = s.client.listVolumeGroups(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list volume groups: %w", err)
		}
		if s.cfg.Metrics.VolumeGroups.Enabled {
			stats, statsErrors, statsErr := fetchStatsConcurrently(ctx, "volume group", snapshot.volumeGroups, func(volumeGroup nutanixVolumeGroup) string { return volumeGroup.ID }, s.client.getVolumeGroupStats)
			if statsErr != nil {
				return snapshot, statsErr
			}
			for i := range snapshot.volumeGroups {
				snapshot.volumeGroups[i].Stats = stats[i]
			}
			for _, statsError := range statsErrors {
				s.warnPartialFailure(statsError)
			}
		}
	}

	additional, additionalErr := s.client.collectAdditionalMetrics(ctx, additionalMetricsRequest{
		DataProtection:    s.cfg.Metrics.DataProtection.Enabled,
		Files:             s.cfg.Metrics.Files.Enabled,
		Microsegmentation: s.cfg.Metrics.Microsegmentation.Enabled,
		Networking:        s.cfg.Metrics.Networking.Enabled,
		Objects:           s.cfg.Metrics.Objects.Enabled,
		PrismCentral:      s.cfg.Metrics.PrismCentral.Enabled,
		VMs:               snapshot.vms,
	})
	if additionalErr != nil {
		return snapshot, fmt.Errorf("failed to collect additional Prism Central metrics: %w", additionalErr)
	}
	snapshot.additional = additional.Metrics
	for _, partialErr := range additional.Errors {
		s.warnPartialFailure(partialErr)
	}

	return snapshot, nil
}

func (s *scraper) warnPartialFailure(err error) {
	if s.logger != nil {
		s.logger.Warn("Nutanix API collection partially failed", zap.Error(err))
	}
}

func fetchStatsConcurrently[T any](
	ctx context.Context,
	entityType string,
	entities []T,
	entityID func(T) string,
	fetch func(context.Context, T) ([]metricStat, error),
) ([][]metricStat, []error, error) {
	stats := make([][]metricStat, len(entities))
	errorsByIndex := make([]error, len(entities))
	sem := make(chan struct{}, v4StatsWorkers)
	var wg sync.WaitGroup
	for i := range entities {
		if err := ctx.Err(); err != nil {
			break
		}
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			value, err := fetch(ctx, entities[index])
			if err == nil {
				stats[index] = value
				return
			}
			if !isSkippableStatsError(err) && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				errorsByIndex[index] = fmt.Errorf("failed to get %s stats for %q: %w", entityType, entityID(entities[index]), err)
			}
		}(i)
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	return stats, compactErrors(errorsByIndex), nil
}

func isSkippableStatsError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "CLU-10008") ||
		strings.Contains(msg, "CLUSTERMGMT_SERVICE_NOT_SUPPORTED_ENTITY_ERROR") ||
		strings.Contains(msg, "VMM-30102") ||
		strings.Contains(msg, "VM_INVALID_ARGUMENT")
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "404") || strings.Contains(message, "resource not found")
}

type metricBuilder struct {
	metrics    pmetric.Metrics
	scope      pmetric.ScopeMetrics
	byName     map[string]pmetric.Metric
	apiVersion string
	startTime  pcommon.Timestamp
	now        pcommon.Timestamp
}

func newMetricBuilder(client nutanixClient, startTime pcommon.Timestamp, apiVersion string) *metricBuilder {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	attrs := rm.Resource().Attributes()
	attrs.PutStr("service.name", "nutanix-prism")
	attrs.PutStr("service.instance.id", fmt.Sprintf("%s:%d", client.serverAddress(), client.serverPort()))
	attrs.PutStr("server.address", client.serverAddress())
	attrs.PutInt("server.port", client.serverPort())
	attrs.PutStr("nutanix.prism.api.version", apiVersion)
	scope := rm.ScopeMetrics().AppendEmpty()
	scope.Scope().SetName(metadata.ScopeName)

	return &metricBuilder{
		metrics:    metrics,
		scope:      scope,
		byName:     map[string]pmetric.Metric{},
		startTime:  startTime,
		now:        pcommon.NewTimestampFromTime(time.Now()),
		apiVersion: apiVersion,
	}
}

func (b *metricBuilder) addSnapshot(snapshot prismSnapshot, cfg *Config) {
	if cfg.Metrics.Clusters.Enabled {
		b.addClusterMetrics(snapshot, cfg)
	}
	if cfg.Metrics.Disks.Enabled {
		b.addDiskMetrics(snapshot.disks)
	}
	if cfg.Metrics.Hosts.Enabled {
		b.addHostMetrics(snapshot.hosts, snapshot.vms, snapshot.disks, cfg.Metrics.Disks.Enabled)
	}
	if cfg.Metrics.Networking.Enabled {
		b.addSubnetMetrics(snapshot.subnets)
	}
	if cfg.Metrics.StorageContainers.Enabled {
		b.addStorageContainerMetrics(snapshot.storageContainers)
	}
	if cfg.Metrics.VMs.Enabled {
		b.addVMStats(snapshot.vms)
	}
	if cfg.Metrics.VolumeGroups.Enabled {
		b.addVolumeGroupMetrics(snapshot.volumeGroups)
	}
	b.addAdditionalMetrics(snapshot.additional)
}

func (b *metricBuilder) addAdditionalMetrics(metrics []additionalMetric) {
	for _, metric := range metrics {
		if len(metric.Stats) > 0 {
			b.addEntityStats(metric.Name, metric.Description, metric.Attributes, metric.Stats)
			continue
		}
		b.addGauge(metric.Name, metric.Description, metric.Unit, metric.Attributes, metric.Value)
	}
}

func (b *metricBuilder) addClusterMetrics(snapshot prismSnapshot, cfg *Config) {
	b.addGauge("nutanix.cluster.count", "Number of managed Nutanix clusters", "{cluster}", nil, float64(len(snapshot.clusters)))
	for _, cluster := range snapshot.clusters {
		attrs := clusterAttrs(cluster)
		b.addInfo("nutanix.cluster.info", "Nutanix cluster information", attrs)
		b.addEntityStats("nutanix.cluster.stat", "Nutanix cluster statistic", attrs, cluster.Stats)

		clusterVMs := filterVMsByCluster(snapshot.vms, cluster)
		clusterVGs := filterVolumeGroupsByCluster(snapshot.volumeGroups, cluster)
		clusterHosts := filterHostsByCluster(snapshot.hosts, cluster)
		b.addVMCounts(attrs, clusterVMs)
		b.addVolumeGroupCounts(attrs, clusterVGs)
		b.addGauge("nutanix.host.count", "Number of Nutanix hosts", "{host}", attrs, float64(len(clusterHosts)))
		if cfg.Metrics.StorageContainers.Enabled {
			b.addStorageContainerCounts(attrs, filterStorageContainersByCluster(snapshot.storageContainers, cluster))
		}
		if cfg.Metrics.Disks.Enabled {
			b.addDiskCounts(attrs, filterDisksByCluster(snapshot.disks, cluster))
		}
		if cfg.Metrics.Networking.Enabled {
			b.addSubnetCounts(attrs, filterSubnetsByCluster(snapshot.subnets, cluster))
		}
	}
}

func (b *metricBuilder) addHostMetrics(hosts []nutanixHost, vms []nutanixVM, disks []nutanixDisk, disksEnabled bool) {
	b.addGauge("nutanix.host.count", "Number of Nutanix hosts", "{host}", nil, float64(len(hosts)))
	for _, host := range hosts {
		attrs := hostAttrs(host)
		b.addEntityStats("nutanix.host.stat", "Nutanix host statistic", attrs, host.Stats)
		b.addHostVMCounts(attrs, filterPoweredOnVMsByHost(vms, host))
		if disksEnabled {
			b.addDiskCounts(attrs, filterDisksByHost(disks, host))
		}
	}
}

func (b *metricBuilder) addStorageContainerMetrics(storageContainers []nutanixStorageContainer) {
	b.addStorageContainerCounts(nil, storageContainers)
	for _, storageContainer := range storageContainers {
		attrs := storageContainerAttrs(storageContainer)
		b.addEntityStats("nutanix.storage.container.stat", "Nutanix storage container statistic", attrs, storageContainer.Stats)
	}
}

func (b *metricBuilder) addStorageContainerCounts(baseAttrs map[string]string, storageContainers []nutanixStorageContainer) {
	b.addGauge("nutanix.storage.container.count", "Number of Nutanix storage containers", "{storage_container}", baseAttrs, float64(len(storageContainers)))
	for _, encrypted := range []string{"true", "false"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.storage.container.encrypted"] = encrypted
		b.addGauge("nutanix.storage.container.count", "Number of Nutanix storage containers", "{storage_container}", attrs, float64(countStorageContainers(storageContainers, encrypted, 0)))
	}
	for _, replicationFactor := range []int{1, 2, 3} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.storage.container.replication_factor"] = strconv.Itoa(replicationFactor)
		b.addGauge("nutanix.storage.container.count", "Number of Nutanix storage containers", "{storage_container}", attrs, float64(countStorageContainers(storageContainers, "", replicationFactor)))
	}
}

func (b *metricBuilder) addVMStats(vms []nutanixVM) {
	for i := range vms {
		attrs := vmAttrs(vms[i])
		b.addEntityStats("nutanix.vm.stat", "Nutanix VM statistic", attrs, vms[i].Stats)
	}
}

func (b *metricBuilder) addVolumeGroupMetrics(volumeGroups []nutanixVolumeGroup) {
	b.addVolumeGroupCounts(nil, volumeGroups)
	for _, volumeGroup := range volumeGroups {
		attrs := volumeGroupAttrs(volumeGroup)
		b.addEntityStats("nutanix.volume_group.stat", "Nutanix volume group statistic", attrs, volumeGroup.Stats)
	}
}

func (b *metricBuilder) addVolumeGroupCounts(baseAttrs map[string]string, volumeGroups []nutanixVolumeGroup) {
	b.addGauge("nutanix.volume_group.count", "Number of Nutanix volume groups", "{volume_group}", baseAttrs, float64(len(volumeGroups)))
	for _, status := range []string{"shared", "not_shared"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.volume_group.sharing_status"] = status
		b.addGauge("nutanix.volume_group.count", "Number of Nutanix volume groups", "{volume_group}", attrs, float64(countVolumeGroupsBySharingStatus(volumeGroups, status)))
	}
}

func (b *metricBuilder) addVMCounts(baseAttrs map[string]string, vms []nutanixVM) {
	b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", baseAttrs, float64(len(vms)))

	for _, powerState := range []string{"on", "off"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.vm.power_state"] = powerState
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", attrs, float64(countVMsByPowerState(vms, powerState)))
	}
	for _, bootType := range []string{"legacy", "uefi"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.vm.boot.type"] = bootType
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", attrs, float64(countVMsByString(vms, func(vm nutanixVM) string { return vm.BootType }, bootType)))
	}
	for _, protectionType := range []string{"unprotected", "pd_protected", "rule_protected"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.vm.protection.type"] = protectionType
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", attrs, float64(countVMsByString(vms, func(vm nutanixVM) string { return vm.ProtectionType }, protectionType)))
	}
	for _, present := range []bool{true, false} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.vm.gpu.present"] = strconv.FormatBool(present)
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", attrs, float64(countVMsByBool(vms, func(vm nutanixVM) bool { return vm.HasGPU }, present)))
	}
	for _, state := range []string{"installed", "enabled", "reachable", "vss_snapshot_capable"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.vm.guest_tools.state"] = state
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", attrs, float64(countVMsWithGuestToolsState(vms, state)))
	}

	b.addGauge("nutanix.vm.vcpu.count", "Number of vCPUs assigned to Nutanix VMs", "{vcpu}", baseAttrs, sumVMVCPUs(vms))
	b.addGauge("nutanix.vm.memory.assigned", "Memory assigned to Nutanix VMs", "By", baseAttrs, sumVMMemoryBytes(vms))
	b.addGauge("nutanix.vm.disk.count", "Number of Nutanix VM disks", "{disk}", baseAttrs, float64(countVMDisks(vms, "")))
	for _, bus := range []string{"ide", "sata", "scsi"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.vm.disk.bus"] = bus
		b.addGauge("nutanix.vm.disk.count", "Number of Nutanix VM disks", "{disk}", attrs, float64(countVMDisks(vms, bus)))
	}
	b.addGauge("nutanix.vm.nic.count", "Number of Nutanix VM NICs", "{nic}", baseAttrs, float64(countVMNICs(vms)))
}

func (b *metricBuilder) addDiskMetrics(disks []nutanixDisk) {
	b.addDiskCounts(nil, disks)
	for i := range disks {
		b.addEntityStats("nutanix.disk.stat", "Nutanix physical disk statistic", diskAttrs(disks[i]), disks[i].Stats)
	}
}

func (b *metricBuilder) addDiskCounts(baseAttrs map[string]string, disks []nutanixDisk) {
	b.addGauge("nutanix.disk.count", "Number of Nutanix physical disks", "{disk}", baseAttrs, float64(len(disks)))
	for _, tier := range []string{"ssd_pcie", "ssd_sata", "das_sata", "ssd_mem_nvme", "cloud"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.disk.storage_tier"] = tier
		b.addGauge("nutanix.disk.count", "Number of Nutanix physical disks", "{disk}", attrs, float64(countDisksByTier(disks, tier)))
	}
}

func (b *metricBuilder) addSubnetMetrics(subnets []nutanixSubnet) {
	b.addSubnetCounts(nil, subnets)
}

func (b *metricBuilder) addSubnetCounts(baseAttrs map[string]string, subnets []nutanixSubnet) {
	b.addGauge("nutanix.subnet.count", "Number of Nutanix subnets", "{subnet}", baseAttrs, float64(len(subnets)))
	for _, subnetType := range []string{"overlay", "vlan"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.subnet.type"] = subnetType
		b.addGauge("nutanix.subnet.count", "Number of Nutanix subnets", "{subnet}", attrs, float64(countSubnetsByString(subnets, func(subnet nutanixSubnet) string { return subnet.SubnetType }, subnetType)))
	}
	for _, mode := range []string{"advanced", "basic"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.subnet.type"] = "vlan"
		attrs["nutanix.subnet.networking_mode"] = mode
		b.addGauge("nutanix.subnet.count", "Number of Nutanix subnets", "{subnet}", attrs, float64(countSubnetsByNetworkingMode(filterSubnetsByType(subnets, "vlan"), mode)))
	}
	for _, external := range []bool{true, false} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.subnet.external"] = strconv.FormatBool(external)
		b.addGauge("nutanix.subnet.count", "Number of Nutanix subnets", "{subnet}", attrs, float64(countSubnetsByBool(subnets, func(subnet nutanixSubnet) *bool { return subnet.External }, external)))
	}
}

func (b *metricBuilder) addHostVMCounts(baseAttrs map[string]string, poweredOnVMs []nutanixVM) {
	b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", baseAttrs, float64(len(poweredOnVMs)))
	for _, powerState := range []string{"on", "off"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.vm.power_state"] = powerState
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", attrs, float64(countVMsByPowerState(poweredOnVMs, powerState)))
	}
	b.addGauge("nutanix.vm.vcpu.count", "Number of vCPUs assigned to Nutanix VMs", "{vcpu}", baseAttrs, sumVMVCPUs(poweredOnVMs))
	b.addGauge("nutanix.vm.memory.assigned", "Memory assigned to Nutanix VMs", "By", baseAttrs, sumVMMemoryBytes(poweredOnVMs))
	b.addGauge("nutanix.vm.disk.count", "Number of Nutanix VM disks", "{disk}", baseAttrs, float64(countVMDisks(poweredOnVMs, "")))
	for _, bus := range []string{"ide", "sata", "scsi"} {
		diskAttrs := cloneAttrs(baseAttrs)
		diskAttrs["nutanix.vm.disk.bus"] = bus
		b.addGauge("nutanix.vm.disk.count", "Number of Nutanix VM disks", "{disk}", diskAttrs, float64(countVMDisks(poweredOnVMs, bus)))
	}
	b.addGauge("nutanix.vm.nic.count", "Number of Nutanix VM NICs", "{nic}", baseAttrs, float64(countVMNICs(poweredOnVMs)))
	for _, bootType := range []string{"legacy", "uefi"} {
		stateAttrs := cloneAttrs(baseAttrs)
		stateAttrs["nutanix.vm.boot.type"] = bootType
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", stateAttrs, float64(countVMsByString(poweredOnVMs, func(vm nutanixVM) string { return vm.BootType }, bootType)))
	}
	for _, protectionType := range []string{"unprotected", "pd_protected", "rule_protected"} {
		stateAttrs := cloneAttrs(baseAttrs)
		stateAttrs["nutanix.vm.protection.type"] = protectionType
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", stateAttrs, float64(countVMsByString(poweredOnVMs, func(vm nutanixVM) string { return vm.ProtectionType }, protectionType)))
	}
	for _, present := range []bool{true, false} {
		stateAttrs := cloneAttrs(baseAttrs)
		stateAttrs["nutanix.vm.gpu.present"] = strconv.FormatBool(present)
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", stateAttrs, float64(countVMsByBool(poweredOnVMs, func(vm nutanixVM) bool { return vm.HasGPU }, present)))
	}
	for _, state := range []string{"installed", "enabled", "reachable", "vss_snapshot_capable"} {
		stateAttrs := cloneAttrs(baseAttrs)
		stateAttrs["nutanix.vm.guest_tools.state"] = state
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", stateAttrs, float64(countVMsWithGuestToolsState(poweredOnVMs, state)))
	}
}

func (b *metricBuilder) addEntityStats(metricName, description string, baseAttrs map[string]string, stats []metricStat) {
	statKind := "v4.stats"
	if b.apiVersion == "v2.0" {
		statKind = "v2.stats"
	}
	for _, stat := range stats {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.stat.name"] = sanitizeAttributeValue(stat.Name)
		attrs["nutanix.stat.kind"] = statKind
		b.addGauge(metricName, description, "", attrs, stat.Value)
	}
}

func (b *metricBuilder) addInfo(name, description string, attrs map[string]string) {
	b.addGauge(name, description, "1", attrs, 1)
}

func (b *metricBuilder) addGauge(name, description, unit string, attrs map[string]string, value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}

	metric, ok := b.byName[name]
	if !ok {
		metric = b.scope.Metrics().AppendEmpty()
		metric.SetName(name)
		metric.SetDescription(description)
		metric.SetUnit(unit)
		metric.SetEmptyGauge()
		b.byName[name] = metric
	}

	dp := metric.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(b.startTime)
	dp.SetTimestamp(b.now)
	dp.SetDoubleValue(value)
	for k, v := range attrs {
		if v != "" {
			dp.Attributes().PutStr(k, v)
		}
	}
}

func clusterAttrs(cluster nutanixCluster) map[string]string {
	return map[string]string{
		"nutanix.cluster.id":   cluster.ID,
		"nutanix.cluster.name": cluster.Name,
	}
}

func hostAttrs(host nutanixHost) map[string]string {
	return map[string]string{
		"nutanix.host.id":      host.ID,
		"nutanix.host.name":    host.Name,
		"nutanix.cluster.id":   host.ClusterID,
		"nutanix.cluster.name": host.ClusterName,
	}
}

func storageContainerAttrs(storageContainer nutanixStorageContainer) map[string]string {
	return map[string]string{
		"nutanix.storage.container.id":                 storageContainer.ID,
		"nutanix.storage.container.name":               storageContainer.Name,
		"nutanix.storage.container.encrypted":          boolString(storageContainer.Encrypted),
		"nutanix.storage.container.replication_factor": positiveIntString(storageContainer.ReplicationFactor),
		"nutanix.cluster.id":                           storageContainer.ClusterID,
		"nutanix.cluster.name":                         storageContainer.ClusterName,
	}
}

func vmAttrs(vm nutanixVM) map[string]string {
	return map[string]string{
		"nutanix.vm.id":          vm.ID,
		"nutanix.vm.name":        vm.Name,
		"nutanix.host.id":        vm.HostID,
		"nutanix.cluster.id":     vm.ClusterID,
		"nutanix.vm.power_state": vm.PowerState,
	}
}

func volumeGroupAttrs(volumeGroup nutanixVolumeGroup) map[string]string {
	return map[string]string{
		"nutanix.volume_group.id":             volumeGroup.ID,
		"nutanix.volume_group.name":           volumeGroup.Name,
		"nutanix.volume_group.sharing_status": volumeGroup.SharingStatus,
		"nutanix.cluster.id":                  volumeGroup.ClusterID,
	}
}

func diskAttrs(disk nutanixDisk) map[string]string {
	return map[string]string{
		"nutanix.disk.id":           disk.ID,
		"nutanix.disk.serial":       disk.Serial,
		"nutanix.disk.storage_tier": disk.StorageTier,
		"nutanix.host.id":           disk.HostID,
		"nutanix.host.name":         disk.HostName,
		"nutanix.cluster.id":        disk.ClusterID,
		"nutanix.cluster.name":      disk.ClusterName,
	}
}

func filterVMsByCluster(vms []nutanixVM, cluster nutanixCluster) []nutanixVM {
	if cluster.ID == "" {
		return nil
	}
	var filtered []nutanixVM
	for i := range vms {
		if vms[i].ClusterID == cluster.ID {
			filtered = append(filtered, vms[i])
		}
	}
	return filtered
}

func filterPoweredOnVMsByHost(vms []nutanixVM, host nutanixHost) []nutanixVM {
	if host.ID == "" {
		return nil
	}
	var filtered []nutanixVM
	for i := range vms {
		if vms[i].HostID == host.ID && vms[i].PowerState == "on" {
			filtered = append(filtered, vms[i])
		}
	}
	return filtered
}

func filterVolumeGroupsByCluster(volumeGroups []nutanixVolumeGroup, cluster nutanixCluster) []nutanixVolumeGroup {
	if cluster.ID == "" {
		return nil
	}
	var filtered []nutanixVolumeGroup
	for _, vg := range volumeGroups {
		if vg.ClusterID == cluster.ID {
			filtered = append(filtered, vg)
		}
	}
	return filtered
}

func filterHostsByCluster(hosts []nutanixHost, cluster nutanixCluster) []nutanixHost {
	if cluster.ID == "" {
		return nil
	}
	var filtered []nutanixHost
	for _, host := range hosts {
		if host.ClusterID == cluster.ID {
			filtered = append(filtered, host)
		}
	}
	return filtered
}

func filterStorageContainersByCluster(containers []nutanixStorageContainer, cluster nutanixCluster) []nutanixStorageContainer {
	if cluster.ID == "" {
		return nil
	}
	var filtered []nutanixStorageContainer
	for _, container := range containers {
		if container.ClusterID == cluster.ID {
			filtered = append(filtered, container)
		}
	}
	return filtered
}

func filterDisksByCluster(disks []nutanixDisk, cluster nutanixCluster) []nutanixDisk {
	if cluster.ID == "" {
		return nil
	}
	var filtered []nutanixDisk
	for i := range disks {
		if disks[i].ClusterID == cluster.ID {
			filtered = append(filtered, disks[i])
		}
	}
	return filtered
}

func filterDisksByHost(disks []nutanixDisk, host nutanixHost) []nutanixDisk {
	var filtered []nutanixDisk
	for i := range disks {
		if host.ID != "" && disks[i].HostID == host.ID {
			filtered = append(filtered, disks[i])
		}
	}
	return filtered
}

func filterSubnetsByCluster(subnets []nutanixSubnet, cluster nutanixCluster) []nutanixSubnet {
	if cluster.ID == "" {
		return nil
	}
	var filtered []nutanixSubnet
	for _, subnet := range subnets {
		for _, clusterID := range subnet.ClusterIDs {
			if clusterID == cluster.ID {
				filtered = append(filtered, subnet)
				break
			}
		}
	}
	return filtered
}

func filterSubnetsByType(subnets []nutanixSubnet, subnetType string) []nutanixSubnet {
	var filtered []nutanixSubnet
	for _, subnet := range subnets {
		if subnet.SubnetType == subnetType {
			filtered = append(filtered, subnet)
		}
	}
	return filtered
}

func filterManagedClusters(clusters []nutanixCluster) []nutanixCluster {
	filtered := make([]nutanixCluster, 0, len(clusters))
	for _, cluster := range clusters {
		if !isPrismCentralCluster(cluster) {
			filtered = append(filtered, cluster)
		}
	}
	return filtered
}

func countVMsByPowerState(vms []nutanixVM, powerState string) int {
	count := 0
	for i := range vms {
		if vms[i].PowerState == powerState {
			count++
		}
	}
	return count
}

func countVMsByString(vms []nutanixVM, value func(nutanixVM) string, expected string) int {
	count := 0
	for i := range vms {
		if value(vms[i]) == expected {
			count++
		}
	}
	return count
}

func countVMsByBool(vms []nutanixVM, value func(nutanixVM) bool, expected bool) int {
	count := 0
	for i := range vms {
		if value(vms[i]) == expected {
			count++
		}
	}
	return count
}

func countVMsWithGuestToolsState(vms []nutanixVM, state string) int {
	count := 0
	for i := range vms {
		var value *bool
		switch state {
		case "installed":
			value = vms[i].GuestTools.Installed
		case "enabled":
			value = vms[i].GuestTools.Enabled
		case "reachable":
			value = vms[i].GuestTools.Reachable
		case "vss_snapshot_capable":
			value = vms[i].GuestTools.VSSSnapshotCapable
		}
		if value != nil && *value {
			count++
		}
	}
	return count
}

func countStorageContainers(containers []nutanixStorageContainer, encrypted string, replicationFactor int) int {
	count := 0
	for _, container := range containers {
		if encrypted != "" && boolString(container.Encrypted) != encrypted {
			continue
		}
		if replicationFactor != 0 && container.ReplicationFactor != replicationFactor {
			continue
		}
		count++
	}
	return count
}

func countVolumeGroupsBySharingStatus(volumeGroups []nutanixVolumeGroup, status string) int {
	count := 0
	for _, volumeGroup := range volumeGroups {
		if volumeGroup.SharingStatus == status {
			count++
		}
	}
	return count
}

func countDisksByTier(disks []nutanixDisk, tier string) int {
	count := 0
	for i := range disks {
		if disks[i].StorageTier == tier {
			count++
		}
	}
	return count
}

func countSubnetsByString(subnets []nutanixSubnet, value func(nutanixSubnet) string, expected string) int {
	count := 0
	for _, subnet := range subnets {
		if value(subnet) == expected {
			count++
		}
	}
	return count
}

func countSubnetsByNetworkingMode(subnets []nutanixSubnet, mode string) int {
	count := 0
	for _, subnet := range subnets {
		if subnet.AdvancedNetworking == nil {
			continue
		}
		if (*subnet.AdvancedNetworking && mode == "advanced") || (!*subnet.AdvancedNetworking && mode == "basic") {
			count++
		}
	}
	return count
}

func countSubnetsByBool(subnets []nutanixSubnet, value func(nutanixSubnet) *bool, expected bool) int {
	count := 0
	for _, subnet := range subnets {
		actual := value(subnet)
		if actual != nil && *actual == expected {
			count++
		}
	}
	return count
}

func sumVMVCPUs(vms []nutanixVM) float64 {
	var total float64
	for i := range vms {
		total += float64(vms[i].NumSockets * vms[i].NumCoresPerSocket)
	}
	return total
}

func sumVMMemoryBytes(vms []nutanixVM) float64 {
	var total float64
	for i := range vms {
		total += float64(vms[i].MemoryBytes)
	}
	return total
}

func countVMDisks(vms []nutanixVM, bus string) int {
	count := 0
	for i := range vms {
		for _, diskBus := range vms[i].DiskBuses {
			if bus == "" || diskBus == bus {
				count++
			}
		}
	}
	return count
}

func countVMNICs(vms []nutanixVM) int {
	count := 0
	for i := range vms {
		count += vms[i].NICCount
	}
	return count
}
