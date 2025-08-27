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

	ticker := time.NewTicker(100 * time.Millisecond)
	quit := make(chan struct{})
	t.Cleanup(func() {
		close(quit)
	})

	syslogTestMessage := "syslog information level log for testing"
	go func() {
		for {
			select {
			case <-ticker.C:
				t.Logf("Sending syslog message")
				writer.Emerg(syslogTestMessage)
				writer.Alert(syslogTestMessage)
				writer.Crit(syslogTestMessage)
				writer.Err(syslogTestMessage)
				writer.Info(syslogTestMessage)
				t.Logf("Sent log message to syslog in other goroutine")

				cmd := exec.Command("logger", syslogTestMessage)
				require.NoError(t, cmd.Run())
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		t.Logf("Checking for sent log messages")
		foundSyslog := false
		if tc.HECReceiverSink.LogRecordCount() > 0 {
			t.Logf("Found syslogs")
			for _, log := range tc.HECReceiverSink.AllLogs() {
				for i := range log.ResourceLogs().Len() {
					for j := range log.ResourceLogs().At(i).ScopeLogs().Len() {
						for k := range log.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().Len() {
							if strings.Contains(log.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().At(k).Body().Str(), syslogTestMessage) {
								foundSyslog = true
							}
						}
					}
				}
			}
		} else {
			t.Logf("No syslogs found")
		}
		require.Greater(c, tc.HECReceiverSink.LogRecordCount(), 0)
		require.True(c, foundSyslog)
	}, 20*time.Second, 500*time.Millisecond)
}
