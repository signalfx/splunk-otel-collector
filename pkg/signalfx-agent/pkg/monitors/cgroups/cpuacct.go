package cgroups

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"

	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

func (m *Monitor) getCPUAcctMetrics(controllerPath string, pathFilter filter.StringFilter) []*datapoint.Datapoint {
	if !m.Output.HasEnabledMetricInGroup(groupCpuacct) && !m.Output.HasEnabledMetricInGroup(groupCpuacctPerCPU) {
		return nil
	}

	var dps []*datapoint.Datapoint

	walkControllerHierarchy(controllerPath, func(cgroupName string, files []string) {
		if !pathFilter.Matches(cgroupName) {
			return
		}

		for _, f := range files {
			file := f
			var parseFunc func(io.Reader) (sfxclient.Collector, error)

			switch filepath.Base(f) {
			case "cpuacct.usage":
				parseFunc = parseCPUAcctUsageFile

			case "cpuacct.usage_user":
				parseFunc = parseCPUAcctUsageUserFile

			case "cpuacct.usage_sys":
				parseFunc = parseCPUAcctUsageSystemFile

			case "cpuacct.usage_percpu":
				if !m.Output.HasEnabledMetricInGroup(groupCpuacctPerCPU) {
					continue
				}

				parseFunc = parseCPUAcctUsagePerCPUFile

			case "cpuacct.usage_all":
				if !m.Output.HasEnabledMetricInGroup(groupCpuacctPerCPU) {
					continue
				}

				parseFunc = parseCPUAcctUsageAllFile

			default:
				continue
			}

			err := withOpenFile(file, func(fd *os.File) {
				usage, err := parseFunc(fd)
				if err != nil {
					m.logger.WithError(err).Errorf("Failed to parse %s", file)
					return
				}

				usageDPs := usage.Datapoints()
				for i := range usageDPs {
					usageDPs[i].Dimensions["cgroup"] = cgroupName
				}
				dps = append(dps, usageDPs...)
			})
			if err != nil {
				m.logger.WithError(err).Errorf("Could not process %s", file)
			}
		}
	})

	return dps
}

type CPUUsage struct {
	User   *int64
	System *int64
	Total  *int64
}

func (u *CPUUsage) Datapoints() []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint
	if u.Total != nil {
		dps = append(dps, sfxclient.CumulativeP(cgroupCpuacctUsageNs, map[string]string{}, u.Total))
	}

	if u.User != nil {
		dps = append(dps, sfxclient.CumulativeP(cgroupCpuacctUsageUserNs, map[string]string{}, u.User))
	}

	if u.System != nil {
		dps = append(dps, sfxclient.CumulativeP(cgroupCpuacctUsageSystemNs, map[string]string{}, u.System))
	}

	return dps
}

type PerCPUUsage struct {
	CPUUsage
	CPU int
}

func (u *PerCPUUsage) Datapoints() []*datapoint.Datapoint {
	dps := u.CPUUsage.Datapoints()
	for i := range dps {
		dps[i].Metric += "_per_cpu"
		dps[i].Dimensions["cpu"] = strconv.Itoa(u.CPU)
	}
	return dps
}

type PerCPUUsageSet []PerCPUUsage

func (ps PerCPUUsageSet) Datapoints() []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint

	for _, pcpu := range ps {
		dps = append(dps, pcpu.Datapoints()...)
	}

	return dps
}

func parseCPUAcctUsageFile(fileReader io.Reader) (sfxclient.Collector, error) {
	usage, err := parseSingleInt(fileReader)
	if err != nil {
		return nil, err
	}

	return &CPUUsage{Total: &usage}, nil
}

func parseCPUAcctUsageUserFile(fileReader io.Reader) (sfxclient.Collector, error) {
	usage, err := parseSingleInt(fileReader)
	if err != nil {
		return nil, err
	}

	return &CPUUsage{User: &usage}, nil
}

func parseCPUAcctUsageSystemFile(fileReader io.Reader) (sfxclient.Collector, error) {
	usage, err := parseSingleInt(fileReader)
	if err != nil {
		return nil, err
	}

	return &CPUUsage{System: &usage}, nil
}

func parseCPUAcctUsagePerCPUFile(fileReader io.Reader) (sfxclient.Collector, error) {
	usageBytes, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(usageBytes), " ")
	var usages PerCPUUsageSet
	for i, us := range parts {
		usage, err := strconv.ParseInt(strings.TrimSpace(us), 10, 64)
		if err != nil {
			continue
		}
		usages = append(usages, PerCPUUsage{CPUUsage: CPUUsage{Total: &usage}, CPU: i})
	}
	return usages, nil
}

func parseCPUAcctUsageAllFile(fileReader io.Reader) (sfxclient.Collector, error) {
	var usages []PerCPUUsage

	lineScanner := bufio.NewScanner(fileReader)
	for lineScanner.Scan() {
		lineParts := strings.Split(lineScanner.Text(), " ")
		if len(lineParts) != 3 {
			return nil, fmt.Errorf("unexpected line in cpuacct.usage_all: %v", lineScanner.Text())
		}
		if lineParts[0] == "cpu" {
			continue
		}

		user, _ := strconv.ParseInt(lineParts[1], 10, 64)
		system, _ := strconv.ParseInt(lineParts[2], 10, 64)

		cpu, err := strconv.ParseInt(lineParts[0], 10, 64)
		if err != nil {
			return nil, err
		}
		usage := PerCPUUsage{CPUUsage: CPUUsage{User: &user, System: &system}, CPU: int(cpu)}

		usages = append(usages, usage)
	}

	return PerCPUUsageSet(usages), nil
}
