package service

import (
	"github.com/vmware/govmomi/vim25/types"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

// Returned by a perfFetcher implementation every time points are fetched.
// Paginates through metrics in blocks of inventory objects. Max number of inv
// objects per iteration is determined by pageSize.
type invIterator struct {
	gateway    IGateway
	inv        []*model.InventoryObject
	pageSize   int
	numInvObjs int
	numPages   int
	pageNum    int
	maxSample  int32
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
