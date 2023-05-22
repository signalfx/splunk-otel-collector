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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestIncludeTemplatedConfigs(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()

	csPort := testutils.GetAvailablePort(t)
	collector, shutdown := tc.SplunkOtelCollector("", func(collector testutils.Collector) testutils.Collector {
		if cc, ok := collector.(*testutils.CollectorContainer); ok {
			testdata, err := filepath.Abs(filepath.Join(".", "testdata"))
			require.NoError(t, err)
			collector = cc.WithMount(testdata, "/testdata")
		}

		return collector.WithArgs(
			// setting this directly to not rely on `/etc/config.yaml` container default
			"--config", path.Join(".", "testdata", "include_templated.yaml"),
		).WithEnv(
			map[string]string{
				"SPLUNK_DEBUG_CONFIG_SERVER_PORT": fmt.Sprintf("%d", csPort),
			},
		)
	})
	defer shutdown()

	require.Eventually(t, func() bool {
		for _, log := range tc.ObservedLogs.All() {
			if strings.Contains(log.Message,
				`Set config to [testdata/include_templated.yaml]`,
			) {
				return true
			}
		}
		return false
	}, 20*time.Second, time.Second)

	expectedConfig := map[string]any{
		"receivers": map[string]any{
			"hostmetrics": map[string]any{
				"collection_interval": "1s",
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
				"endpoint": tc.OTLPEndpoint,
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
	effective := collector.EffectiveConfig(t, csPort)
	if !testutils.CollectorImageIsSet() {
		// default collector process uses --set service.telemetry args
		delete(effective["service"].(map[string]any), "telemetry")
	}

	require.Equal(t, expectedConfig, effective)

	expectedResourceMetrics := tc.ResourceMetrics("hostmetrics.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
