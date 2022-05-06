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
	"github.com/apache/pulsar-client-go/pulsar"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
)

//func TestNewMetricsExporter_err_version(t *testing.T) {
//	c := Config{ProtocolVersion: "0.0.0", Encoding: defaultEncoding}
//	mexp, err := newMetricsExporter(c, componenttest.NewNopExporterCreateSettings(), metricsMarshalers())
//	assert.Error(t, err)
//	assert.Nil(t, mexp)
//}

func TestNewMetricsExporter_err_encoding(t *testing.T) {
	c := Config{Encoding: "bar"}
	mexp, err := newMetricsExporter(c, componenttest.NewNopExporterCreateSettings(), metricsMarshalers())
	assert.EqualError(t, err, errUnrecognizedEncoding.Error())
	assert.Nil(t, mexp)
}

func TestMetricsDataPusher(t *testing.T) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	defer client.Close()

	producer, err := client.CreateProducer(pulsar.ProducerOptions{})

	p := pulsarMetricsProducer{
		producer:  producer,
		marshaler: newPdataMetricsMarshaler(pmetric.NewProtoMarshaler(), defaultEncoding),
	}
	t.Cleanup(func() {
		require.NoError(t, p.Close(context.Background()))
	})
	err1 := p.metricsDataPusher(context.Background(), testdata.GenerateMetricsTwoMetrics())
	require.NoError(t, err1)
}

func TestMetricsDataPusher_err(t *testing.T) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	defer client.Close()

	producer, err := client.CreateProducer(pulsar.ProducerOptions{})

	p := pulsarMetricsProducer{
		producer:  producer,
		marshaler: newPdataMetricsMarshaler(pmetric.NewProtoMarshaler(), defaultEncoding),
		logger:    zap.NewNop(),
	}
	t.Cleanup(func() {
		require.NoError(t, p.Close(context.Background()))
	})
	md := testdata.GenerateMetricsTwoMetrics()
	err1 := p.metricsDataPusher(context.Background(), md)
	assert.EqualError(t, err1, err.Error())
}

//func TestMetricsDataPusher_marshal_error(t *testing.T) {
//	expErr := fmt.Errorf("failed to marshal")
//	p := pulsarMetricsProducer{
//		marshaler: &metricsErrorMarshaler{err: expErr},
//		logger:    zap.NewNop(),
//	}
//	md := testdata.GenerateMetricsTwoMetrics()
//	err := p.metricsDataPusher(context.Background(), md)
//	require.Error(t, err)
//	assert.Contains(t, err.Error(), expErr.Error())
//}

type metricsErrorMarshaler struct {
	err error
}

func (e metricsErrorMarshaler) Marshal(_ pmetric.Metrics, _ string) ([]*pulsar.ProducerMessage, error) {
	return nil, e.err
}

func (e metricsErrorMarshaler) Encoding() string {
	panic("implement me")
}
