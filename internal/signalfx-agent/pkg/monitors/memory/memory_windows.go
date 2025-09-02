package memory

import (
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
)

func (m *Monitor) makeMemoryDatapoints(memInfo *mem.VirtualMemoryStat, swapInfo *mem.SwapMemoryStat, dimensions map[string]string) []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		datapoint.New("memory.utilization", dimensions, datapoint.NewFloatValue(memInfo.UsedPercent), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.used", dimensions, datapoint.NewIntValue(int64(memInfo.Used)), datapoint.Gauge, time.Time{}),           //nolint:gosec
		datapoint.New("memory.available", dimensions, datapoint.NewIntValue(int64(memInfo.Available)), datapoint.Gauge, time.Time{}), //nolint:gosec
		datapoint.New("memory.total", dimensions, datapoint.NewIntValue(int64(memInfo.Total)), datapoint.Gauge, time.Time{}),         //nolint:gosec
		datapoint.New("memory.swap_total", dimensions, datapoint.NewIntValue(int64(swapInfo.Total)), datapoint.Gauge, time.Time{}),   //nolint:gosec
		datapoint.New("memory.swap_free", dimensions, datapoint.NewIntValue(int64(swapInfo.Free)), datapoint.Gauge, time.Time{}),     //nolint:gosec
		datapoint.New("memory.swap_used", dimensions, datapoint.NewIntValue(int64(swapInfo.Used)), datapoint.Gauge, time.Time{}),     //nolint:gosec
	}
}
