// Copyright  Splunk, Inc.
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
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestFileProvider(t *testing.T) {
	testdataPath, err := filepath.Abs(path.Join(".", "testdata"))
	require.NoError(t, err)
	testutils.AssertAllMetricsReceived(
		t, "memory.yaml", "", nil,
		[]testutils.CollectorBuilder{
			func(collector testutils.Collector) testutils.Collector {
				if cc, ok := collector.(*testutils.CollectorContainer); ok {
					collector = cc.WithMount(testdataPath, "/testdata")
				}
				return collector.WithArgs("--config", "file:./testdata/file_config.yaml")
			},
		},
	)
}
