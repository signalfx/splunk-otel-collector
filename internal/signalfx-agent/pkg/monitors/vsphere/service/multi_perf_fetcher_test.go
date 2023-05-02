package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

func TestTooManyInvObjects(t *testing.T) {
	g := fakePaginatorGateway{}
	_, err := g.queryPerf(invObjs(100), 1)
	require.NotNil(t, err)
}

func TestNumPages(t *testing.T) {
	p := multiPagePerfFetcher{
		pageSize: 3,
	}
	require.Equal(t, 0, p.numPages(0))
	require.Equal(t, 1, p.numPages(1))
	require.Equal(t, 1, p.numPages(2))
	require.Equal(t, 1, p.numPages(3))
	require.Equal(t, 2, p.numPages(4))
}

func TestMultiPage1(t *testing.T) {
	p := multiPagePerfFetcher{pageSize: 4, gateway: &fakePaginatorGateway{}, log: testLog}
	it := p.invIterator(invObjs(4), 1)
	page, hasNext, err := it.nextInvPage()
	require.Nil(t, err)
	require.False(t, hasNext)
	require.Equal(t, 4, len(page))
}

func TestMultiPage2(t *testing.T) {
	p := multiPagePerfFetcher{pageSize: 4, gateway: &fakePaginatorGateway{}, log: testLog}
	it := p.invIterator(invObjs(5), 1)
	page, hasNext, err := it.nextInvPage()
	require.Nil(t, err)
	require.True(t, hasNext)
	require.Equal(t, 4, len(page))

	page, hasNext, err = it.nextInvPage()
	require.Nil(t, err)
	require.False(t, hasNext)
	require.Equal(t, 1, len(page))
}

func invObjs(n int) []*model.InventoryObject {
	var objs []*model.InventoryObject
	for i := 0; i < n; i++ {
		objs = append(objs, &model.InventoryObject{})
	}
	return objs
}

var _ IGateway = (*fakePaginatorGateway)(nil)

type fakePaginatorGateway struct{}

func (g *fakePaginatorGateway) retrievePerformanceManager() (*mo.PerformanceManager, error) {
	panic("implement me")
}

func (g *fakePaginatorGateway) topLevelFolderRef() types.ManagedObjectReference {
	panic("implement me")
}

func (g *fakePaginatorGateway) retrieveRefProperties(
	types.ManagedObjectReference,
	interface{},
) error {
	panic("implement me")
}

func (g *fakePaginatorGateway) queryAvailablePerfMetric(
	types.ManagedObjectReference,
) (*types.QueryAvailablePerfMetricResponse, error) {
	panic("implement me")
}

func (g *fakePaginatorGateway) queryPerfProviderSummary(types.ManagedObjectReference) (
	*types.QueryPerfProviderSummaryResponse, error,
) {
	panic("implement me")
}

func (g *fakePaginatorGateway) queryPerf(invObjs []*model.InventoryObject, maxSample int32) (*types.QueryPerfResponse, error) {
	// simulate api failure if too many inv objects are passed in
	if len(invObjs) > 10 {
		return nil, errors.New("too many inv objects")
	}
	var metrics []types.BasePerfEntityMetricBase
	// otherwise return one metric per inv object
	for range invObjs {
		metrics = append(metrics, &types.PerfEntityMetric{})
	}
	return &types.QueryPerfResponse{
		Returnval: metrics,
	}, nil
}

func (g *fakePaginatorGateway) retrieveCurrentTime() (*time.Time, error) {
	panic("implement me")
}

func (g *fakePaginatorGateway) vcenterName() string {
	panic("implement me")
}
