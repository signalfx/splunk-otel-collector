package cgroups

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"

	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

func (m *Monitor) getMemoryMetrics(controllerPath string, pathFilter filter.StringFilter) []*datapoint.Datapoint {
	if !m.Output.HasEnabledMetricInGroup(groupMemory) {
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
			case "memory.stat":
				parseFunc = parseMemoryStatFile

			case "memory.limit_in_bytes":
				parseFunc = parseMemoryLimitInBytesFile

			case "memory.failcnt":
				parseFunc = parseMemoryFailcntFile

			case "memory.max_usage_in_bytes":
				parseFunc = parseMemoryMaxUsageInBytesFile

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

var knownCumulatives = map[string]bool{
	"pgpgin":           true,
	"pgpgout":          true,
	"pgfault":          true,
	"pgmajfault":       true,
	"total_pgpgin":     true,
	"total_pgpgout":    true,
	"total_pgfault":    true,
	"total_pgmajfault": true,
}

func parseMemoryStatFile(fileReader io.Reader) ([]*datapoint.Datapoint, error) {
	var dps []*datapoint.Datapoint

	lineScanner := bufio.NewScanner(fileReader)
	for lineScanner.Scan() {
		lineParts := strings.Split(lineScanner.Text(), " ")
		if len(lineParts) != 2 {
			return nil, fmt.Errorf("malformed line in memory.stat file: %v", lineScanner.Text())
		}

		key := lineParts[0]
		val, err := strconv.ParseInt(lineParts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse memory.stat value %s: %v", lineParts[1], err)
		}

		typ := datapoint.Gauge
		if knownCumulatives[key] {
			typ = datapoint.Counter
		}

		dps = append(dps, datapoint.New("cgroup.memory_stat_"+key, map[string]string{}, datapoint.NewIntValue(val), typ, time.Time{}))
	}

	return dps, nil
}

func parseMemoryLimitInBytesFile(fileReader io.Reader) ([]*datapoint.Datapoint, error) {
	limit, err := parseSingleInt(fileReader)
	if err != nil {
		return nil, err
	}

	return []*datapoint.Datapoint{
		sfxclient.Gauge(cgroupMemoryLimitInBytes, map[string]string{}, limit),
	}, nil
}

func parseMemoryFailcntFile(fileReader io.Reader) ([]*datapoint.Datapoint, error) {
	count, err := parseSingleInt(fileReader)
	if err != nil {
		return nil, err
	}

	return []*datapoint.Datapoint{
		sfxclient.Cumulative(cgroupMemoryFailcnt, map[string]string{}, count),
	}, nil
}

func parseMemoryMaxUsageInBytesFile(fileReader io.Reader) ([]*datapoint.Datapoint, error) {
	limit, err := parseSingleInt(fileReader)
	if err != nil {
		return nil, err
	}

	return []*datapoint.Datapoint{
		sfxclient.Gauge(cgroupMemoryMaxUsageInBytes, map[string]string{}, limit),
	}, nil
}
