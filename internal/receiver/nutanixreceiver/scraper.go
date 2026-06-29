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
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
)

type scraper struct {
	cfg       *Config
	client    nutanixClient
	startTime pcommon.Timestamp
}

type prismSnapshot struct {
	clusters          []nutanixCluster
	hosts             []nutanixHost
	storageContainers []nutanixStorageContainer
	vms               []nutanixVM
	volumeGroups      []nutanixVolumeGroup
}

func newScraper(_ receiver.Settings, cfg *Config) *scraper {
	return &scraper{cfg: cfg}
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

	builder := newMetricBuilder(s.client, s.startTime)
	builder.addSnapshot(snapshot, s.cfg)
	return builder.metrics, nil
}

func (s *scraper) fetchSnapshot(ctx context.Context) (prismSnapshot, error) {
	var snapshot prismSnapshot
	var err error

	if s.cfg.Metrics.Clusters.Enabled {
		snapshot.clusters, err = s.client.listClusters(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list clusters: %w", err)
		}
		for i := range snapshot.clusters {
			stats, statsErr := s.client.getClusterStats(ctx, snapshot.clusters[i])
			if statsErr != nil {
				if isSkippableStatsError(statsErr) {
					continue
				}
				return snapshot, fmt.Errorf("failed to get cluster stats for %q: %w", snapshot.clusters[i].ID, statsErr)
			}
			snapshot.clusters[i].Stats = stats
		}
	}

	if s.cfg.Metrics.Hosts.Enabled || s.cfg.Metrics.Clusters.Enabled {
		snapshot.hosts, err = s.client.listHosts(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list hosts: %w", err)
		}
		if s.cfg.Metrics.Hosts.Enabled {
			for i := range snapshot.hosts {
				stats, statsErr := s.client.getHostStats(ctx, snapshot.hosts[i])
				if statsErr != nil {
					if isSkippableStatsError(statsErr) {
						continue
					}
					return snapshot, fmt.Errorf("failed to get host stats for %q: %w", snapshot.hosts[i].ID, statsErr)
				}
				snapshot.hosts[i].Stats = stats
			}
		}
	}

	if s.cfg.Metrics.StorageContainers.Enabled {
		snapshot.storageContainers, err = s.client.listStorageContainers(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list storage containers: %w", err)
		}
		for i := range snapshot.storageContainers {
			stats, statsErr := s.client.getStorageContainerStats(ctx, snapshot.storageContainers[i])
			if statsErr != nil {
				if isSkippableStatsError(statsErr) {
					continue
				}
				return snapshot, fmt.Errorf("failed to get storage container stats for %q: %w", snapshot.storageContainers[i].ID, statsErr)
			}
			snapshot.storageContainers[i].Stats = stats
		}
	}

	if s.cfg.Metrics.VMs.Enabled || s.cfg.Metrics.Clusters.Enabled || s.cfg.Metrics.Hosts.Enabled {
		snapshot.vms, err = s.client.listVMs(ctx)
		if err != nil {
			return snapshot, fmt.Errorf("failed to list vms: %w", err)
		}
		if s.cfg.Metrics.VMs.Enabled {
			statsByVM, statsErr := s.client.listVMStats(ctx)
			if statsErr != nil {
				if isSkippableStatsError(statsErr) {
					statsByVM = map[string][]metricStat{}
				} else {
					return snapshot, fmt.Errorf("failed to list vm stats: %w", statsErr)
				}
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
			for i := range snapshot.volumeGroups {
				stats, statsErr := s.client.getVolumeGroupStats(ctx, snapshot.volumeGroups[i])
				if statsErr != nil {
					if isSkippableStatsError(statsErr) {
						continue
					}
					return snapshot, fmt.Errorf("failed to get volume group stats for %q: %w", snapshot.volumeGroups[i].ID, statsErr)
				}
				snapshot.volumeGroups[i].Stats = stats
			}
		}
	}

	return snapshot, nil
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

type metricBuilder struct {
	metrics   pmetric.Metrics
	scope     pmetric.ScopeMetrics
	byName    map[string]pmetric.Metric
	startTime pcommon.Timestamp
	now       pcommon.Timestamp
}

func newMetricBuilder(client nutanixClient, startTime pcommon.Timestamp) *metricBuilder {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	attrs := rm.Resource().Attributes()
	attrs.PutStr("service.name", "nutanix-prism")
	attrs.PutStr("service.instance.id", fmt.Sprintf("%s:%d", client.serverAddress(), client.serverPort()))
	attrs.PutStr("server.address", client.serverAddress())
	attrs.PutInt("server.port", client.serverPort())
	attrs.PutStr("nutanix.prism.api.version", "v4")

	return &metricBuilder{
		metrics:   metrics,
		scope:     rm.ScopeMetrics().AppendEmpty(),
		byName:    map[string]pmetric.Metric{},
		startTime: startTime,
		now:       pcommon.NewTimestampFromTime(time.Now()),
	}
}

func (b *metricBuilder) addSnapshot(snapshot prismSnapshot, cfg *Config) {
	if cfg.Metrics.Clusters.Enabled {
		b.addClusterMetrics(snapshot)
	}
	if cfg.Metrics.Hosts.Enabled {
		b.addHostMetrics(snapshot.hosts, snapshot.vms)
	}
	if cfg.Metrics.StorageContainers.Enabled {
		b.addStorageContainerMetrics(snapshot.storageContainers)
	}
	if cfg.Metrics.VMs.Enabled {
		b.addVMStats(snapshot.vms)
	}
	if cfg.Metrics.VolumeGroups.Enabled {
		b.addVolumeGroupMetrics(snapshot.volumeGroups, cfg.Metrics.Clusters.Enabled)
	}
}

func (b *metricBuilder) addClusterMetrics(snapshot prismSnapshot) {
	for _, cluster := range snapshot.clusters {
		attrs := clusterAttrs(cluster)
		b.addInfo("nutanix.cluster.info", "Nutanix cluster information", attrs)
		b.addEntityStats("nutanix.cluster.stat", "Nutanix cluster Prism v4 statistic", attrs, cluster.Stats)

		clusterVMs := filterVMsByCluster(snapshot.vms, cluster)
		clusterVGs := filterVolumeGroupsByCluster(snapshot.volumeGroups, cluster)
		b.addVMCounts(attrs, clusterVMs)
		b.addGauge("nutanix.volume_group.count", "Number of Nutanix volume groups", "{volume_group}", attrs, float64(len(clusterVGs)))
	}
}

func (b *metricBuilder) addHostMetrics(hosts []nutanixHost, vms []nutanixVM) {
	for _, host := range hosts {
		attrs := hostAttrs(host)
		b.addEntityStats("nutanix.host.stat", "Nutanix host Prism v4 statistic", attrs, host.Stats)
		b.addHostVMCounts(attrs, filterPoweredOnVMsByHost(vms, host))
	}
}

func (b *metricBuilder) addStorageContainerMetrics(storageContainers []nutanixStorageContainer) {
	b.addGauge("nutanix.storage.container.count", "Number of Nutanix storage containers", "{storage_container}", nil, float64(len(storageContainers)))
	for _, storageContainer := range storageContainers {
		attrs := storageContainerAttrs(storageContainer)
		b.addEntityStats("nutanix.storage.container.stat", "Nutanix storage container Prism v4 statistic", attrs, storageContainer.Stats)
	}
}

func (b *metricBuilder) addVMStats(vms []nutanixVM) {
	for i := range vms {
		attrs := vmAttrs(vms[i])
		b.addEntityStats("nutanix.vm.stat", "Nutanix VM Prism v4 statistic", attrs, vms[i].Stats)
	}
}

func (b *metricBuilder) addVolumeGroupMetrics(volumeGroups []nutanixVolumeGroup, clustersEnabled bool) {
	if !clustersEnabled {
		b.addGauge("nutanix.volume_group.count", "Number of Nutanix volume groups", "{volume_group}", nil, float64(len(volumeGroups)))
	}
	for _, volumeGroup := range volumeGroups {
		attrs := volumeGroupAttrs(volumeGroup)
		b.addEntityStats("nutanix.volume_group.stat", "Nutanix volume group Prism v4 statistic", attrs, volumeGroup.Stats)
	}
}

func (b *metricBuilder) addVMCounts(baseAttrs map[string]string, vms []nutanixVM) {
	b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", baseAttrs, float64(len(vms)))

	for _, powerState := range []string{"on", "off"} {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.vm.power_state"] = powerState
		b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", attrs, float64(countVMsByPowerState(vms, powerState)))
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

func (b *metricBuilder) addHostVMCounts(baseAttrs map[string]string, poweredOnVMs []nutanixVM) {
	attrs := cloneAttrs(baseAttrs)
	attrs["nutanix.vm.power_state"] = "on"
	b.addGauge("nutanix.vm.count", "Number of Nutanix VMs", "{vm}", attrs, float64(len(poweredOnVMs)))
	b.addGauge("nutanix.vm.vcpu.count", "Number of vCPUs assigned to Nutanix VMs", "{vcpu}", baseAttrs, sumVMVCPUs(poweredOnVMs))
	b.addGauge("nutanix.vm.memory.assigned", "Memory assigned to Nutanix VMs", "By", baseAttrs, sumVMMemoryBytes(poweredOnVMs))
	b.addGauge("nutanix.vm.disk.count", "Number of Nutanix VM disks", "{disk}", baseAttrs, float64(countVMDisks(poweredOnVMs, "")))
	for _, bus := range []string{"ide", "sata", "scsi"} {
		diskAttrs := cloneAttrs(baseAttrs)
		diskAttrs["nutanix.vm.disk.bus"] = bus
		b.addGauge("nutanix.vm.disk.count", "Number of Nutanix VM disks", "{disk}", diskAttrs, float64(countVMDisks(poweredOnVMs, bus)))
	}
	b.addGauge("nutanix.vm.nic.count", "Number of Nutanix VM NICs", "{nic}", baseAttrs, float64(countVMNICs(poweredOnVMs)))
}

func (b *metricBuilder) addEntityStats(metricName, description string, baseAttrs map[string]string, stats []metricStat) {
	for _, stat := range stats {
		attrs := cloneAttrs(baseAttrs)
		attrs["nutanix.stat.name"] = sanitizeAttributeValue(stat.Name)
		attrs["nutanix.stat.kind"] = "v4.stats"
		b.addGauge(metricName, description, "1", attrs, stat.Value)
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
		"nutanix.storage.container.id":   storageContainer.ID,
		"nutanix.storage.container.name": storageContainer.Name,
		"nutanix.cluster.id":             storageContainer.ClusterID,
		"nutanix.cluster.name":           storageContainer.ClusterName,
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
		"nutanix.volume_group.id":   volumeGroup.ID,
		"nutanix.volume_group.name": volumeGroup.Name,
		"nutanix.cluster.id":        volumeGroup.ClusterID,
	}
}

func filterVMsByCluster(vms []nutanixVM, cluster nutanixCluster) []nutanixVM {
	if cluster.ID == "" {
		return vms
	}
	var filtered []nutanixVM
	for i := range vms {
		if vms[i].ClusterID == "" || vms[i].ClusterID == cluster.ID {
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
		return volumeGroups
	}
	var filtered []nutanixVolumeGroup
	for _, vg := range volumeGroups {
		if vg.ClusterID == "" || vg.ClusterID == cluster.ID {
			filtered = append(filtered, vg)
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
