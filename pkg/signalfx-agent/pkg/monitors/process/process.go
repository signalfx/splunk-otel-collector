package process

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/prometheus/procfs"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"true" acceptsEndpoints:"false"`

	// A list of process command names to match and send metrics about.  This
	// is the name contained in /proc/<pid>/comm and is limited to just 15
	// characters on Linux.  Only one of `processes` or `executables` must
	// match a process for it to have metrics generated about it.
	// This is an [overridable
	// set](https://docs.splunk.com/observability/gdi/smart-agent/smart-agent-resources.html#filtering-data-using-the-smart-agent)
	// that supports regexp and glob values.
	Processes []string `yaml:"processes"`

	// A list of process executable paths to match against and send metrics
	// about.  This is the binary executable that is symlinked in the
	// /proc/<pid>/exe file on Linux.  This must match the full path.  Only one
	// of `processes` or `executables` must match a process for it to have
	// metrics generated about it.
	// This is an [overridable
	// set](https://docs.splunk.com/observability/gdi/smart-agent/smart-agent-resources.html#filtering-data-using-the-smart-agent)
	// that supports regexp and glob values.
	Executables []string `yaml:"executables"`
}

func (c *Config) Validate() error {
	if len(c.Processes) == 0 && len(c.Executables) == 0 {
		return errors.New("one of processes or executables must be specified")
	}
	return nil
}

// Monitor for Utilization
type Monitor struct {
	Output types.FilteringOutput
	cancel func()
	logger logrus.FieldLogger
}

// Configure is the main function of the monitor, it will report host metadata
// on a varied interval
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	var fs procfs.FS
	var err error
	if conf.ProcPath == "" {
		fs, err = procfs.NewDefaultFS()
	} else {
		fs, err = procfs.NewFS(conf.ProcPath)
	}
	if err != nil {
		return err
	}

	var processNameFilter filter.StringFilter
	if len(conf.Processes) > 0 {
		processNameFilter, err = filter.NewOverridableStringFilter(conf.Processes)
		if err != nil {
			return err
		}
	}

	var executablesFilter filter.StringFilter
	if len(conf.Executables) > 0 {
		executablesFilter, err = filter.NewOverridableStringFilter(conf.Executables)
		if err != nil {
			return err
		}
	}

	utils.RunOnInterval(ctx, func() {
		dps, err := gatherProcessMetrics(&fs, processNameFilter, executablesFilter)
		if err != nil {
			m.logger.WithError(err).Error("Failed to gather process metrics")
			return
		}

		m.Output.SendDatapoints(dps...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

func gatherProcessMetrics(fs *procfs.FS, processNameFilter filter.StringFilter, executableFilter filter.StringFilter) ([]*datapoint.Datapoint, error) {
	procs, err := fs.AllProcs()
	if err != nil {
		return nil, err
	}

	var dps []*datapoint.Datapoint
	for i := range procs {
		if !matchesFilter(&procs[i], processNameFilter, executableFilter) {
			continue
		}

		stat, err := procs[i].Stat()
		if err != nil {
			continue
		}

		exec, err := procs[i].Executable()
		if err != nil {
			continue
		}

		dims := map[string]string{
			"pid":        strconv.Itoa(procs[i].PID),
			"command":    stat.Comm,
			"executable": exec,
		}

		dps = append(dps, []*datapoint.Datapoint{
			sfxclient.CumulativeF(processCPUTimeSeconds, dims, stat.CPUTime()),
			sfxclient.Gauge(processRssMemoryBytes, dims, int64(stat.ResidentMemory())),
		}...)
	}

	return dps, nil
}

func matchesFilter(proc *procfs.Proc, nameFilter filter.StringFilter, execFilter filter.StringFilter) bool {
	if nameFilter != nil {
		comm, err := proc.Comm()
		if err == nil {
			if matches := nameFilter.Matches(comm); matches {
				return true
			}
		}
	}
	if execFilter != nil {
		exec, err := proc.Executable()
		if err == nil {
			if matches := execFilter.Matches(exec); matches {
				return true
			}
		}
	}
	return false
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
