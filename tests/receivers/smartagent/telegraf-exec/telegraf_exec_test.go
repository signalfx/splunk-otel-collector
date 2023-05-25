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
	"path"
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestTelegrafExecWithGoScript(t *testing.T) {
	collectorImage := testutils.GetCollectorImageOrSkipTest(t)
	platform := "amd64"
	if testutils.CollectorImageIsForArm(t) {
		platform = "arm64"
	}
	testutils.AssertAllMetricsReceived(
		t, "all.yaml", "", nil,
		[]testutils.CollectorBuilder{func(collector testutils.Collector) testutils.Collector {
			cc := collector.(*testutils.CollectorContainer)
			cc.Container = cc.Container.WithContext(
				path.Join(".", "testdata", "exec"),
			).WithBuildArgs(map[string]*string{
				"IMAGE_PLATFORM":              &platform,
				"SPLUNK_OTEL_COLLECTOR_IMAGE": &collectorImage,
			})
			return cc.WithArgs("--config", "/etc/config.yaml")
		}},
	)
}
