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
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestPrometheusExporterProvidesInternalMetrics(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	testutils.AssertAllMetricsReceived(
		t, "internal.yaml", "internal_metrics_config.yaml", nil, nil,
	)
}

func TestPrometheusExporterProvidesOTelInternalMetrics(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	testutils.AssertAllMetricsReceived(
		t, "otel_internal.yaml", "internal_metrics_config.yaml", nil, []testutils.CollectorBuilder{
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithArgs("--feature-gates", "+telemetry.useOtelForInternalMetrics")
			},
		},
	)
}

func TestPrometheusExporterScrapesTargets(t *testing.T) {
	httpd := []testutils.Container{testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "httpd"),
	).WithName("httpd").WithExposedPorts("8000:80").WillWaitForPorts("80")}
	testutils.AssertAllMetricsReceived(
		t, "httpd.yaml", "httpd_metrics_config.yaml", httpd, nil,
	)
}

func TestPrometheusExporterScrapesTargetsWithFilter(t *testing.T) {
	httpd := []testutils.Container{testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "httpd"),
	).WithName("httpd").WithExposedPorts("8000:80").WillWaitForPorts("80")}
	testutils.AssertAllMetricsReceived(
		t, "httpd_filtered.yaml", "httpd_metrics_config_with_filter.yaml", httpd, nil,
	)
}
