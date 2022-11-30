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
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCollectorProcessWithMultipleTemplateConfigs(t *testing.T) {
	logCore, logs := observer.New(zap.DebugLevel)
	logger := zap.New(logCore)

	csPort := testutils.GetAvailablePort(t)
	collector, err := testutils.NewCollectorProcess().
		WithArgs("--config", path.Join(".", "testdata", "templated.yaml")).
		WithLogger(logger).
		WithEnv(map[string]string{
			"SPLUNK_DEBUG_CONFIG_SERVER_PORT": fmt.Sprintf("%d", csPort),
		}).
		Build()

	require.NotNil(t, collector)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, collector.Shutdown())
	}()

	err = collector.Start()
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		for _, log := range logs.All() {
			if strings.Contains(log.Message,
				`Set config to [testdata/templated.yaml]`,
			) {
				return true
			}
		}
		return false
	}, 20*time.Second, time.Second)

	require.Eventually(t, func() bool {
		for _, log := range logs.All() {
			// Confirm collector starts and runs successfully
			if strings.Contains(log.Message, "Everything is ready. Begin running and processing data.") {
				return true
			}
		}
		return false
	}, 20*time.Second, time.Second)

	expectedConfig := map[string]any{
		"receivers": map[string]any{
			"hostmetrics": map[string]any{
				"collection_interval": "10s",
				"scrapers": map[string]any{
					"cpu":        nil,
					"disk":       nil,
					"filesystem": nil,
					"memory":     nil,
					"network":    nil,
				},
			},
		},
		"processors": map[string]any{
			"resourcedetection": map[string]any{
				"detectors": []any{"system"},
			},
		},
		"exporters": map[string]any{
			"otlp": map[string]any{
				"endpoint": "localhost:23456",
				"tls": map[string]any{
					"insecure": true,
				},
			},
		},
		"service": map[string]any{
			"pipelines": map[string]any{
				"metrics": map[string]any{
					"processors": []any{"resourcedetection"},
					"receivers":  []any{"hostmetrics"},
					"exporters":  []any{"otlp"},
				},
			},
		},
	}

	require.Equal(t, expectedConfig, collector.EffectiveConfig(t, csPort))
}
