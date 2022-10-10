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

package tests

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCollectdKafkaReceiversProvideAllMetrics(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	kafka := testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "kafka"),
	).WithEnv(map[string]string{
		"KAFKA_ZOOKEEPER_CONNECT": "zookeeper:2181",
	}).WithNetworks("kafka")

	_, stop := tc.Containers(
		testutils.NewContainer().WithImage(
			"zookeeper:3.5",
		).WithName("zookeeper").WithNetworks(
			"kafka",
		).WithExposedPorts("2181:2181").WillWaitForPorts("2181"),
		kafka.WithName("kafka-broker").WithEnvVar(
			"START_AS", "broker",
		).WithExposedPorts("7099:7099", "9092:9092").WillWaitForPorts("7099", "9092"),

		kafka.WithName("kafka-topic-creator").WithEnvVar(
			"START_AS", "create-topic",
		).WillWaitForLogs(`Created topic "sfx-employee".`),

		kafka.WithName("kafka-producer").WithEnv(map[string]string{
			"START_AS": "producer", "KAFKA_BROKER": "kafka-broker:9092", "JMX_PORT": "8099",
		}).WithExposedPorts("8099:8099").WillWaitForPorts("8099"),

		kafka.WithName("kafka-consumer").WithEnv(map[string]string{
			"START_AS": "consumer", "KAFKA_BROKER": "kafka-broker:9092", "JMX_PORT": "9099",
		}).WithExposedPorts("9099:9099").WillWaitForPorts("9099"),
	)
	defer stop()

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
			ttc := testutils.NewTestcase(tt)
			expectedResourceMetrics := ttc.ResourceMetrics(args.resourceMetricsFilename)

			_, shutdown := ttc.SplunkOtelCollector(args.collectorConfigFilename)
			defer shutdown()

			require.NoError(tt, ttc.OTLPReceiverSink.AssertAllMetricsReceived(tt, *expectedResourceMetrics, 30*time.Second))
		})
	}
}
