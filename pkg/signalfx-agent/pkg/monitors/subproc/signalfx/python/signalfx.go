// Package python contains a subproc implementation that works better with
// SignalFx datapoint than the collectd/python monitor, the latter being
// constrained to the collectd python interface.
package python

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc/signalfx"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &PyMonitor{
			MonitorCore: subproc.New(),
		}
	}, &Config{})
}

// PyConfig is an interface for passing in Config structs derrived from the Python Config struct
type PyConfig interface {
	config.MonitorCustomConfig
	PythonConfig() *Config
}

// Config specifies configurations that are specific to the individual python based monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	// Host will be filled in by auto-discovery if this monitor has a discovery
	// rule.
	Host string `yaml:"host" json:"host,omitempty"`
	// Port will be filled in by auto-discovery if this monitor has a discovery
	// rule.
	Port uint16 `yaml:"port" json:"port,omitempty"`
	// Path to the Python script that implements the monitoring logic.
	ScriptFilePath string `yaml:"scriptFilePath" json:"scriptFilePath"`
	// By default, the agent will use its bundled Python runtime (version 2.7).
	// If you wish to use a Python runtime that already exists on the system,
	// specify the full path to the `python` binary here, e.g.
	// `/usr/bin/python3`.
	PythonBinary string `yaml:"pythonBinary" json:"pythonBinary"`
	// The PYTHONPATH that will be used when importing the script specified at
	// `scriptFilePath`.  The directory of `scriptFilePath` will always be
	// included in the path.
	PythonPath              []string `yaml:"pythonPath" json:"pythonPath"`
	config.AdditionalConfig `yaml:",inline" json:"-" neverLog:"true"`
}

// MarshalJSON flattens out the CustomConfig provided by the user into a single
// map so that it is simpler to access config in Python.
func (c Config) MarshalJSON() ([]byte, error) {
	type ConfigX Config // prevent recursion
	b, err := json.Marshal(ConfigX(c))
	if err != nil {
		return nil, err
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	// Don't need this.
	delete(m, "OtherConfig")

	for k, v := range c.AdditionalConfig {
		m[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(m)
}

// PyMonitor that runs python monitors as a subprocess
type PyMonitor struct {
	*subproc.MonitorCore

	Output types.Output
}

// Configure starts the subprocess and configures the plugin
func (m *PyMonitor) Configure(conf *Config) error {
	runtimeConf := subproc.DefaultPythonRuntimeConfig("sfxmonitor")
	if conf.PythonBinary != "" {
		runtimeConf.Binary = conf.PythonBinary
		runtimeConf.Env = os.Environ()
	} else {
		// Pass down the default runtime binary to the Python script if it
		// needs it
		conf.PythonBinary = runtimeConf.Binary
	}
	if len(conf.PythonPath) > 0 {
		runtimeConf.Env = append(runtimeConf.Env, "PYTHONPATH="+strings.Join(conf.PythonPath, ":"))
	}

	handler := &signalfx.JSONHandler{
		Output: m.Output,
		Logger: m.Logger(),
	}
	return m.MonitorCore.ConfigureInSubproc(conf, runtimeConf, handler)
}
