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
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/nutanixreceiver/internal/metadata"
)

func TestScraper(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = "prism.example.com"
	cfg.Username = "readonly"
	cfg.Password = "secret"

	s := newScraper(receivertest.NewNopSettings(receivertest.NopType), cfg)
	s.client = newFakeNutanixClient()
	s.startTime = pcommon.NewTimestampFromTime(time.Now())

	md, err := s.scrape(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, md.DataPointCount(), 25)

	resourceAttrs := md.ResourceMetrics().At(0).Resource().Attributes()
	require.Equal(t, "nutanix-prism", resourceAttrs.AsRaw()["service.name"])
	require.Equal(t, "prism.example.com", resourceAttrs.AsRaw()["server.address"])
	require.Equal(t, int64(9440), resourceAttrs.AsRaw()["server.port"])
	require.Equal(t, "v4", resourceAttrs.AsRaw()["nutanix.prism.api.version"])
	require.Equal(t, metadata.ScopeName, md.ResourceMetrics().At(0).ScopeMetrics().At(0).Scope().Name())

	require.InDelta(t, 42.0, findGaugeValue(t, md, "nutanix.cluster.stat", map[string]string{
		"nutanix.cluster.name": "cluster-a",
		"nutanix.stat.name":    "controllerNumIops",
		"nutanix.stat.kind":    "v4.stats",
	}), 0.001)

	require.InDelta(t, 2.0, findGaugeValue(t, md, "nutanix.vm.count", map[string]string{
		"nutanix.cluster.name": "cluster-a",
	}), 0.001)

	require.InDelta(t, 1.0, findGaugeValue(t, md, "nutanix.vm.count", map[string]string{
		"nutanix.host.name":      "host-a",
		"nutanix.vm.power_state": "on",
	}), 0.001)
	require.InDelta(t, 1.0, findGaugeValue(t, md, "nutanix.vm.count", map[string]string{
		"nutanix.host.name": "host-a",
	}), 0.001)

	require.InDelta(t, 6.0, findGaugeValue(t, md, "nutanix.vm.vcpu.count", map[string]string{
		"nutanix.cluster.name": "cluster-a",
	}), 0.001)

	require.InDelta(t, 6442450944.0, findGaugeValue(t, md, "nutanix.vm.memory.assigned", map[string]string{
		"nutanix.cluster.name": "cluster-a",
	}), 0.001)

	require.InDelta(t, 1.0, findGaugeValue(t, md, "nutanix.vm.disk.count", map[string]string{
		"nutanix.cluster.name": "cluster-a",
		"nutanix.vm.disk.bus":  "scsi",
	}), 0.001)

	require.InDelta(t, 1.0, findGaugeValue(t, md, "nutanix.storage.container.stat", map[string]string{
		"nutanix.storage.container.name": "container-a",
		"nutanix.stat.name":              "controllerNumIops",
	}), 0.001)

	require.InDelta(t, 5.0, findGaugeValue(t, md, "nutanix.volume_group.stat", map[string]string{
		"nutanix.volume_group.name": "vg-a",
		"nutanix.stat.name":         "controllerNumIOPS",
	}), 0.001)
}

func TestScraperExtendedInventoryMetrics(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = "prism.example.com"
	cfg.Username = "readonly"
	cfg.Password = "secret"
	cfg.Metrics.Disks.Enabled = true
	cfg.Metrics.Networking.Enabled = true

	s := newScraper(receivertest.NewNopSettings(receivertest.NopType), cfg)
	s.client = newFakeNutanixClient()
	s.startTime = pcommon.NewTimestampFromTime(time.Now())

	md, err := s.scrape(context.Background())
	require.NoError(t, err)

	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.cluster.count", nil), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.host.count", map[string]string{
		"nutanix.cluster.id": "cluster-1",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.vm.count", map[string]string{
		"nutanix.cluster.id":   "cluster-1",
		"nutanix.vm.boot.type": "uefi",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.vm.count", map[string]string{
		"nutanix.cluster.id":         "cluster-1",
		"nutanix.vm.protection.type": "rule_protected",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.vm.count", map[string]string{
		"nutanix.cluster.id":           "cluster-1",
		"nutanix.vm.guest_tools.state": "installed",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.vm.count", map[string]string{
		"nutanix.host.id":        "host-1",
		"nutanix.vm.gpu.present": "true",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.storage.container.count", map[string]string{
		"nutanix.storage.container.encrypted": "true",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.volume_group.count", map[string]string{
		"nutanix.volume_group.sharing_status": "shared",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.disk.count", map[string]string{
		"nutanix.disk.storage_tier": "ssd_pcie",
	}), 0.001)
	require.InDelta(t, 7, findGaugeValue(t, md, "nutanix.disk.stat", map[string]string{
		"nutanix.disk.id":   "disk-1",
		"nutanix.stat.name": "diskNumIops",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.subnet.count", map[string]string{
		"nutanix.cluster.id":  "cluster-1",
		"nutanix.subnet.type": "overlay",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.subnet.count", map[string]string{
		"nutanix.subnet.type":            "vlan",
		"nutanix.subnet.networking_mode": "basic",
	}), 0.001)
	require.InDelta(t, 1, findGaugeValue(t, md, "nutanix.subnet.count", map[string]string{
		"nutanix.subnet.external": "true",
	}), 0.001)
}

func TestScraperPrismElementUsesVMInventoryStats(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = "prism.example.com"
	cfg.APIVersion = "v2.0"
	cfg.Username = "readonly"
	cfg.Password = "secret"

	client := newFakeNutanixClient()
	s := newScraper(receivertest.NewNopSettings(receivertest.NopType), cfg)
	s.client = client
	s.startTime = pcommon.NewTimestampFromTime(time.Now())

	md, err := s.scrape(context.Background())
	require.NoError(t, err)
	require.Zero(t, client.vmStatsCalls)
	require.InDelta(t, 3, findGaugeValue(t, md, "nutanix.vm.stat", map[string]string{
		"nutanix.vm.id":     "vm-1",
		"nutanix.stat.name": "hypervisorNumIops",
	}), 0.001)
}

func TestScraperSkipsUnsupportedEntityStats(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = "prism.example.com"
	cfg.Username = "readonly"
	cfg.Password = "secret"

	client := newFakeNutanixClient()
	client.clusters = append(client.clusters, nutanixCluster{
		ID:        "pc-cluster",
		Name:      "prism-central",
		Functions: []string{"PRISM_CENTRAL"},
	})
	client.clusterStatsErrors = map[string]error{
		"pc-cluster": errors.New(`{"code":"CLU-10008","errorGroup":"CLUSTERMGMT_SERVICE_NOT_SUPPORTED_ENTITY_ERROR"}`),
	}

	s := newScraper(receivertest.NewNopSettings(receivertest.NopType), cfg)
	s.client = client
	s.startTime = pcommon.NewTimestampFromTime(time.Now())

	md, err := s.scrape(context.Background())
	require.NoError(t, err)

	require.InDelta(t, 42.0, findGaugeValue(t, md, "nutanix.cluster.stat", map[string]string{
		"nutanix.cluster.name": "cluster-a",
		"nutanix.stat.name":    "controllerNumIops",
	}), 0.001)
	require.False(t, hasGaugeDataPoint(md, "nutanix.cluster.info", map[string]string{
		"nutanix.cluster.name": "prism-central",
	}))
	require.False(t, hasGaugeDataPoint(md, "nutanix.cluster.stat", map[string]string{
		"nutanix.cluster.name": "prism-central",
	}))
}

func TestClusterFiltersRequireExactClusterReference(t *testing.T) {
	cluster := nutanixCluster{ID: "cluster-1"}
	vms := []nutanixVM{
		{ID: "matched", ClusterID: "cluster-1"},
		{ID: "unknown"},
		{ID: "other", ClusterID: "cluster-2"},
	}
	volumeGroups := []nutanixVolumeGroup{
		{ID: "matched", ClusterID: "cluster-1"},
		{ID: "unknown"},
		{ID: "other", ClusterID: "cluster-2"},
	}

	require.Equal(t, []nutanixVM{vms[0]}, filterVMsByCluster(vms, cluster))
	require.Equal(t, []nutanixVolumeGroup{volumeGroups[0]}, filterVolumeGroupsByCluster(volumeGroups, cluster))
}

func TestScraperContinuesAfterClusterStatsError(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = "prism.example.com"
	cfg.Username = "readonly"
	cfg.Password = "secret"

	client := newFakeNutanixClient()
	client.clusterStatsErrors = map[string]error{
		"cluster-1": errors.New("boom"),
	}

	s := newScraper(receivertest.NewNopSettings(receivertest.NopType), cfg)
	s.client = client
	s.startTime = pcommon.NewTimestampFromTime(time.Now())

	md, err := s.scrape(context.Background())
	require.NoError(t, err)
	require.True(t, hasGaugeDataPoint(md, "nutanix.cluster.info", map[string]string{
		"nutanix.cluster.id": "cluster-1",
	}))
	require.False(t, hasGaugeDataPoint(md, "nutanix.cluster.stat", map[string]string{
		"nutanix.cluster.id": "cluster-1",
	}))
}

func TestScraperExplainsPrismElementV4NotFound(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = "prism.example.com"
	cfg.Username = "readonly"
	cfg.Password = "secret"

	client := newFakeNutanixClient()
	client.clustersError = errors.New("HTTP/1.1 404 NOT FOUND: resource not found")
	s := newScraper(receivertest.NewNopSettings(receivertest.NopType), cfg)
	s.client = client

	_, err := s.scrape(context.Background())
	require.ErrorContains(t, err, "configure api_version v2.0 when endpoint is Prism Element")
}

func TestScraperSkipsInvalidVMStatsSelect(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = "prism.example.com"
	cfg.Username = "readonly"
	cfg.Password = "secret"

	client := newFakeNutanixClient()
	client.vmStatsError = errors.New(`{"code":"VMM-30102","errorGroup":"VM_INVALID_ARGUMENT","argumentsMap":{"argument_key":"$select","argument_value":"EMPTY"}}`)

	s := newScraper(receivertest.NewNopSettings(receivertest.NopType), cfg)
	s.client = client
	s.startTime = pcommon.NewTimestampFromTime(time.Now())

	md, err := s.scrape(context.Background())
	require.NoError(t, err)

	require.InDelta(t, 2.0, findGaugeValue(t, md, "nutanix.vm.count", map[string]string{
		"nutanix.cluster.name": "cluster-a",
	}), 0.001)
	require.False(t, hasGaugeDataPoint(md, "nutanix.vm.stat", map[string]string{
		"nutanix.vm.name": "vm-a",
	}))
}

func TestScraperContinuesAfterVMStatsError(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = "prism.example.com"
	cfg.Username = "readonly"
	cfg.Password = "secret"

	client := newFakeNutanixClient()
	client.vmStatsError = errors.New("boom")

	s := newScraper(receivertest.NewNopSettings(receivertest.NopType), cfg)
	s.client = client
	s.startTime = pcommon.NewTimestampFromTime(time.Now())

	md, err := s.scrape(context.Background())
	require.NoError(t, err)
	require.InDelta(t, 2, findGaugeValue(t, md, "nutanix.vm.count", map[string]string{
		"nutanix.cluster.id": "cluster-1",
	}), 0.001)
	require.False(t, hasGaugeDataPoint(md, "nutanix.vm.stat", map[string]string{
		"nutanix.vm.id": "vm-1",
	}))
}

func TestFetchStatsConcurrentlyIsBounded(t *testing.T) {
	entities := make([]int, 25)
	var active atomic.Int32
	var maximum atomic.Int32

	stats, errs, err := fetchStatsConcurrently(context.Background(), "test", entities, strconv.Itoa, func(context.Context, int) ([]metricStat, error) {
		current := active.Add(1)
		defer active.Add(-1)
		for {
			observed := maximum.Load()
			if current <= observed || maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		return []metricStat{{Name: "value", Value: 1}}, nil
	})

	require.NoError(t, err)
	require.Empty(t, errs)
	require.Len(t, stats, len(entities))
	require.Greater(t, maximum.Load(), int32(1))
	require.LessOrEqual(t, maximum.Load(), int32(v4StatsWorkers))
}

type fakeNutanixClient struct {
	clustersError      error
	vmStatsError       error
	clusterStatsErrors map[string]error
	additional         additionalSnapshot
	clusters           []nutanixCluster
	disks              []nutanixDisk
	hosts              []nutanixHost
	subnets            []nutanixSubnet
	storageContainers  []nutanixStorageContainer
	vms                []nutanixVM
	volumeGroups       []nutanixVolumeGroup
	vmStatsCalls       int
}

func newFakeNutanixClient() *fakeNutanixClient {
	return &fakeNutanixClient{
		clusters: []nutanixCluster{
			{ID: "cluster-1", Name: "cluster-a", Stats: []metricStat{
				{Name: "controllerNumIops", Value: 42},
				{Name: "hypervisorCpuUsagePpm", Value: 9000},
			}},
		},
		hosts: []nutanixHost{
			{ID: "host-1", Name: "host-a", ClusterID: "cluster-1", ClusterName: "cluster-a", Stats: []metricStat{
				{Name: "controllerNumIops", Value: 11},
				{Name: "memoryCapacityBytes", Value: 2048},
			}},
		},
		storageContainers: []nutanixStorageContainer{
			{ID: "container-1", Name: "container-a", ClusterID: "cluster-1", ClusterName: "cluster-a", Encrypted: boolRef(true), ReplicationFactor: 2, Stats: []metricStat{
				{Name: "controllerNumIops", Value: 1},
			}},
		},
		vms: []nutanixVM{
			{
				ID:                "vm-1",
				Name:              "vm-a",
				ClusterID:         "cluster-1",
				HostID:            "host-1",
				PowerState:        "on",
				NumSockets:        2,
				NumCoresPerSocket: 1,
				MemoryBytes:       4294967296,
				DiskBuses:         []string{"scsi"},
				NICCount:          1,
				BootType:          "uefi",
				HasGPU:            true,
				ProtectionType:    "rule_protected",
				GuestTools: nutanixGuestTools{
					Installed:          boolRef(true),
					Enabled:            boolRef(true),
					Reachable:          boolRef(true),
					VSSSnapshotCapable: boolRef(false),
				},
				Stats: []metricStat{{Name: "hypervisorNumIops", Value: 3}},
			},
			{
				ID:                "vm-2",
				Name:              "vm-b",
				ClusterID:         "cluster-1",
				HostID:            "host-1",
				PowerState:        "off",
				NumSockets:        2,
				NumCoresPerSocket: 2,
				MemoryBytes:       2147483648,
				DiskBuses:         []string{"sata"},
				NICCount:          0,
				BootType:          "legacy",
				ProtectionType:    "unprotected",
				Stats:             []metricStat{{Name: "hypervisorNumIops", Value: 0}},
			},
		},
		volumeGroups: []nutanixVolumeGroup{
			{ID: "vg-1", Name: "vg-a", ClusterID: "cluster-1", SharingStatus: "shared", Stats: []metricStat{{Name: "controllerNumIOPS", Value: 5}}},
		},
		disks: []nutanixDisk{
			{ID: "disk-1", Serial: "serial-1", ClusterID: "cluster-1", ClusterName: "cluster-a", HostID: "host-1", HostName: "host-a", StorageTier: "ssd_pcie", Stats: []metricStat{{Name: "diskNumIops", Value: 7}}},
		},
		subnets: []nutanixSubnet{
			{ID: "subnet-1", Name: "overlay-a", ClusterIDs: []string{"cluster-1"}, SubnetType: "overlay", AdvancedNetworking: boolRef(true), External: boolRef(false)},
			{ID: "subnet-2", Name: "vlan-a", ClusterIDs: []string{"cluster-1"}, SubnetType: "vlan", AdvancedNetworking: boolRef(false), External: boolRef(true)},
		},
	}
}

func boolRef(value bool) *bool {
	return &value
}

func (f *fakeNutanixClient) serverAddress() string {
	return "prism.example.com"
}

func (f *fakeNutanixClient) serverPort() int64 {
	return 9440
}

func (f *fakeNutanixClient) listClusters(context.Context) ([]nutanixCluster, error) {
	if f.clustersError != nil {
		return nil, f.clustersError
	}
	return f.clusters, nil
}

func (f *fakeNutanixClient) listHosts(context.Context) ([]nutanixHost, error) {
	return f.hosts, nil
}

func (f *fakeNutanixClient) listStorageContainers(context.Context) ([]nutanixStorageContainer, error) {
	return f.storageContainers, nil
}

func (f *fakeNutanixClient) listVMs(context.Context) ([]nutanixVM, error) {
	return f.vms, nil
}

func (f *fakeNutanixClient) listVolumeGroups(context.Context) ([]nutanixVolumeGroup, error) {
	return f.volumeGroups, nil
}

func (f *fakeNutanixClient) listDisks(context.Context) ([]nutanixDisk, error) {
	return f.disks, nil
}

func (f *fakeNutanixClient) listSubnets(context.Context) ([]nutanixSubnet, error) {
	return f.subnets, nil
}

func (f *fakeNutanixClient) getClusterStats(_ context.Context, cluster nutanixCluster) ([]metricStat, error) {
	if err := f.clusterStatsErrors[cluster.ID]; err != nil {
		return nil, err
	}
	return cluster.Stats, nil
}

func (f *fakeNutanixClient) getHostStats(_ context.Context, host nutanixHost) ([]metricStat, error) {
	return host.Stats, nil
}

func (f *fakeNutanixClient) getStorageContainerStats(_ context.Context, storageContainer nutanixStorageContainer) ([]metricStat, error) {
	return storageContainer.Stats, nil
}

func (f *fakeNutanixClient) listVMStats(context.Context) (map[string][]metricStat, error) {
	f.vmStatsCalls++
	if f.vmStatsError != nil {
		return nil, f.vmStatsError
	}
	stats := map[string][]metricStat{}
	for i := range f.vms {
		stats[f.vms[i].ID] = f.vms[i].Stats
	}
	return stats, nil
}

func (f *fakeNutanixClient) getVolumeGroupStats(_ context.Context, volumeGroup nutanixVolumeGroup) ([]metricStat, error) {
	return volumeGroup.Stats, nil
}

func (f *fakeNutanixClient) getDiskStats(_ context.Context, disk nutanixDisk) ([]metricStat, error) {
	return disk.Stats, nil
}

func (f *fakeNutanixClient) collectAdditionalMetrics(context.Context, additionalMetricsRequest) (additionalSnapshot, error) {
	return f.additional, nil
}

func findGaugeValue(t *testing.T, md pmetric.Metrics, metricName string, attrs map[string]string) float64 {
	t.Helper()
	if value, ok := findGaugeValueForAttrs(md, metricName, attrs); ok {
		return value
	}
	t.Fatalf("metric %q with attributes %v not found", metricName, attrs)
	return 0
}

func hasGaugeDataPoint(md pmetric.Metrics, metricName string, attrs map[string]string) bool {
	_, ok := findGaugeValueForAttrs(md, metricName, attrs)
	return ok
}

func findGaugeValueForAttrs(md pmetric.Metrics, metricName string, attrs map[string]string) (float64, bool) {
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				if metric.Name() != metricName {
					continue
				}
				for l := 0; l < metric.Gauge().DataPoints().Len(); l++ {
					dp := metric.Gauge().DataPoints().At(l)
					if hasAttrs(dp, attrs) {
						return dp.DoubleValue(), true
					}
				}
			}
		}
	}
	return 0, false
}

func hasAttrs(dp pmetric.NumberDataPoint, attrs map[string]string) bool {
	for k, want := range attrs {
		got, ok := dp.Attributes().Get(k)
		if !ok || got.Str() != want {
			return false
		}
	}
	return true
}
