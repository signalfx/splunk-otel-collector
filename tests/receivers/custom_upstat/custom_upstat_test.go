// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration

package tests

import (
	"errors"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCustomUpstatIntegration(t *testing.T) {
	path, err := filepath.Abs(path.Join(".", "testdata", "custom"))
	require.NoError(t, err)
	testutils.AssertAllMetricsReceived(t, "all.yaml", "custom_upstat.yaml",
		nil, []testutils.CollectorBuilder{func(collector testutils.Collector) testutils.Collector {
			collector, err := collector.WithBoundDirectory(path, "/var/collectd-python/upstat")
			if errors.Is(testutils.ErrUnsupportedFeature, err) {
				// we are running in process
				collector = collector.WithEnv(map[string]string{"PLUGIN_FOLDER": path})
			} else {
				// we are running with a container
				collector = collector.WithEnv(map[string]string{"PLUGIN_FOLDER": "/var/collectd-python/upstat"})
			}
			return collector
		}})
}
