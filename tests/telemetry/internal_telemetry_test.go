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

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const (
	internalTelemetryConfigFile                   = "internal_telemetry_config.yaml"
	internalTelemetryExpectedFile                 = "expected_internal_telemetry.assert.yaml"
	internalTelemetryPrometheusReaderConfigFile   = "internal_telemetry_prometheus_reader_config.yaml"
	internalTelemetryPrometheusReaderExpectedFile = "expected_internal_telemetry_prometheus_reader.assert.yaml"
)

func TestInternalTelemetryPrometheusNames(t *testing.T) {
	testutils.RunMetricsCollectionTest(t, internalTelemetryConfigFile, internalTelemetryExpectedFile,
		testutils.WithMetricsAssertionFile(internalTelemetryExpectedFile),
	)
}

func TestInternalTelemetryPrometheusReaderNames(t *testing.T) {
	testutils.RunMetricsCollectionTest(t, internalTelemetryPrometheusReaderConfigFile, internalTelemetryPrometheusReaderExpectedFile,
		testutils.WithMetricsAssertionFile(internalTelemetryPrometheusReaderExpectedFile),
	)
}
