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
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestCreateDefaultConfig(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, configtest.CheckConfigStruct(cfg))
	assert.Equal(t, defaultBroker, cfg.Brokers)
	assert.Equal(t, "", cfg.Topic)
}

func TestCreateMetricsExport(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Brokers = "invalid:9092"
	//cfg.ProtocolVersion = "2.0.0"
	// this disables contacting the broker so we can successfully create the exporter
	//cfg.Metadata.Full = false
	mf := pulsarExporterFactory{metricsMarshalers: metricsMarshalers()}
	mr, err := mf.createMetricsExporter(context.Background(), componenttest.NewNopExporterCreateSettings(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, mr)
}

func TestCreateMetricsExporter_err(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Brokers = "invalid:9092"
	//cfg.ProtocolVersion = "2.0.0"
	mf := pulsarExporterFactory{metricsMarshalers: metricsMarshalers()}
	mr, err := mf.createMetricsExporter(context.Background(), componenttest.NewNopExporterCreateSettings(), cfg)
	require.Error(t, err)
	assert.Nil(t, mr)
}
