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

func TestJMXReceiverProvidesAllJVMMetrics(t *testing.T) {
	//os.Setenv("SPLUNK_OTEL_COLLECTOR_IMAGE", "otelcol:latest")
	containers := []testutils.Container{
		testutils.NewContainer().WithContext(
			path.Join(".", "testdata", "server"),
		).WithExposedPorts("7199:7199").WithName("jmx").WillWaitForPorts("7199"),
	}

	//tmp_dir := "/etc/otel/collector/tmp"
	//local_tmp := filepath.Join(".", "tmp")
	//mount_dir, err := filepath.Abs(local_tmp)
	//require.NoError(t, err)
	//os.MkdirAll(mount_dir, 1777)
	//defer os.RemoveAll(local_tmp)

	testutils.AssertAllMetricsReceived(t, "all.yaml",
		"all_metrics_config.yaml", containers,
		[]testutils.CollectorBuilder{
			func(collector testutils.Collector) testutils.Collector {
				// JMX requires a local directory that can be written to, so we must mount a local dir
				// that the collector has write access to.

				return collector.WillFail(true)
			},
		})
}
