package tests

import (
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCollectorProcessWithMultipleConfigs(t *testing.T) {
	logCore, logs := observer.New(zap.DebugLevel)
	logger := zap.New(logCore)
	collector, err := testutils.NewCollectorProcess().
		WithArgs("--config", path.Join(".", "testdata", "receivers.yaml"),
			"--config", path.Join(".", "testdata", "processors.yaml"),
			"--config", path.Join(".", "testdata", "exporters.yaml"),
			"--config", path.Join(".", "testdata", "services.yaml")).
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

	require.Eventually(t, func() bool {
		// Need expected configs for comparison to config server contents. Sections in the config server can be in
		// different orders, so ensure the proper sections are in place without enforcing order requirements.
		expectedBodySections := []string{
			"receivers:\n  hostmetrics:\n    collection_interval: 10s\n    scrapers:\n      cpu: null\n      disk: null\n      filesystem: null\n      memory: null\n      network: null\n",
			"processors:\n  resourcedetection:\n    detectors:\n    - system\n",
			"exporters:\n  otlp:\n    endpoint: localhost:23456\n    tls:\n      insecure: true\n",
			"service:\n  pipelines:\n    metrics:\n      exporters:\n      - otlp\n      processors:\n      - resourcedetection\n      receivers:\n      - hostmetrics",
		}
		// There aren't any keys or tokens in test configs, so initial and effective have the same contents.
		// We don't need to test the token removal functionality for the effective server here since config server
		// tests already cover it.
		configServerURLs := []string{
			"http://localhost:55554/debug/configz/initial",
			"http://localhost:55554/debug/configz/effective",
		}

		for _, url := range configServerURLs {
			resp, err := http.Get(url)
			if err != nil {
				return false
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			sb := string(body)

			for _, bodySection := range expectedBodySections {
				if !strings.Contains(sb, bodySection) {
					return false
				}
			}
		}

		return true
	}, 20*time.Second, time.Second)

	err = collector.Shutdown()
	require.NoError(t, err)
}
