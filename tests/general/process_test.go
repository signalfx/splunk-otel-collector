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
	"io"
	"net/http"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCollectorProcessWithMultipleConfigs(t *testing.T) {
	logCore, logs := observer.New(zap.DebugLevel)
	logger := zap.New(logCore)

	csPort := testutils.GetAvailablePort(t)
	collector, err := testutils.NewCollectorProcess().
		WithArgs("--config", path.Join(".", "testdata", "receivers.yaml"),
			"--config", path.Join(".", "testdata", "processors.yaml"),
			"--config", path.Join(".", "testdata", "exporters.yaml"),
			"--config", path.Join(".", "testdata", "services.yaml")).
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

	_, ok := collector.(*testutils.CollectorProcess)
	require.True(t, ok)

	require.Eventually(t, func() bool {
		for _, log := range logs.All() {
			if strings.Contains(log.Message,
				`Set config to [testdata/receivers.yaml,testdata/processors.yaml,testdata/exporters.yaml,testdata/services.yaml]`,
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
	for _, tc := range []struct {
		expected map[string]any
		endpoint string
	}{
		{expected: map[string]any{"file": expectedConfig}, endpoint: "initial"},
		{expected: expectedConfig, endpoint: "effective"},
	} {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/debug/configz/%s", csPort, tc.endpoint))
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		actual := map[string]any{}
		require.NoError(t, yaml.Unmarshal(body, &actual))

		require.Equal(t, tc.expected, confmap.NewFromStringMap(actual).ToStringMap())
	}

}
