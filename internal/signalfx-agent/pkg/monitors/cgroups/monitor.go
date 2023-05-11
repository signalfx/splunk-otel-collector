package cgroups

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/prometheus/procfs"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"false"`

	// The cgroup names to include/exclude, based on the full hierarchy path.
	// This is an [overridable
	// set](https://docs.splunk.com/observability/gdi/smart-agent/smart-agent-resources.html#filtering-data-using-the-smart-agent).
	// If not provided, this defaults to all cgroups.
	// E.g. to monitor all Docker container cgroups, you could use a value of
	// `["/docker/*"]`.
	Cgroups []string `yaml:"cgroups"`
}

func (c *Config) Validate() error {
	if runtime.GOOS != "linux" {
		return errors.New("the cgroup monitor only works on Linux")
	}
	return nil
}

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

type Monitor struct {
	Output types.FilteringOutput
	cancel context.CancelFunc
	logger logrus.FieldLogger
}

// Configure the monitor and start collection
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	cgroups := conf.Cgroups
	if len(cgroups) == 0 {
		cgroups = []string{"*"}
	}

	pathFilter, err := filter.NewOverridableStringFilter(cgroups)
	if err != nil {
		return err
	}

	utils.RunOnInterval(ctx, func() {
		controllerPaths, err := getCgroupControllerPaths(conf.ProcPath)
		if err != nil {
			m.logger.WithError(err).Error("Failed to get cgroup controller roots")
			return
		}

		var dps []*datapoint.Datapoint

		if controllerPaths.CPUAcct != "" {
			dps = append(dps, m.getCPUMetrics(controllerPaths.CPU, pathFilter)...)
			dps = append(dps, m.getCPUAcctMetrics(controllerPaths.CPUAcct, pathFilter)...)
			dps = append(dps, m.getMemoryMetrics(controllerPaths.Memory, pathFilter)...)
		}

		m.Output.SendDatapoints(dps...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Cgroups controllers can only have one hierarchy, so even it is mounted
// multiple times, all of the mounts are a mirror of each other.  Therefore,
// having a single path for each controller is fine.
type CgroupControllerPaths struct {
	CPU     string
	CPUAcct string
	Memory  string
}

func getCgroupControllerPaths(procPath string) (*CgroupControllerPaths, error) {
	var fs procfs.FS
	var err error
	if procPath == "" {
		fs, err = procfs.NewDefaultFS()
	} else {
		fs, err = procfs.NewFS(procPath)
	}
	if err != nil {
		return nil, err
	}

	initProc, err := fs.Proc(1)
	if err != nil {
		return nil, fmt.Errorf("could not get init process: %v", err)
	}

	mounts, err := initProc.MountInfo()
	if err != nil {
		return nil, fmt.Errorf("could not get mount info for init process: %v", err)
	}

	var paths CgroupControllerPaths

	for _, mount := range mounts {
		if mount.FSType != "cgroup" {
			continue
		}

		path := mount.MountPoint

		for opt := range mount.SuperOptions {
			switch opt {
			case "cpuacct":
				paths.CPUAcct = path
			case "cpu":
				paths.CPU = path
			case "memory":
				paths.Memory = path
			}
		}
	}

	return &paths, nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
