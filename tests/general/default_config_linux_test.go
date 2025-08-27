// Copyright Splunk, Inc.
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
	"log/syslog"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDefaultLogConfig(t *testing.T) {
	tc := testutils.NewHECTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownHECReceiverSink()

	syslogTestMessage := "syslog information level log for testing"
	path, err := filepath.Abs("../../cmd/otelcol/config/collector/logs_config_linux.yaml")
	require.NoError(t, err)

	_, shutdown := tc.SplunkOtelCollectorProcess(path,
		func(collector testutils.Collector) testutils.Collector {
			env := map[string]string{
				"SPLUNK_ACCESS_TOKEN":                "not.real",
				"SPLUNK_FILE_STORAGE_EXTENSION_PATH": t.TempDir(),
				"SPLUNK_HEC_TOKEN":                   "not.real",
				"SPLUNK_INGEST_URL":                  "not.real",
				"SPLUNK_LISTEN_INTERFACE":            "127.0.0.1",
			}
			return collector.WithEnv(env)
		},
	)
	defer shutdown()

	writer, err := syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, "otelcol")
	require.NoError(t, err)
	defer writer.Close()

	// The channel is required for synchronizing between the syslog writer and the
	// check for logs written to syslog. Without it, the logs may be written before
	// the check occurs, meaning the test is waiting for some other process to write to
	// syslog, resulting in flakiness.
	ticker := time.NewTicker(100 * time.Millisecond)
	quit := make(chan struct{})
	t.Cleanup(func() {
		close(quit)
	})
	go func() {
		for {
			select {
			case <-ticker.C:
				writer.Emerg(syslogTestMessage)
				writer.Alert(syslogTestMessage)
				writer.Crit(syslogTestMessage)
				writer.Err(syslogTestMessage)
				writer.Info(syslogTestMessage)
				t.Log("Sent log message to syslog in other goroutine")

				cmd := exec.Command("logger", syslogTestMessage)
				require.NoError(t, cmd.Run())
				t.Log("Sent log message to syslog via logger command in other goroutine")

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		t.Logf("Checking for sent log messages")
		if tc.HECReceiverSink.LogRecordCount() > 0 {
			for _, log := range tc.HECReceiverSink.AllLogs() {
				for i := range log.ResourceLogs().Len() {
					for j := range log.ResourceLogs().At(i).ScopeLogs().Len() {
						for k := range log.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().Len() {
							t.Log("Received another syslog:")
							t.Logf("Timestamp: %s", log.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().At(k).Timestamp().String())
							t.Logf("Body: %s", log.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().At(k).Body().Str())
							t.Logf("Attributes: %v", log.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().At(k).Attributes().AsRaw())
							t.Logf("Event name: %s", log.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().At(k).EventName())
							t.Logf("Severity text: %s", log.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().At(k).SeverityText())
							if strings.Contains(log.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().At(k).Body().Str(), syslogTestMessage) {
								t.Logf("Found the syslog sent")
								require.True(c, true)
							}
						}
					}
				}
			}
			t.Logf("Didn't find log, but there was more than 0")
		}
		require.Greater(c, tc.HECReceiverSink.LogRecordCount(), 0)
	}, 20*time.Second, 500*time.Millisecond)
}
