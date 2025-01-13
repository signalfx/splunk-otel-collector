// Copyright  Splunk, Inc.
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

//go:build integration

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestEnvProvider(t *testing.T) {
	config := `receivers:
  hostmetrics:
    collection_interval: 1s
    scrapers:
      memory:

exporters:
  otlp:
    endpoint: ${env:OTLP_ENDPOINT}
    tls:
      insecure: true

service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      exporters: [otlp]
`
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollector("", func(collector testutils.Collector) testutils.Collector {
		return collector.WithEnv(
			map[string]string{"SOME_ENV_VAR": config},
		).WithArgs("--config", "env:SOME_ENV_VAR")
	})
	defer shutdown()

	missingMetrics := map[string]struct{}{
		"system.memory.usage": {},
	}

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		for i := 0; i < len(tc.OTLPReceiverSink.AllMetrics()); i++ {
			m := tc.OTLPReceiverSink.AllMetrics()[i]
			for j := 0; j < m.ResourceMetrics().Len(); j++ {
				rm := m.ResourceMetrics().At(j)
				for k := 0; k < rm.ScopeMetrics().Len(); k++ {
					sm := rm.ScopeMetrics().At(k)
					for l := 0; l < sm.Metrics().Len(); l++ {
						delete(missingMetrics, sm.Metrics().At(l).Name())
					}
				}
			}
		}
		msg := "Missing metrics:\n"
		for k := range missingMetrics {
			msg += fmt.Sprintf("- %q\n", k)
		}
		assert.Len(tt, missingMetrics, 0, msg)
	}, 30*time.Second, 1*time.Second)
}
