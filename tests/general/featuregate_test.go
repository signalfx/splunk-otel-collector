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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestFeatureGateCommand(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	c, shutdown := tc.SplunkOtelCollectorContainer(
		"", func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(
				map[string]string{
					"SPLUNK_REALM":        "noop",
					"SPLUNK_ACCESS_TOKEN": "noop",
				},
			)
		},
	)
	defer shutdown()

	t.Run("list all feature gates", func(t *testing.T) {
		sc, stdout, _ := c.Container.AssertExec(t, 15*time.Second, "/otelcol", "featuregate")
		require.Zero(t, sc)
		require.Contains(t, stdout, "ID")
		require.Contains(t, stdout, "Enabled")
		require.Contains(t, stdout, "Stage")
		require.Contains(t, stdout, "Description")
		require.Contains(t, stdout, "splunk.opamp.enabled")
	})

	t.Run("show feature gate", func(t *testing.T) {
		sc, stdout, _ := c.Container.AssertExec(t, 15*time.Second,
			"/otelcol", "featuregate", "splunk.opamp.enabled",
		)
		require.Zero(t, sc)
		require.Contains(t, stdout, "Feature: splunk.opamp.enabled")
		require.Contains(t, stdout, "Enabled:")
		require.Contains(t, stdout, "Stage: Beta")
		require.Contains(t, stdout, "Description:")
		require.Contains(t, stdout, "From Version:")
	})

	t.Run("help", func(t *testing.T) {
		sc, stdout, _ := c.Container.AssertExec(t, 15*time.Second, "/otelcol", "featuregate", "--help")
		require.Zero(t, sc)
		require.Contains(t, stdout, "Display information about available feature gates and their status")
		require.Contains(t, stdout, "otelcol featuregate [feature-id] [flags]")
	})

	t.Run("unknown feature gate", func(t *testing.T) {
		sc, _, stderr := c.Container.AssertExec(t, 15*time.Second,
			"/otelcol", "featuregate", "does.not.exist",
		)
		require.Equal(t, 1, sc)
		require.Contains(t, stderr, `feature "does.not.exist" not found`)
	})

	t.Run("too many arguments", func(t *testing.T) {
		sc, _, stderr := c.Container.AssertExec(t, 15*time.Second,
			"/otelcol", "featuregate", "splunk.opamp.enabled", "extra-argument",
		)
		require.Equal(t, 1, sc)
		require.Contains(t, stderr, "accepts at most 1 arg(s), received 2")
	})
}
