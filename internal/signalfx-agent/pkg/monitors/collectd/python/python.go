// Package python contains a monitor that runs Collectd Python plugins
// directly in a subprocess.  It uses the logic in pkg/monitors/subproc
// to do most of the work of managing a Python subprocess and doing the
// configuration/shutdown calls.
package python

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/mailru/easyjson"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	mpCollectd "github.com/signalfx/ingest-protocols/protocol/collectd"
	collectdformat "github.com/signalfx/ingest-protocols/protocol/collectd/format"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc/signalfx"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils/collectdutil"
)

const messageTypeValueList subproc.MessageType = 100

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

// CommonConfig for all Python-based monitors
type CommonConfig struct {
	// Path to a python binary that should be used to execute the Python code.
	// If not set, a built-in runtime will be used.  Can include arguments to
	// the binary as well.
	PythonBinary string `yaml:"pythonBinary" json:"pythonBinary"`
}

// Config specifies configurations that are specific to the individual python based monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	CommonConfig         `yaml:",inline"`
	// Host will be filled in by auto-discovery if this monitor has a discovery
	// rule.  It can then be used under pluginConfig by the template
	// `{{.Host}}`
	Host string `yaml:"host"`
	// Port will be filled in by auto-discovery if this monitor has a discovery
	// rule.  It can then be used under pluginConfig by the template
	// `{{.Port}}`
	Port uint16 `yaml:"port"`
	// Corresponds to the ModuleName option in collectd-python
	ModuleName string `yaml:"moduleName" json:"moduleName"`
	// Corresponds to a set of ModulePath options in collectd-python
	ModulePaths []string `yaml:"modulePaths" json:"modulePaths"`
	// This is a yaml form of the collectd config.
	PluginConfig map[string]interface{} `yaml:"pluginConfig" json:"pluginConfig" neverLog:"true"`
	// A set of paths to [../types.db files](https://collectd.org/documentation/manpages/types.db.5.shtml)
	// that are needed by your plugin.  If not specified, the runner will use
	// the global collectd ../types.db file.
	TypesDBPaths []string `yaml:"typesDBPaths" json:"typesDBPaths"`
}

// PythonConfig returns the embedded python.CoreConfig struct from the interface
func (c *Config) PythonConfig() *Config {
	return c
}

// PyMonitor that runs collectd python plugins directly
type PyMonitor struct {
	*subproc.MonitorCore

	Output types.FilteringOutput
}

// Configure starts the subprocess and configures the plugin
func (m *PyMonitor) Configure(conf PyConfig) error {
	// get the python config from the supplied config
	pyconf := conf.PythonConfig()
	if len(pyconf.TypesDBPaths) == 0 {
		pyconf.TypesDBPaths = append(pyconf.TypesDBPaths, collectd.DefaultTypesDBPath())
	}

	for k := range pyconf.PluginConfig {
		if v, ok := pyconf.PluginConfig[k].(string); ok {
			if v == "" {
				continue
			}

			template, err := template.New("nested").Parse(v)
			if err != nil {
				m.Logger().WithError(err).Errorf("Could not parse value '%s' as template", v)
				continue
			}

			out := bytes.Buffer{}
			// fill in any templates with the whole config struct passed into this method
			err = template.Option("missingkey=error").Execute(&out, conf)
			if err != nil {
				m.Logger().WithFields(log.Fields{
					"template": v,
					"error":    err,
					"context":  spew.Sdump(conf),
				}).Error("Could not render nested config template")
				continue
			}

			var result interface{} = out.String()
			if i, err := strconv.Atoi(result.(string)); err == nil {
				result = i
			}

			pyconf.PluginConfig[k] = result
		}
	}

	runtimeConf := subproc.DefaultPythonRuntimeConfig("sfxcollectd")
	pyBin := conf.PythonConfig().PythonBinary
	if pyBin != "" {
		args := strings.Fields(pyBin)
		runtimeConf.Binary = args[0]
		runtimeConf.Env = os.Environ()
		runtimeConf.Args = append(runtimeConf.Args, args[1:]...)
	}

	return m.MonitorCore.ConfigureInSubproc(pyconf, runtimeConf, m)
}

func (m *PyMonitor) ProcessMessages(ctx context.Context, dataReader subproc.MessageReceiver) {
	for {
		m.Logger().Debug("Waiting for messages")
		msgType, payloadReader, err := dataReader.RecvMessage()

		if ctx.Err() != nil {
			return
		}

		m.Logger().Debugf("Got message of type %d", msgType)

		// This is usually due to the pipe being closed
		if err != nil {
			m.Logger().WithError(err).Error("Could not receive messages")
			return
		}

		if err := m.handleMessage(msgType, payloadReader); err != nil {
			m.Logger().WithError(err).Error("Could not handle message from Python")
			continue
		}
	}
}

func (m *PyMonitor) handleMessage(msgType subproc.MessageType, payloadReader io.Reader) error {
	switch msgType {
	case messageTypeValueList:
		var valueList collectdformat.JSONWriteFormat
		if err := easyjson.UnmarshalFromReader(payloadReader, &valueList); err != nil {
			return err
		}

		dps := make([]*datapoint.Datapoint, 0)
		events := make([]*event.Event, 0)

		collectdutil.ConvertWriteFormat((*mpCollectd.JSONWriteFormat)(&valueList), &dps, &events)

		m.Output.SendDatapoints(dps...)
		for i := range events {
			m.Output.SendEvent(events[i])
		}

	case subproc.MessageTypeLog:
		return m.HandleLogMessage(payloadReader)
	default:
		return fmt.Errorf("unknown message type received %d", msgType)
	}

	return nil
}

// HandleLogMessage just passes through the reader and logger to the main JSON
// implementation
func (m *PyMonitor) HandleLogMessage(logReader io.Reader) error {
	return signalfx.HandleLogMessage(logReader, m.Logger())
}
