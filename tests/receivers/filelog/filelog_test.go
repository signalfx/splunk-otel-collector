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
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestFilelogReceiverSyslogFormat(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorContainer(
		"syslog_config.yaml",
		func(c testutils.Collector) testutils.Collector {
			cc := c.(*testutils.CollectorContainer)
			syslog, err := filepath.Abs(filepath.Join(".", "testdata", "syslog"))
			require.NoError(t, err)
			cc.Container = cc.Container.WithMount(testcontainers.BindMount(syslog, "/opt/syslog"))
			return cc
		},
	)

	defer shutdown()

	expectedResourceLogs := tc.ResourceLogs("syslog.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedResourceLogs, 30*time.Second))
}
