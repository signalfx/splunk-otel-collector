package cgroups

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"

	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

func (m *Monitor) getCPUMetrics(controllerPath string, pathFilter filter.StringFilter) []*datapoint.Datapoint {
	if !m.Output.HasEnabledMetricInGroup(groupCPU) {
		return nil
	}

	var dps []*datapoint.Datapoint

	walkControllerHierarchy(controllerPath, func(cgroupName string, files []string) {
		if !pathFilter.Matches(cgroupName) {
			return
		}

		for _, f := range files {
			var parseFunc func(io.Reader) ([]*datapoint.Datapoint, error)

			switch filepath.Base(f) {
			case "cpu.stat":
				parseFunc = parseCPUStatFile
			case "cpu.shares":
				parseFunc = parseCPUSharesFile
			case "cpu.cfs_period_us":
				parseFunc = parseCPUCFSPeriodUSFile
			case "cpu.cfs_quota_us":
				parseFunc = parseCPUCFSQuotaUSFile
			default:
				continue
			}

			var err error = nil
			var fileDPs []*datapoint.Datapoint

			err = withOpenFile(f, func(fd *os.File) {
				fileDPs, err = parseFunc(fd)
			})

			if err != nil {
				m.logger.WithError(err).Errorf("Failed to process %s", f)
				continue
			}

			for i := range fileDPs {
				fileDPs[i].Dimensions["cgroup"] = cgroupName
			}

			dps = append(dps, fileDPs...)
		}
	})

	return dps
}

func parseCPUStatFile(fileReader io.Reader) ([]*datapoint.Datapoint, error) {
	var dps []*datapoint.Datapoint

	lineScanner := bufio.NewScanner(fileReader)
	for lineScanner.Scan() {
		lineParts := strings.Split(lineScanner.Text(), " ")
		if len(lineParts) != 2 {
			return nil, fmt.Errorf("malformed line in cpu.stat file: %v", lineScanner.Text())
		}

		key := lineParts[0]
		val, err := strconv.ParseInt(lineParts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse cpu.stat value %s: %v", lineParts[1], err)
		}

		dps = append(dps, sfxclient.Cumulative("cgroup.cpu_stat_"+key, map[string]string{}, val))
	}

	return dps, nil
}

func parseCPUSharesFile(fileReader io.Reader) ([]*datapoint.Datapoint, error) {
	shares, err := parseSingleInt(fileReader)
	if err != nil {
		return nil, err
	}

	return []*datapoint.Datapoint{
		sfxclient.Gauge(cgroupCPUShares, map[string]string{}, shares),
	}, nil
}

func parseCPUCFSPeriodUSFile(fileReader io.Reader) ([]*datapoint.Datapoint, error) {
	period, err := parseSingleInt(fileReader)
	if err != nil {
		return nil, err
	}

	return []*datapoint.Datapoint{
		sfxclient.Gauge(cgroupCPUCfsPeriodUs, map[string]string{}, period),
	}, nil
}

func parseCPUCFSQuotaUSFile(fileReader io.Reader) ([]*datapoint.Datapoint, error) {
	quota, err := parseSingleInt(fileReader)
	if err != nil {
		return nil, err
	}

	return []*datapoint.Datapoint{
		sfxclient.Gauge(cgroupCPUCfsQuotaUs, map[string]string{}, quota),
	}, nil
}
