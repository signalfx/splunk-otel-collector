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

//go:build smartagent_integration

package tests

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCustomUpstatIntegration(t *testing.T) {
	path, err := filepath.Abs(path.Join(".", "testdata", "upstat"))
	require.NoError(t, err)

	testutils.CheckGoldenFileWithMount(t, "custom_upstat.yaml", "custom_upstat_expected.yaml", [][]string{
		{filepath.Join(path, "collectd-upstat.py"), "/var/collectd-python/upstat/collectd-upstat.py"},
		{filepath.Join(path, "upstat_types.db"), "/var/collectd-python/upstat/upstat_types.db"},
	},
		pmetrictest.IgnoreMetricAttributeValue("host"),
		pmetrictest.IgnoreTimestamp())
}
