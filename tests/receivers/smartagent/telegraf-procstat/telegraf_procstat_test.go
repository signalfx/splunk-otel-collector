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

//go:build smartagent_integration

package tests

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestTelegrafProcstatReceiverProvidesAllMetrics(t *testing.T) {
	testutils.RunMetricsCollectionTest(t, "all_metrics_config.yaml", "expected_all.yaml",
		testutils.WithCompareMetricsOptions(
			pmetrictest.IgnoreMetricsOrder(),
			pmetrictest.IgnoreTimestamp(),
			pmetrictest.IgnoreMetricValues(),
		),
	)
}
