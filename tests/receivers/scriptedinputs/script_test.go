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
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestScriptReceiverCpu(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_cpu.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("CPU\\s+pctUser\\s+pctNice\\s+pctSystem\\s+pctIowait\\s+pctIdle\\nall(\\s*\\d{1,3}.\\d{1,3}){5}\\n0(\\s*\\d{1,3}.\\d{1,3}){5}"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverDf(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_df.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		fmt.Printf("Received log entry - \n%s", lr.Body().Str())
		assert.Regexp(tt, regexp.MustCompile("Filesystem\\s+Type\\s+Size\\s+Used\\s+Avail\\s+Use%\\s+Inodes\\s+IUsed\\s+IFree\\s+IUse%\\s+MountedOn"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverHardware(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_hardware.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("KEY\\s+VALUE"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverInterfaces(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_interfaces.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("Name\\s+MAC\\s+inetAddr\\s+inet6Addr\\s+Collisions\\s+RXbytes\\s+RXerrors\\s+RXdropped\\s+TXbytes\\s+TXerrors\\s+TXdropped\\s+Speed\\s+Duplex"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverIostat(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_iostat.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("Device\\s+r/s\\s+rkB/s\\s+rrqm/s\\s+%rrqm\\s+r_await\\s+rareq-sz\\s+w/s\\s+wkB/s\\s+wrqm/s\\s+%wrqm\\s+w_await\\s+wareq-sz\\s+d/s\\s+dkB/s\\s+drqm/s\\s+%drqm\\s+d_await\\s+dareq-sz\\s+(f/s\\s+)?(f_await\\s+)?aqu-sz\\s+%util"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverLsof(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_lsof.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("COMMAND\\s+PID\\s+USER\\s+FD\\s+TYPE\\s+DEVICE\\s+SIZE\\s+NODE\\s+NAME"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverNetstat(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_netstat.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("Proto\\s+Recv-Q\\s+Send-Q\\s+LocalAddress\\s+ForeignAddress\\s+State"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverOpenPorts(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_openPorts.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("Proto\\s+Port"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverPackage(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_package.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("NAME\\s+VERSION\\s+RELEASE\\s+ARCH\\s+VENDOR\\s+GROUP"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverProtocol(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_protocol.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("IPdropped\\s+TCPrexmits\\s+TCPreorder\\s+TCPpktRecv\\s+TCPpktSent\\s+UDPpktLost\\s+UDPunkPort\\s+UDPpktRecv\\s+UDPpktSent"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverPs(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_ps.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("USER\\s+PID\\s+%CPU\\s+%MEM\\s+VSZ\\s+RSS\\s+TTY\\s+STAT\\s+START\\s+TIME\\s+COMMAND\\s+ARGS"), lr.Body().Str())
	}, 30*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}

func TestScriptReceiverTop(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("script_config_top.yaml")
	defer shutdown()

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if tc.OTLPReceiverSink.LogRecordCount() == 0 {
			assert.Fail(tt, "no logs received")
			return
		}
		receivedOTLPLogs := tc.OTLPReceiverSink.AllLogs()

		lr := receivedOTLPLogs[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
		assert.Regexp(tt, regexp.MustCompile("PID\\s+USER\\s+PR\\s+NI\\s+VIRT\\s+RES\\s+SHR\\s+S\\s+pctCPU\\s+pctMEM\\s+cpuTIME\\s+COMMAND"), lr.Body().Str())
	}, 10*time.Second, 10*time.Millisecond, "Failed to receive expected logs")
}
