// Copyright Splunk, Inc.
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

//go:build discovery_integration_kafkametrics

package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/internal/discoverytest"
)

func TestIntegrationKafkaMetricsAutoDiscovery(t *testing.T) {
	t.Skip("Skipping test as a known issue fixed in 0.124.0")
	tests := map[string]struct {
		configFileName     string
		logMessageToAssert string
	}{
		"Successful Discovery test": {
			configFileName:     "docker_observer_without_ssl_kafkametrics_config.yaml",
			logMessageToAssert: `kafkametrics receiver is working!`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			otelConfigPath, err := filepath.Abs(filepath.Join(".", "testdata", test.configFileName))
			require.NoError(t, err)
			discoverytest.Run(t, "kafkametrics", otelConfigPath, test.logMessageToAssert)
		})
	}
}
