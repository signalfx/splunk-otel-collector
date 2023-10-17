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
	"path"
	"runtime"
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	if runtime.GOOS == "darwin" {
		t.Skip("unable to share sockets between mac and d4m vm: https://github.com/docker/for-mac/issues/483#issuecomment-758836836")
	}
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	kafka := testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "kafka"),
	).WithEnv(map[string]string{
		"KAFKA_ZOOKEEPER_CONNECT": "zookeeper:2181",
	}).WithNetworks("kafka")

    kafkaZookeeper := testutils.NewContainer().WithImage("zookeeper:3.5").WithName("zookeeper").WithNetworks("kafka").WithExposedPorts("2181:2181").WillWaitForPorts("2181")

    kafkaBroker := kafka.WithName("kafka-broker").WithEnvVar("START_AS", "broker").WithExposedPorts("7099:7099", "9092:9092").WillWaitForPorts("7099", "9092")

    kafkaConsumer := kafka.WithName("kafka-consumer").WithEnv(map[string]string{"START_AS": "consumer", "KAFKA_BROKER": "kafka-broker:9092", "JMX_PORT": "9099",}).WithExposedPorts("9099:9099").WillWaitForPorts("9099")

    kafkaProducer := kafka.WithName("kafka-producer").WithEnv(map[string]string{"START_AS": "producer", "KAFKA_BROKER": "kafka-broker:9092", "JMX_PORT": "8099",}).WithExposedPorts("8099:8099").WillWaitForPorts("8099")

    kafkaTopicCreator := kafka.WithName("kafka-topic-creator").WithEnvVar("START_AS", "create-topic",).WillWaitForLogs(`Created topic "sfx-employee".`)

    containers := []testutils.Container{kafkaZookeeper, kafkaBroker, kafkaConsumer, kafkaProducer, kafkaTopicCreator}

    collector := []testutils.CollectorBuilder {
                func(c testutils.Collector) testutils.Collector {
			        cc := c.(*testutils.CollectorContainer)
			        cc.Container = cc.Container.WithBinds("/var/run/docker.sock:/var/run/docker.sock:ro")
			        cc.Container = cc.Container.WillWaitForLogs("Discovering for next")
			        cc.Container = cc.Container.WithUser(fmt.Sprintf("999:%d", testutils.GetDockerGID(t)))
			        return cc
                },
                func (c testutils.Collector) testutils.Collector {
                    return c.WithEnv(map[string]string {
                        "SPLUNK_DISCOVERY_DURATION": "20S",
                        "SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
                    }).WithArgs(
				        "--discovery",
				        "--set", `splunk.discovery.extensions.k8s_observer.enabled=false`,
				        "--set", `splunk.discovery.extensions.host_observer.enabled=false`,
                    )
                },
            }

	for _, args := range []struct {
		name                    string
		resourceMetricsFilename string
		collectorConfigFilename string
	} {
		{"broker metrics", "all_broker.yaml", "otlp_exporter.yaml"},
		{"producer metrics", "all_producer.yaml", "otlp_exporter.yaml"},
		{"consumer metrics", "all_consumer.yaml", "otlp_exporter.yaml"},
	} {
		t.Run(args.name, func(tt *testing.T) {
			testutils.AssertAllMetricsReceived(tt, args.resourceMetricsFilename, args.collectorConfigFilename, containers, collector)
	    })
    }
}
