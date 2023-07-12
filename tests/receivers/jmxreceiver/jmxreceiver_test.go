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
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestJMXReceiverProvidesAllJVMMetrics(t *testing.T) {

	containers := []testutils.Container{
		testutils.NewContainer().WithContext(
			path.Join(".", "testdata", "server"),
		).WithExposedPorts("7199:7199").WithName("jmx").WillWaitForPorts("7199"),
	}

	jmx_gatherer_path := downloadJMXGatherer(t)

	testutils.AssertAllMetricsReceived(t, "all.yaml",
		"all_metrics_config.yaml", containers,
		[]testutils.CollectorBuilder{
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithEnv(map[string]string{
					"JMX_PATH": jmx_gatherer_path,
				})
			},
		})

	deleteJMXGatherer(t, jmx_gatherer_path)
}

func downloadJMXGatherer(t *testing.T) string {
	jmx_version, err := os.ReadFile("../../../internal/buildscripts/packaging/jmx-metric-gatherer-release.txt")
	require.NoError(t, err)

	jmx_file_name := "opentelemetry-jmx-metrics.jar"
	remote_url := fmt.Sprintf("https://github.com/open-telemetry/opentelemetry-java-contrib/releases/download/v$version/$jmx_filename",
		jmx_version, jmx_file_name)

	resp, err := http.Get(remote_url)
	require.NoError(t, err)
	defer resp.Body.Close()

	jmx_gatherer_path := path.Join(".", "testdata", jmx_file_name)
	out, err := os.Create(jmx_gatherer_path)
	require.NoError(t, err)
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	require.NoError(t, err)

	return jmx_gatherer_path
}

func deleteJMXGatherer(t *testing.T, jmx_gatherer_path string) {
	err := os.Remove(jmx_gatherer_path)
	require.NoError(t, err)
}
