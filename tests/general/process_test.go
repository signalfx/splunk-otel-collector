package tests

import (
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

	err = collector.Shutdown()
	require.NoError(t, err)
}
