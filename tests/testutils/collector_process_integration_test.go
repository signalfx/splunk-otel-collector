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
//go:build testutils && testutilsintegration

package testutils

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils/telemetry"
)

func TestCollectorPath(t *testing.T) {
	p, err := findCollectorPath()
	require.NoError(t, err)
	require.NotEmpty(t, p)
	assert.True(t, strings.HasSuffix(p, "/bin/otelcol"))
}

func TestConfigPathNotRequiredUponBuildWithArgs(t *testing.T) {
	withArgs := NewCollectorProcess().WithArgs("arg_one", "arg_two")

	collector, err := withArgs.Build()
	require.NoError(t, err)
	require.NotNil(t, collector)
}

func TestCollectorProcessConfigSourced(t *testing.T) {
	tc := NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.SplunkOtelCollectorProcess("collector_process_config.yaml")
	defer shutdown()

	expectedMetrics, err := telemetry.LoadResourceMetrics(filepath.Join(".", "testdata", "expected_host_metrics.yaml"))
	require.NoError(t, err)
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedMetrics, 10*time.Second))
}
