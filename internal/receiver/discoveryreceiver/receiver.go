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
	"sync"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	mnoop "go.opentelemetry.io/otel/metric/noop"
	tnoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

const (
	observerNameAttr = "discovery.observer.name"
	observerTypeAttr = "discovery.observer.type"
)

var (
	_ receiver.Logs = (*discoveryReceiver)(nil)
)

type discoveryReceiver struct {
	nextLogsConsumer    consumer.Logs
	nextMetricsConsumer consumer.Metrics
	receiverCreator     receiver.Metrics
	alreadyLogged       *sync.Map
	endpointTracker     *endpointTracker
	sentinel            chan struct{}
	metricsConsumer     *metricsConsumer
	statementEvaluator  *statementEvaluator
	logger              *zap.Logger
	config              *Config
	obsreportReceiver   *receiverhelper.ObsReport
	pLogs               chan plog.Logs
	observables         map[component.ID]observer.Observable
	loopFinished        *sync.WaitGroup
	settings            receiver.Settings
}

func newDiscoveryReceiver(
	settings receiver.Settings,
	config *Config,
) (*discoveryReceiver, error) {
	obsReceiver, err := receiverhelper.NewObsReport(receiverhelper.ObsReportSettings{
		ReceiverID:             settings.ID,
		Transport:              "none",
		ReceiverCreateSettings: settings,
	})
	if err != nil {
		return nil, err
	}

	d := &discoveryReceiver{
		config:            config,
		obsreportReceiver: obsReceiver,
		logger:            settings.TelemetrySettings.Logger,
		settings:          settings,
		pLogs:             make(chan plog.Logs),
		sentinel:          make(chan struct{}, 1),
		loopFinished:      &sync.WaitGroup{},
		alreadyLogged:     &sync.Map{},
	}

	return d, nil
}

func (d *discoveryReceiver) Start(ctx context.Context, host component.Host) (err error) {
	if d.observables, err = d.observablesFromHost(host); err != nil {
		return fmt.Errorf("failed obtaining observables from host: %w", err)
	}

	var correlations *correlationStore
	if d.nextLogsConsumer != nil {
		correlations = newCorrelationStore(d.logger, d.config.CorrelationTTL)
		if d.nextLogsConsumer != nil {
			d.endpointTracker = newEndpointTracker(d.observables, d.config, d.logger, d.pLogs, correlations)
			d.endpointTracker.start()
		}

		if d.statementEvaluator, err = newStatementEvaluator(d.logger, d.settings.ID, d.config, correlations); err != nil {
			return fmt.Errorf("failed creating statement evaluator: %w", err)
		}
	}

	d.metricsConsumer = newMetricsConsumer(d.logger, d.config, correlations, d.nextMetricsConsumer)

	if err = d.createAndSetReceiverCreator(); err != nil {
		return fmt.Errorf("failed creating internal receiver_creator: %w", err)
	}

	if d.nextLogsConsumer != nil {
		loopStarted := &sync.WaitGroup{}
		loopStarted.Add(1)
		d.loopFinished.Add(1)
		go d.consumerLoop(loopStarted)
		// wait until we know consumer loop is running before starting receiver creator
		// so as not to miss any resulting telemetry
		loopStarted.Wait()
		d.logger.Debug("log consumer initializing initialized")
	}

	if err = d.receiverCreator.Start(ctx, host); err != nil {
		return fmt.Errorf("failed starting internal receiver_creator: %w", err)
	}
	d.logger.Debug("started receiver_creator receiver")
	return
}

func (d *discoveryReceiver) Shutdown(ctx context.Context) error {
	if d.endpointTracker != nil {
		d.endpointTracker.stop()
		d.logger.Debug("discovery receiver shutting down")
		close(d.sentinel)
		d.loopFinished.Wait()
		close(d.pLogs)
		d.logger.Debug("finished shutdown")
	}

	if d.receiverCreator != nil {
		if err := d.receiverCreator.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed shutting down internal receiver_creator: %w", err)
		}
	}

	return nil
}

func (d *discoveryReceiver) consumerLoop(loopStarted *sync.WaitGroup) {
	loopStarted.Done()
	defer d.loopFinished.Done()
	for {
		select {
		case <-d.sentinel:
			d.logger.Debug("halting consumer loop.")
			return
		case pLog, ok := <-d.pLogs:
			if !ok {
				return
			}
			ctx := d.obsreportReceiver.StartLogsOp(context.Background())
			err := d.nextLogsConsumer.ConsumeLogs(context.Background(), pLog)
			if err != nil {
				d.logger.Info("logsConsumer failed consumption", zap.Error(err))
			}
			d.obsreportReceiver.EndLogsOp(ctx, typeStr, pLog.LogRecordCount(), err)
		}
	}
}

func (d *discoveryReceiver) createAndSetReceiverCreator() error {
	receiverCreatorFactory, receiverCreatorConfig, err := d.config.receiverCreatorFactoryAndConfig()
	if err != nil {
		return err
	}
	id := component.MustNewIDWithName(receiverCreatorFactory.Type().String(), d.settings.ID.String())
	ts := component.TelemetrySettings{
		Logger:         d.logger,
		TracerProvider: tnoop.NewTracerProvider(),
		MeterProvider:  mnoop.NewMeterProvider(),
	}
	if d.statementEvaluator != nil {
		// TODO: Introduce a wrapper logger that combines the receiver_creator logger with the statement evaluator logger
		//   in a way that we avoid flooding the logs with errors but still provide enough information to debug issues.
		ts.Logger = d.statementEvaluator.evaluatedLogger.With(
			zap.String("kind", "receiver"),
			zap.String("name", id.String()),
		)
	}
	receiverCreatorSettings := receiver.Settings{
		ID:                id,
		TelemetrySettings: ts,
		BuildInfo: component.BuildInfo{
			Command: "discovery",
			Version: "latest",
		},
	}
	if d.receiverCreator, err = receiverCreatorFactory.CreateMetrics(
		context.Background(), receiverCreatorSettings, receiverCreatorConfig, d.metricsConsumer,
	); err != nil {
		return err
	}
	return nil
}

// observablesFromHost finds configured `watch_observers` extension instances from the host
// by their ComponentID. It is based on the equivalent logic in the Receiver Creator:
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/d6042eda45ec9d8a5df1ae553388eaca67d9d16c/receiver/receivercreator/receiver.go#L79
func (d *discoveryReceiver) observablesFromHost(host component.Host) (map[component.ID]observer.Observable, error) {
	watchObservables := map[component.ID]observer.Observable{}
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
