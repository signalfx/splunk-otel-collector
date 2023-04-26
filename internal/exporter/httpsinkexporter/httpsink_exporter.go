// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpsinkexporter

import (
	"context"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/jaegertracing/jaeger/model"
	jaegertranslator "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

var spanMarshaler = &jsonpb.Marshaler{}
var metricMarshaler = &pmetric.JSONMarshaler{}

// httpSinkExporter ...
type httpSinkExporter struct {
	sink *sink

	metrics chan pmetric.Metrics
	spans   chan *model.Batch
}

func newExporter(logger *zap.Logger, endpoint string) *httpSinkExporter {
	return &httpSinkExporter{sink: newSink(logger, endpoint)}
}

func (e *httpSinkExporter) ConsumeTraces(_ context.Context, td ptrace.Traces) error {
	// TODO: replace jaeger with OTLP
	batches, err := jaegertranslator.ProtoFromTraces(td)
	if err != nil {
		return err
	}
	for _, batch := range batches {
		go func(b *model.Batch) {
			e.spans <- b
		}(batch)
	}
	return nil
}

func (e *httpSinkExporter) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
	go func(m pmetric.Metrics) {
		e.metrics <- m
	}(md)
	return nil
}

func (e *httpSinkExporter) Start(ctx context.Context, _ component.Host) error {
	e.sink.start(ctx)
	go e.fanOutSpans()
	go e.fanOutMetrics()
	return nil
}

// Shutdown stops the exporter and is invoked during shutdown.
func (e *httpSinkExporter) Shutdown(ctx context.Context) error {
	return e.sink.shutdown(ctx)
}

func (e *httpSinkExporter) fanOutSpans() error {
	e.spans = make(chan *model.Batch)
	for {
		batch := <-e.spans
		clients := e.sink.clients(typeSpans)
		for _, c := range clients {
			if !c.stopped {
				go func(c *client) {
					c.spans <- batch
				}(c)
			}
		}
	}
}

func (e *httpSinkExporter) fanOutMetrics() error {
	e.metrics = make(chan pmetric.Metrics)
	for {
		metrics := <-e.metrics
		clients := e.sink.clients(typeMetrics)
		for _, c := range clients {
			if !c.stopped {
				go func(c *client) {
					c.metrics <- metrics
				}(c)
			}
		}
	}
}
