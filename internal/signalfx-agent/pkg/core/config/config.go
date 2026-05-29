// Package config contains configuration structures and related helper logic for all
// agent components.
package config


// Config is the top level config struct for configurations that are common to all platforms
type Config struct {
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

// AdditionalConfig is the type that should be used for any "catch-all" config
// fields in a monitor/observer.  That field should be marked as
// `yaml:",inline"`.  It will receive special handling when config is rendered
// to merge all values from multiple decoding rounds.
type AdditionalConfig map[string]interface{}
