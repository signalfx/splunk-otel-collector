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

package pulsarexporter

import (
	"context"
	"testing"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/exporter/pulsarexporter/testdata"
)

func TestNewMetricsExporter_err_encoding(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Encoding = "bar"
	mexp, err := newMetricsExporter(*cfg, componenttest.NewNopExporterCreateSettings(), metricsMarshalers())
	assert.EqualError(t, err, errUnrecognizedEncoding.Error())
	assert.Nil(t, mexp)
}

func TestNewMetricsExporter_err_auth_type(t *testing.T) {
	c := Config{
		Authentication: Authentication{TLS: &configtls.TLSClientSetting{
			TLSSetting: configtls.TLSSetting{
				CAFile:   "",
				CertFile: "",
				KeyFile:  "",
			},
		},
		},
		Encoding: defaultEncoding,
		Producer: Producer{
			CompressionType: "none",
		},
	}
	mexp, err := newMetricsExporter(c, componenttest.NewNopExporterCreateSettings(), metricsMarshalers())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load TLS config")
	assert.Nil(t, mexp)
}

func TestMetricsDataPusher(t *testing.T) {
	producer := &mockProducer{topic: defaultName, name: defaultName}
	ms := testdata.GenerateMetricsTwoMetrics()
	err := producer.metricsDataPusher(context.Background(), ms)
	require.NoError(t, err)
}


type mockMetricsMarshaler struct {
	err error
	encoding string
}

func (e *mockMetricsMarshaler) Marshal(_ pmetric.Metrics, _ string) ([]*pulsar.ProducerMessage, error) {
	return nil, nil
}

func (e *mockMetricsMarshaler) Encoding() string {
	return e.encoding
}

type mockProducer struct {
	topic string
	name  string
}

func (c *mockProducer) Topic() string {
	return c.topic
}

func (c *mockProducer) Name() string {
	return c.name
}

func (e *mockProducer) metricsDataPusher(ctx context.Context, md pmetric.Metrics) error {
	return nil
}

func (c *mockProducer) Send(context.Context, *pulsar.ProducerMessage) (pulsar.MessageID, error) {
	return nil, nil
}

func (c *mockProducer) SendAsync(context.Context, *pulsar.ProducerMessage, func(pulsar.MessageID, *pulsar.ProducerMessage, error)) error{
	return nil
}

func (c *mockProducer) LastSequenceID() int64 {
	return 1
}

func (c *mockProducer) Flush() error {
	return nil
}

func (c *mockProducer) Close() {
}
