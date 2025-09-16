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

//go:build discovery_integration_mongodb

package tests

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/internal/discoverytest"
)

func TestIntegrationMongoDBAutoDiscovery(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("Integration tests are only run on linux architecture: https://github.com/signalfx/splunk-otel-collector/blob/main/.github/workflows/integration-test.yml#L35")
	}

	tests := map[string]struct {
		configFileName     string
		logMessageToAssert string
	}{
		"Partial Discovery test": {
			configFileName:     "docker_observer_without_ssl_with_wrong_authentication_mongodb_config.yaml",
			logMessageToAssert: "Make sure your user credentials are correctly specified as environment variables.```SPLUNK_DISCOVERY_RECEIVERS_mongodb_CONFIG_username=\"<username>\"SPLUNK_DISCOVERY_RECEIVERS_mongodb_CONFIG_password=\"<password>\"```",
		},
		"Successful Discovery test": {
			configFileName:     "docker_observer_without_ssl_mongodb_config.yaml",
			logMessageToAssert: `mongodb receiver is working!`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			otelConfigPath, err := filepath.Abs(filepath.Join(".", "testdata", test.configFileName))
			require.NoError(t, err)
			discoverytest.Run(t, "mongodb", otelConfigPath, test.logMessageToAssert)
		})
	}
}
