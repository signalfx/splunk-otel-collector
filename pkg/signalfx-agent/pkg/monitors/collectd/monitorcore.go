//go:build linux
// +build linux

package collectd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// MonitorCore contains common data/logic for collectd monitors, mainly
// stuff related to templating of the plugin config files.  This should
// generally not be used directly, but rather one of the structs that embeds
// this: StaticMonitorCore or ServiceMonitorCore.
type MonitorCore struct {
	Template *template.Template
	Output   types.FilteringOutput
	// Where to write the plugin config to on the filesystem
	configFilename           string
	config                   config.MonitorCustomConfig
	monitorID                types.MonitorID
	lock                     sync.Mutex
	UsesGenericJMX           bool
	collectdInstanceOverride *Manager
	logger                   log.FieldLogger
}

// NewMonitorCore creates a new initialized but unconfigured MonitorCore with
// the given template.
func NewMonitorCore(template *template.Template) *MonitorCore {
	return &MonitorCore{
		Template: template,
		logger:   log.StandardLogger(),
	}
}

// Init generates a unique file name for each distinct monitor instance
func (mc *MonitorCore) Init() error {
	InjectTemplateFuncs(mc.Template)

	return nil
}

// SetCollectdInstance allows you to override the instance of collectd used by
// this monitor
func (mc *MonitorCore) SetCollectdInstance(instance *Manager) {
	mc.collectdInstanceOverride = instance
}

func (mc *MonitorCore) collectdInstance() *Manager {
	if mc.collectdInstanceOverride != nil {
		return mc.collectdInstanceOverride
	}
	return MainInstance()
}

// SetConfigurationAndRun sets the configuration to be used when rendering
// templates, and writes config before queueing a collectd restart.
func (mc *MonitorCore) SetConfigurationAndRun(conf config.MonitorCustomConfig) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mConf := conf.MonitorConfigCore()
	mc.monitorID = mConf.MonitorID
	mc.logger = mc.logger.WithFields(log.Fields{"monitorType": conf.MonitorConfigCore().Type, "monitorID": string(mc.monitorID)})

	if mConf.IsolatedCollectd {
		cconf := *MainInstance().Config()
		cconf.WriteServerPort = 0
		cconf.WriteServerQuery = "?monitorID=" + string(mConf.MonitorID)
		cconf.InstanceName = "monitor-" + string(mConf.MonitorID)
		cconf.ReadThreads = 10
		cconf.WriteThreads = 1
		cconf.WriteQueueLimitHigh = 10000
		cconf.WriteQueueLimitLow = 10000
		cconf.IntervalSeconds = mConf.IntervalSeconds
		cconf.Logger = mc.logger
		mc.logger.Info(fmt.Sprintf("starting isolated configd instance %q", cconf.InstanceName))
		mc.SetCollectdInstance(InitCollectd(&cconf))
	}

	mc.config = conf
	mc.configFilename = fmt.Sprintf("20-%s.%s.conf", mc.Template.Name(), string(mc.monitorID))

	if err := mc.WriteConfigForPlugin(); err != nil {
		return err
	}
	return mc.SetConfiguration()
}

// SetConfiguration adds various fields from the config to the template context
// but does not render the config.
func (mc *MonitorCore) SetConfiguration() error {
	return mc.collectdInstance().ConfigureFromMonitor(mc.monitorID, mc.Output, mc.UsesGenericJMX)
}

// WriteConfigForPlugin will render the config template to the filesystem and
// queue a collectd restart
func (mc *MonitorCore) WriteConfigForPlugin() error {
	pluginConfigText := bytes.Buffer{}

	err := mc.Template.Execute(&pluginConfigText, mc.config)
	if err != nil {
		return fmt.Errorf("Could not render collectd config file for %s.  Context was %#v %w",
			mc.Template.Name(), mc.config, err)
	}

	mc.logger.WithFields(log.Fields{
		"renderPath": mc.renderPath(),
		"context":    mc.config,
	}).Debug("Writing collectd plugin config file")

	if err := WriteConfFile(pluginConfigText.String(), mc.renderPath()); err != nil {
		mc.logger.WithFields(log.Fields{
			"error": err,
			"path":  mc.renderPath(),
		}).Error("Could not render collectd plugin config")
		return err
	}

	return nil
}

func (mc *MonitorCore) renderPath() string {
	return filepath.Join(mc.collectdInstance().ManagedConfigDir(), mc.configFilename)
}

// RemoveConfFile deletes the collectd config file for this monitor
func (mc *MonitorCore) RemoveConfFile() {
	os.Remove(mc.renderPath())
}

// Shutdown removes the config file and restarts collectd
func (mc *MonitorCore) Shutdown() {
	mc.logger.WithFields(log.Fields{
		"path": mc.renderPath(),
	}).Debug("Removing collectd plugin config")

	mc.RemoveConfFile()
	mc.collectdInstance().MonitorDidShutdown(mc.monitorID)
}
