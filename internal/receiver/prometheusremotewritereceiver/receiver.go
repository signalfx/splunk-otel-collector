// Copyright Splunk, Inc.
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

package prometheusremotewritereceiver

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
)

var _ receiver.Metrics = (*simplePrometheusWriteReceiver)(nil)

// simplePrometheusWriteReceiver implements the receiver.Metrics for PrometheusRemoteWrite protocol.
type simplePrometheusWriteReceiver struct {
	server       *PrometheusRemoteWriteServer
	reporter     iReporter
	nextConsumer consumer.Metrics
	cancel       context.CancelFunc
	config       *Config
	settings     receiver.CreateSettings
	sync.Mutex
}

// newPrometheusRemoteWriteReceiver creates the PrometheusRemoteWrite receiver with the given parameters.
func newPrometheusRemoteWriteReceiver(
	settings receiver.CreateSettings,
	config *Config,
	nextConsumer consumer.Metrics,
) (*simplePrometheusWriteReceiver, error) {
	if nextConsumer == nil {
		return nil, component.ErrNilNextConsumer
	}

	rep, err := newReporter(settings)
	if err != nil {
		return nil, err
	}

	r := &simplePrometheusWriteReceiver{
		settings:     settings,
		config:       config,
		nextConsumer: nextConsumer,
		reporter:     rep,
	}
	return r, nil
}

// Start starts an HTTP server that can process Prometheus Remote Write Requests
func (receiver *simplePrometheusWriteReceiver) Start(ctx context.Context, host component.Host) error {
	metricsChannel := make(chan pmetric.Metrics, receiver.config.BufferSize)
	cfg := &ServerConfig{
		HTTPServerSettings: receiver.config.HTTPServerSettings,
		Path:               receiver.config.ListenPath,
		Mc:                 metricsChannel,
		Reporter:           receiver.reporter,
		Host:               host,
	}
	ctx, receiver.cancel = context.WithCancel(ctx)
	server, err := NewPrometheusRemoteWriteServer(ctx, cfg)
	if err != nil {
		return err
	}
	receiver.Lock()
	defer receiver.Unlock()
	if nil != receiver.server {
		err := receiver.server.Close()
		if err != nil {
			return err
		}
	}
	receiver.server = server

	go receiver.startServer(host)
	go receiver.manageServerLifecycle(ctx, metricsChannel)

	return nil
}

func (receiver *simplePrometheusWriteReceiver) startServer(host component.Host) {
	receiver.Lock()
	prometheusRemoteWriteServer := receiver.server
	if prometheusRemoteWriteServer == nil {
		host.ReportFatalError(fmt.Errorf("start called on null prometheusRemoteWriteServer for receiver %s", typeString))
	}
	receiver.Unlock()
	if err := prometheusRemoteWriteServer.ListenAndServe(); err != nil {
		// our receiver swallows http's ErrServeClosed, and we should only get "concerning" issues at this point in the code.
		host.ReportFatalError(err)
		receiver.reporter.OnDebugf("Error in %s/%s listening on %s/%s: %s", typeString, receiver.settings.ID, prometheusRemoteWriteServer.Addr, prometheusRemoteWriteServer.Path, err)
	}
}

func (receiver *simplePrometheusWriteReceiver) manageServerLifecycle(ctx context.Context, metricsChannel <-chan pmetric.Metrics) {
	reporter := receiver.reporter
	for {
		select {
		case metrics, stillOpen := <-metricsChannel:
			if !stillOpen {
				return
			}
			metricContext := reporter.StartMetricsOp(ctx)
			err := receiver.flush(metricContext, metrics)
			if err != nil {
				reporter.OnTranslationError(metricContext, err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// Shutdown stops the PrometheusSimpleRemoteWrite receiver.
func (receiver *simplePrometheusWriteReceiver) Shutdown(context.Context) error {
	receiver.Lock()
	defer receiver.Unlock()
	if receiver.cancel == nil {
		return nil
	}
	defer receiver.cancel()
	if receiver.server != nil {
		return receiver.server.Close()
	}
	return nil
}

func (receiver *simplePrometheusWriteReceiver) flush(ctx context.Context, metrics pmetric.Metrics) error {
	err := receiver.nextConsumer.ConsumeMetrics(ctx, metrics)
	receiver.reporter.OnMetricsProcessed(ctx, metrics.DataPointCount(), err)
	return err
}
