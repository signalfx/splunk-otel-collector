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

func TestCollectorProcessWithMultipleTemplateConfigs(t *testing.T) {
	logCore, logs := observer.New(zap.DebugLevel)
	logger := zap.New(logCore)
	collector, err := testutils.NewCollectorProcess().
		WithArgs("--config", path.Join(".", "testdata", "templated.yaml")).
		WithLogger(logger).
		Build()

	require.NotNil(t, collector)
	require.NoError(t, err)

	err = collector.Start()
	require.NoError(t, err)

	_, ok := collector.(*testutils.CollectorProcess)
	require.True(t, ok)

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
	for _, tc := range []struct {
		expected map[string]any
		endpoint string
	}{
		{expected: expectedConfig, endpoint: "effective"},
	} {
		resp, err := http.Get(fmt.Sprintf("http://localhost:55554/debug/configz/%s", tc.endpoint))
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		actual := map[string]any{}
		require.NoError(t, yaml.Unmarshal(body, &actual))

		require.Equal(t, tc.expected, confmap.NewFromStringMap(actual).ToStringMap())
	}

	require.NoError(t, collector.Shutdown())
}
