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
	Logger               log.FieldLogger `yaml:"-"`
	WriteServerIPAddr    string          `yaml:"writeServerIPAddr" default:"127.9.8.7"`
	WriteServerQuery     string          `yaml:"-"`
	InstanceName         string          `yaml:"-"`
	BundleDir            string          `yaml:"-"`
	ConfigDir            string          `yaml:"configDir" default:"/var/run/signalfx-agent/collectd"`
	LogLevel             string          `yaml:"logLevel" default:"notice"`
	WriteQueueLimitHigh  int             `yaml:"writeQueueLimitHigh" default:"500000"`
	IntervalSeconds      int             `yaml:"intervalSeconds" default:"0"`
	WriteQueueLimitLow   int             `yaml:"writeQueueLimitLow" default:"400000"`
	WriteThreads         int             `yaml:"writeThreads" default:"2"`
	ReadThreads          int             `yaml:"readThreads" default:"5"`
	Timeout              int             `yaml:"timeout" default:"40"`
	WriteServerPort      uint16          `yaml:"writeServerPort" default:"0"`
	DisableCollectd      bool            `yaml:"disableCollectd" default:"false"`
	HasGenericJMXMonitor bool            `yaml:"-"`
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
