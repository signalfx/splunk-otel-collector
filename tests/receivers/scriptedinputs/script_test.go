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
}
