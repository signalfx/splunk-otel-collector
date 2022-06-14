// Copyright  The OpenTelemetry Authors
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

package pulsarexporter

import (
	"context"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var errUnrecognizedEncoding = fmt.Errorf("unrecognized encoding")

// pulsarMetricsExporter produce metrics messages to pulsar
type pulsarMetricsExporter struct {
	producer  pulsar.Producer
	marshaler MetricsMarshaler
	logger    *zap.Logger
	topic     string
}

func newMetricsExporter(config Config, set component.ExporterCreateSettings, marshalers map[string]MetricsMarshaler) (*pulsarMetricsExporter, error) {
	marshaler := marshalers[config.Encoding]
	if marshaler == nil {
		return nil, errUnrecognizedEncoding
	}
	producer, err := newPulsarProducer(config)

	if err != nil {
		return nil, err
	}

	return &pulsarMetricsExporter{
		producer:  producer,
		topic:     config.Topic,
		marshaler: marshaler,
		logger:    set.Logger,
	}, nil
}

func newPulsarProducer(config Config) (pulsar.Producer, error) {
	// Get pulsar client options
	clientOptions, clientOptionsErr := config.getClientOptions()
	if clientOptionsErr != nil {
		return nil, clientOptionsErr
	}

	// Initiate pulsar client
	client, clientErr := pulsar.NewClient(clientOptions)
	if clientErr != nil {
		return nil, clientErr
	}

	// Get pulsar pruducer options
	producerOptions, producerOptionsErr := config.getProducerOptions()
	if producerOptionsErr != nil {
		return nil, producerOptionsErr
	}

	// Initiate pulsar producer
	producer, producerErr := client.CreateProducer(producerOptions)
	if producerErr != nil {
		return nil, producerErr
	}

	return producer, nil
}

func (e *pulsarMetricsExporter) metricsDataPusher(ctx context.Context, md pmetric.Metrics) error {
	messages, err := e.marshaler.Marshal(md)
	if err != nil {
		return consumererror.NewPermanent(err)
	}
	var errors error
	for _, element := range messages {
		e.producer.SendAsync(ctx, element, func(_ pulsar.MessageID, _ *pulsar.ProducerMessage, err error) {
			if err != nil {
				errors = multierr.Append(errors, err)
			}
		})
	}
	if errors == nil {
		return nil
	}
	return fmt.Errorf("pulsar producer failed to send metric data due to error: %w", errors)
}

func (e *pulsarMetricsExporter) Close(context.Context) error {
	e.producer.Close()
	return nil
}
