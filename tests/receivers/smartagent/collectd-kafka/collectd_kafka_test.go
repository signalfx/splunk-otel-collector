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

func TestCollectdKafkaReceiversProvideAllMetrics(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	for _, args := range []struct {
		name                    string
		resourceMetricsFilename string
		collectorConfigFilename string
	}{
		{"broker metrics", "all_broker.yaml", "all_broker_metrics_config.yaml"},
		{"producer metrics", "all_producer.yaml", "all_producer_metrics_config.yaml"},
		{"consumer metrics", "all_consumer.yaml", "all_consumer_metrics_config.yaml"},
	} {
		t.Run(args.name, func(tt *testing.T) {
			testutils.AssertAllMetricsReceived(tt, args.resourceMetricsFilename, args.collectorConfigFilename, nil, nil)
		})
	}
}
