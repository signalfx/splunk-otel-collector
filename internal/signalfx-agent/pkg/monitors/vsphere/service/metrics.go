package service

import (
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

type MetricsSvc struct {
	log     logrus.FieldLogger
	gateway IGateway
}

func NewMetricsService(gateway IGateway, log logrus.FieldLogger) *MetricsSvc {
	return &MetricsSvc{log: log, gateway: gateway}
}

// Retrieves metric metadata (PerformanceManager) from vCenter, indexes that data by the id (key) of each
// metric (aka perf counter), and returns the index.
func (svc *MetricsSvc) RetrievePerfCounterIndex() (model.MetricInfosByKey, error) {
	pm, err := svc.gateway.retrievePerformanceManager()
	if err != nil {
		svc.log.WithError(err).Error("retrievePerformanceManager failed")
		return nil, err
	}
	return indexPerfCountersByKey(pm.PerfCounter), nil
}

// Walks through the passed-in inventory. For each inventory object, queries the SDK for available metrics for the
// inventory object and assigns the available metrics to the inventory object for later querying the SDK for performance data.
func (svc *MetricsSvc) PopulateInvMetrics(inv *model.Inventory) {
	for _, invObj := range inv.Objects {
		resp, err := svc.gateway.queryAvailablePerfMetric(invObj.Ref)
		if err != nil {
			svc.log.WithField("invObj.ref", invObj.Ref).WithError(err).Error("populateInvMetrics: queryAvailablePerfMetric failed")
			continue
		}
		invObj.MetricIds = resp.Returnval
	}
}

// Indexes performance counters by their key (id) for lookup during points retrieval.
func indexPerfCountersByKey(perfCounters []types.PerfCounterInfo) model.MetricInfosByKey {
	idx := make(model.MetricInfosByKey)
	for _, counter := range perfCounters {
		metricName := getMetricName(counter)
		idx[counter.Key] = model.MetricInfo{MetricName: metricName, PerfCounterInfo: counter}
	}
	return idx
}

// Builds a human readable metric name with a suffix indicating the units (if appropriate).
func getMetricName(perfCounterInfo types.PerfCounterInfo) string {
	group := perfCounterInfo.GroupInfo.GetElementDescription().Key
	name := perfCounterInfo.NameInfo.GetElementDescription().Key
	orig := group + "_" + name
	suffix := string(getMetricUnits(group, name))
	return fixupMetricName(orig) + suffix
}

// Converts a camel cased, dotted metric name to snake case, and adds the vsphere prefix.
func fixupMetricName(in string) string {
	const prefix = "vsphere."
	return prefix + dotsToUnderscores(camelToSnakeCase(in))
}

var camelRegexp = regexp.MustCompile("[[:upper:]]+")

// Converts a camel cased name to snake case.
func camelToSnakeCase(in string) string {
	return camelRegexp.ReplaceAllStringFunc(in, func(s string) string {
		return "_" + strings.ToLower(s)
	})
}

var dotRegexp = regexp.MustCompile(`\.`)

// Replaces dots with underscores.
func dotsToUnderscores(in string) string {
	return dotRegexp.ReplaceAllString(in, "_")
}

type units string

const (
	percent = units("percent")
	ms      = units("ms")
	mhz     = units("mhz")
	kbs     = units("kbs") // kilobytes per second
	kb      = units("kb")
	mb      = units("mb")
	tb      = units("tb")
	joules  = units("joules")
	watts   = units("watts")
	seconds = units("seconds")
)

// Given a group name and vCenter name for the metric, returns the units.
func getMetricUnits(groupName string, name string) units {
	groupUnits, ok := metricUnits[groupName]
	if !ok {
		return ""
	}
	suffix, ok := groupUnits[name]
	if !ok {
		return ""
	}
	return "_" + suffix
}

var metricUnits = map[string]map[string]units{
	"cpu": {
		"coreUtilization":        percent,
		"costop":                 ms,
		"demand":                 mhz,
		"demandEntitlementRatio": percent,
		"entitlement":            mhz,
		"idle":                   ms,
		"latency":                percent,
		"maxlimited":             ms,
		"overlap":                ms,
		"readiness":              percent,
		"ready":                  ms,
		"reservedCapacity":       mhz,
		"run":                    ms,
		"swapwait":               ms,
		"system":                 ms,
		"totalCapacity":          mhz,
		"usage":                  percent,
		"used":                   ms,
		"utilization":            percent,
		"wait":                   ms,
	},
	"datastore": {
		"datastoreVMObservedLatency":     ms,
		"maxTotalLatency":                ms,
		"read":                           kbs,
		"sizeNormalizedDatastoreLatency": ms,
		"totalReadLatency":               ms,
		"totalWriteLatency":              ms,
		"write":                          kbs,
	},
	"disk": {
		"deviceLatency":      ms,
		"deviceReadLatency":  ms,
		"deviceWriteLatency": ms,
		"kernelLatency":      ms,
		"kernelReadLatency":  ms,
		"kernelWriteLatency": ms,
		"maxTotalLatency":    ms,
		"queueLatency":       ms,
		"queueReadLatency":   ms,
		"queueWriteLatency":  ms,
		"read":               kbs,
		"totalReadLatency":   ms,
		"totalWriteLatency":  ms,
		"usage":              kbs,
		"write":              kbs,
	},
	"hbr": {
		"hbrNetRx": kbs,
		"hbrNetTx": kbs,
	},
	"mem": {
		"active":                 kb,
		"activewrite":            kb,
		"compressed":             kb,
		"compressionRate":        kbs,
		"consumed":               kb,
		"decompressionRate":      kbs,
		"entitlement":            kb,
		"granted":                kb,
		"heap":                   kb,
		"heapfree":               kb,
		"latency":                percent,
		"llSwapIn":               kb,
		"llSwapInRate":           kbs,
		"llSwapOut":              kb,
		"llSwapOutRate":          kbs,
		"llSwapUsed":             kb,
		"lowfreethreshold":       kb,
		"overhead":               kb,
		"overheadMax":            kb,
		"overheadTouched":        kb,
		"reservedCapacity":       mb,
		"shared":                 kb,
		"sharedcommon":           kb,
		"swapin":                 kb,
		"swapinRate":             kbs,
		"swapout":                kb,
		"swapoutRate":            kbs,
		"swapped":                kb,
		"swaptarget":             kb,
		"swapused":               kb,
		"sysUsage":               kb,
		"totalCapacity":          mb,
		"unreserved":             kb,
		"usage":                  percent,
		"vmfs.pbc.capMissRatio":  percent,
		"vmfs.pbc.overhead":      kb,
		"vmfs.pbc.size":          mb,
		"vmfs.pbc.sizeMax":       mb,
		"vmfs.pbc.workingSet":    tb,
		"vmfs.pbc.workingSetMax": tb,
		"vmmemctl":               kb,
		"vmmemctltarget":         kb,
		"zero":                   kb,
		"zipSaved":               kb,
		"zipped":                 kb,
	},
	"net": {
		"bytesRx":     kbs,
		"bytesTx":     kbs,
		"received":    kbs,
		"transmitted": kbs,
		"usage":       kbs,
	},
	"power": {
		"energy":   joules,
		"power":    watts,
		"powerCap": watts,
	},
	"rescpu": {
		"actav15":      percent,
		"actav5":       percent,
		"actpk1":       percent,
		"actpk15":      percent,
		"actpk5":       percent,
		"maxLimited1":  percent,
		"maxLimited15": percent,
		"maxLimited5":  percent,
		"runav1":       percent,
		"runav15":      percent,
		"runav5":       percent,
		"runpk1":       percent,
		"runpk15":      percent,
		"runpk5":       percent,
		"samplePeriod": ms,
	},
	"storageAdapter": {
		"maxTotalLatency":   ms,
		"read":              kbs,
		"totalReadLatency":  ms,
		"totalWriteLatency": ms,
		"write":             kbs,
	},
	"storagePath": {
		"maxTotalLatency":   ms,
		"read":              kbs,
		"totalReadLatency":  ms,
		"totalWriteLatency": ms,
		"write":             kbs,
	},
	"sys": {
		"osUptime":               seconds,
		"resourceCpuAct1":        percent,
		"resourceCpuAct5":        percent,
		"resourceCpuAllocMax":    mhz,
		"resourceCpuMaxLimited1": percent,
		"resourceCpuMaxLimited5": percent,
		"resourceCpuRun1":        percent,
		"resourceCpuRun5":        percent,
		"resourceCpuUsage":       mhz,
		"resourceMemAllocMax":    kb,
		"resourceMemAllocMin":    kb,
		"resourceMemConsumed":    kb,
		"resourceMemCow":         kb,
		"resourceMemMapped":      kb,
		"resourceMemOverhead":    kb,
		"resourceMemShared":      kb,
		"resourceMemSwapped":     kb,
		"resourceMemTouched":     kb,
		"resourceMemZero":        kb,
		"uptime":                 seconds,
	},
	"virtualDisk": {
		"read":              kbs,
		"totalReadLatency":  ms,
		"totalWriteLatency": ms,
		"write":             kbs,
	},
}
