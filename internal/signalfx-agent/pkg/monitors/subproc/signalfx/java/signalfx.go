package java

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
		return &Monitor{
			MonitorCore: subproc.New(),
		}
	}, &Config{})
}

// CustomConfig is embedded in Config struct to catch all extra config to pass to Java
type CustomConfig map[string]interface{}

// Config specifies configurations that are specific to the individual Java based monitor
type Config struct {
	CustomConfig         `yaml:",inline" json:"-" neverLog:"true"`
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	Host                 string   `yaml:"host" json:"host,omitempty"`
	JarFilePath          string   `yaml:"jarFilePath" json:"jarFilePath"`
	JavaBinary           string   `yaml:"javaBinary" json:"javaBinary"`
	MainClass            string   `yaml:"mainClass" json:"mainClass"`
	ClassPath            []string `yaml:"classPath" json:"classPath"`
	ExtraJavaArgs        []string `yaml:"extraJavaArgs" json:"extraJavaArgs"`
	Port                 uint16   `yaml:"port" json:"port,omitempty"`
}

// MarshalJSON flattens out the CustomConfig provided by the user into a single
// map so that it is simpler to access config in Java.
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

	for k, v := range c.CustomConfig {
		m[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(m)
}

// Monitor that runs Java monitors as a subprocess
type Monitor struct {
	*subproc.MonitorCore

	Output types.Output
}

func NewMonitorCore() *Monitor {
	return &Monitor{
		MonitorCore: subproc.New(),
	}
}

// Configure starts the subprocess and configures the plugin
func (m *Monitor) Configure(conf *Config) error {
	runtimeConf := subproc.DefaultJavaRuntimeConfig(conf.JarFilePath)
	if conf.JavaBinary != "" {
		runtimeConf.Binary = conf.JavaBinary
		runtimeConf.Env = os.Environ()
	} else {
		// Pass down the default runtime binary to the Java app if it
		// needs it
		conf.JavaBinary = runtimeConf.Binary
	}
	classPath := append([]string(nil), conf.ClassPath...)
	if len(conf.MainClass) > 0 {
		classPath = append(classPath, conf.JarFilePath)
	}

	if len(classPath) > 0 {
		runtimeConf.Args = append(runtimeConf.Args, []string{
			"-cp",
			strings.Join(classPath, ";"),
		}...)
	}

	runtimeConf.Args = append(runtimeConf.Args, conf.ExtraJavaArgs...)

	// This has to go last on the args
	if len(conf.MainClass) > 0 {
		runtimeConf.Args = append(runtimeConf.Args, conf.MainClass)
	} else {
		runtimeConf.Args = append(runtimeConf.Args, []string{
			"-jar",
			conf.JarFilePath,
		}...)
	}

	handler := &signalfx.JSONHandler{
		Output: m.Output,
		Logger: m.Logger(),
	}
	return m.MonitorCore.ConfigureInSubproc(conf, runtimeConf, handler)
}
