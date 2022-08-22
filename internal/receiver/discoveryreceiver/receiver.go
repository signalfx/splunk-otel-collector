// Copyright Splunk, Inc.
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

package discoveryreceiver

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/obsreport"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

var _ component.LogsReceiver = (*discoveryReceiver)(nil)

// Discovery receiver is a metrics consumer from the internal receiver creator
// but it will be evaluating them for log record conversion
var _ consumer.Metrics = (*discoveryReceiver)(nil)

type discoveryReceiver struct {
	logsConsumer      consumer.Logs
	logger            *zap.Logger
	config            *Config
	obsreportReceiver *obsreport.Receiver
	observables       map[config.ComponentID]observer.Observable
	settings          component.ReceiverCreateSettings
}

func newDiscoveryReceiver(
	settings component.ReceiverCreateSettings,
	config *Config,
	consumer consumer.Logs,
) (*discoveryReceiver, error) { // nolint:unparam
	d := &discoveryReceiver{
		config: config,
		obsreportReceiver: obsreport.NewReceiver(obsreport.ReceiverSettings{
			ReceiverID:             config.ID(),
			Transport:              "none",
			ReceiverCreateSettings: settings,
		}),
		logger:       settings.TelemetrySettings.Logger,
		settings:     settings,
		logsConsumer: consumer,
	}

	return d, nil
}

func (d *discoveryReceiver) Start(ctx context.Context, host component.Host) (err error) {
	if d.observables, err = d.observablesFromHost(host); err != nil {
		return fmt.Errorf("failed obtaining observables from host: %w", err)
	}
	return nil
}

func (d *discoveryReceiver) Shutdown(ctx context.Context) error {
	return nil
}

func (d *discoveryReceiver) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (d *discoveryReceiver) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// TODO: md -> plog.Logs and evaluation
	return nil
}

// observablesFromHost finds configured `watch_observers` extension instances from the host
// by their ComponentID. It is based on the equivalent logic in the Receiver Creator:
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/d6042eda45ec9d8a5df1ae553388eaca67d9d16c/receiver/receivercreator/receiver.go#L79
func (d *discoveryReceiver) observablesFromHost(host component.Host) (map[config.ComponentID]observer.Observable, error) {
	watchObservables := map[config.ComponentID]observer.Observable{}
	for _, obs := range d.config.WatchObservers {
		for cid, ext := range host.GetExtensions() {
			if cid != obs {
				continue
			}

			observable, ok := ext.(observer.Observable)
			if !ok {
				return nil, fmt.Errorf("extension %q in watch_observers is not an observer", obs.String())
			}
			watchObservables[obs] = observable
		}
	}

	// Make sure all specified watch_observers are present
	for _, obs := range d.config.WatchObservers {
		if watchObservables[obs] == nil {
			return nil, fmt.Errorf("failed to find observer %q as a configured extension", obs)
		}
	}
	if len(watchObservables) == 0 {
		d.logger.Warn("no observers were configured so discoveryreceiver will be inactive")
	}

	return watchObservables, nil
}
