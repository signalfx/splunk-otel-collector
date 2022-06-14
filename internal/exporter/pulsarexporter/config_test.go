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
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/service/servicetest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.NoError(t, err)

	factory := NewFactory()
	factories.Exporters[typeStr] = factory
	cfg, err := servicetest.LoadConfigAndValidate(filepath.Join("testdata", "config.yaml"), factories)
	require.NoError(t, err)
	require.Equal(t, 1, len(cfg.Exporters))

	configActual := cfg.Exporters[config.NewComponentID(typeStr)].(*Config)
	assert.Equal(t, &Config{
		ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
		TimeoutSettings: exporterhelper.TimeoutSettings{
			Timeout: 5 * time.Second,
		},
		RetrySettings: exporterhelper.RetrySettings{
			Enabled:         true,
			InitialInterval: 5 * time.Second,
			MaxInterval:     30 * time.Second,
			MaxElapsedTime:  5 * time.Minute,
		},
		QueueSettings: exporterhelper.QueueSettings{
			Enabled:      true,
			NumConsumers: 10,
			QueueSize:    5000,
		},
		Topic:    "otlp_metrics",
		Encoding: "otlp_proto",
		Broker:   "pulsar+ssl://localhost:6651",
		Producer: Producer{
			Name: "otel-pulsar",
			SendTimeout: 0,
			DisableBlockIfQueueFull: false,
			MaxPendingMessages: 100,
			HashingScheme: "java_string_hash",
			CompressionType: "zstd",
			CompressionLevel: defaultCompressionLevel,
			BatcherBuilderType: 1,
			DisableBatching: false,
			BatchingMaxPublishDelay: 10,
			BatchingMaxMessages: 1000,
			BatchingMaxSize: 128000,
			PartitionsAutoDiscoveryInterval: 1,
		},
		Authentication: Authentication{TLS: &configtls.TLSClientSetting{
			InsecureSkipVerify: true,
			TLSSetting: configtls.TLSSetting{
				CAFile:   "",
				CertFile: "",
				KeyFile:  "",
			},
		},
		},
	}, configActual)
}