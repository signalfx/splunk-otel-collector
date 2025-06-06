// Code generated by monitor-code-gen. DO NOT EDIT.

package cpu

import (
	"github.com/signalfx/golib/v3/datapoint"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
)

const monitorType = "cpu"

var groupSet = map[string]bool{}

const (
	cpuIdle               = "cpu.idle"
	cpuInterrupt          = "cpu.interrupt"
	cpuNice               = "cpu.nice"
	cpuNumProcessors      = "cpu.num_processors"
	cpuSoftirq            = "cpu.softirq"
	cpuSteal              = "cpu.steal"
	cpuSystem             = "cpu.system"
	cpuUser               = "cpu.user"
	cpuUtilization        = "cpu.utilization"
	cpuUtilizationPerCore = "cpu.utilization_per_core"
	cpuWait               = "cpu.wait"
)

var metricSet = map[string]monitors.MetricInfo{
	cpuIdle:               {Type: datapoint.Counter},
	cpuInterrupt:          {Type: datapoint.Counter},
	cpuNice:               {Type: datapoint.Counter},
	cpuNumProcessors:      {Type: datapoint.Gauge},
	cpuSoftirq:            {Type: datapoint.Counter},
	cpuSteal:              {Type: datapoint.Counter},
	cpuSystem:             {Type: datapoint.Counter},
	cpuUser:               {Type: datapoint.Counter},
	cpuUtilization:        {Type: datapoint.Gauge},
	cpuUtilizationPerCore: {Type: datapoint.Gauge},
	cpuWait:               {Type: datapoint.Counter},
}

var defaultMetrics = map[string]bool{
	cpuIdle:          true,
	cpuNumProcessors: true,
	cpuUtilization:   true,
}

var groupMetricsMap = map[string][]string{}

var monitorMetadata = monitors.Metadata{
	MonitorType:     "cpu",
	DefaultMetrics:  defaultMetrics,
	Metrics:         metricSet,
	SendUnknown:     false,
	Groups:          groupSet,
	GroupMetricsMap: groupMetricsMap,
	SendAll:         false,
}
