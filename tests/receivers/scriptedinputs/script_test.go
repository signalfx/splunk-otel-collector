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
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func checkLog(t *testing.T, configFile string, rg *regexp.Regexp) {
	f := otlpreceiver.NewFactory()
	port := testutils.GetAvailablePort(t)
	c := f.CreateDefaultConfig().(*otlpreceiver.Config)
	c.GRPC.NetAddr.Endpoint = fmt.Sprintf("localhost:%d", port)
	sink := &consumertest.LogsSink{}
	receiver, err := f.CreateLogsReceiver(context.Background(), receivertest.NewNopCreateSettings(), c, sink)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, receiver.Shutdown(context.Background()))
	})
	logger, _ := zap.NewDevelopment()

	dockerHost := "0.0.0.0"
	p, err := testutils.NewCollectorProcess().
		WithConfigPath(filepath.Join("testdata", configFile)).
		WithLogger(logger).
		WithEnv(map[string]string{"OTLP_ENDPOINT": fmt.Sprintf("%s:%d", dockerHost, port)}).
		Build()
	require.NoError(t, err)
	require.NoError(t, p.Start())
	t.Cleanup(func() {
		require.NoError(t, p.Shutdown())
	})

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(sink.AllLogs()) == 0 {
			assert.Fail(tt, "No logs collected")
			return
		}
		latestLogRecord := sink.AllLogs()[len(sink.AllLogs())-1].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Body().AsString()
		assert.True(tt, rg.MatchString(latestLogRecord), "no match, got this", latestLogRecord)
	}, 30*time.Second, 1*time.Second)
}

func TestScriptReceiverCpu(t *testing.T) {
	rg := regexp.MustCompile(`CPU\s+pctUser\s+pctNice\s+pctSystem\s+pctIowait\s+pctIdle\nall(\s*\d{1,3}.\d{1,3}){5}\n0(\s*\d{1,3}.\d{1,3}){5}`)
	checkLog(t, "script_config_cpu.yaml", rg)
}

func TestScriptReceiverDf(t *testing.T) {
	rg := regexp.MustCompile(`Filesystem\s+Type\s+Size\s+Used\s+Avail\s+Use%\s+Inodes\s+IUsed\s+IFree\s+IUse%\s+MountedOn`)
	checkLog(t, "script_config_df.yaml", rg)
}

func TestScriptReceiverHardware(t *testing.T) {
	rg := regexp.MustCompile(`KEY\s+VALUE`)
	checkLog(t, "script_config_hardware.yaml", rg)
}

func TestScriptReceiverInterfaces(t *testing.T) {
	rg := regexp.MustCompile(`Name\s+MAC\s+inetAddr\s+inet6Addr\s+Collisions\s+RXbytes\s+RXerrors\s+RXdropped\s+TXbytes\s+TXerrors\s+TXdropped\s+Speed\s+Duplex`)
	checkLog(t, "script_config_interfaces.yaml", rg)
}

func TestScriptReceiverIostat(t *testing.T) {
	rg := regexp.MustCompile(`Device\s+r/s\s+rkB/s\s+rrqm/s\s+%rrqm\s+r_await\s+rareq-sz\s+w/s\s+wkB/s\s+wrqm/s\s+%wrqm\s+w_await\s+wareq-sz\s+d/s\s+dkB/s\s+drqm/s\s+%drqm\s+d_await\s+dareq-sz\s+aqu-sz\s+%util`)
	checkLog(t, "script_config_iostat.yaml", rg)
}

func TestScriptReceiverLsof(t *testing.T) {
	rg := regexp.MustCompile(`COMMAND\s+PID\s+USER\s+FD\s+TYPE\s+DEVICE\s+SIZE\s+NODE\s+NAME`)
	checkLog(t, "script_config_lsof.yaml", rg)
}

func TestScriptReceiverNetstat(t *testing.T) {
	rg := regexp.MustCompile(`Proto\s+Recv-Q\s+Send-Q\s+LocalAddress\s+ForeignAddress\s+State`)
	checkLog(t, "script_config_netstat.yaml", rg)
}

func TestScriptReceiverOpenPorts(t *testing.T) {
	rg := regexp.MustCompile(`Proto\s+Port`)
	checkLog(t, "script_config_openPorts.yaml", rg)
}

func TestScriptReceiverPackage(t *testing.T) {
	rg := regexp.MustCompile(`NAME\s+VERSION\s+RELEASE\s+ARCH\s+VENDOR\s+GROUP`)
	checkLog(t, "script_config_package.yaml", rg)
}

func TestScriptReceiverProtocol(t *testing.T) {
	rg := regexp.MustCompile(`IPdropped\s+TCPrexmits\s+TCPreorder\s+TCPpktRecv\s+TCPpktSent\s+UDPpktLost\s+UDPunkPort\s+UDPpktRecv\s+UDPpktSent`)
	checkLog(t, "script_config_protocol.yaml", rg)
}

func TestScriptReceiverPs(t *testing.T) {
	rg := regexp.MustCompile(`USER\s+PID\s+%CPU\s+%MEM\s+VSZ\s+RSS\s+TTY\s+STAT\s+START\s+TIME\s+COMMAND\s+ARGS`)
	checkLog(t, "script_config_ps.yaml", rg)
}

func TestScriptReceiverTop(t *testing.T) {
	rg := regexp.MustCompile(`PID\s+USER\s+PR\s+NI\s+VIRT\s+RES\s+SHR\s+S\s+pctCPU\s+pctMEM\s+cpuTIME\s+COMMAND`)
	checkLog(t, "script_config_top.yaml", rg)
}

func TestScriptReceiverWho(t *testing.T) {
	rg := regexp.MustCompile(`USERNAME\s+LINE\s+HOSTNAME\s+TIME`)
	checkLog(t, "script_config_who.yaml", rg)
}
