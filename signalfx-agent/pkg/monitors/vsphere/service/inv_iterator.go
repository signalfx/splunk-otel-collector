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
	"github.com/vmware/govmomi/vim25/types"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

// Returned by a perfFetcher implementation every time points are fetched.
// Paginates through metrics in blocks of inventory objects. Max number of inv
// objects per iteration is determined by pageSize.
type invIterator struct {
	inv        []*model.InventoryObject
	maxSample  int32
	pageSize   int
	numInvObjs int
	numPages   int
	gateway    IGateway

	pageNum int
}

func (it *invIterator) nextInvPage() (
	metrics []types.BasePerfEntityMetricBase,
	hasNext bool,
	err error,
) {
	startIdx := it.pageNum * it.pageSize
	endIdx := startIdx + it.pageSize
	if endIdx > it.numInvObjs {
		endIdx = it.numInvObjs
	}
	slice := it.inv[startIdx:endIdx]
	perf, err := it.gateway.queryPerf(slice, it.maxSample)
	if err != nil {
		return nil, false, err
	}
	it.pageNum++
	return perf.Returnval, it.pageNum < it.numPages, nil
}
