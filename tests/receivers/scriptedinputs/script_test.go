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

func TestScriptReceiverDf(t *testing.T) {
	//*bandwidth.sh - no result on MAC

	testutils.AssertValidLogsHeader(t, "cpu.yaml", "script_config_cpu.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "cpu.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/cpu.sh"))
				}
				return c
			},
		},
	)

	testutils.AssertValidLogsHeader(t, "df.yaml", "script_config_df.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "df.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/df.sh"))
				}
				return c
			},
		},
	)

	testutils.AssertValidLogsHeader(t, "hardware.yaml", "script_config_hardware.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "hardware.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/hardware.sh"))
				}
				return c
			},
		},
	)

	testutils.AssertValidLogsHeader(t, "iostat.yaml", "script_config_iostat.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "iostat.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/iostat.sh"))
				}
				return c
			},
		},
	)

	testutils.AssertValidLogsHeader(t, "lastlog.yaml", "script_config_lastlog.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "lastlog.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/lastlog.sh"))
				}
				return c
			},
		},
	)

	testutils.AssertValidLogsHeader(t, "lsof.yaml", "script_config_lsof.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "lsof.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/lsof.sh"))
				}
				return c
			},
		},
	)

	testutils.AssertValidLogsHeader(t, "netstat.yaml", "script_config_netstat.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "netstat.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/netstat.sh"))
				}
				return c
			},
		},
	)
	// nfsiostat.sh - no result on MAC

	testutils.AssertValidLogsHeader(t, "openPorts.yaml", "script_config_openPorts.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "openPorts.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/openPorts.sh"))
				}
				return c
			},
		},
	)

	// openPortsEnhanced.sh - no header

	testutils.AssertValidLogsHeader(t, "package.yaml", "script_config_package.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "package.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/package.sh"))
				}
				return c
			},
		},
	)
	// passwd.sh - no header

	testutils.AssertValidLogsHeader(t, "protocol.yaml", "script_config_protocol.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "protocol.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/protocol.sh"))
				}
				return c
			},
		},
	)

	testutils.AssertValidLogsHeader(t, "ps.yaml", "script_config_ps.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "ps.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/ps.sh"))
				}
				return c
			},
		},
	)
	// rlog.sh - no result on MAC

	// selinuxChecker.sh - no result on MAC

	// service.sh - hangs on MAC

	// sshdChecker.sh - no result on MAC

	// time.sh - no header

	testutils.AssertValidLogsHeader(t, "top.yaml", "script_config_top.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "top.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/top.sh"))
				}
				return c
			},
		},
	)

	// update.sh - hangs on MAC

	// uptime.sh - no result on MAC

	testutils.AssertValidLogsHeader(t, "usersWithLoginPrivs.yaml", "script_config_usersWithLoginPrivs.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "usersWithLoginPrivs.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/usersWithLoginPrivs.sh"))
				}
				return c
			},
		},
	)
	// header.sh - no result on MAC

	// vmstat.sh - no result on MAC

	testutils.AssertValidLogsHeader(t, "who.yaml", "script_config_who.yaml",
		nil, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				df, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "internal", "receiver", "scriptedinputsreceiver", "scripts", "who.sh"))
				require.NoError(t, err)
				if cc, ok := c.(*testutils.CollectorContainer); ok {
					cc.Container = cc.Container.WithMount(testcontainers.BindMount(df, "/etc/otel/collector/scripts/who.sh"))
				}
				return c
			},
		},
	)
}
