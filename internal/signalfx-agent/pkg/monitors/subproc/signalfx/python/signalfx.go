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
	config.AdditionalConfig `yaml:",inline" json:"-" neverLog:"true"`
	config.MonitorConfig    `yaml:",inline" acceptsEndpoints:"true"`
	Host                    string   `yaml:"host" json:"host,omitempty"`
	ScriptFilePath          string   `yaml:"scriptFilePath" json:"scriptFilePath"`
	PythonBinary            string   `yaml:"pythonBinary" json:"pythonBinary"`
	PythonPath              []string `yaml:"pythonPath" json:"pythonPath"`
	Port                    uint16   `yaml:"port" json:"port,omitempty"`
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
	runtimeConf := subproc.DefaultPythonRuntimeConfig(conf.BundleDir, "sfxmonitor")
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
