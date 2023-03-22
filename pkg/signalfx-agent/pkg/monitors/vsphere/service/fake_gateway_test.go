package service

import (
	"fmt"
	"time"

	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

const fakeMetricKey = 42

var perfMetricSeriesValue = int64(111)

type fakeGateway struct {
	metricIDCounter     int32
	typeCounts          typeCounter
	numMetricsPerInvObj int
}

type typeCounter struct {
	cluster int
	compute int
	host    int
	vm      int
}

func newFakeGateway(numMetricsToFake int) *fakeGateway {
	return &fakeGateway{numMetricsPerInvObj: numMetricsToFake}
}

func (g *fakeGateway) topLevelFolderRef() types.ManagedObjectReference {
	return types.ManagedObjectReference{
		Value: "top",
	}
}

func (g *fakeGateway) retrievePerformanceManager() (*mo.PerformanceManager, error) {
	return &mo.PerformanceManager{
		PerfCounter: []types.PerfCounterInfo{{
			Key:       fakeMetricKey,
			GroupInfo: &types.ElementDescription{Key: "cpu"},
			NameInfo:  &types.ElementDescription{Key: "coreUtilization"},
			StatsType: "delta",
		}},
	}, nil
}

func (g *fakeGateway) retrieveRefProperties(mor types.ManagedObjectReference, dst interface{}) error {
	switch t := dst.(type) {
	case *mo.Folder:
		if mor.Value == "top" {
			t.Self = mor
			t.ChildEntity = []types.ManagedObjectReference{
				{Type: model.DatacenterType, Value: "dc-1"},
			}
		} else {
			t.Self = mor
			cluster := g.createRef(model.ClusterComputeType, "cluster", &g.typeCounts.cluster)
			freeStandingHost := g.createRef(model.ComputeType, "compute", &g.typeCounts.compute)
			t.ChildEntity = []types.ManagedObjectReference{cluster, freeStandingHost}
		}
	case *mo.ClusterComputeResource:
		t.Self = mor
		t.Name = "foo cluster"
		hostRef := g.createRef(model.HostType, "host", &g.typeCounts.host)
		t.ComputeResource.Host = []types.ManagedObjectReference{hostRef}
	case *mo.Datacenter:
		t.Self = mor
		t.Name = "foo dc"
	case *mo.HostSystem:
		t.Self = mor
		t.Name = "4.4.4.4"
		t.Config = &types.HostConfigInfo{
			Product: types.AboutInfo{
				OsType: "foo os type",
			},
		}
		vmRef := g.createRef(model.VMType, "vm", &g.typeCounts.vm)
		t.Vm = []types.ManagedObjectReference{vmRef}
	case *mo.VirtualMachine:
		t.Self = mor
		t.Name = "foo vm"
		t.Config = &types.VirtualMachineConfigInfo{
			GuestId: "foo guest id",
		}
		t.Guest = &types.GuestInfo{
			IpAddress:     "1.2.3.4",
			HostName:      "foo.host.name",
			GuestFamily:   "fooFam",
			GuestFullName: "fooFullName",
		}
	case *mo.ComputeResource:
		t.Self = mor
		hostRef := g.createRef(model.HostType, "freehost", &g.typeCounts.host)
		t.Host = []types.ManagedObjectReference{hostRef}
	default:
		return fmt.Errorf("type not found %v", t)
	}
	return nil
}

// noinspection GoUnusedParameter
func (g *fakeGateway) queryAvailablePerfMetric(ref types.ManagedObjectReference) (*types.QueryAvailablePerfMetricResponse, error) {
	counterID := g.metricIDCounter
	g.metricIDCounter++
	return &types.QueryAvailablePerfMetricResponse{
		Returnval: []types.PerfMetricId{
			{CounterId: counterID, Instance: fmt.Sprintf("instance-%d", counterID)},
		},
	}, nil
}

// noinspection GoUnusedParameter
func (g *fakeGateway) queryPerfProviderSummary(mor types.ManagedObjectReference) (*types.QueryPerfProviderSummaryResponse, error) {
	panic("implement me")
}

// noinspection GoUnusedParameter
func (g *fakeGateway) queryPerf(inv []*model.InventoryObject, maxSample int32) (*types.QueryPerfResponse, error) {
	var ret []types.BasePerfEntityMetricBase
	counter := 0
	for range inv {
		for i := 0; i < g.numMetricsPerInvObj; i++ {
			m := &types.PerfEntityMetric{
				Value: []types.BasePerfMetricSeries{
					&types.PerfMetricIntSeries{
						PerfMetricSeries: types.PerfMetricSeries{
							Id: types.PerfMetricId{
								CounterId: fakeMetricKey,
							},
						},
						Value: []int64{perfMetricSeriesValue},
					},
				},
				SampleInfo: []types.PerfSampleInfo{{Timestamp: time.Time{}}},
				PerfEntityMetricBase: types.PerfEntityMetricBase{
					Entity: types.ManagedObjectReference{Value: fmt.Sprintf("ref-%d", counter)},
				},
			}
			ret = append(ret, m)
			counter++
		}
	}
	return &types.QueryPerfResponse{Returnval: ret}, nil
}

func (g *fakeGateway) createRef(key string, prefix string, counter *int) types.ManagedObjectReference {
	out := types.ManagedObjectReference{Type: key, Value: fmt.Sprintf("%s-%d", prefix, *counter)}
	*counter++
	return out
}

func (g *fakeGateway) retrieveCurrentTime() (*time.Time, error) {
	panic("implement me")
}

func (g *fakeGateway) vcenterName() string {
	return "my-vc"
}
