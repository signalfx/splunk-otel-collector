package hostmetadata

import (
	"context"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/metadata/hostmetadata"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/metadata"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

const (
	errNotAWS        = "not an aws box"
	uptimeMetricName = "sfxagent.hostmetadata"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{Monitor: metadata.Monitor{},
			startTime: time.Now()}
	}, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"true" acceptsEndpoints:"false"`
}

// Monitor for host-metadata
type Monitor struct {
	metadata.Monitor
	startTime time.Time
	cancel    func()
	logger    logrus.FieldLogger
}

// Configure is the main function of the monitor, it will report host metadata
// on a varied interval
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	// metadatafuncs are the functions to collect host metadata.
	// putting them directly in the array raised issues with the return type of info
	// By placing them inside of anonymous functions I can return (info, error)
	metadataFuncs := []func() (info, error){
		func() (info, error) { return hostmetadata.GetCPU() },
		func() (info, error) { return hostmetadata.GetMemory() },
		func() (info, error) { return hostmetadata.GetOS() },
	}

	intervals := []time.Duration{
		// on startup with some 0-60s dither
		time.Duration(rand.Int63n(60)) * time.Second,
		// 1 minute after the previous because sometimes pieces of metadata
		// aren't available immediately on startup like aws identity information
		time.Duration(60) * time.Second,
		// 1 hour after the previous with some 0-60s dither
		time.Duration(rand.Int63n(60)+3600) * time.Second,
		// 1 day after the previous with some 0-10m dither
		time.Duration(rand.Int63n(600)+86400) * time.Second,
	}

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	m.logger.Debugf("Waiting %f seconds to emit metadata", intervals[0].Seconds())

	// gather metadata on intervals
	utils.RunOnArrayOfIntervals(ctx,
		func() { m.ReportMetadataProperties(metadataFuncs) },
		intervals, utils.RepeatLast)

	// emit metadata metric
	utils.RunOnInterval(ctx,
		m.ReportUptimeMetric,
		time.Duration(conf.IntervalSeconds)*time.Second,
	)

	return nil
}

// info is an interface to the structs returned by the metadata packages in golib
type info interface {
	ToStringMap() map[string]string
}

// ReportMetadataProperties emits properties about the host
func (m *Monitor) ReportMetadataProperties(metadataFuncs []func() (info, error)) {
	for _, f := range metadataFuncs {
		meta, err := f()

		if err != nil {
			// suppress the not an aws box error message it is expected
			if err.Error() == errNotAWS {
				m.logger.Debug(err)
			} else {
				m.logger.WithError(err).Errorf("an error occurred while gathering metrics")
			}
			continue
		}

		// get the properties as a map
		properties := meta.ToStringMap()

		// emit each key/value pair
		for k, v := range properties {
			m.EmitProperty(k, v)
		}
	}
}

// ReportUptimeMetric report metrics
func (m *Monitor) ReportUptimeMetric() {
	dims := map[string]string{
		"plugin":         monitorType,
		"signalfx_agent": os.Getenv(constants.AgentVersionEnvVar),
	}

	if collectdVersion := os.Getenv(constants.CollectdVersionEnvVar); collectdVersion != "" {
		dims["collectd"] = collectdVersion
	}

	if osInfo, err := hostmetadata.GetOS(); err == nil {
		kernelName := strings.ToLower(osInfo.HostKernelName)
		dims["kernel_name"] = kernelName
		dims["kernel_release"] = osInfo.HostKernelRelease
		dims["kernel_version"] = osInfo.HostKernelVersion

		switch kernelName {
		case "windows":
			dims["os_version"] = osInfo.HostKernelRelease
		case "linux":
			dims["os_version"] = osInfo.HostLinuxVersion
		}
	}

	m.Output.SendDatapoints(
		datapoint.New(
			uptimeMetricName,
			dims,
			datapoint.NewFloatValue(time.Since(m.startTime).Seconds()),
			datapoint.Counter,
			time.Now(),
		),
	)
}
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
