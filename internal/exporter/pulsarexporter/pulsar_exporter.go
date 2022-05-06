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
	"log"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

var errUnrecognizedEncoding = fmt.Errorf("unrecognized encoding")

// pulsarMetricsProducer uses sarama to produce metrics messages to pulsar
type pulsarMetricsProducer struct {
	producer  pulsar.Producer
	marshaler MetricsMarshaler
	logger    *zap.Logger
}

type pulsarErrors struct {
	err string
}

func (pe pulsarErrors) Error() string {
	return fmt.Sprintf("Failed to deliver messages due to %s", pe.err)
}

func (e *pulsarMetricsProducer) metricsDataPusher(_ context.Context, md pmetric.Metrics) error {
	messages, err := e.marshaler.Marshal(md)
	if err != nil {
		return consumererror.NewPermanent(err)
	}
	for _, element := range messages { // (1) pulsar at the time of implementing this did not support sending an array of messages. (2) Use array index if needed to propagate back the errors
		_, err = e.producer.Send(context.Background(), element)
	}
	if err != nil {
		log.Fatal(err)
		return pulsarErrors{err.Error()}
	}
	return nil
}

func newPulsarProducer(config Config) (pulsar.Producer, error) {

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                        config.Brokers,
		OperationTimeout:           30 * time.Second,
		ConnectionTimeout:          30 * time.Second,
		TLSAllowInsecureConnection: true,
	})
	if err != nil {
		log.Fatalf("Could not instantiate Pulsar client: %v", err)
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: config.Topic,
	})

	if err != nil {
		return nil, err
	}
	return producer, nil
}

func newMetricsExporter(config Config, set component.ExporterCreateSettings, marshalers map[string]MetricsMarshaler) (*pulsarMetricsProducer, error) {
	marshaler := marshalers[config.Encoding]
	if marshaler == nil {
		return nil, errUnrecognizedEncoding
	}
	producer, err := newPulsarProducer(config)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		return nil, err
	}

	return &pulsarMetricsProducer{
		producer:  producer,
		marshaler: marshaler,
		logger:    set.Logger,
	}, nil
}

func (e *pulsarMetricsProducer) Close(context.Context) error {
	e.producer.Close()
	return nil // TODO: check return
}
