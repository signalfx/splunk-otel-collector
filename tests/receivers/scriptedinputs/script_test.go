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

package scriptedinputs

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/signalfx/splunk-otel-collector/tests/testutils/telemetry"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestScriptReceiver(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_cpu.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "cpu.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}
