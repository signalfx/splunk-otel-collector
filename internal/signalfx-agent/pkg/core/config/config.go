// Package config contains configuration structures and related helper logic for all
// agent components.
package config

import (
	"path/filepath"
	"runtime"
)

// Config is the top level config struct for configurations that are common to all platforms
type Config struct {
	// Path to the directory holding bundled monitor dependencies. This will
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
