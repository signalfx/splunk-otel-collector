// Package config contains configuration structures and related helper logic for all
// agent components.
package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/hashstructure"
	log "github.com/sirupsen/logrus"
)

// Config is the top level config struct for configurations that are common to all platforms
type Config struct {
	// Configuration of the managed collectd subprocess
	Collectd CollectdConfig `yaml:"collectd" default:"{}"`
	// Path to the directory holding the agent dependencies.  This will
	// normally be derived automatically. Overrides the envvar
	// SIGNALFX_BUNDLE_DIR if set.
	BundleDir string `yaml:"bundleDir"`
	// This exists purely to give the user a place to put common yaml values to
	// reference in other parts of the config file.
	Scratch interface{} `yaml:"scratch" neverLog:"omit"`
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

// CollectdConfig high-level configurations
type CollectdConfig struct {
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
	BundleDir string `yaml:"-"`
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
