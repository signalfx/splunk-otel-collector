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
	"errors"
	"net"
	"net/http"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/prometheusremotewritereceiver/internal"
)

var _ receiver.Metrics = (*simplePrometheusWriteReceiver)(nil)

// simplePrometheusWriteReceiver implements the receiver.Metrics for PrometheusRemoteWrite protocol.
type simplePrometheusWriteReceiver struct {
	server       *http.Server
	reporter     internal.Reporter
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

func (receiver *simplePrometheusWriteReceiver) buildTransportServer(ctx context.Context, config *internal.ServerConfig) (*http.Server, error) {
	server, err := internal.NewPrometheusRemoteWriteServer(ctx, config)
	return server.Server, err

}

// Start starts an HTTP server that can process Prometheus Remote Write Requests
func (receiver *simplePrometheusWriteReceiver) Start(ctx context.Context, host component.Host) error {
	metricsChannel := make(chan pmetric.Metrics, receiver.config.BufferSize)
	cfg := &internal.ServerConfig{
		HTTPServerSettings: receiver.config.HTTPServerSettings,
		Path:               receiver.config.ListenPath,
		Mc:                 metricsChannel,
		Reporter:           receiver.reporter,
		Host:               host,
	}
	ctx, receiver.cancel = context.WithCancel(ctx)
	server, err := receiver.buildTransportServer(ctx, cfg)
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
	defer receiver.Unlock()
	if err := receiver.server.ListenAndServe(); err != nil {
		if !errors.Is(err, net.ErrClosed) {
			host.ReportFatalError(err)
		}
	}
}

func (receiver *simplePrometheusWriteReceiver) manageServerLifecycle(ctx context.Context, metricsChannel chan pmetric.Metrics) {
	for {
		select {
		case metrics := <-metricsChannel:
			receiver.Lock()
			metricContext := receiver.reporter.StartMetricsOp(ctx)
			err := receiver.flush(metricContext, metrics)
			if err != nil {
				receiver.reporter.OnTranslationError(metricContext, err)
				receiver.Unlock()
				close(metricsChannel)
				return
			}
			receiver.Unlock()
		case <-ctx.Done():
			close(metricsChannel)
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
	return receiver.server.Close()
}

func (receiver *simplePrometheusWriteReceiver) flush(ctx context.Context, metrics pmetric.Metrics) error {
	err := receiver.nextConsumer.ConsumeMetrics(ctx, metrics)
	receiver.reporter.OnMetricsProcessed(ctx, metrics.DataPointCount(), err)
	return err
}
