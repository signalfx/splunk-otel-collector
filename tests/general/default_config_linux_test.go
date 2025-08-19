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
	"log"
	"log/syslog"
	"path/filepath"
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
	t.Setenv("SPLUNK_FILE_STORAGE_EXTENSION_PATH", t.TempDir())

	path, err := filepath.Abs("../../cmd/otelcol/config/collector/logs_config_linux.yaml")
	require.NoError(t, err)

	_, shutdown := tc.SplunkOtelCollectorProcess(path)
	defer shutdown()

	// Establish a connection to the syslog daemon.
	// The priority here acts as a default for the Writer if not specified in method calls.
	writer, err := syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, "otelcol")
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
