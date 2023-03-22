package filesystems

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	gopsutil "github.com/shirou/gopsutil/disk"
	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

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
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"false"`

	// Path to the root of the host filesystem.  Useful when running in a
	// container and the host filesystem is mounted in some subdirectory under
	// /.  The disk usage metrics emitted will be based at this path.
	HostFSPath string `yaml:"hostFSPath"`

	// The filesystem types to include/exclude.  This is an [overridable
	// set](https://docs.splunk.com/observability/gdi/smart-agent/smart-agent-resources.html#filtering-data-using-the-smart-agent).
	// If this is not set, the default value is the set of all
	// **non-logical/virtual filesystems** on the system.  On Linux this list
	// is determined by reading the `/proc/filesystems` file and choosing the
	// filesystems that do not have the `nodev` modifier.
	FSTypes []string `yaml:"fsTypes"`

	// The mount paths to include/exclude. This is an [overridable
	// set](https://docs.splunk.com/observability/gdi/smart-agent/smart-agent-resources.html#filtering-data-using-the-smart-agent).
	// NOTE: If you are using the hostFSPath option you should not include the
	// `/hostfs/` mount in the filter.  If both this and `fsTypes` is
	// specified, the two filters combine in an AND relationship.
	MountPoints []string `yaml:"mountPoints"`

	// Set to true to emit the "mode" dimension, which represents
	// whether the mount is "rw" or "ro".
	SendModeDimension bool `yaml:"sendModeDimension" default:"false"`
}

// Monitor for Utilization
type Monitor struct {
	Output            types.FilteringOutput
	cancel            func()
	conf              *Config
	hostFSPath        string
	fsTypes           *filter.OverridableStringFilter
	mountPoints       *filter.OverridableStringFilter
	sendModeDimension bool
	logger            log.FieldLogger
}

// returns common dimensions map for every filesystem
func (m *Monitor) getCommonDimensions(partition *gopsutil.PartitionStat) map[string]string {
	dims := map[string]string{
		"mountpoint": strings.Replace(partition.Mountpoint, " ", "_", -1),
		"device":     strings.Replace(partition.Device, " ", "_", -1),
		"fs_type":    strings.Replace(partition.Fstype, " ", "_", -1),
	}
	if m.sendModeDimension {
		var mode string
		opts := strings.Split(partition.Opts, ",")
		for _, opt := range opts {
			if opt == "ro" || opt == "rw" {
				mode = opt
				break
			}
		}
		if mode == "ro" || mode == "rw" {
			dims["mode"] = mode
		} else {
			m.logger.Infof("Failed to parse filesystem 'mode' for device `%s` - mountpoint `%s`", partition.Device, partition.Mountpoint)
		}
	}
	// sanitize hostfs path in mountpoint
	if m.hostFSPath != "" {
		dims["mountpoint"] = strings.Replace(dims["mountpoint"], m.hostFSPath, "", 1)
		if dims["mountpoint"] == "" {
			dims["mountpoint"] = "/"
		}
	}

	return dims
}

func (m *Monitor) makeInodeDatapoints(dimensions map[string]string, disk *gopsutil.UsageStat) []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		datapoint.New(dfInodesFree, dimensions, datapoint.NewIntValue(int64(disk.InodesFree)), datapoint.Gauge, time.Time{}),
		datapoint.New(dfInodesUsed, dimensions, datapoint.NewIntValue(int64(disk.InodesUsed)), datapoint.Gauge, time.Time{}),
		datapoint.New(percentInodesFree, dimensions, datapoint.NewIntValue(int64(100-disk.InodesUsedPercent)), datapoint.Gauge, time.Time{}),
		datapoint.New(percentInodesUsed, dimensions, datapoint.NewIntValue(int64(disk.InodesUsedPercent)), datapoint.Gauge, time.Time{}),
		// TODO: implement percent_inodes.reserved and df_inodes.reserved.
		// collectd doesn't appear to get the inode reserved stats right on
		// linux due to the glibc implementation that sets f_favail = f_ffree
		// in statvfs() (the value for these metrics are always 0 in collectd).
		// See https://sourceware.org/git/?p=glibc.git;a=blob;f=sysdeps/unix/sysv/linux/internal_statvfs.c;h=99dce0749086eaad5d592e210e2ad2176159874a;hb=HEAD#l84
	}
}

func (m *Monitor) makeDFComplex(dimensions map[string]string, disk *gopsutil.UsageStat) []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		datapoint.New(dfComplexFree, dimensions, datapoint.NewIntValue(int64(disk.Free)), datapoint.Gauge, time.Time{}),
		datapoint.New(dfComplexUsed, dimensions, datapoint.NewIntValue(int64(disk.Used)), datapoint.Gauge, time.Time{}),
		datapoint.New(dfComplexReserved, dimensions, datapoint.NewIntValue(int64(disk.Total-disk.Used-disk.Free)), datapoint.Gauge, time.Time{}),
	}
}

func (m *Monitor) makePercentBytes(dimensions map[string]string, disk *gopsutil.UsageStat) []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		datapoint.New(percentBytesFree, dimensions, datapoint.NewFloatValue(100-disk.UsedPercent), datapoint.Gauge, time.Time{}),
		datapoint.New(percentBytesUsed, dimensions, datapoint.NewFloatValue(disk.UsedPercent), datapoint.Gauge, time.Time{}),
		datapoint.New(percentBytesReserved, dimensions, datapoint.NewFloatValue((float64(disk.Total-disk.Used-disk.Free)/float64(disk.Total))*100.0), datapoint.Gauge, time.Time{}),
	}
}

// emitDatapoints emits a set of memory datapoints
func (m *Monitor) emitDatapoints() {
	// If the user has specified some fsTypes in the config then get all
	// partitions, otherwise omit the logical (virtual) filesystems by default.
	all := false
	if m.fsTypes != nil || m.mountPoints != nil {
		all = true
	}

	partitions, err := getPartitions(all)
	if err != nil && len(partitions) == 0 {
		m.logger.WithError(err).Errorf("failed to collect any mountpoints")
		return
	}

	if err != nil {
		m.logger.Debugf("failed to collect all mountpoints: %v", err)
	}

	dps := make([]*datapoint.Datapoint, 0)

	var used uint64
	var total uint64
	for i := range partitions {
		partition := partitions[i]

		// skip it if the filesystem doesn't match
		if m.fsTypes != nil && !m.fsTypes.Matches(partition.Fstype) {
			m.logger.Debugf("skipping mountpoint `%s` with fs type `%s`", partition.Mountpoint, partition.Fstype)
			continue
		}

		var mount string
		if m.hostFSPath != "" {
			mount = strings.Replace(partition.Mountpoint, m.hostFSPath, "", 1)
			if mount == "" {
				mount = "/"
			}
		} else {
			mount = partition.Mountpoint
		}

		// skip it if the mountpoint doesn't match
		if m.mountPoints != nil && !m.mountPoints.Matches(mount) {
			m.logger.Debugf("skipping mountpoint '%s'", partition.Mountpoint)
			continue
		}

		// if we can't collect usage stats about the mountpoint then skip it
		disk, err := getUsage(m.hostFSPath, partition.Mountpoint)
		if err != nil {
			m.logger.WithError(err).WithField("mountpoint", partition.Mountpoint).Debug("failed to collect usage for mountpoint")
			continue
		}

		// get common dimensions according to reportByDevice config
		commonDims := m.getCommonDimensions(&partition)

		// disk utilization
		dps = append(dps, datapoint.New(diskUtilization,
			commonDims,
			datapoint.NewFloatValue(disk.UsedPercent),
			datapoint.Gauge,
			time.Time{}),
		)

		dps = append(dps, m.makeDFComplex(commonDims, disk)...)

		if m.Output.HasEnabledMetricInGroup(groupPercentage) {
			dps = append(dps, m.makePercentBytes(commonDims, disk)...)
		}

		// inodes are not available on windows
		if runtime.GOOS != "windows" && m.Output.HasEnabledMetricInGroup(groupInodes) {
			dps = append(dps, m.makeInodeDatapoints(commonDims, disk)...)
		}

		// update totals
		used += disk.Used
		total += (disk.Used + disk.Free)
	}

	diskSummary, err := calculateUtil(float64(used), float64(total))
	if err != nil {
		m.logger.WithError(err).Errorf("failed to calculate utilization data")
	} else {
		dps = append(dps, datapoint.New(diskSummaryUtilization, nil, datapoint.NewFloatValue(diskSummary), datapoint.Gauge, time.Time{}))
	}

	m.Output.SendDatapoints(dps...)
}

// Configure is the main function of the monitor, it will report host metadata
// on a varied interval
func (m *Monitor) Configure(conf *Config) error {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// save shallow copy of conf to monitor for quick reference
	confCopy := *conf
	m.conf = &confCopy

	// configure filters
	var err error
	if len(m.conf.FSTypes) > 0 {
		m.fsTypes, err = filter.NewOverridableStringFilter(m.conf.FSTypes)
		if err != nil {
			return err
		}
	}

	// strip trailing / from HostFSPath so when we do string replacement later
	// we are left with a path starting at /
	m.hostFSPath = strings.TrimRight(m.conf.HostFSPath, "/")

	// configure filters
	if len(m.conf.MountPoints) > 0 {
		m.mountPoints, err = filter.NewOverridableStringFilter(m.conf.MountPoints)
	}

	// return an error if we can't set the filter
	if err != nil {
		return err
	}

	m.sendModeDimension = m.conf.SendModeDimension

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		m.emitDatapoints()
	}, time.Duration(m.conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}

func calculateUtil(used float64, total float64) (percent float64, err error) {
	if total == 0 {
		percent = 0
	} else {
		percent = used / total * 100
	}

	if percent < 0 {
		err = fmt.Errorf("percent %v < 0 total: %v used: %v", percent, used, total)
	}

	if percent > 100 {
		err = fmt.Errorf("percent %v > 100 total: %v used: %v ", percent, used, total)
	}
	return percent, err
}
