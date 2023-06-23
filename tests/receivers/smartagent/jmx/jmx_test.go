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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

var cassandra = []testutils.Container{
	testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "server"),
	).WithExposedPorts("7199:7199").WithName("cassandra").WillWaitForPorts("7199"),
}

func TestJmxReceiverProvidesAllMetrics(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	testutils.AssertAllMetricsReceived(
		t, "all.yaml", "all_metrics_config.yaml", cassandra,
		[]testutils.CollectorBuilder{
			func(collector testutils.Collector) testutils.Collector {
				p, err := filepath.Abs(filepath.Join(".", "testdata", "script.groovy"))
				require.NoError(t, err)
				return collector.WithMount(p, "/opt/script.groovy")
			},
		},
	)
}
