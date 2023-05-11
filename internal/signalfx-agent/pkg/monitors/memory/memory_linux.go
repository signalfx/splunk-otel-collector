package memory

import (
	"time"

	"github.com/shirou/gopsutil/mem"
	"github.com/signalfx/golib/v3/datapoint"
)

func (m *Monitor) makeMemoryDatapoints(memInfo *mem.VirtualMemoryStat, swapInfo *mem.SwapMemoryStat, dimensions map[string]string) []*datapoint.Datapoint {
	used := memInfo.Total - memInfo.Free - memInfo.Buffers - (memInfo.Cached - memInfo.SReclaimable) - memInfo.Slab

	return []*datapoint.Datapoint{
		datapoint.New("memory.utilization", dimensions, datapoint.NewFloatValue(float64(used)/float64(memInfo.Total)*100), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.used", dimensions, datapoint.NewIntValue(int64(used)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.buffered", dimensions, datapoint.NewIntValue(int64(memInfo.Buffers)), datapoint.Gauge, time.Time{}),
		// for some reason gopsutil decided to add slab_reclaimable to cached which collectd does not
		datapoint.New("memory.cached", dimensions, datapoint.NewIntValue(int64(memInfo.Cached-memInfo.SReclaimable)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.slab_recl", dimensions, datapoint.NewIntValue(int64(memInfo.SReclaimable)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.slab_unrecl", dimensions, datapoint.NewIntValue(int64(memInfo.Slab-memInfo.SReclaimable)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.free", dimensions, datapoint.NewIntValue(int64(memInfo.Free)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.total", dimensions, datapoint.NewIntValue(int64(memInfo.Total)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.swap_total", dimensions, datapoint.NewIntValue(int64(swapInfo.Total)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.swap_free", dimensions, datapoint.NewIntValue(int64(swapInfo.Free)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.swap_used", dimensions, datapoint.NewIntValue(int64(swapInfo.Used)), datapoint.Gauge, time.Time{}),
	}
}
