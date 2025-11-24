package docker

import (
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/signalfx/golib/v3/sfxclient" //nolint:staticcheck // SA1019: deprecated package still in use
)

var memoryStatCounters = map[string]bool{
	"pgfault":          true,
	"pgmajfault":       true,
	"pgpgin":           true,
	"pgpgout":          true,
	"total_pgfault":    true,
	"total_pgmajfault": true,
	"total_pgpgin":     true,
	"total_pgpgout":    true,
}

var basicBlockIOMetrics = map[string]bool{
	"blkio.io_service_bytes_recursive.read":  true,
	"blkio.io_service_bytes_recursive.write": true,
}

// ConvertStatsToMetrics converts a docker container stats object into an array of datapoints
func ConvertStatsToMetrics(container *container.InspectResponse, parsed *container.StatsResponse, enhancedMetricsConfig EnhancedMetricsConfig) ([]*datapoint.Datapoint, error) {
	var dps []*datapoint.Datapoint
	dps = append(dps, convertBlkioStats(&parsed.BlkioStats, enhancedMetricsConfig.EnableExtraBlockIOMetrics)...)
	dps = append(dps, convertCPUStats(&parsed.CPUStats, &parsed.PreCPUStats, enhancedMetricsConfig.EnableExtraCPUMetrics)...)
	dps = append(dps, convertMemoryStats(&parsed.MemoryStats, enhancedMetricsConfig.EnableExtraMemoryMetrics)...)
	dps = append(dps, convertNetworkStats(&parsed.Networks, enhancedMetricsConfig.EnableExtraNetworkMetrics)...)

	now := time.Now()
	for i := range dps {
		dps[i].Timestamp = now

		if dps[i].Dimensions == nil {
			dps[i].Dimensions = map[string]string{}
		}
		// Set to preserve compatibility with docker-collectd plugin
		dps[i].Dimensions["plugin"] = "docker"
		name := strings.TrimPrefix(container.Name, "/")
		dps[i].Dimensions["container_name"] = name
		// Duplicate container_name in plugin_instance to maintain compat with
		// collectd-docker plugin
		dps[i].Dimensions["plugin_instance"] = name
		dps[i].Dimensions["container_image"] = container.Config.Image
		dps[i].Dimensions["container_id"] = container.ID
		dps[i].Dimensions["container_hostname"] = container.Config.Hostname
	}

	return dps, nil
}

func convertBlkioStats(stats *container.BlkioStats, enhancedMetrics bool) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	for k, v := range map[string][]container.BlkioStatEntry{
		"io_service_bytes_recursive": stats.IoServiceBytesRecursive,
		"io_serviced_recursive":      stats.IoServicedRecursive,
		"io_queue_recursive":         stats.IoQueuedRecursive,
		"io_service_time_recursive":  stats.IoServiceTimeRecursive,
		"io_wait_time_recursive":     stats.IoWaitTimeRecursive,
		"io_merged_recursive":        stats.IoMergedRecursive,
		"io_time_recursive":          stats.IoTimeRecursive,
		"sectors_recursive":          stats.SectorsRecursive,
	} {
		for _, bs := range v {
			if bs.Op == "" {
				continue
			}
			metricName := "blkio." + k + "." + strings.ToLower(bs.Op)

			if _, exists := basicBlockIOMetrics[metricName]; enhancedMetrics || exists {
				dims := map[string]string{
					"device_major": strconv.FormatUint(bs.Major, 10),
					"device_minor": strconv.FormatUint(bs.Minor, 10),
				}
				out = append(out, sfxclient.Cumulative("blkio."+k+"."+strings.ToLower(bs.Op), dims, int64(bs.Value))) //nolint:gosec
			}
		}
	}

	return out
}

func convertCPUStats(stats, prior *container.CPUStats, enhancedMetrics bool) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	out = append(out, []*datapoint.Datapoint{
		sfxclient.Cumulative("cpu.usage.total", nil, int64(stats.CPUUsage.TotalUsage)), //nolint:gosec
		sfxclient.Cumulative("cpu.usage.system", nil, int64(stats.SystemUsage)),        //nolint:gosec
	}...)

	// Except two metrics above, everything else will be added only when enhnacedMetrics is enabled
	if enhancedMetrics {
		out = append(out, []*datapoint.Datapoint{
			sfxclient.Cumulative("cpu.usage.kernelmode", nil, int64(stats.CPUUsage.UsageInKernelmode)), //nolint:gosec
			sfxclient.Cumulative("cpu.usage.usermode", nil, int64(stats.CPUUsage.UsageInUsermode)),     //nolint:gosec
		}...)

		for i, v := range stats.CPUUsage.PercpuUsage {
			dims := map[string]string{
				"core": "cpu" + strconv.Itoa(i),
			}
			out = append(out, sfxclient.Cumulative("cpu.percpu.usage", dims, int64(v))) //nolint:gosec
		}

		out = append(out, []*datapoint.Datapoint{
			sfxclient.Cumulative("cpu.throttling_data.periods", nil, int64(stats.ThrottlingData.Periods)),                    //nolint:gosec
			sfxclient.Cumulative("cpu.throttling_data.throttled_periods", nil, int64(stats.ThrottlingData.ThrottledPeriods)), //nolint:gosec
			sfxclient.Cumulative("cpu.throttling_data.throttled_time", nil, int64(stats.ThrottlingData.ThrottledTime)),       //nolint:gosec
		}...)

		out = append(out, sfxclient.GaugeF("cpu.percent", nil, calculateCPUPercent(prior, stats)))
	}

	return out
}

// Copied from
// https://github.com/docker/cli/blob/dbd96badb6959c2b7070664aecbcf0f7c299c538/cli/command/container/stats_helpers.go
func calculateCPUPercent(previous, v *container.CPUStats) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUUsage.TotalUsage) - float64(previous.CPUUsage.TotalUsage)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.SystemUsage) - float64(previous.SystemUsage)
		onlineCPUs  = float64(v.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(v.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return cpuPercent
}

func convertMemoryStats(stats *container.MemoryStats, enhancedMetrics bool) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	// If not present, default value will be 0.
	bufferCacheUsage := stats.Stats["total_cache"]

	out = append(out, sfxclient.Gauge("memory.usage.limit", nil, int64(stats.Limit))) //nolint:gosec
	if stats.PrivateWorkingSet == 0 {
		// See discussion at https://github.com/signalfx/signalfx-agent/issues/1009
		out = append(out, sfxclient.Gauge("memory.usage.total", nil, int64(stats.Usage-bufferCacheUsage))) //nolint:gosec
	} else {
		// This is used for Windows containers
		out = append(out, sfxclient.Gauge("memory.usage.total", nil, int64(stats.PrivateWorkingSet))) //nolint:gosec
	}

	// Except two metrics above, everything else will be added only when enhnacedMetrics is enabled
	if enhancedMetrics {
		out = append(out, []*datapoint.Datapoint{
			sfxclient.Gauge("memory.usage.max", nil, int64(stats.MaxUsage)), //nolint:gosec
			sfxclient.GaugeF("memory.percent", nil,
				// If cache is not present it will use the default value of 0
				100.0*(float64(stats.Usage)-float64(stats.Stats["cache"]))/float64(stats.Limit)),
		}...)

		for k, v := range stats.Stats {
			if _, exists := memoryStatCounters[k]; exists {
				out = append(out, sfxclient.Cumulative("memory.stats."+k, nil, int64(v))) //nolint:gosec
			} else {
				out = append(out, sfxclient.Gauge("memory.stats."+k, nil, int64(v))) //nolint:gosec
			}
		}
	}

	return out
}

func convertNetworkStats(stats *map[string]container.NetworkStats, enhancedMetrics bool) []*datapoint.Datapoint {
	if stats == nil {
		return nil
	}
	var out []*datapoint.Datapoint
	for k, s := range *stats {
		dims := map[string]string{
			"interface": k,
		}

		out = append(out, []*datapoint.Datapoint{
			sfxclient.Cumulative("network.usage.rx_bytes", dims, int64(s.RxBytes)), //nolint:gosec
			sfxclient.Cumulative("network.usage.tx_bytes", dims, int64(s.TxBytes)), //nolint:gosec
		}...)

		if enhancedMetrics {
			out = append(out, []*datapoint.Datapoint{
				sfxclient.Cumulative("network.usage.rx_dropped", dims, int64(s.RxDropped)), //nolint:gosec
				sfxclient.Cumulative("network.usage.rx_errors", dims, int64(s.RxErrors)),   //nolint:gosec
				sfxclient.Cumulative("network.usage.rx_packets", dims, int64(s.RxPackets)), //nolint:gosec
				sfxclient.Cumulative("network.usage.tx_dropped", dims, int64(s.TxDropped)), //nolint:gosec
				sfxclient.Cumulative("network.usage.tx_errors", dims, int64(s.TxErrors)),   //nolint:gosec
				sfxclient.Cumulative("network.usage.tx_packets", dims, int64(s.TxPackets)), //nolint:gosec
			}...)
		}
	}

	return out
}
