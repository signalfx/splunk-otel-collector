// Copyright 2021, OpenTelemetry Authors
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
	"reflect"
	"strings"
	"sync"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/extension/smartagentextension"
)

const setOutputErrMsg = "unable to set Output field of monitor"

type Receiver struct {
	logger       *zap.Logger
	config       *Config
	monitor      interface{}
	nextConsumer consumer.MetricsConsumer

	startOnce sync.Once
	stopOnce  sync.Once
}

var _ component.MetricsReceiver = (*Receiver)(nil)

var (
	rusToZap           *logrusToZap
	collectdLock       sync.Mutex
	configuredCollectd bool
)

func init() {
	rusToZap = newLogrusToZap()
}

func NewReceiver(logger *zap.Logger, config Config, nextConsumer consumer.MetricsConsumer) *Receiver {
	return &Receiver{
		logger:       logger,
		config:       &config,
		nextConsumer: nextConsumer,
	}
}

func (r *Receiver) Start(_ context.Context, host component.Host) error {
	err := r.config.validate()
	if err != nil {
		return fmt.Errorf("config validation failed for %q: %w", r.config.Name(), err)
	}

	configCore := r.config.monitorConfig.MonitorConfigCore()
	monitorType := configCore.Type
	monitorName := strings.ReplaceAll(r.config.Name(), "/", "")
	configCore.MonitorID = types.MonitorID(monitorName)

	// source logger set to the standard logrus logger because it is assumed that is what the monitor is using.
	rusToZap.redirect(logrusKey{
		Logger:      logrus.StandardLogger(),
		monitorType: r.config.monitorConfig.MonitorConfigCore().Type,
	}, r.logger)

	r.monitor, err = r.createMonitor(monitorType, host.GetExtensions())
	if err != nil {
		return fmt.Errorf("failed creating monitor %q: %w", monitorType, err)
	}

	err = componenterror.ErrAlreadyStarted
	r.startOnce.Do(func() {
		// starts the monitor
		err = config.CallConfigure(r.monitor, r.config.monitorConfig)
	})
	return err
}

func (r *Receiver) Shutdown(context.Context) error {
	defer rusToZap.unRedirect(logrusKey{
		Logger:      logrus.StandardLogger(),
		monitorType: r.config.monitorConfig.MonitorConfigCore().Type,
	}, r.logger)

	err := componenterror.ErrAlreadyStopped
	if r.monitor == nil {
		err = fmt.Errorf("smartagentreceiver's Shutdown() called before Start() or with invalid monitor state")
	} else if shutdownable, ok := (r.monitor).(monitors.Shutdownable); !ok {
		err = fmt.Errorf("invalid monitor state at Shutdown(): %#v", r.monitor)
	} else {
		r.stopOnce.Do(func() {
			shutdownable.Shutdown()
			err = nil
		})
	}
	return err
}

func (r *Receiver) createMonitor(
	monitorType string,
	extensions map[configmodels.Extension]component.ServiceExtension) (interface{}, error) {
	// retrieve registered MonitorFactory from agent's registration store
	monitorFactory, ok := monitors.MonitorFactories[monitorType]
	if !ok {
		return nil, fmt.Errorf("unable to find MonitorFactory for %q", monitorType)
	}

	monitor := monitorFactory() // monitor is a pointer to a monitor struct

	output := NewOutput(*r.config, r.nextConsumer, r.logger)
	set, err := SetStructFieldWithExplicitType(
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

	collectdLock.Lock()
	defer collectdLock.Unlock()
	// Note, that this receiver has to configure main collectd even for collectd/custom monitor,
	// despite the fact that, that monitor stands up its own instance of collectd to prevent this
	// panic "Main collectd instance should not be accessed before being configured".
	if r.config.monitorConfig.MonitorConfigCore().IsCollectdBased() && !configuredCollectd {
		r.setUpCollectdConfig(extensions)
		if err := collectd.ConfigureMainCollectd(r.getCollectdConfig()); err != nil {
			return nil, err
		}
		configuredCollectd = true
	}

	return monitor, nil
}

func (r *Receiver) setUpCollectdConfig(extensions map[configmodels.Extension]component.ServiceExtension) {
	f := smartagentextension.NewFactory()
	defaultCfg := f.CreateDefaultConfig().(*smartagentextension.Config)
	r.config.collectdConfig = defaultCfg.CollectdConfig
	// Do a lookup for any smartagent extensions to pick up common collectd options
	// to be applied across instances of the receiver.
	for c := range extensions {
		if c.Type() != f.Type() {
			continue
		}

		cfg, ok := c.(*smartagentextension.Config)
		if !ok {
			continue
		}
		r.config.collectdConfig = cfg.CollectdConfig

		// If there are multiple extensions configured, pick the first one. Ideally,
		// there would only be one extension.
		break
	}
}

// Returns a configuration for collectd. Defaults provided by the receiver will overridden by
// options specified on the smartagent extension.
func (r *Receiver) getCollectdConfig() *config.CollectdConfig {
	return &config.CollectdConfig{
		DisableCollectd:      false,
		Timeout:              r.config.collectdConfig.Timeout,
		ReadThreads:          r.config.collectdConfig.ReadThreads,
		WriteThreads:         r.config.collectdConfig.WriteThreads,
		WriteQueueLimitHigh:  r.config.collectdConfig.WriteQueueLimitHigh,
		WriteQueueLimitLow:   r.config.collectdConfig.WriteQueueLimitLow,
		LogLevel:             r.config.collectdConfig.LogLevel,
		IntervalSeconds:      r.config.collectdConfig.IntervalSeconds,
		WriteServerIPAddr:    r.config.collectdConfig.WriteServerIPAddr,
		WriteServerPort:      r.config.collectdConfig.WriteServerPort,
		ConfigDir:            r.config.collectdConfig.ConfigDir,
		BundleDir:            r.config.collectdConfig.BundleDir,
		HasGenericJMXMonitor: true,
		InstanceName:         "",
		WriteServerQuery:     "",
	}
}
