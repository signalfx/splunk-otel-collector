// Copyright  Splunk, Inc.
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
