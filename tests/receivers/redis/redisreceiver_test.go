// Copyright  Splunk, Inc.
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

func TestRedisReceiverProvidesAllMetrics(t *testing.T) {
	server := testutils.NewContainer().WithContext(path.Join(".", "testdata", "server")).WithExposedPorts("6379:6379").WithName("redis-server").WillWaitForPorts("6379").WillWaitForLogs("Ready to accept connections")
	containers := []testutils.Container{server}
	testutils.AssertAllMetricsReceived(t, "all.yaml", "all_metrics_config.yaml", containers, nil)
}

func TestRedisReceiverProvidesAllMetricsWithServer(t *testing.T) {
	server := testutils.NewContainer().WithContext(path.Join(".", "testdata", "server")).WithExposedPorts("6379:6379").WithNetworks("redis_network").WithName("redis-server").WillWaitForLogs("Ready to accept connections")
	client := testutils.NewContainer().WithContext(path.Join(".", "testdata", "client")).WithName("redis-client").WithNetworks("redis_network").WillWaitForLogs("redis client started")
	containers := []testutils.Container{server, client}
	testutils.AssertAllMetricsReceived(t, "all_server.yaml", "all_metrics_config.yaml", containers, nil)
}
