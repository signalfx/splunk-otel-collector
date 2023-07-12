// Copyright The OpenTelemetry Authors
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

// Taken from https://github.com/open-telemetry/opentelemetry-collector/blob/v0.66.0/confmap/converter/overwritepropertiesconverter/properties_test.go
// to prevent breaking changes.
package configconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func TestOverwritePropertiesConverter_Empty(t *testing.T) {
	pmp := NewOverwritePropertiesConverter(nil)
	conf := confmap.NewFromStringMap(map[string]interface{}{"foo": "bar"})
	assert.NoError(t, pmp.Convert(context.Background(), conf))
	assert.Equal(t, map[string]interface{}{"foo": "bar"}, conf.ToStringMap())
}

func TestOverwritePropertiesConverter(t *testing.T) {
	props := []string{
		"processors.batch.timeout=2s",
		"processors.batch/foo.timeout=3s",
		"processors.batch/bar.send_batch_size=200000",
		"receivers.otlp.protocols.grpc.endpoint=localhost:1818",
		"exporters.kafka.brokers=foo:9200,foo2:9200",
		"exporters.otlp.protocols={grpc: {endpoint: localhost:1819}}",
		"processors.filter.metrics.include.metric_names=[metric.one, metric.two]",
	}

	pmp := NewOverwritePropertiesConverter(props)
	conf := confmap.New()
	require.NoError(t, pmp.Convert(context.Background(), conf))
	keys := conf.AllKeys()
	assert.Len(t, keys, 7)
	assert.Equal(t, "2s", conf.Get("processors::batch::timeout"))
	assert.Equal(t, "3s", conf.Get("processors::batch/foo::timeout"))
	assert.Equal(t, 200000, conf.Get("processors::batch/bar::send_batch_size"))
	assert.Equal(t, "foo:9200,foo2:9200", conf.Get("exporters::kafka::brokers"))
	assert.Equal(t, "localhost:1818", conf.Get("receivers::otlp::protocols::grpc::endpoint"))
	assert.Equal(t, "localhost:1819", conf.Get("exporters::otlp::protocols::grpc::endpoint"))
	assert.Equal(t, []any{"metric.one", "metric.two"}, conf.Get("processors::filter::metrics::include::metric_names"))
}

func TestOverwritePropertiesConverter_InvalidProperty(t *testing.T) {
	pmp := NewOverwritePropertiesConverter([]string{"=2s"})
	conf := confmap.New()
	assert.Error(t, pmp.Convert(context.Background(), conf))

	pmp = NewOverwritePropertiesConverter([]string{"key={:false"})
	conf = confmap.New()
	assert.EqualError(t, pmp.Convert(context.Background(), conf), "error unmarshalling \"key\" value: yaml: did not find expected node content")
}
