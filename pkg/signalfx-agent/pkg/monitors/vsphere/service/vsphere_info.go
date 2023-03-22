package service

import (
	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

// Encapsulates services necessary to retrieve the inventory and available metrics, and build the metric index.
type VSphereInfoService struct {
	inventorySvc *InventorySvc
	metricsSvc   *MetricsSvc
}

func NewVSphereInfoService(inventorySvc *InventorySvc, metricsSvc *MetricsSvc) *VSphereInfoService {
	return &VSphereInfoService{inventorySvc: inventorySvc, metricsSvc: metricsSvc}
}

// Retrieves the inventory, available metrics, and metric index.
func (svc *VSphereInfoService) RetrieveVSphereInfo() (*model.VsphereInfo, error) {
	inv, err := svc.retrievePopulatedInventory()
	if err != nil {
		return nil, err
	}
	idx, err := svc.metricsSvc.RetrievePerfCounterIndex()
	if err != nil {
		return nil, err
	}
	return &model.VsphereInfo{Inv: inv, PerfCounterIndex: idx}, nil
}

// Retrieves the inventory and populates each inventory object with its available metrics.
func (svc *VSphereInfoService) retrievePopulatedInventory() (*model.Inventory, error) {
	inv, err := svc.inventorySvc.RetrieveInventory()
	if err != nil {
		return nil, err
	}
	svc.metricsSvc.PopulateInvMetrics(inv)
	return inv, nil
}
