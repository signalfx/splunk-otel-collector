package kubeletmetrics

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/signalfx-agent/pkg/utils/k8sutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func convertContainerMetrics(c *v1alpha1.ContainerStats, status *v1.ContainerStatus, dims map[string]string) []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint

	if status != nil {
		dims["container_id"] = k8sutil.StripContainerID(status.ContainerID)
		dims["container_image"] = status.Image
	}
	dims["container_spec_name"] = c.Name

	if c.CPU != nil {
		if c.CPU.UsageCoreNanoSeconds != nil {
			dps = append(dps, sfxclient.CumulativeF(containerCPUUtilization, dims, float64(*c.CPU.UsageCoreNanoSeconds)/10_000_000.0))
		}
	}

	if c.Memory != nil {
		if c.Memory.AvailableBytes != nil {
			dps = append(dps, sfxclient.Gauge(containerMemoryAvailableBytes, dims, int64(*c.Memory.AvailableBytes)))
		}
		if c.Memory.UsageBytes != nil {
			dps = append(dps, sfxclient.Gauge(containerMemoryUsageBytes, dims, int64(*c.Memory.UsageBytes)))
		}
		if c.Memory.WorkingSetBytes != nil {
			dps = append(dps, sfxclient.Gauge(containerMemoryWorkingSetBytes, dims, int64(*c.Memory.WorkingSetBytes)))
		}
		if c.Memory.RSSBytes != nil {
			dps = append(dps, sfxclient.Gauge(containerMemoryRssBytes, dims, int64(*c.Memory.RSSBytes)))
		}
		if c.Memory.PageFaults != nil {
			dps = append(dps, sfxclient.Cumulative(containerMemoryPageFaults, dims, int64(*c.Memory.PageFaults)))
		}
		if c.Memory.MajorPageFaults != nil {
			dps = append(dps, sfxclient.Cumulative(containerMemoryMajorPageFaults, dims, int64(*c.Memory.MajorPageFaults)))
		}
	}

	if c.Rootfs != nil {
		if c.Rootfs.AvailableBytes != nil {
			// uint64 -> int64 conversion should be safe since disk sizes
			// aren't going to get that big for a long time.
			dps = append(dps, sfxclient.Gauge(containerFsAvailableBytes, dims, int64(*c.Rootfs.AvailableBytes)))
		}
		if c.Rootfs.CapacityBytes != nil {
			dps = append(dps, sfxclient.Gauge(containerFsCapacityBytes, dims, int64(*c.Rootfs.CapacityBytes)))
		}
		if c.Rootfs.UsedBytes != nil {
			dps = append(dps, sfxclient.Gauge(containerFsUsageBytes, dims, int64(*c.Rootfs.UsedBytes)))
		}
	}

	return dps
}
