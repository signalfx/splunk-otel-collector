package converter

import (
	"math"
	"strconv"
	"strings"
	"time"

	info "github.com/google/cadvisor/info/v1"
	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"
	stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"

	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// InfoProvider provides a swappable interface to actually get the cAdvisor
// metrics
type InfoProvider interface {
	// Get information about all subcontainers of the specified container (includes self).
	SubcontainersInfo(containerName string) ([]info.ContainerInfo, error)
	// Get information about the machine.
	GetMachineInfo() (*info.MachineInfo, error)
	// Get Ephemeral storage usage stats from Pod
	GetEphemeralStatsFromPods() ([]stats.PodStats, error)
}

// metricValue describes a single metric value for a given set of label values
// within a parent containerMetric.
type metricValue struct {
	value  datapoint.Value
	labels []string
}

type metricValues []metricValue

type containerMetric struct {
	name        string
	help        string
	valueType   datapoint.MetricType
	extraLabels []string
	getValues   func(s *info.ContainerStats) metricValues
}

type containerSpecMetric struct {
	containerMetric
	getValues func(s *info.ContainerInfo) metricValues
}

type machineInfoMetric struct {
	containerMetric
	getValues func(s *info.MachineInfo) metricValues
}

type podStatusMetric struct {
	name      string
	valueType datapoint.MetricType
	getValues func(s *stats.FsStats) metricValues
}

// CadvisorCollector metric collector and converter
type CadvisorCollector struct {
	infoProvider               InfoProvider
	containerMetrics           []containerMetric
	containerSpecMetrics       []containerSpecMetric
	containerSpecMemMetrics    []containerSpecMetric
	containerSpecCPUMetrics    []containerSpecMetric
	machineInfoMetrics         []machineInfoMetric
	podEphemeralStorageMetrics []podStatusMetric
	sendDPs                    func(...*datapoint.Datapoint)
	defaultDimensions          map[string]string
	logger                     log.FieldLogger
}

// fsValues is a helper method for assembling per-filesystem stats.
func fsValues(fsStats []info.FsStats, valueFn func(*info.FsStats) datapoint.Value) metricValues {
	values := make(metricValues, 0, len(fsStats))
	for i := range fsStats {
		values = append(values, metricValue{
			value:  valueFn(&fsStats[i]),
			labels: []string{fsStats[i].Device},
		})
	}
	return values
}

func networkValues(net []info.InterfaceStats, valueFn func(*info.InterfaceStats) datapoint.Value) metricValues {
	values := make(metricValues, 0, len(net))
	for i := range net {
		values = append(values, metricValue{
			value:  valueFn(&net[i]),
			labels: []string{net[i].Name},
		})
	}
	return values
}

func getContainerMetrics() []containerMetric {
	var metrics = []containerMetric{
		{
			name:      "container_last_seen",
			help:      "Last time a container was seen by the exporter",
			valueType: datapoint.Timestamp,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(time.Now().UnixNano())}}
			},
		},
		{
			name:      "container_cpu_user_seconds_total",
			help:      "Cumulative user cpu time consumed in nanoseconds.",
			valueType: datapoint.Counter,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Cpu.Usage.User))}}
			},
		},
		{
			name:      "container_cpu_system_seconds_total",
			help:      "Cumulative system cpu time consumed in nanoseconds.",
			valueType: datapoint.Counter,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Cpu.Usage.System))}}
			},
		},
		{
			name:      "container_cpu_usage_seconds_total",
			help:      "Cumulative cpu time consumed per cpu in nanoseconds.",
			valueType: datapoint.Counter,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Cpu.Usage.Total))}}
			},
		},
		{
			name:      "container_cpu_utilization",
			help:      "Cumulative cpu utilization in percentages.",
			valueType: datapoint.Counter,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Cpu.Usage.Total / 10000000))}}
			},
		},
		{
			name:      "container_cpu_percent",
			help:      "Cumulative cpu utilization as a percentage of the host total CPU available",
			valueType: datapoint.Counter,
			getValues: func(s *info.ContainerStats) metricValues {
				cpuCount := utils.MaxInt(len(s.Cpu.Usage.PerCpu), 1)
				return metricValues{{value: datapoint.NewIntValue(int64(int(s.Cpu.Usage.Total) / (10000000 * cpuCount)))}}
			},
		},
		{
			name:        "container_cpu_utilization_per_core",
			help:        "Cumulative cpu utilization in percentages per core",
			valueType:   datapoint.Counter,
			extraLabels: []string{"cpu"},
			getValues: func(s *info.ContainerStats) metricValues {
				metricValues := make(metricValues, len(s.Cpu.Usage.PerCpu))
				for index, coreUsage := range s.Cpu.Usage.PerCpu {
					if coreUsage > 0 {
						metricValues[index] = metricValue{value: datapoint.NewIntValue(int64(coreUsage / 10000000)), labels: []string{"cpu" + strconv.Itoa(index)}}
					} else {
						metricValues[index] = metricValue{value: datapoint.NewIntValue(int64(0)), labels: []string{strconv.Itoa(index)}}
					}
				}
				return metricValues
			},
		},
		{
			name:      "container_cpu_cfs_periods",
			help:      "Total number of elapsed CFS enforcement intervals.",
			valueType: datapoint.Counter,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Cpu.CFS.Periods))}}
			},
		},
		{
			name:      "container_cpu_cfs_throttled_periods",
			help:      "Total number of times tasks in the cgroup have been throttled",
			valueType: datapoint.Counter,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Cpu.CFS.ThrottledPeriods))}}
			},
		},
		{
			name:      "container_cpu_cfs_throttled_time",
			help:      "Total time duration, in nanoseconds, for which tasks in the cgroup have been throttled.",
			valueType: datapoint.Counter,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Cpu.CFS.ThrottledTime))}}
			},
		},
		{
			name:      "container_memory_rss",
			help:      "RSS memory used by the container",
			valueType: datapoint.Gauge,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Memory.RSS))}}
			},
		},
		{
			name:      "container_memory_failcnt",
			help:      "Number of memory usage hits limits",
			valueType: datapoint.Counter,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Memory.Failcnt))}}
			},
		},
		{
			name:      "container_memory_usage_bytes",
			help:      "Current memory usage in bytes.",
			valueType: datapoint.Gauge,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Memory.Usage))}}
			},
		},
		{
			name:      "container_memory_working_set_bytes",
			help:      "Current working set in bytes.",
			valueType: datapoint.Gauge,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(s.Memory.WorkingSet))}}
			},
		},
		{
			name:        "container_memory_failures_total",
			help:        "Cumulative count of memory allocation failures.",
			valueType:   datapoint.Counter,
			extraLabels: []string{"type", "scope"},
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{
					{
						value:  datapoint.NewIntValue(int64(s.Memory.ContainerData.Pgfault)),
						labels: []string{"pgfault", "container"},
					},
					{
						value:  datapoint.NewIntValue(int64(s.Memory.ContainerData.Pgmajfault)),
						labels: []string{"pgmajfault", "container"},
					},
					{
						value:  datapoint.NewIntValue(int64(s.Memory.HierarchicalData.Pgfault)),
						labels: []string{"pgfault", "hierarchy"},
					},
					{
						value:  datapoint.NewIntValue(int64(s.Memory.HierarchicalData.Pgmajfault)),
						labels: []string{"pgmajfault", "hierarchy"},
					},
				}
			},
		},
		{
			name:        "container_fs_limit_bytes",
			help:        "Number of bytes that can be consumed by the container on this filesystem.",
			valueType:   datapoint.Gauge,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.Limit))
				})
			},
		},
		{
			name:        "container_fs_usage_bytes",
			help:        "Number of bytes that are consumed by the container on this filesystem.",
			valueType:   datapoint.Gauge,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.Usage))
				})
			},
		},
		{
			name:        "container_fs_reads_total",
			help:        "Cumulative count of reads completed",
			valueType:   datapoint.Gauge,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.ReadsCompleted))
				})
			},
		},
		{
			name:        "container_fs_sector_reads_total",
			help:        "Cumulative count of sector reads completed",
			valueType:   datapoint.Counter,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.SectorsRead))
				})
			},
		},
		{
			name:        "container_fs_reads_merged_total",
			help:        "Cumulative count of reads merged",
			valueType:   datapoint.Counter,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.ReadsMerged))
				})
			},
		},
		{
			name:        "container_fs_read_seconds_total",
			help:        "Cumulative count of seconds spent reading",
			valueType:   datapoint.Counter,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.ReadTime / uint64(time.Second)))
				})
			},
		},
		{
			name:        "container_fs_writes_total",
			help:        "Cumulative count of writes completed",
			valueType:   datapoint.Counter,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.WritesCompleted))
				})
			},
		},
		{
			name:        "container_fs_sector_writes_total",
			help:        "Cumulative count of sector writes completed",
			valueType:   datapoint.Counter,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.SectorsWritten))
				})
			},
		},
		{
			name:        "container_fs_writes_merged_total",
			help:        "Cumulative count of writes merged",
			valueType:   datapoint.Counter,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.WritesMerged))
				})
			},
		},
		{
			name:        "container_fs_write_seconds_total",
			help:        "Cumulative count of seconds spent writing",
			valueType:   datapoint.Counter,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.WriteTime / uint64(time.Second)))
				})
			},
		},
		{
			name:        "container_fs_io_current",
			help:        "Number of I/Os currently in progress",
			valueType:   datapoint.Gauge,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.IoInProgress))
				})
			},
		},
		{
			name:        "container_fs_io_time_seconds_total",
			help:        "Cumulative count of seconds spent doing I/Os",
			valueType:   datapoint.Counter,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.IoTime / uint64(time.Second)))
				})
			},
		},
		{
			name:        "container_fs_io_time_weighted_seconds_total",
			help:        "Cumulative weighted I/O time in seconds",
			valueType:   datapoint.Counter,
			extraLabels: []string{"device"},
			getValues: func(s *info.ContainerStats) metricValues {
				return fsValues(s.Filesystem, func(fs *info.FsStats) datapoint.Value {
					return datapoint.NewIntValue(int64(fs.WeightedIoTime / uint64(time.Second)))
				})
			},
		},
		{
			name:        "pod_network_receive_bytes_total",
			help:        "Cumulative count of bytes received",
			valueType:   datapoint.Counter,
			extraLabels: []string{"interface"},
			getValues: func(s *info.ContainerStats) metricValues {
				return networkValues(s.Network.Interfaces, func(is *info.InterfaceStats) datapoint.Value {
					return datapoint.NewIntValue(int64(is.RxBytes))
				})
			},
		},
		{
			name:        "pod_network_receive_packets_total",
			help:        "Cumulative count of packets received",
			valueType:   datapoint.Counter,
			extraLabels: []string{"interface"},
			getValues: func(s *info.ContainerStats) metricValues {
				return networkValues(s.Network.Interfaces, func(is *info.InterfaceStats) datapoint.Value {
					return datapoint.NewIntValue(int64(is.RxPackets))
				})
			},
		},
		{
			name:        "pod_network_receive_packets_dropped_total",
			help:        "Cumulative count of packets dropped while receiving",
			valueType:   datapoint.Counter,
			extraLabels: []string{"interface"},
			getValues: func(s *info.ContainerStats) metricValues {
				return networkValues(s.Network.Interfaces, func(is *info.InterfaceStats) datapoint.Value {
					return datapoint.NewIntValue(int64(is.RxDropped))
				})
			},
		},
		{
			name:        "pod_network_receive_errors_total",
			help:        "Cumulative count of errors encountered while receiving",
			valueType:   datapoint.Counter,
			extraLabels: []string{"interface"},
			getValues: func(s *info.ContainerStats) metricValues {
				return networkValues(s.Network.Interfaces, func(is *info.InterfaceStats) datapoint.Value {
					return datapoint.NewIntValue(int64(is.RxErrors))
				})
			},
		},
		{
			name:        "pod_network_transmit_bytes_total",
			help:        "Cumulative count of bytes transmitted",
			valueType:   datapoint.Counter,
			extraLabels: []string{"interface"},
			getValues: func(s *info.ContainerStats) metricValues {
				return networkValues(s.Network.Interfaces, func(is *info.InterfaceStats) datapoint.Value {
					return datapoint.NewIntValue(int64(is.TxBytes))
				})
			},
		},
		{
			name:        "pod_network_transmit_packets_total",
			help:        "Cumulative count of packets transmitted",
			valueType:   datapoint.Counter,
			extraLabels: []string{"interface"},
			getValues: func(s *info.ContainerStats) metricValues {
				return networkValues(s.Network.Interfaces, func(is *info.InterfaceStats) datapoint.Value {
					return datapoint.NewIntValue(int64(is.TxPackets))
				})
			},
		},
		{
			name:        "pod_network_transmit_packets_dropped_total",
			help:        "Cumulative count of packets dropped while transmitting",
			valueType:   datapoint.Counter,
			extraLabels: []string{"interface"},
			getValues: func(s *info.ContainerStats) metricValues {
				return networkValues(s.Network.Interfaces, func(is *info.InterfaceStats) datapoint.Value {
					return datapoint.NewIntValue(int64(is.TxDropped))
				})
			},
		},
		{
			name:        "pod_network_transmit_errors_total",
			help:        "Cumulative count of errors encountered while transmitting",
			valueType:   datapoint.Counter,
			extraLabels: []string{"interface"},
			getValues: func(s *info.ContainerStats) metricValues {
				return networkValues(s.Network.Interfaces, func(is *info.InterfaceStats) datapoint.Value {
					return datapoint.NewIntValue(int64(is.TxErrors))
				})
			},
		},
		{
			name:        "container_tasks_state",
			help:        "Number of tasks in given state",
			extraLabels: []string{"state"},
			valueType:   datapoint.Gauge,
			getValues: func(s *info.ContainerStats) metricValues {
				return metricValues{
					{
						value:  datapoint.NewIntValue(int64(s.TaskStats.NrSleeping)),
						labels: []string{"sleeping"},
					},
					{
						value:  datapoint.NewIntValue(int64(s.TaskStats.NrRunning)),
						labels: []string{"running"},
					},
					{
						value:  datapoint.NewIntValue(int64(s.TaskStats.NrStopped)),
						labels: []string{"stopped"},
					},
					{
						value:  datapoint.NewIntValue(int64(s.TaskStats.NrUninterruptible)),
						labels: []string{"uninterruptible"},
					},
					{
						value:  datapoint.NewIntValue(int64(s.TaskStats.NrIoWait)),
						labels: []string{"iowaiting"},
					},
				}
			},
		},
	}

	return metrics
}

func getContainerSpecCPUMetrics() []containerSpecMetric {
	var metrics = []containerSpecMetric{
		{
			containerMetric: containerMetric{
				name:        "container_spec_cpu_shares",
				help:        "",
				valueType:   datapoint.Gauge,
				extraLabels: []string{},
			},
			getValues: func(container *info.ContainerInfo) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(container.Spec.Cpu.Limit))}}
			},
		},
		{
			containerMetric: containerMetric{
				name:        "container_spec_cpu_quota",
				help:        "",
				valueType:   datapoint.Gauge,
				extraLabels: []string{},
			},
			getValues: func(container *info.ContainerInfo) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(container.Spec.Cpu.Quota))}}
			},
		},
		{
			containerMetric: containerMetric{
				name:        "container_spec_cpu_period",
				help:        "",
				valueType:   datapoint.Gauge,
				extraLabels: []string{},
			},
			getValues: func(container *info.ContainerInfo) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(container.Spec.Cpu.Period))}}
			},
		},
	}

	return metrics
}

func getContainerSpecMemMetrics() []containerSpecMetric {
	var metrics = []containerSpecMetric{
		{
			containerMetric: containerMetric{
				name:        "container_spec_memory_limit_bytes",
				help:        "",
				valueType:   datapoint.Gauge,
				extraLabels: []string{},
			},
			getValues: func(container *info.ContainerInfo) metricValues {
				limit := int64(container.Spec.Memory.Limit)
				// Workaround a strange issue where cadvisor reports
				// ridiculously high memory values in older k8s versions.
				if limit >= math.MaxUint64/3 {
					return metricValues{}
				}

				return metricValues{{value: datapoint.NewIntValue(limit)}}
			},
		},
		{
			containerMetric: containerMetric{
				name:        "container_spec_memory_swap_limit_bytes",
				help:        "",
				valueType:   datapoint.Gauge,
				extraLabels: []string{},
			},
			getValues: func(container *info.ContainerInfo) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(container.Spec.Memory.SwapLimit))}}
			},
		},
	}

	return metrics
}

func getMachineInfoMetrics() []machineInfoMetric {
	var metrics = []machineInfoMetric{
		{
			containerMetric: containerMetric{
				name:        "machine_cpu_frequency_khz",
				help:        "",
				valueType:   datapoint.Gauge,
				extraLabels: []string{},
			},
			getValues: func(machineInfo *info.MachineInfo) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(machineInfo.CpuFrequency))}}
			},
		},
		{
			containerMetric: containerMetric{
				name:        "machine_cpu_cores",
				help:        "",
				valueType:   datapoint.Gauge,
				extraLabels: []string{},
			},
			getValues: func(machineInfo *info.MachineInfo) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(machineInfo.NumCores))}}
			},
		},
		{
			containerMetric: containerMetric{
				name:        "machine_memory_bytes",
				help:        "",
				valueType:   datapoint.Gauge,
				extraLabels: []string{},
			},
			getValues: func(machineInfo *info.MachineInfo) metricValues {
				return metricValues{{value: datapoint.NewIntValue(int64(machineInfo.MemoryCapacity))}}
			},
		},
	}
	return metrics
}

func getContainerSpecMetrics() []containerSpecMetric {
	var metrics = []containerSpecMetric{
		{
			containerMetric: containerMetric{
				name:        "container_start_time_seconds",
				help:        "",
				valueType:   datapoint.Gauge,
				extraLabels: []string{},
			},
			getValues: func(container *info.ContainerInfo) metricValues {
				return metricValues{{value: datapoint.NewIntValue(container.Spec.CreationTime.Unix())}}
			},
		},
	}

	return metrics
}

func getPodEphemeralStorageMetrics() []podStatusMetric {
	var metrics = []podStatusMetric{
		{
			name:      "pod_ephemeral_storage_capacity_bytes",
			valueType: datapoint.Gauge,
			getValues: func(es *stats.FsStats) metricValues {
				if es.CapacityBytes != nil {
					return metricValues{{value: datapoint.NewIntValue(int64(*es.CapacityBytes))}}
				}
				return metricValues{}
			},
		},
		{
			name:      "pod_ephemeral_storage_used_bytes",
			valueType: datapoint.Counter,
			getValues: func(es *stats.FsStats) metricValues {
				if es.UsedBytes != nil {
					return metricValues{{value: datapoint.NewIntValue(int64(*es.UsedBytes))}}
				}
				return metricValues{}

			},
		},
	}

	return metrics
}

// NewCadvisorCollector creates new CadvisorCollector
func NewCadvisorCollector(
	infoProvider InfoProvider,
	sendDPs func(...*datapoint.Datapoint),
	defaultDimensions map[string]string,
	logger log.FieldLogger) *CadvisorCollector {

	return &CadvisorCollector{
		infoProvider:               infoProvider,
		containerMetrics:           getContainerMetrics(),
		containerSpecCPUMetrics:    getContainerSpecCPUMetrics(),
		containerSpecMemMetrics:    getContainerSpecMemMetrics(),
		containerSpecMetrics:       getContainerSpecMetrics(),
		machineInfoMetrics:         getMachineInfoMetrics(),
		podEphemeralStorageMetrics: getPodEphemeralStorageMetrics(),
		sendDPs:                    sendDPs,
		defaultDimensions:          defaultDimensions,
		logger:                     logger,
	}
}

// Collect fetches the stats from all containers and delivers them as
// Prometheus metrics. It implements prometheus.PrometheusCollector.
func (c *CadvisorCollector) Collect(hasPodEphemeralStorageStatsGroupEnabled bool) {
	c.collectMachineInfo()
	c.collectVersionInfo()
	c.collectContainersInfo()
	if hasPodEphemeralStorageStatsGroupEnabled {
		c.collectEphemeralStorageStatsFromPod()
	}
}

func (c *CadvisorCollector) preprocessAndSendDatapoints(dps ...*datapoint.Datapoint) {
	// This is the filtering in place trick from https://github.com/golang/go/wiki/SliceTricks#filter-in-place
	n := 0
	for i := range dps {
		dp := dps[i]
		dims := dp.Dimensions

		if dims["container_name"] == "POD" {
			isNetworkMetric := strings.HasPrefix(dp.Metric, "pod_network_")
			if !isNetworkMetric {
				// Skip any non-networking metrics on the POD container since
				// they are very low value.
				continue
			}
			// Get rid of container_name from pod network metrics since they
			// pertain to the entire pod.
			delete(dims, "container_name")
		}

		dims["metric_source"] = "kubernetes"

		for k, v := range c.defaultDimensions {
			dims[k] = v
		}

		// remove high cardinality dimensions
		delete(dims, "id")
		delete(dims, "name")

		dps[n] = dp
		n++
	}

	c.sendDPs(dps[:n]...)
}

func copyDims(dims map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range dims {
		newMap[k] = v
	}
	return newMap
}

func (c *CadvisorCollector) collectContainersInfo() {
	now := time.Now()
	containers, err := c.infoProvider.SubcontainersInfo("/")
	if err != nil {
		c.logger.WithError(err).Error("Couldn't get cAdvisor container stats")
		return
	}
	dpsOut := make([]*datapoint.Datapoint, 0)
	for i := range containers {
		container := containers[i]
		dims := make(map[string]string)

		image := container.Spec.Image
		if len(image) > 0 {
			dims["container_image"] = image
		}

		dims["container_id"] = container.Id
		dims["container_spec_name"] = container.Spec.Labels["io.kubernetes.container.name"]
		// TODO: Remove this once everybody is migrated to neoagent v2 and
		// change built-in dashboards to use container_spec_name
		dims["container_name"] = container.Spec.Labels["io.kubernetes.container.name"]
		dims["kubernetes_pod_uid"] = container.Spec.Labels["io.kubernetes.pod.uid"]
		dims["kubernetes_pod_name"] = container.Spec.Labels["io.kubernetes.pod.name"]
		dims["kubernetes_namespace"] = container.Spec.Labels["io.kubernetes.pod.namespace"]

		// Container spec
		for _, cm := range c.containerSpecMetrics {
			for _, metricValue := range cm.getValues(&container) {
				newDims := copyDims(dims)

				// Add extra dimensions
				for i, label := range cm.extraLabels {
					newDims[label] = metricValue.labels[i]
				}

				dpsOut = append(dpsOut, datapoint.New(cm.name, newDims, metricValue.value, cm.valueType, now))
			}
		}

		if container.Spec.HasCpu {
			for i := range c.containerSpecCPUMetrics {
				for _, metricValue := range c.containerSpecCPUMetrics[i].getValues(&container) {
					newDims := copyDims(dims)

					// Add extra dimensions
					for i, label := range c.containerSpecCPUMetrics[i].extraLabels {
						newDims[label] = metricValue.labels[i]
					}

					dpsOut = append(dpsOut, datapoint.New(c.containerSpecCPUMetrics[i].name, newDims, metricValue.value, c.containerSpecCPUMetrics[i].valueType, now))
				}
			}
		}

		if container.Spec.HasMemory {
			for i := range c.containerSpecMemMetrics {
				cm := c.containerSpecMemMetrics[i]
				for _, metricValue := range cm.getValues(&container) {
					newDims := copyDims(dims)

					// Add extra dimensions
					for i, label := range cm.extraLabels {
						newDims[label] = metricValue.labels[i]
					}

					dpsOut = append(dpsOut, datapoint.New(cm.name, newDims, metricValue.value, cm.valueType, now))
				}
			}
		}

		// Now for the actual metrics
		if len(container.Stats) > 0 {
			// only get the latest stats from this container. note/warning: the stats array contains historical statistics in earliest-to-latest order
			lastStatIndex := len(container.Stats) - 1
			stat := container.Stats[lastStatIndex]

			for _, cm := range c.containerMetrics {
				for _, metricValue := range cm.getValues(stat) {
					newDims := copyDims(dims)

					// Add extra dimensions
					for i, label := range cm.extraLabels {
						newDims[label] = metricValue.labels[i]
					}

					dpsOut = append(dpsOut, datapoint.New(cm.name, newDims, metricValue.value, cm.valueType, now))
				}
			}
		}
	}

	c.preprocessAndSendDatapoints(dpsOut...)
}

func (c *CadvisorCollector) collectVersionInfo() {}

func (c *CadvisorCollector) collectMachineInfo() {
	now := time.Now()
	machineInfo, err := c.infoProvider.GetMachineInfo()
	if err != nil {
		//c.errors.Set(1)
		c.logger.Errorf("Couldn't get machine info: %s", err)
		return
	}
	if machineInfo == nil {
		return
	}

	dims := make(map[string]string)

	dps := make([]*datapoint.Datapoint, 0)
	for _, cm := range c.machineInfoMetrics {
		for _, metricValue := range cm.getValues(machineInfo) {
			newDims := copyDims(dims)

			// Add extra dimensions
			for i, label := range cm.extraLabels {
				newDims[label] = metricValue.labels[i]
			}

			dps = append(dps, datapoint.New(cm.name, newDims, metricValue.value, cm.valueType, now))
		}
	}

	c.preprocessAndSendDatapoints(dps...)
}

func (c *CadvisorCollector) collectEphemeralStorageStatsFromPod() {
	now := time.Now()
	podStats, err := c.infoProvider.GetEphemeralStatsFromPods()

	if err != nil {
		c.logger.Errorf("Couldn't get Pod stats: %s", err)
		return
	}

	if podStats == nil {
		return
	}

	dps := make([]*datapoint.Datapoint, 0)
	for _, podStat := range podStats {
		dims := make(map[string]string)

		dims["kubernetes_pod_uid"] = podStat.PodRef.UID
		dims["kubernetes_pod_name"] = podStat.PodRef.Name
		dims["kubernetes_namespace"] = podStat.PodRef.Namespace

		for _, pm := range c.podEphemeralStorageMetrics {
			for _, metricValue := range pm.getValues(podStat.EphemeralStorage) {
				dps = append(dps, datapoint.New(pm.name, dims, metricValue.value, pm.valueType, now))
			}
		}
	}

	c.preprocessAndSendDatapoints(dps...)
}
