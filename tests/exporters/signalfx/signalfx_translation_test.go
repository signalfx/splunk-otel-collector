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
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestSignalFxExporterTranslatesOTelCPUMetrics(t *testing.T) {
	testutils.AssertAllMetricsReceived(
		t, "cpu_translations.yaml", "cpu_translations_config.yaml", nil,
	)
}

func TestSignalFxExporterTranslatesOTelMemoryMetrics(t *testing.T) {
	testutils.AssertAllMetricsReceived(
		t, "memory_translations.yaml", "memory_translations_config.yaml", nil,
	)
}
