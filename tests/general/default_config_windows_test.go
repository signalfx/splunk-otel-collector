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
	"fmt"
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

	path, err := filepath.Abs("../../cmd/otelcol/config/collector/splunk_logs_config_windows.yaml")
	require.NoError(t, err)

	testMessage := fmt.Sprintf("otelcol-fips-default-log-config-%s", tc.ID)
	require.NoError(t, exec.Command(
		"eventcreate.exe",
		"/T", "INFORMATION",
		"/ID", "100",
		"/L", "APPLICATION",
		"/D", testMessage,
	).Run())

	_, shutdown := tc.SplunkOtelCollectorProcess(path,
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(map[string]string{
				"NO_WINDOWS_SERVICE":         "1",
				"SPLUNK_PLATFORM_TOKEN":      "test-token",
				"SPLUNK_PLATFORM_URL":        strings.Replace(tc.HECEndpointForCollector, "0.0.0.0", "127.0.0.1", 1),
				"SPLUNK_PLATFORM_LOGS_INDEX": "test-index",
				"SPLUNK_LISTEN_INTERFACE":    "127.0.0.1",
			})
		},
	)
	defer shutdown()

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		found := false
		for _, logs := range tc.HECReceiverSink.AllLogs() {
			for i := range logs.ResourceLogs().Len() {
				for j := range logs.ResourceLogs().At(i).ScopeLogs().Len() {
					for k := range logs.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().Len() {
						if strings.Contains(logs.ResourceLogs().At(i).ScopeLogs().At(j).LogRecords().At(k).Body().Str(), testMessage) {
							found = true
						}
					}
				}
			}
		}
		require.True(c, found, "Windows event log message was not received")
	}, time.Minute, 500*time.Millisecond)
}
