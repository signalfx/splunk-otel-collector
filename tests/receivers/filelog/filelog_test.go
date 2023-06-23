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

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestFilelogReceiverSyslogFormat(t *testing.T) {
	testutils.AssertAllLogsReceived(t, "syslog.yaml", "syslog_config.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				syslog, err := filepath.Abs(filepath.Join(".", "testdata", "syslog"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(syslog, "/opt/syslog"))
					syslog = "/opt/syslog"
				}
				return c.WithEnv(map[string]string{"LOGFILE_PATH": syslog})
			},
		},
	)
}
