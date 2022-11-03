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
	"github.com/apache/pulsar-client-go/pulsar"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// MetricsMarshaler marshals metrics into Message array
type MetricsMarshaler interface {
	// Marshal serializes metrics into pulsar's ProducerMessage
	Marshal(metrics pmetric.Metrics) ([]*pulsar.ProducerMessage, error)

	// Encoding returns encoding name
	Encoding() string
}

// metricsMarshalers returns map of supported encodings and MetricsMarshaler
func metricsMarshalers() map[string]MetricsMarshaler {
	otlpPb := newPdataMetricsMarshaler(&pmetric.ProtoMarshaler{}, defaultEncoding)
	otlpJSON := newPdataMetricsMarshaler(&pmetric.JSONMarshaler{}, "otlp_json")
	return map[string]MetricsMarshaler{
		otlpPb.Encoding():   otlpPb,
		otlpJSON.Encoding(): otlpJSON,
	}
}
