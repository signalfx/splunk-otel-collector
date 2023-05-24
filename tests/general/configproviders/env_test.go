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
	"testing"

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
	testutils.AssertAllMetricsReceived(
		t, "memory.yaml", "", nil,
		[]testutils.CollectorBuilder{
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithEnv(
					map[string]string{"SOME_ENV_VAR": config},
				).WithArgs("--config", "env:SOME_ENV_VAR")
			},
		},
	)
}
