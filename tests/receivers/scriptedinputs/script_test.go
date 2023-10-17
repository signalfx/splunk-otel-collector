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

func TestScriptReceiverCpu(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_cpu.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "cpu.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverDf(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_df.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "df.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverHardware(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_hardware.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "hardware.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverInterfaces(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_interfaces.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "interfaces.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverIostat(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_iostat.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "iostat.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverLsof(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_lsof.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "lsof.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverNetstat(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_netstat.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "netstat.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverOpenPorts(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_openPorts.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "openPorts.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverPackage(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_package.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "package.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverProtocol(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_protocol.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "protocol.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverPs(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_ps.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "ps.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}

func TestScriptReceiverTop(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_top.yaml")
	defer shutdown()

	expectedLogs, err := telemetry.LoadResourceLogs(filepath.Join("testdata", "resource_logs", "top.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedLogs, 10*time.Second))
}
