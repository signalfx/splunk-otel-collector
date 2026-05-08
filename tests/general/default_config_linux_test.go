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

	writer, err := syslog.New(syslog.LOG_DAEMON, "otelcol")
	require.NoError(t, err)
	defer writer.Close()
	quit := make(chan struct{})
	t.Cleanup(func() {
		close(quit)
	})

	syslogTestMessage := "syslog information level log for testing"
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				// From testing, both ways of sending syslog messages is required to
				// trigger writing to syslog. If only one is included, for some reason
				// /var/log/syslog doesn't get written to until some external syslog
				// message is sent, in which case the message sent here would appear.
				writer.Emerg(syslogTestMessage)
				cmd := exec.Command("logger", syslogTestMessage)
				require.NoError(t, cmd.Run())
			}
		}
	}()

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		foundSyslog := false
		if tc.HECReceiverSink.LogRecordCount() > 0 {
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
		}
		require.Positive(c, tc.HECReceiverSink.LogRecordCount())
		require.True(c, foundSyslog)
	}, 20*time.Second, 500*time.Millisecond)
}
