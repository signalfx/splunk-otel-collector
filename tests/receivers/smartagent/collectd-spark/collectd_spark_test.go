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
	"context"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCollectdSparkReceiverProvidesAllMetrics(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	spark := testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "server"),
	).WithNetworks(
		"spark",
	)

	containers, stop := tc.Containers(
		spark.WithName("spark-master").WithExposedPorts(
			"4040:4040", "7077:7077", "8080:8080",
		).WillWaitForPorts("7077", "8080").WillWaitForLogs(
			"I have been elected leader!",
		).WithCmd("bin/spark-class", "org.apache.spark.deploy.master.Master"),
		spark.WithName("spark-worker").WithExposedPorts(
			"8081:8081",
		).WillWaitForPorts("8081").WillWaitForLogs(
			"Successfully registered with master",
		).WithCmd("bin/spark-class", "org.apache.spark.deploy.worker.Worker", "spark://spark-master:7077"),
	)
	defer stop()

	master := containers[0]
	rc, _, err := master.Exec(context.Background(), []string{"sh", "-c", "nc -lk 9999 &"})
	require.NoError(t, err)
	require.Zero(t, rc)

	rc, _, err = master.Exec(context.Background(), []string{
		"sh", "-c", "bin/spark-submit --master spark://spark-master:7077 --conf spark.driver.host=spark-master " +
			"examples/src/main/python/streaming/network_wordcount.py spark-master 9999 &",
	})
	require.NoError(t, err)
	require.Zero(t, rc)

	for _, args := range []struct {
		name                    string
		resourceMetricsFilename string
		collectorConfigFilename string
	}{
		{"master metrics", "all_master.yaml", "all_master_metrics_config.yaml"},
		{"worker metrics", "all_worker.yaml", "all_worker_metrics_config.yaml"},
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
