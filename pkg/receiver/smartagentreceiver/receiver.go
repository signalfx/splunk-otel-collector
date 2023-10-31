// Copyright OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package smartagentreceiver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/shirou/gopsutil/v3/common"
	saconfig "github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils/hostfs"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	otelcolextension "go.opentelemetry.io/collector/extension"
	otelcolreceiver "go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension"
)

const setOutputErrMsg = "unable to set output field of monitor"
const systemTypeKey = "system.type"

type receiver struct {
	monitor             any
	nextMetricsConsumer consumer.Metrics
	nextLogsConsumer    consumer.Logs
	nextTracesConsumer  consumer.Traces
	logger              *zap.Logger
	config              *Config
	params              otelcolreceiver.CreateSettings
	sync.Mutex
}

var _ otelcolreceiver.Metrics = (*receiver)(nil)

var (
	saConfig                 *saconfig.Config
	nonWordCharacters        = regexp.MustCompile(`[^\w]+`)
	logrusShim               *logrusToZap
	configureCollectdOnce    sync.Once
	configureEnvironmentOnce sync.Once
	configureLogrusOnce      sync.Once
)

func newReceiver(params otelcolreceiver.CreateSettings, config Config) *receiver {
	return &receiver{
		logger: params.Logger,
		params: params,
		config: &config,
	}
}

func (r *receiver) registerMetricsConsumer(metricsConsumer consumer.Metrics) {
	r.Lock()
	defer r.Unlock()
	r.nextMetricsConsumer = metricsConsumer
}

func (r *receiver) registerLogsConsumer(logsConsumer consumer.Logs) {
	r.Lock()
	defer r.Unlock()
	r.nextLogsConsumer = logsConsumer
}

func (r *receiver) registerTracesConsumer(tracesConsumer consumer.Traces) {
	r.Lock()
	defer r.Unlock()
	r.nextTracesConsumer = tracesConsumer
}

func (r *receiver) Start(_ context.Context, host component.Host) error {
	// subsequent Start() invocations should noop
	if r.monitor != nil {
		return nil
	}

	err := r.config.validate()
	if err != nil {
		return fmt.Errorf("config validation failed for %q: %w", r.params.ID.String(), err)
	}

	configCore := r.config.monitorConfig.MonitorConfigCore()
	monitorType := configCore.Type
	monitorID := nonWordCharacters.ReplaceAllString(r.params.ID.String(), "")
	configCore.MonitorID = types.MonitorID(monitorID)

	configureLogrusOnce.Do(func() {
		// we need a default logger that doesn't tie to a particular receiver instance
		// but still uses the underlying service TelemetrySettings.Logger:
		logrusShim = newLogrusToZap(r.logger.With(zap.String("name", "default")))
	})

	// source logger set to the logrus StandardLogger because it is assumed that the monitor's is derived from it
	logrusShim.redirect(monitorLogrus{
		Logger:      logrus.StandardLogger(),
		monitorType: r.config.monitorConfig.MonitorConfigCore().Type,
		monitorID:   monitorID,
	}, r.logger)

	if !r.config.acceptsEndpoints {
		r.logger.Debug("This Smart Agent monitor does not use Host/Port config fields. If either are set, they will be ignored.", zap.String("monitor_type", monitorType))
	}
	r.monitor, err = r.createMonitor(monitorType, host)
	if err != nil {
		return fmt.Errorf("failed creating monitor %q: %w", monitorType, err)
	}

	configCore.ProcPath = saConfig.ProcPath
	configCore.BundleDir = saConfig.BundleDir

	return saconfig.CallConfigure(r.monitor, r.config.monitorConfig)
}

func (r *receiver) Shutdown(context.Context) error {
	if r.monitor == nil {
		return fmt.Errorf("smartagentreceiver's Shutdown() called before Start() or with invalid monitor state")
	} else if shutdownable, ok := (r.monitor).(monitors.Shutdownable); !ok {
		return fmt.Errorf("invalid monitor state at Shutdown(): %#v", r.monitor)
	} else {
		shutdownable.Shutdown()
	}
	return nil
}

func (r *receiver) createMonitor(monitorType string, host component.Host) (monitor any, err error) {
	// retrieve registered MonitorFactory from agent's registration store
	monitorFactory, ok := monitors.MonitorFactories[monitorType]
	if !ok {
		return nil, fmt.Errorf("unable to find MonitorFactory for %q", monitorType)
	}

	monitor = monitorFactory() // monitor is a pointer to a monitor struct

	// Make metadata nil if we aren't using built in filtering and then none of
	// the new filtering logic will apply.
	metadata, ok := monitors.MonitorMetadatas[monitorType]
	if !ok || metadata == nil {
		// This indicates a programming error in not specifying metadata, not
		// bad user input
		return nil, fmt.Errorf("could not find monitor metadata of type %s", monitorType)
	}
	monitorFiltering, err := newMonitorFiltering(r.config.monitorConfig, metadata, r.logger)
	if err != nil {
		return nil, err
	}

	output, err := newOutput(
		*r.config, monitorFiltering, r.nextMetricsConsumer, r.nextLogsConsumer, r.nextTracesConsumer, host, r.params,
	)
	if err != nil {
		return nil, err
	}
	set, err := setStructFieldWithExplicitType(
		monitor, "Output", output,
		reflect.TypeOf((*types.Output)(nil)).Elem(),
		reflect.TypeOf((*types.FilteringOutput)(nil)).Elem(),
	)
	if !set || err != nil {
		if err != nil {
			return nil, fmt.Errorf("%s: %w", setOutputErrMsg, err)
		}
		return nil, fmt.Errorf("%s", setOutputErrMsg)
	}

	for k, v := range r.config.monitorConfig.MonitorConfigCore().ExtraDimensions {
		output.AddExtraDimension(k, v)
	}

	output.AddExtraDimension(systemTypeKey, stripMonitorTypePrefix(monitorType))

	// Configure SmartAgentConfigProvider to gather any global config overrides and
	// set required envs.
	configureEnvironmentOnce.Do(func() {
		r.setUpSmartAgentConfigProvider(host.GetExtensions())
		setUpEnvironment()
	})

	if r.config.monitorConfig.MonitorConfigCore().IsCollectdBased() {
		configureCollectdOnce.Do(func() {
			r.logger.Info("Configuring collectd")
			err = collectd.ConfigureMainCollectd(&saConfig.Collectd)
		})
	}

	return monitor, err
}

func stripMonitorTypePrefix(s string) string {
	idx := strings.Index(s, "/")
	if idx == -1 {
		return s
	}
	return s[idx+1:]
}

func (r *receiver) setUpSmartAgentConfigProvider(extensions map[component.ID]otelcolextension.Extension) {
	// If smartagent extension is not configured, use the default config.
	f := smartagentextension.NewFactory()
	saConfig = &f.CreateDefaultConfig().(*smartagentextension.Config).Config

	// Do a lookup for any smartagent extensions to pick up common collectd options
	// to be applied across instances of the receiver.
	var foundAtLeastOne bool
	var multipleSAExtensions bool
	var chosenExtension component.ID
	for c, ext := range extensions {
		if c.Type() != f.Type() {
			continue
		}

		if foundAtLeastOne {
			multipleSAExtensions = true
			continue
		}

		var cfgProvider smartagentextension.SmartAgentConfigProvider
		cfgProvider, foundAtLeastOne = ext.(smartagentextension.SmartAgentConfigProvider)
		if !foundAtLeastOne {
			continue
		}
		saConfig = cfgProvider.SmartAgentConfig()
		chosenExtension = c
		r.logger.Info("Smart Agent Config provider configured", zap.Stringer("extension_name", chosenExtension))
	}

	if multipleSAExtensions {
		r.logger.Warn(fmt.Sprintf("multiple smartagent extensions found, using %v", chosenExtension))
	}
}

func setUpEnvironment() {
	if runtime.GOOS != "windows" { // Agent bundle doesn't include jre for Windows
		if _, ok := os.LookupEnv("JAVA_HOME"); !ok {
			os.Setenv("JAVA_HOME", filepath.Join(saConfig.BundleDir, "jre"))
		}
	}

	hostfs.SetEnvMap(common.EnvMap{
		common.HostProcEnvKey: saConfig.ProcPath,
		common.HostEtcEnvKey:  saConfig.EtcPath,
		common.HostVarEnvKey:  saConfig.VarPath,
		common.HostRunEnvKey:  saConfig.RunPath,
		common.HostSysEnvKey:  saConfig.SysPath,
	})
}
