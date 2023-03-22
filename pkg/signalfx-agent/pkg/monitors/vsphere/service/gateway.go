package service

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

// A thin wrapper around the vmomi SDK so that callers don't have to use it directly.
type IGateway interface {
	queryPerf(invObjs []*model.InventoryObject, maxSample int32) (*types.QueryPerfResponse, error)
	retrievePerformanceManager() (*mo.PerformanceManager, error)
	topLevelFolderRef() types.ManagedObjectReference
	retrieveRefProperties(mor types.ManagedObjectReference, dst interface{}) error
	queryAvailablePerfMetric(ref types.ManagedObjectReference) (*types.QueryAvailablePerfMetricResponse, error)
	queryPerfProviderSummary(mor types.ManagedObjectReference) (*types.QueryPerfProviderSummaryResponse, error)
	retrieveCurrentTime() (*time.Time, error)
	vcenterName() string
}

type Gateway struct {
	ctx    context.Context
	client *govmomi.Client
	log    log.FieldLogger
	vcName string
}

func NewGateway(ctx context.Context, client *govmomi.Client, log log.FieldLogger) *Gateway {
	return &Gateway{
		ctx:    ctx,
		client: client,
		log:    log,
		vcName: client.Client.URL().Host,
	}
}

func (g *Gateway) retrievePerformanceManager() (*mo.PerformanceManager, error) {
	var pm mo.PerformanceManager
	err := mo.RetrieveProperties(
		g.ctx,
		g.client,
		g.client.ServiceContent.PropertyCollector,
		*g.client.Client.ServiceContent.PerfManager,
		&pm,
	)
	return &pm, err
}

func (g *Gateway) topLevelFolderRef() types.ManagedObjectReference {
	return g.client.ServiceContent.RootFolder
}

func (g *Gateway) retrieveRefProperties(ref types.ManagedObjectReference, dst interface{}) error {
	return mo.RetrieveProperties(
		g.ctx,
		g.client,
		g.client.ServiceContent.PropertyCollector,
		ref,
		dst,
	)
}

func (g *Gateway) queryAvailablePerfMetric(ref types.ManagedObjectReference) (*types.QueryAvailablePerfMetricResponse, error) {
	req := types.QueryAvailablePerfMetric{
		This:       *g.client.Client.ServiceContent.PerfManager,
		Entity:     ref,
		IntervalId: model.RealtimeMetricsInterval,
	}
	return methods.QueryAvailablePerfMetric(g.ctx, g.client.Client, &req)
}

func (g *Gateway) queryPerfProviderSummary(ref types.ManagedObjectReference) (*types.QueryPerfProviderSummaryResponse, error) {
	req := types.QueryPerfProviderSummary{
		This:   *g.client.Client.ServiceContent.PerfManager,
		Entity: ref,
	}
	return methods.QueryPerfProviderSummary(g.ctx, g.client.Client, &req)
}

func (g *Gateway) queryPerf(invObjs []*model.InventoryObject, maxSample int32) (*types.QueryPerfResponse, error) {
	numObjs := len(invObjs)
	if numObjs == 0 {
		// empty inventory, return empty response
		// passing an empty spec to the api causes an error
		g.log.Warn("empty inventory, skipping QueryPerf")
		return &types.QueryPerfResponse{}, nil
	}

	specs := make([]types.PerfQuerySpec, 0, numObjs)
	for _, invObj := range invObjs {
		specs = append(specs, types.PerfQuerySpec{
			Entity:     invObj.Ref,
			MaxSample:  maxSample,
			IntervalId: model.RealtimeMetricsInterval,
			MetricId:   invObj.MetricIds,
		})
	}
	queryPerf := types.QueryPerf{
		This:      *g.client.Client.ServiceContent.PerfManager,
		QuerySpec: specs,
	}
	return methods.QueryPerf(g.ctx, g.client.Client, &queryPerf)
}

func (g *Gateway) retrieveCurrentTime() (*time.Time, error) {
	return methods.GetCurrentTime(g.ctx, g.client)
}

func (g *Gateway) vcenterName() string {
	return g.vcName
}
