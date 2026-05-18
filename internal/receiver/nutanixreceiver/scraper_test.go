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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
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

type fakeNutanixClient struct {
	clusters          []nutanixCluster
	hosts             []nutanixHost
	storageContainers []nutanixStorageContainer
	vms               []nutanixVM
	volumeGroups      []nutanixVolumeGroup
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
			{ID: "container-1", Name: "container-a", ClusterID: "cluster-1", ClusterName: "cluster-a", Stats: []metricStat{
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
				Stats:             []metricStat{{Name: "hypervisorNumIops", Value: 3}},
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
				Stats:             []metricStat{{Name: "hypervisorNumIops", Value: 0}},
			},
		},
		volumeGroups: []nutanixVolumeGroup{
			{ID: "vg-1", Name: "vg-a", ClusterID: "cluster-1", Stats: []metricStat{{Name: "controllerNumIOPS", Value: 5}}},
		},
	}
}

func (f *fakeNutanixClient) serverAddress() string {
	return "prism.example.com"
}

func (f *fakeNutanixClient) serverPort() int64 {
	return 9440
}

func (f *fakeNutanixClient) listClusters(context.Context) ([]nutanixCluster, error) {
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

func (f *fakeNutanixClient) getClusterStats(_ context.Context, cluster nutanixCluster) ([]metricStat, error) {
	return cluster.Stats, nil
}

func (f *fakeNutanixClient) getHostStats(_ context.Context, host nutanixHost) ([]metricStat, error) {
	return host.Stats, nil
}

func (f *fakeNutanixClient) getStorageContainerStats(_ context.Context, storageContainer nutanixStorageContainer) ([]metricStat, error) {
	return storageContainer.Stats, nil
}

func (f *fakeNutanixClient) listVMStats(context.Context) (map[string][]metricStat, error) {
	stats := map[string][]metricStat{}
	for _, vm := range f.vms {
		stats[vm.ID] = vm.Stats
	}
	return stats, nil
}

func (f *fakeNutanixClient) getVolumeGroupStats(_ context.Context, volumeGroup nutanixVolumeGroup) ([]metricStat, error) {
	return volumeGroup.Stats, nil
}

func findGaugeValue(t *testing.T, md pmetric.Metrics, metricName string, attrs map[string]string) float64 {
	t.Helper()
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
						return dp.DoubleValue()
					}
				}
			}
		}
	}
	t.Fatalf("metric %q with attributes %v not found", metricName, attrs)
	return 0
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
