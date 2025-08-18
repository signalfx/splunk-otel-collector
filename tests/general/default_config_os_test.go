package tests

import (
	"log"
	"log/syslog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDefaultLogConfig(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()
	defer tc.ShutdownHECReceiverSink()

	t.Setenv("SPLUNK_ACCESS_TOKEN", "not.real")
	t.Setenv("SPLUNK_HEC_TOKEN", "not.real")
	t.Setenv("SPLUNK_INGEST_URL", "not.real")
	t.Setenv("SPLUNK_REALM", "not.real")
	t.Setenv("SPLUNK_LISTEN_INTERFACE", "127.0.0.1")

	_, shutdown := tc.SplunkOtelCollectorProcess("logs_config_linux.yaml")
	defer shutdown()

	// Establish a connection to the syslog daemon.
	// The priority here acts as a default for the Writer if not specified in method calls.
	writer, err := syslog.New(syslog.LOG_LOCAL0|syslog.LOG_INFO, "my_service")
	if err != nil {
		log.Fatalf("Unable to connect to syslog: %v", err)
	}
	defer writer.Close()

	writer.Info("This is an informational message.")
	writer.Warning("A warning occurred.")
	writer.Err("An error happened!")
	writer.Debug("This is a debug message (may not be visible depending on syslog configuration).")

	require.Eventually(t, func() bool {
		if len(tc.HECReceiverSink.AllLogs()) > 0 {
			return true
		}
		return false
	}, 20*time.Second, time.Second)
}
