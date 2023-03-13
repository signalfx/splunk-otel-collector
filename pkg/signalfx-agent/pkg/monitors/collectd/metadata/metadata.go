//go:build linux
// +build linux

package metadata

//go:generate ../../../../scripts/collectd-template-to-go metadata.tmpl

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/utils/hostfs"
)

var logger = log.WithFields(log.Fields{"monitorType": monitorType})

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{
			MonitorCore: *collectd.NewMonitorCore(CollectdTemplate),
			logger:      logger,
		}
	}, &Config{})
}

// Config is the monitor-specific config with the generic config embedded
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"true"`
	WriteServerURL       string `yaml:"writeServerURL"`
	// (Deprecated) Please set the agent configuration `procPath` instead of
	// this monitor configuration option.
	// The path to the proc filesystem. Useful to override in containerized
	// environments.
	ProcFSPath string `yaml:"procFSPath" default:""`
	// (Deprecated) Please set the agent configuration `etcPath` instead of this
	// monitor configuration option.
	// The path to the main host config dir. Useful to override in
	// containerized environments.
	EtcPath string `yaml:"etcPath" default:""`
	// Collect the cpu utilization per core, reported as `cpu.utilization_per_core`.
	PerCoreCPUUtil bool `yaml:"perCoreCPUUtil"`
	// A directory where the metadata plugin can persist the history of
	// successful host metadata syncs so that host metadata is not sent
	// redundantly.
	PersistencePath string `yaml:"persistencePath" default:"/var/run/signalfx-agent"`
	// If true, process "top" information will not be sent.  This can be useful
	// if you have an extremely high number of processes and performance of the
	// plugin is poor.  This defaults to `false`, but should be set to `true`
	// if using the `processlist` monitor since that duplicates this
	// functionality.
	OmitProcessInfo bool `yaml:"omitProcessInfo"`
	// Set this to a non-zero value to enable the DogStatsD listener as part of
	// this monitor.  The listener will accept metrics on the DogStatsD format,
	// and sends them as SignalFx datapoints to our backend.  Setting to a value
	// setting the `DogStatsDPort` to `0` will result in a random port assignment.
	// **Note: The listener emits directly to SignalFx and will not be subject to
	// filters configured with the SignalFx Smart Agent.  Internal stats about the
	// SignalFx Smart Agent will not reflect datapoints set through the DogStatsD listener**
	DogStatsDPort *uint `yaml:"dogStatsDPort"`
	// This is only required when running the DogStatsD listener.  Set this to
	// your SignalFx access token.
	Token string `yaml:"token" neverLog:"true"`
	// Optionally override the default ip that the DogStatsD listener listens
	// on.  (**default**: "0.0.0.0")
	DogStatsDIP string `yaml:"dogStatsDIP"`
	// This is optional only used when running the DogStatsD listener.
	// By default the DogStatsD listener will emit to SignalFx Ingest.
	// (**default**: "https://ingest.signalfx.com")
	IngestEndpoint string `yaml:"ingestEndpoint"`
	// Set this to enable verbose logging from the monitor
	Verbose bool `yaml:"verbose"`
}

// Validate will check the config for correctness.
func (c *Config) Validate() error {
	if c.DogStatsDPort != nil && c.Token == "" {
		return errors.New("you must configure 'token' with your SignalFx access token when running the DogStatsD listener")
	}
	if c.DogStatsDPort == nil && (c.Token != "" || c.IngestEndpoint != "" || c.DogStatsDIP != "") {
		return errors.New("optional DogStatsD configurations have been set, but the DogStatsDPort is not configured")
	}
	return nil
}

func (c *Config) GetExtraMetrics() []string {
	var out []string
	if c.PerCoreCPUUtil {
		out = append(out, cpuUtilizationPerCore)
	}
	return out
}

var _ config.ExtraMetrics = &Config{}

// Monitor is the main type that represents the monitor
type Monitor struct {
	collectd.MonitorCore
	logger log.FieldLogger
}

// Configure configures and runs the plugin in collectd
func (m *Monitor) Configure(conf *Config) error {
	m.logger = m.logger.WithField("monitorID", conf.MonitorID)
	conf.WriteServerURL = collectd.MainInstance().WriteServerURL()
	if conf.ProcFSPath != "" {
		m.logger.Warningf("please set the `procPath` top level agent configuration instead of the monitor level configuration")
	} else {
		// get top level configuration for /proc path
		conf.ProcFSPath = hostfs.HostProc()
	}
	if conf.EtcPath != "" {
		m.logger.Warningf("Please set the `etcPath` top level agent configuration instead of the monitor level configuration")
	} else {
		// get top level configuration for /etc path
		conf.EtcPath = hostfs.HostEtc()
	}
	return m.SetConfigurationAndRun(conf)
}
