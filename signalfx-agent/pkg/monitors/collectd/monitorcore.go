// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	"github.com/pkg/errors"
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
}

// NewMonitorCore creates a new initialized but unconfigured MonitorCore with
// the given template.
func NewMonitorCore(template *template.Template) *MonitorCore {
	return &MonitorCore{
		Template: template,
	}
}

// Init generates a unique file name for each distinct monitor instance
func (bm *MonitorCore) Init() error {
	InjectTemplateFuncs(bm.Template)

	return nil
}

// SetCollectdInstance allows you to override the instance of collectd used by
// this monitor
func (bm *MonitorCore) SetCollectdInstance(instance *Manager) {
	bm.collectdInstanceOverride = instance
}

func (bm *MonitorCore) collectdInstance() *Manager {
	if bm.collectdInstanceOverride != nil {
		return bm.collectdInstanceOverride
	}
	return MainInstance()
}

// SetConfigurationAndRun sets the configuration to be used when rendering
// templates, and writes config before queueing a collectd restart.
func (bm *MonitorCore) SetConfigurationAndRun(conf config.MonitorCustomConfig) error {
	bm.lock.Lock()
	defer bm.lock.Unlock()

	bm.config = conf
	bm.monitorID = conf.MonitorConfigCore().MonitorID

	bm.configFilename = fmt.Sprintf("20-%s.%s.conf", bm.Template.Name(), string(bm.monitorID))

	if err := bm.WriteConfigForPlugin(); err != nil {
		return err
	}
	return bm.SetConfiguration(conf)
}

// SetConfiguration adds various fields from the config to the template context
// but does not render the config.
func (bm *MonitorCore) SetConfiguration(conf config.MonitorCustomConfig) error {
	return bm.collectdInstance().ConfigureFromMonitor(bm.monitorID, bm.Output, bm.UsesGenericJMX)
}

// WriteConfigForPlugin will render the config template to the filesystem and
// queue a collectd restart
func (bm *MonitorCore) WriteConfigForPlugin() error {
	pluginConfigText := bytes.Buffer{}

	err := bm.Template.Execute(&pluginConfigText, bm.config)
	if err != nil {
		return errors.Wrapf(err, "Could not render collectd config file for %s.  Context was %#v",
			bm.Template.Name(), bm.config)
	}

	log.WithFields(log.Fields{
		"renderPath": bm.renderPath(),
		"context":    bm.config,
	}).Debug("Writing collectd plugin config file")

	if err := WriteConfFile(pluginConfigText.String(), bm.renderPath()); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  bm.renderPath(),
		}).Error("Could not render collectd plugin config")
		return err
	}

	return nil
}

func (bm *MonitorCore) renderPath() string {
	return filepath.Join(bm.collectdInstance().ManagedConfigDir(), bm.configFilename)
}

// RemoveConfFile deletes the collectd config file for this monitor
func (bm *MonitorCore) RemoveConfFile() {
	os.Remove(bm.renderPath())
}

// Shutdown removes the config file and restarts collectd
func (bm *MonitorCore) Shutdown() {
	log.WithFields(log.Fields{
		"path": bm.renderPath(),
	}).Debug("Removing collectd plugin config")

	bm.RemoveConfFile()
	bm.collectdInstance().MonitorDidShutdown(bm.monitorID)
}
