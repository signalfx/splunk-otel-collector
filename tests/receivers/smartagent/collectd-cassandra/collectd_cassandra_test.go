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

func TestCollectdCassandraReceiverProvidesAllMetrics(t *testing.T) {
	testutils.AssertAllMetricsReceived(
		t, "all.yaml", "all_metrics_config.yaml", nil, nil,
	)
}

func TestCollectdCassandraReceiverProvidesDefaultMetrics(t *testing.T) {
	testutils.AssertAllMetricsReceived(
		t, "default.yaml", "default_metrics_config.yaml", nil, nil,
	)
}
