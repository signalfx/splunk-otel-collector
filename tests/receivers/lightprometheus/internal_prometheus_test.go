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
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestInternalPrometheusMetrics(t *testing.T) {
	t.Skip("covered by normal light prometheus receiver, skip until we get the exact delta in the expected values updated")
	testutils.SkipIfNotContainerTest(t) // TODO: enhance internal metric settings detection for process config
	testutils.AssertAllMetricsReceived(
		t, "internal.yaml", "internal_metrics_config.yaml", nil, nil,
	)
}

func TestHttpdBasicAuth(t *testing.T) {
	testutils.CheckGoldenFile(t, "httpd_basic_auth.yaml", "httpd_basic_auth_expected.yaml",
		pmetrictest.IgnoreScopeVersion(),
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreStartTimestamp(),
		pmetrictest.IgnoreMetricValues(),
		pmetrictest.IgnoreResourceMetricsOrder(),
		pmetrictest.IgnoreScopeMetricsOrder(),
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreMetricDataPointsOrder(),
	)
}
