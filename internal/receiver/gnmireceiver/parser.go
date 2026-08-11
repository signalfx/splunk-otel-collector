// Copyright Splunk, Inc.
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

package gnmireceiver

import (
	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// metricParser converts gNMI SubscribeResponse messages into OTel metrics.
//
// This is a stub: the value decoding, path/prefix handling, and type/unit
// resolution are implemented in a follow-up change. It currently returns no
// metrics so the connection engine can be exercised end-to-end without the
// conversion logic.
type metricParser struct{}

func newMetricParser() *metricParser {
	return &metricParser{}
}

// parse converts a single gNMI SubscribeResponse into metrics.
//
// TODO: implement value decoding. The error return will become meaningful
// once path/value parsing is added; the caller already handles it.
//
//nolint:unparam // error is always nil until value decoding is implemented
func (p *metricParser) parse(_ *gnmipb.SubscribeResponse) (pmetric.Metrics, error) {
	return pmetric.NewMetrics(), nil
}
