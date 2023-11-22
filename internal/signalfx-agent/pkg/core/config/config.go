// Package config contains configuration structures and related helper logic for all
// agent components.
package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/hashstructure"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config/sources"
	"github.com/signalfx/signalfx-agent/pkg/core/config/validation"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

const (
	TraceExportFormatSAPM = "sapm"
)

// Config is the top level config struct for configurations that are common to all platforms
type Config struct {
	// Deprecated: this setting has no effect and will be removed.
	// The access token for the org that should receive the metrics emitted by
	// the agent.
	SignalFxAccessToken string `yaml:"signalFxAccessToken" neverLog:"true"`
	// Deprecated: this setting has no effect and will be removed.
	// The URL of SignalFx ingest server.  Should be overridden if using the
	// SignalFx Gateway.  If not set, this will be determined by the
	// `signalFxRealm` option below.  If you want to send trace spans to a
	// different location, set the `traceEndpointUrl` option.  If you want to
	// send events to a different location, set the `eventEndpointUrl` option.
	IngestURL string `yaml:"ingestUrl"`
	// Deprecated: this setting has no effect and will be removed.
	// The full URL (including path) to the event ingest server.  If this is
	// not set, all events will be sent to the same place as `ingestUrl`
	// above.
	EventEndpointURL string `yaml:"eventEndpointUrl"`
	// Deprecated: this setting has no effect and will be removed.
	// The full URL (including path) to the trace ingest server.  If this is
	// not set, all trace spans will be sent to the same place as `ingestUrl`
	// above.
	TraceEndpointURL string `yaml:"traceEndpointUrl"`
	// Deprecated: this setting has no effect and will be removed.
	// The SignalFx API base URL.  If not set, this will determined by the
	// `signalFxRealm` option below.
	APIURL string `yaml:"apiUrl"`
	// Deprecated: this setting has no effect and will be removed.
	// The SignalFx Realm that the organization you want to send to is a part
	// of.  This defaults to the original realm (`us0`) but if you are setting
	// up the agent for the first time, you quite likely need to change this.
	SignalFxRealm string `yaml:"signalFxRealm"`
	// Deprecated: this setting has no effect and will be removed.
	// The hostname that will be reported as the `host` dimension. If blank,
	// this will be auto-determined by the agent based on a reverse lookup of
	// the machine's IP address.
	Hostname string `yaml:"hostname"`
	// Deprecated: this setting has no effect and will be removed.
	// If true (the default), and the `hostname` option is not set, the
	// hostname will be determined by doing a reverse DNS query on the IP
	// address that is returned by querying for the bare hostname.  This is
	// useful in cases where the hostname reported by the kernel is a short
	// name. (**default**: `true`)
	UseFullyQualifiedHost *bool `yaml:"useFullyQualifiedHost"`
	// Deprecated: this setting has no effect and will be removed.
	// Our standard agent model is to collect metrics for services running on
	// the same host as the agent.  Therefore, host-specific dimensions (e.g.
	// `host`, `AWSUniqueId`, etc) are automatically added to every datapoint
	// that is emitted from the agent by default.  Set this to true if you are
	// using the agent primarily to monitor things on other hosts.  You can set
	// this option at the monitor level as well.
	DisableHostDimensions bool `yaml:"disableHostDimensions"`
	// Deprecated: this setting has no effect and will be removed.
	// How often to send metrics to SignalFx.  Monitors can override this
	// individually.
	IntervalSeconds int `yaml:"intervalSeconds"`
	// Deprecated: this setting has no effect and will be removed.
	// This flag sets the HTTP timeout duration for metadata queries from AWS, Azure and GCP.
	// This should be a duration string that is accepted by https://golang.org/pkg/time/#ParseDuration
	CloudMetadataTimeout timeutil.Duration `yaml:"cloudMetadataTimeout"`
	// Deprecated: this setting has no effect and will be removed.
	// Dimensions (key:value pairs) that will be added to every datapoint emitted by the agent.
	// To specify that all metrics should be high-resolution, add the dimension `sf_hires: 1`
	GlobalDimensions map[string]string `yaml:"globalDimensions"`
	// Deprecated: this setting has no effect and will be removed.
	// Tags (key:value pairs) that will be added to every span emitted by the agent.
	GlobalSpanTags map[string]string `yaml:"globalSpanTags"`
	// Deprecated: this setting has no effect and will be removed.
	// The logical environment/cluster that this agent instance is running in.
	// All of the services that this instance monitors should be in the same
	// environment as well. This value, if provided, will be synced as a
	// property onto the `host` dimension, or onto any cloud-provided specific
	// dimensions (`AWSUniqueId`, `gcp_id`, and `azure_resource_id`) when
	// available. Example values: "prod-usa", "dev"
	Cluster string `yaml:"cluster"`
	// Deprecated: this setting has no effect and will be removed.
	// If true, force syncing of the `cluster` property on the `host` dimension,
	// even when cloud-specific dimensions are present.
	SyncClusterOnHostDimension bool `yaml:"syncClusterOnHostDimension"`
	// Deprecated: this setting has no effect and will be removed.
	// If true, a warning will be emitted if a discovery rule contains
	// variables that will never possibly match a rule.  If using multiple
	// observers, it is convenient to set this to false to suppress spurious
	// errors.
	ValidateDiscoveryRules *bool `yaml:"validateDiscoveryRules"`
	// Deprecated: this setting has no effect and will be removed.
	// A list of observers to use (see observer config)
	Observers []ObserverConfig `yaml:"observers"`
	// Deprecated: this setting has no effect and will be removed.
	// A list of monitors to use (see monitor config)
	Monitors []MonitorConfig `yaml:"monitors"`
	// Deprecated: this setting has no effect and will be removed.
	// Configuration of the datapoint/event writer
	Writer WriterConfig `yaml:"writer"`
	// Deprecated: this setting has no effect and will be removed.
	// Log configuration
	Logging LogConfig `yaml:"logging"`
	// Configuration of the managed collectd subprocess
	Collectd CollectdConfig `yaml:"collectd" default:"{}"`
	// Deprecated: this setting has no effect and will be removed.
	// This must be unset or explicitly set to true. In prior versions of the
	// agent, there was a filtering mechanism that relied heavily on an
	// external whitelist.json file to determine which metrics were sent by
	// default.  This is all inherent to the agent now and the old style of
	// filtering is no longer available.
	EnableBuiltInFiltering *bool `yaml:"enableBuiltInFiltering"`
	// Deprecated: this setting has no effect and will be removed.
	// A list of metric filters that will include metrics.  These
	// filters take priority over the filters specified in `metricsToExclude`.
	MetricsToInclude []MetricFilter `yaml:"metricsToInclude"`
	// Deprecated: this setting has no effect and will be removed.
	// A list of metric filters
	MetricsToExclude []MetricFilter `yaml:"metricsToExclude"`
	// Deprecated: this setting has no effect and will be removed.
	// A list of properties filters
	PropertiesToExclude []PropertyFilterConfig `yaml:"propertiesToExclude"`

	// Deprecated: this setting has no effect and will be removed.
	// The host on which the internal status server will listen.  The internal
	// status HTTP server serves internal metrics and diagnostic information
	// about the agent and can be scraped by the `internal-metrics` monitor.
	// Can be set to `0.0.0.0` if you want to monitor the agent from another
	// host.  If you set this to blank/null, the internal status server will
	// not be started.  See `internalStatusPort`.
	InternalStatusHost string `yaml:"internalStatusHost"`
	// Deprecated: this setting has no effect and will be removed.
	// The port on which the internal status server will listen.  See
	// `internalStatusHost`.
	InternalStatusPort uint16 `yaml:"internalStatusPort"`
	// Deprecated: this setting has no effect and will be removed.
	// Enables Go pprof endpoint on port 6060 that serves profiling data for
	// development
	EnableProfiling bool `yaml:"profiling"`
	// Deprecated: this setting has no effect and will be removed.
	// The host/ip address for the pprof profile server to listen on.
	// `profiling` must be enabled for this to have any effect.
	ProfilingHost string `yaml:"profilingHost"`
	// Deprecated: this setting has no effect and will be removed.
	// The port for the pprof profile server to listen on. `profiling` must be
	// enabled for this to have any effect.
	ProfilingPort int `yaml:"profilingPort"`
	// Path to the directory holding the agent dependencies.  This will
	// normally be derived automatically. Overrides the envvar
	// SIGNALFX_BUNDLE_DIR if set.
	BundleDir string `yaml:"bundleDir"`
	// This exists purely to give the user a place to put common yaml values to
	// reference in other parts of the config file.
	Scratch interface{} `yaml:"scratch" neverLog:"omit"`
	// Deprecated: this setting has no effect and will be removed.
	// Configuration of remote config stores
	Sources sources.SourceConfig `yaml:"configSources"`
	// Path to the host's `/proc` filesystem.
	// This is useful for containerized environments.
	ProcPath string `yaml:"procPath" default:"/proc"`
	// Path to the host's `/etc` directory.
	// This is useful for containerized environments.
	EtcPath string `yaml:"etcPath" default:"/etc"`
	// Path to the host's `/var` directory.
	// This is useful for containerized environments.
	VarPath string `yaml:"varPath" default:"/var"`
	// Path to the host's `/run` directory.
	// This is useful for containerized environments.
	RunPath string `yaml:"runPath" default:"/run"`
	// Path to the host's `/sys` directory.
	// This is useful for containerized environments.
	SysPath string `yaml:"sysPath" default:"/sys"`
}

// Validate everything that we can about the main config
func (c *Config) validate() error {
	if err := validation.ValidateStruct(c); err != nil {
		return err
	}

	if err := c.Collectd.Validate(); err != nil {
		return err
	}

	return nil
}

// Deprecated: this setting has no effect and will be removed.
// LogConfig contains configuration related to logging
type LogConfig struct {
	// Valid levels include `debug`, `info`, `warn`, `error`.  Note that
	// `debug` logging may leak sensitive configuration (e.g. passwords) to the
	// agent output.
	Level string `yaml:"level"`
	// The log output format to use.  Valid values are: `text`, `json`.
	Format string `yaml:"format" validate:"oneof=text json"`
	// TODO: Support log file output and other log targets
}

// CollectdConfig high-level configurations
type CollectdConfig struct {
	// Deprecated: this setting has no effect and will be removed.
	// If you won't be using any collectd monitors, this can be set to true to
	// prevent collectd from pre-initializing
	DisableCollectd bool `yaml:"disableCollectd" default:"false"`
	// How many read intervals before abandoning a metric. Doesn't affect much
	// in normal usage.
	// See [Timeout](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#timeout_iterations).
	Timeout int `yaml:"timeout" default:"40"`
	// Number of threads dedicated to executing read callbacks. See
	// [ReadThreads](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#readthreads_num)
	ReadThreads int `yaml:"readThreads" default:"5"`
	// Number of threads dedicated to writing value lists to write callbacks.
	// This should be much less than readThreads because writing is batched in
	// the write_http plugin that writes back to the agent.
	// See [WriteThreads](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#writethreads_num).
	WriteThreads int `yaml:"writeThreads" default:"2"`
	// The maximum numbers of values in the queue to be written back to the
	// agent from collectd.  Since the values are written to a local socket
	// that the agent exposes, there should be almost no queuing and the
	// default should be more than sufficient. See
	// [WriteQueueLimitHigh](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#writequeuelimithigh_highnum)
	WriteQueueLimitHigh int `yaml:"writeQueueLimitHigh" default:"500000"`
	// The lowest number of values in the collectd queue before which metrics
	// begin being randomly dropped.  See
	// [WriteQueueLimitLow](https://collectd.org/documentation/manpages/collectd.conf.5.shtml#writequeuelimitlow_lownum)
	WriteQueueLimitLow int `yaml:"writeQueueLimitLow" default:"400000"`
	// Collectd's log level -- info, notice, warning, or err
	LogLevel string `yaml:"logLevel" default:"notice"`
	// A default read interval for collectd plugins.  If zero or undefined,
	// will default to the global agent interval.  Some collectd python
	// monitors do not support overridding the interval at the monitor level,
	// but this setting will apply to them.
	IntervalSeconds int `yaml:"intervalSeconds" default:"0"`
	// The local IP address of the server that the agent exposes to which
	// collectd will send metrics.  This defaults to an arbitrary address in
	// the localhost subnet, but can be overridden if needed.
	WriteServerIPAddr string `yaml:"writeServerIPAddr" default:"127.9.8.7"`
	// The port of the agent's collectd metric sink server.  If set to zero
	// (the default) it will allow the OS to assign it a free port.
	WriteServerPort uint16 `yaml:"writeServerPort" default:"0"`
	// This is where the agent will write the collectd config files that it
	// manages.  If you have secrets in those files, consider setting this to a
	// path on a tmpfs mount.  The files in this directory should be considered
	// transient -- there is no value in editing them by hand.  If you want to
	// add your own collectd config, see the collectd/custom monitor.
	ConfigDir string `yaml:"configDir" default:"/var/run/signalfx-agent/collectd"`

	// The following are propagated from the top-level config
	BundleDir            string `yaml:"-"`
	HasGenericJMXMonitor bool   `yaml:"-"`
	// Assigned by manager, not by user
	InstanceName string `yaml:"-"`
	// A hack to allow custom collectd to easily specify a single monitorID via
	// query parameter
	WriteServerQuery string          `yaml:"-"`
	Logger           log.FieldLogger `yaml:"-"`
}

// Validate the collectd specific config
func (cc *CollectdConfig) Validate() error {
	return nil
}

// Hash calculates a unique hash value for this config struct
func (cc *CollectdConfig) Hash() uint64 {
	hash, err := hashstructure.Hash(cc, nil)
	if err != nil {
		log.WithError(err).Error("Could not get hash of CollectdConfig struct")
		return 0
	}
	return hash
}

// WriteServerURL is the local address served by the agent where collect should
// write datapoints
func (cc *CollectdConfig) WriteServerURL() string {
	return fmt.Sprintf("http://%s:%d/", cc.WriteServerIPAddr, cc.WriteServerPort)
}

// InstanceConfigDir is the directory underneath the ConfigDir that is specific
// to this collectd instance.
func (cc *CollectdConfig) InstanceConfigDir() string {
	return filepath.Join(cc.ConfigDir, cc.InstanceName)
}

// ConfigFilePath returns the path where collectd should render its main config
// file.
func (cc *CollectdConfig) ConfigFilePath() string {
	return filepath.Join(cc.InstanceConfigDir(), "collectd.conf")
}

// ManagedConfigDir returns the dir path where all monitor config should go.
func (cc *CollectdConfig) ManagedConfigDir() string {
	return filepath.Join(cc.InstanceConfigDir(), "managed_config")
}

// Deprecated: no longer in use.
// StoreConfig holds configuration related to config stores (e.g. filesystem,
// zookeeper, etc)
type StoreConfig struct {
	OtherConfig map[string]interface{} `yaml:",inline,omitempty" default:"{}"`
}

// Deprecated: no longer in use.
// ExtraConfig returns generic config as a map
func (sc *StoreConfig) ExtraConfig() map[string]interface{} {
	return sc.OtherConfig
}

// BundlePythonHomeEnvvar returns an envvar string that sets the PYTHONHOME envvar to
// the bundled Python runtime.  It is in a form that is ready to append to
// cmd.Env.
func BundlePythonHomeEnvvar(bundleDir string) string {
	if runtime.GOOS == "windows" {
		return "PYTHONHOME=" + filepath.Join(bundleDir, "python")
	}
	return "PYTHONHOME=" + bundleDir
}

// AdditionalConfig is the type that should be used for any "catch-all" config
// fields in a monitor/observer.  That field should be marked as
// `yaml:",inline"`.  It will receive special handling when config is rendered
// to merge all values from multiple decoding rounds.
type AdditionalConfig map[string]interface{}
