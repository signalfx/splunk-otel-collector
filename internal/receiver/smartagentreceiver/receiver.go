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
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type Receiver struct {
	logger       *zap.Logger
	config       *Config
	monitor      interface{}
	nextConsumer consumer.MetricsConsumer

	startOnce sync.Once
	stopOnce  sync.Once
}

var _ component.MetricsReceiver = (*Receiver)(nil)

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

	r.monitor, err = r.createMonitor(monitorType)
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

func (r *Receiver) createMonitor(monitorType string) (interface{}, error) {
	// retrieve registered MonitorFactory from agent's registration store
	monitorFactory, ok := monitors.MonitorFactories[monitorType]
	if !ok {
		return nil, fmt.Errorf("unable to find MonitorFactory for %q", monitorType)
	}

	monitor := monitorFactory() // monitor is a pointer to a monitor struct

	// monitorOutputValue is the monitor struct's "Output" field value obtained via reflection to be set to
	// an Output instance.
	var monitorOutputValue reflect.Value
	// reflect functions will panic if received Value is not of supported type for receiver methods.
	monitorValue := reflect.Indirect(reflect.ValueOf(monitor))
	// ensure that they are only called when applicable.
	if monitorValue.IsValid() && monitorValue.Type().Kind() == reflect.Struct {
		monitorOutputValue = utils.FindFieldWithEmbeddedStructs(monitor, "Output",
			reflect.TypeOf((*types.Output)(nil)).Elem())
		if !monitorOutputValue.IsValid() {
			monitorOutputValue = utils.FindFieldWithEmbeddedStructs(monitor, "Output",
				reflect.TypeOf((*types.FilteringOutput)(nil)).Elem())
		}
	}

	var output *Output
	if monitorOutputValue.IsValid() {
		output = NewOutput(*r.config, r.nextConsumer, r.logger)
		monitorOutputValue.Set(reflect.ValueOf(output))
	} else {
		return nil, fmt.Errorf("invalid monitor instance: %#v", monitor)
	}

	for k, v := range r.config.monitorConfig.MonitorConfigCore().ExtraDimensions {
		output.AddExtraDimension(k, v)
	}

	return monitor, nil
}
