// Copyright Splunk, Inc.
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

package main

import (
	"context"
	"net"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunFromCmdLine(t *testing.T) {
	tests := []struct {
		name     string
		panicMsg string
		skipMsg  string
		args     []string
		timeout  time.Duration
	}{
		{
			name:    "agent",
			args:    []string{"otelcol", "--config=config/collector/agent_config.yaml"},
			timeout: 15 * time.Second,
		},
		{
			name:    "gateway",
			args:    []string{"otelcol", "--config=config/collector/gateway_config.yaml"},
			timeout: 15 * time.Second,
		},
		{
			name:    "default_discovery",
			args:    []string{"otelcol", "--discovery", "--config=config/collector/agent_config.yaml"},
			timeout: 30 * time.Second,
		},
		// Running the discovery with --dry-run in CI is not desirable because of the use of os.Exit(0) to end the execution.
		// That prevents the test from releasing resources like ports. The test needs to catch the panic to not fail the test,
		// however, the resources won't be properly released for the remaining tests that may use the same resources.
		// Skipping the test by default but keeping it around to deliberate runs on dev box.
		{
			name:     "dry-run_discovery",
			args:     []string{"otelcol", "--discovery", "--dry-run", "--config=config/collector/agent_config.yaml"},
			timeout:  30 * time.Second,
			panicMsg: "unexpected call to os.Exit(0) during test", // os.Exit(0) in the normal execution is expected for '--dry-run'.
			skipMsg:  "Skipping this test by default because --dry-run uses os.Exit(0) to end the execution",
		},
	}

	// Set execution environment
	requiredEnvVars := map[string]string{
		"SPLUNK_ACCESS_TOKEN":     "access_token",
		"SPLUNK_HEC_TOKEN":        "hec_token",
		"SPLUNK_REALM":            "test_realm",
		"SPLUNK_LISTEN_INTERFACE": "127.0.0.1",
		"NO_WINDOWS_SERVICE":      "true", // Avoid using the Windows service manager
	}
	for key, value := range requiredEnvVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipMsg != "" {
				t.Skip(tt.skipMsg)
			}

			// GH darwin runners don't have docker installed, skip discovery tests on them
			// given that the docker_observer is enabled by default.
			if runtime.GOOS == "darwin" && (tt.name == "default_discovery" || tt.name == "dry-run_discovery") {
				if os.Getenv("GITHUB_ACTIONS") == "true" {
					t.Skip("skipping discovery tests on darwin runners since they don't have docker installed")
				}
			}

			testCtx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			otelcolCmdTestCtx = testCtx //nolint:fatcontext

			// Wait for the ConfigServer to be down after the test.
			defer waitForPort(t, "55554")

			if tt.panicMsg != "" {
				assert.PanicsWithValue(t, tt.panicMsg, func() { runFromCmdLine(tt.args) })
				return
			}

			waitForPort(t, "55554")
			runFromCmdLine(tt.args)
		})
	}
}

func waitForPort(t *testing.T, port string) {
	require.Eventually(t, func() bool {
		ln, err := net.Listen("tcp", "localhost:"+port)
		if err == nil {
			ln.Close()
			return true
		}
		return false
	}, 60*time.Second, 500*time.Millisecond)
}
