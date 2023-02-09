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

package memory

import (
	"time"

	"github.com/shirou/gopsutil/mem"
	"github.com/signalfx/golib/v3/datapoint"
)

func (m *Monitor) makeMemoryDatapoints(memInfo *mem.VirtualMemoryStat, swapInfo *mem.SwapMemoryStat, dimensions map[string]string) []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		datapoint.New("memory.utilization", dimensions, datapoint.NewFloatValue(memInfo.UsedPercent), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.used", dimensions, datapoint.NewIntValue(int64(memInfo.Used)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.available", dimensions, datapoint.NewIntValue(int64(memInfo.Available)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.total", dimensions, datapoint.NewIntValue(int64(memInfo.Total)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.swap_total", dimensions, datapoint.NewIntValue(int64(swapInfo.Total)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.swap_free", dimensions, datapoint.NewIntValue(int64(swapInfo.Free)), datapoint.Gauge, time.Time{}),
		datapoint.New("memory.swap_used", dimensions, datapoint.NewIntValue(int64(swapInfo.Used)), datapoint.Gauge, time.Time{}),
	}
}
