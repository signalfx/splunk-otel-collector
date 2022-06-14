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
	"go.uber.org/zap"
	"testing"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/exporter/pulsarexporter/testdata"
)


func TestCreateMetricsExport(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Broker = "pulsar+ssl://localhost:6651"
	mf := pulsarExporterFactory{metricsMarshalers: metricsMarshalers()}
	mr, err := mf.createMetricsExporter(context.Background(), componenttest.NewNopExporterCreateSettings(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, mr)
}

func TestMetricsDataPusher_integration(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Broker = "pulsar+ssl://localhost:6651"
	cfg.Authentication.TLS.InsecureSkipVerify = true
	clientOps, err := cfg.getClientOptions()
	client, err := pulsar.NewClient(clientOps)
	if err != nil {
		t.Fatal(err)
		return
	}

	defer client.Close()
	producerOps, err := cfg.getProducerOptions()
	producer, err := client.CreateProducer(producerOps)
	if err != nil {
		t.Fatal(err)
		return
	}

	p := pulsarMetricsExporter{
		producer:  producer,
		marshaler: newPdataMetricsMarshaler(pmetric.NewProtoMarshaler(), defaultEncoding),
	}
	t.Cleanup(func() {
		require.NoError(t, p.Close(context.Background()))
	})

	ms := testdata.GenerateMetricsTwoMetrics()
	err1 := p.metricsDataPusher(context.Background(), ms)
	require.NoError(t, err1)
}


func TestMetricsDataPusher_err(t *testing.T) {
	expErr := "pulsar producer failed to send metric data due to error"
	cfg := createDefaultConfig().(*Config)
	cfg.Broker = "pulsar+ssl://localhost:6651"
	cfg.Authentication.TLS.InsecureSkipVerify = true
	clientOps, err := cfg.getClientOptions()
	client, err := pulsar.NewClient(clientOps)
	if err != nil {
		t.Fatal(err)
		return
	}

	defer client.Close()
	producerOps, err := cfg.getProducerOptions()
	producer, err := client.CreateProducer(producerOps)
	if err != nil {
		t.Fatal(err)
		return
	}

	p := pulsarMetricsExporter{
		producer:  producer,
		marshaler: newPdataMetricsMarshaler(pmetric.NewProtoMarshaler(), defaultEncoding),
		logger:    zap.NewNop(),
	}
	t.Cleanup(func() {
		require.NoError(t, p.Close(context.Background()))
	})

	md := testdata.GenerateMetricsTwoMetrics()
	err1 := p.metricsDataPusher(context.Background(), md)
	assert.ErrorContains(t, err1,expErr)
}
