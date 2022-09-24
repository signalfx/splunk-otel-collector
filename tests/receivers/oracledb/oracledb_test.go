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

package tests

import (
	"github.com/signalfx/splunk-otel-collector/tests/testutils"
	"path"

	"testing"
	"time"
)

// The Oracle DB container takes close to 10 minutes on a local machine to do the default setup, so the best way to
// account for startup time is to wait for the container to be healthy before continuing test.
var oracledb = []testutils.Container{testutils.NewContainer().WithContext(
	path.Join(".", "testdata", "server"),
).WithName("oracledb").WillWaitForHealth(15 * time.Minute)}

// This test ensures the collector can connect to an Oracle DB, and properly get metrics. It's not intended to
// test the receiver itself.
func TestOracleDBIntegration(t *testing.T) {
	testutils.AssertAllMetricsReceived(
		t, "all.yaml", "all_metrics_config.yaml", oracledb,
	)
}
