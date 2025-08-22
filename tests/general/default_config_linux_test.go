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
	"path/filepath"
	"strings"
	"testing"
	"time"

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
				"SPLUNK_HEC_TOKEN":                   "not.real",
				"SPLUNK_INGEST_URL":                  "not.real",
				"SPLUNK_LISTEN_INTERFACE":            "127.0.0.1",
				"SPLUNK_FILE_STORAGE_EXTENSION_PATH": t.TempDir(),
			}
			return collector.WithEnv(env)
		},
	)
	defer shutdown()

	writer, err := syslog.New(syslog.LOG_INFO, "otelcol")
	require.NoError(t, err)
	defer writer.Close()

	// The channel is required for synchronizing between the syslog writer and the
	// check for logs written to syslog. Without it, the logs may be written before
	// the check occurs, meaning the test is waiting for some other process to write to
	// syslog, resulting in flakiness.
	checked_logs := make(chan bool)
	go func() {
		<-checked_logs
		writer.Emerg(syslogTestMessage)
		t.Log("Sent log message to syslog in other goroutine")
	}()

	logMessageSent := false
	require.Eventually(t, func() bool {
		if !logMessageSent {
			t.Logf("No log message has been sent yet")
			logMessageSent = true
			defer func() {
				checked_logs <- true
				t.Logf("Sent bool to checked logs chan")
			}()
		}

		if len(tc.HECReceiverSink.AllLogs()) > 0 {
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
								return true
							}
						}
					}
				}
			}
			t.Logf("Didn't find log, but there was more than 0")
			return true
		}
		t.Logf("No logs found")
		return false
	}, 20*time.Second, 500*time.Millisecond)
}
