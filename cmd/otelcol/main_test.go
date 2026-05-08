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

// This test is flaky w/ data race validation enabled. The issue is that monitors do not guarantee
// that no more data is going to be sent after shutdown is called. This can cause data races with
// processors and exporters that have been shut down. See https://github.com/signalfx/splunk-otel-collector/pull/7265.
// The build directive below should be removed once the monitors are not supported anymore.
//go:build !race

package main

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunFromCmdLine(t *testing.T) {
	configFiles, err := filepath.Glob(filepath.Join("config", "collector", "*.yml"))
	require.NoError(t, err)
	require.Empty(t, configFiles, "Default configs must end with the extension `.yaml` rather than `.yml`. Please update accordingly.")

	configFiles, err = filepath.Glob(filepath.Join("config", "collector", "*.yaml"))
	require.NoError(t, err)
	require.Len(t, configFiles, 8, "A new test case must be added to TestRunFromCmdLine to validate default configurations")

	tests := []struct {
		name         string
		panicMsg     string
		skipMsg      string
		extraEnvVars map[string]string
		args         []string
		skipWindows  bool
		timeout      time.Duration
		validateOnly bool
	}{
		{
			name:    "agent",
			args:    []string{"otelcol", "--config=config/collector/agent_config.yaml"},
			timeout: 15 * time.Second,
		},
		{
			name: "ecs_ec2",
			args: []string{"otelcol", "--config=config/collector/ecs_ec2_config.yaml"},
			extraEnvVars: map[string]string{
				"ECS_CONTAINER_METADATA_URI_V4": "https://foo.com",
			},
			timeout:      15 * time.Second,
			validateOnly: true,
		},
		{
			name: "fargate",
			args: []string{"otelcol", "--config=config/collector/fargate_config.yaml"},
			extraEnvVars: map[string]string{
				"ECS_CONTAINER_METADATA_URI_V4": "https://foo.com",
			},
			timeout:      15 * time.Second,
			validateOnly: true,
		},
		{
			name:         "full_linux",
			args:         []string{"otelcol", "--config=config/collector/full_config_linux.yaml"},
			timeout:      15 * time.Second,
			validateOnly: true,
			skipWindows:  true,
		},
		{
			name:    "gateway",
			args:    []string{"otelcol", "--config=config/collector/gateway_config.yaml"},
			timeout: 15 * time.Second,
		},
		{
			name:         "logs_linux",
			args:         []string{"otelcol", "--config=config/collector/logs_config_linux.yaml"},
			timeout:      15 * time.Second,
			validateOnly: true,
			skipWindows:  true,
		},
		{
			name:         "otlp_linux",
			args:         []string{"otelcol", "--config=config/collector/otlp_config_linux.yaml"},
			timeout:      15 * time.Second,
			validateOnly: true,
			skipWindows:  true,
		},
		{
			name:         "upstream_agent",
			args:         []string{"otelcol", "--config=config/collector/upstream_agent_config.yaml"},
			timeout:      15 * time.Second,
			validateOnly: true,
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
		"NO_WINDOWS_SERVICE":      "true", // Avoid using the Windows service manager
		"SPLUNK_ACCESS_TOKEN":     "access_token",
		"SPLUNK_HEC_TOKEN":        "hec_token",
		"SPLUNK_REALM":            "test_realm",
		"SPLUNK_LISTEN_INTERFACE": "127.0.0.1",
	}
	for key, value := range requiredEnvVars {
		t.Setenv(key, value)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipWindows && runtime.GOOS == "windows" {
				t.Skip("skipping test on windows")
			}

			if tt.skipMsg != "" {
				t.Skip(tt.skipMsg)
			}

			for key, value := range tt.extraEnvVars {
				t.Setenv(key, value)
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

			defer func() {
				otelcolCmdTestCtx = nil //nolint:fatcontext
			}()

			defer waitForPort(t, "55679")

			args := append([]string{}, tt.args...)
			if tt.validateOnly {
				args = append(args, "validate")
			}

			if tt.panicMsg != "" {
				assert.PanicsWithValue(t, tt.panicMsg, func() { runFromCmdLine(args) })
				return
			}

			waitForPort(t, "55679")
			runFromCmdLine(args)
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
