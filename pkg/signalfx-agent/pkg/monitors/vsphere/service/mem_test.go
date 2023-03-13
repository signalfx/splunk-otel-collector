package service

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

func TestPointsServiceMem(t *testing.T) {
	// Previous impl of returning 1M pts used 730MB before GC, 480MB after
	const maxMBBeforeGc = 50
	const maxMBAfterGc = 1
	requireMaxHeapUsage(t, maxMBBeforeGc, maxMBAfterGc, func() {
		const numInvObjs = 100_000
		const ptsPerInvObj = 10
		numPts := processPoints(numInvObjs, ptsPerInvObj)
		require.Equal(t, 1_000_000, numPts)
	})
}

func processPoints(numInvObjs, numPtsPerInvObj int) int {
	gateway := newFakeGateway(numPtsPerInvObj)
	numPtsReceived := 0
	svc := NewPointsSvc(gateway, testLog, 10, func(pts ...*datapoint.Datapoint) {
		numPtsReceived += len(pts)
	})
	inv := invObjs(numInvObjs)
	_ = svc.FetchPoints(&model.VsphereInfo{Inv: &model.Inventory{Objects: inv}}, 1)
	return numPtsReceived
}

func requireMaxHeapUsage(t *testing.T, maxMBBeforeGc int, maxMBAfterGc int, f func()) {
	var maxBeforeGc, maxAfterGc uint64
	maxBeforeGc = uint64(maxMBBeforeGc) * 1_000_000
	maxAfterGc = uint64(maxMBAfterGc) * 1_000_000
	heapUsageBeforeGc, heapUsageAfterGc := heapInuseDelta(f)
	if heapUsageBeforeGc > maxBeforeGc || heapUsageAfterGc > maxAfterGc {
		t.Log(fmt.Sprintf("heapUsageBeforeGc max: %d actual: %d", maxBeforeGc, heapUsageBeforeGc))
		t.Log(fmt.Sprintf("heapUsageAfterGc max: %d actual: %d", maxAfterGc, heapUsageAfterGc))
		t.Fail()
	}
}

func heapInuseDelta(f func()) (beforeGcDelta, afterGcDelta uint64) {
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startHeap := m.HeapInuse

	f()

	runtime.ReadMemStats(&m)
	beforeGcHeap := m.HeapInuse
	if beforeGcHeap < startHeap {
		return 0, 0
	}

	beforeGcDelta = beforeGcHeap - startHeap
	runtime.GC()

	runtime.ReadMemStats(&m)
	afterGcHeap := m.HeapInuse

	if afterGcHeap < startHeap {
		afterGcDelta = 0
	} else {
		afterGcDelta = afterGcHeap - startHeap
	}

	return beforeGcDelta, afterGcDelta
}
