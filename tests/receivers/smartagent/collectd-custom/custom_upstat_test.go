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
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCustomUpstatIntegration(t *testing.T) {
	t.Skip("Issues with test-containers networking, need to wait for -contrib to update the docker api version for us to update testcontainers-go locally")
	core, observed := observer.New(zap.DebugLevel)
	path, err := filepath.Abs(path.Join(".", "testdata", "upstat"))
	require.NoError(t, err)
	testutils.AssertAllMetricsReceived(t, "all.yaml", "custom_upstat.yaml",
		nil, []testutils.CollectorBuilder{func(collector testutils.Collector) testutils.Collector {
			collector = collector.WithLogger(zap.New(core))
			if cc, ok := collector.(*testutils.CollectorContainer); ok {
				collector = cc.WithMount(path, "/var/collectd-python/upstat")
				return collector.WithEnv(map[string]string{"PLUGIN_FOLDER": "/var/collectd-python/upstat"})
			}
			return collector.WithEnv(map[string]string{"PLUGIN_FOLDER": path})
		}})

	expectedContent := map[string]bool{
		`starting isolated configd instance "monitor-smartagentcollectdcustom"`: false,
		`"name": "smartagent/collectd/custom"`:                                  false,
		`"monitorType": "collectd/custom"`:                                      false,
		`"monitorID": "smartagentcollectdcustom"`:                               false,
	}
	for _, l := range observed.All() {
		for expected := range expectedContent {
			if strings.Contains(l.Message, expected) {
				expectedContent[expected] = true
			}
		}
	}
	for expected, found := range expectedContent {
		assert.True(t, found, expected)
	}
}
