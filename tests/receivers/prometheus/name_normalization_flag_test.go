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

func TestNameNormalization(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	_, stop := tc.Containers(
		testutils.NewContainer().WithContext(
			path.Join(".", "testdata", "httpd"),
		).WithName("httpd").WithExposedPorts("8000:80").WillWaitForPorts("80"),
	)
	defer stop()

	for _, args := range []struct {
		name                    string
		resourceMetricsFilename string
		builder                 testutils.CollectorBuilder
	}{
		{"without flag", "non_normalized_httpd.yaml", nil},
		{"enabled flag", "normalized_httpd.yaml",
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithArgs("--feature-gates=+pkg.translator.prometheus.NormalizeName")
			},
		},
		{"disabled flag", "non_normalized_httpd.yaml",
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithArgs("--feature-gates=-pkg.translator.prometheus.NormalizeName")
			},
		},
	} {
		t.Run(args.name, func(tt *testing.T) {
			var builders []testutils.CollectorBuilder
			if args.builder != nil {
				builders = append(builders, args.builder)
			}
			testutils.AssertAllMetricsReceived(
				t, args.resourceMetricsFilename, "httpd_metrics_config.yaml", nil, builders,
			)
		})
	}
}
