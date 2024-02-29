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

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestSignalFxExporterTranslatesOTelCPUMetrics(t *testing.T) {
	t.Skip("Issues with test-containers networking, need to wait for -contrib to update the docker api version for us to update testcontainers-go locally")
	testutils.AssertAllMetricsReceived(
		t, "cpu_translations.yaml", "cpu_translations_config.yaml", nil,
		[]testutils.CollectorBuilder{
			// Added for implicit resourcedetection processor coverage
			// TODO: add suite
			func(collector testutils.Collector) testutils.Collector {
				machineId, err := filepath.Abs(filepath.Join(".", "testdata", "machine-id"))
				require.NoError(t, err)
				return collector.WithMount(machineId, "/etc/machine-id")
			},
		},
	)
}

func TestSignalFxExporterTranslatesOTelMemoryMetrics(t *testing.T) {
	testutils.AssertAllMetricsReceived(
		t, "memory_translations.yaml", "memory_translations_config.yaml", nil, nil,
	)
}
