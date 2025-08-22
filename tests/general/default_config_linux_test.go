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

	writer, err := syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, "otelcol")
	require.NoError(t, err)
	defer writer.Close()

	// The channel is required for synchronizing between the syslog writer and the
	// check for logs written to syslog. Without it, the logs may be written before
	// the check occurs, meaning the test is waiting for some other process to write to
	// syslog, resulting in flakiness.
	checked_logs := make(chan bool)
	go func() {
		<-checked_logs
		writer.Info(syslogTestMessage)
	}()

	logMessageSent := false
	require.Eventually(t, func() bool {
		if !logMessageSent {
			logMessageSent = true
			defer func() {
				checked_logs <- true
			}()
		}

		if len(tc.HECReceiverSink.AllLogs()) > 0 {
			return true
		}
		return false
	}, 20*time.Second, 500*time.Millisecond)
}
