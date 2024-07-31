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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunFromCmdLine(t *testing.T) {
	tests := []struct {
		name     string
		panicMsg string
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
		{
			name:     "dry-run_discovery",
			args:     []string{"otelcol", "--discovery", "--dry-run", "--config=config/collector/agent_config.yaml"},
			timeout:  30 * time.Second,
			panicMsg: "unexpected call to os.Exit(0) during test", // os.Exit(0) in the normal execution is expected for '--dry-run'.
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
			testCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			otelcolCmdTestCtx = testCtx

			if tt.panicMsg != "" {
				assert.PanicsWithValue(t, tt.panicMsg, func() { runFromCmdLine(tt.args) })
				return
			}

			runFromCmdLine(tt.args)
		})
	}
}
