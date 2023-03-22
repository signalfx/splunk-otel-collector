//go:build linux
// +build linux

package vmem

import (
	"bytes"
	"context"
	"io/ioutil"
	"path"
	"strconv"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/hostfs"
)

var cumulativeCounters = map[string]string{
	"pgpgin":     vmpageIoMemoryIn,
	"pgpgout":    vmpageIoMemoryOut,
	"pswpin":     vmpageIoSwapIn,
	"pswpout":    vmpageIoSwapOut,
	"pgmajfault": vmpageFaultsMajflt,
	"pgfault":    vmpageFaultsMinflt,
}

var gauges = map[string]string{
	"nr_free_pages":      vmpageNumberFreePages,
	"nr_mapped":          vmpageNumberMapped,
	"nr_shmem_pmdmapped": vmpageNumberShmemPmdmapped,
}

func (m *Monitor) parseFileForDatapoints(contents []byte) []*datapoint.Datapoint {
	data := bytes.Fields(contents)
	max := len(data)
	dps := make([]*datapoint.Datapoint, 0, max)

	for i, key := range data {
		// vmstat file structure is (key, value)
		// so every even index is a key and every odd index is the value
		if i%2 == 0 && i+1 < max {
			metricType := datapoint.Gauge
			metricName, ok := gauges[string(key)]
			if !ok {
				metricName, ok = cumulativeCounters[string(key)]
				metricType = datapoint.Counter
			}

			// build and emit the metric if there's a metric name
			if ok {
				val, err := strconv.ParseInt(string(data[i+1]), 10, 64)
				if err != nil {
					m.logger.Errorf("failed to parse value for metric %s", metricName)
					continue
				}
				dps = append(dps, datapoint.New(metricName, nil, datapoint.NewIntValue(val), metricType, time.Time{}))
			}
		}
	}

	return dps
}

// Configure and run the monitor on linux
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	vmstatPath := path.Join(hostfs.HostProc(), "vmstat")

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		contents, err := ioutil.ReadFile(vmstatPath)
		if err != nil {
			m.logger.WithError(err).Errorf("unable to load vmstat file from path '%s'", vmstatPath)
			return
		}
		m.Output.SendDatapoints(m.parseFileForDatapoints(contents)...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}
