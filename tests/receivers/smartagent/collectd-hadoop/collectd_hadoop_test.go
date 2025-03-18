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

//go:build smartagent_integration

package tests

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestCollectdHadoopReceiverProvidesAllMetrics(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	hadoop := testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "server"),
	).WithNetworks("hadoop")

	containers := []testutils.Container{
		hadoop.WithName("hadoop-worker1").WillWaitForLogs("ready"),
		hadoop.WithName("hadoop-master").WithCmd("run_master.sh").WithExposedPorts(
			"8088:8088",
		).WillWaitForPorts("8088").WillWaitForLogs("hadoop is running"),
	}
	_, stop := tc.Containers(containers...)
	defer stop()

	checkMetricsPresence(t, []string{
		"counter.hadoop.cluster.metrics.total_mb",
		"counter.hadoop.cluster.metrics.total_nodes",
		"counter.hadoop.cluster.metrics.total_virtual_cores",
		"gauge.hadoop.cluster.metrics.active_nodes",
		"gauge.hadoop.cluster.metrics.allocated_mb",
		"gauge.hadoop.cluster.metrics.allocated_virtual_cores",
		"gauge.hadoop.cluster.metrics.apps_completed",
		"gauge.hadoop.cluster.metrics.apps_failed",
		"gauge.hadoop.cluster.metrics.apps_killed",
		"gauge.hadoop.cluster.metrics.apps_pending",
		"gauge.hadoop.cluster.metrics.apps_running",
		"gauge.hadoop.cluster.metrics.apps_submitted",
		"gauge.hadoop.cluster.metrics.available_mb",
		"gauge.hadoop.cluster.metrics.available_virtual_cores",
		"gauge.hadoop.cluster.metrics.containers_allocated",
		"gauge.hadoop.cluster.metrics.containers_pending",
		"gauge.hadoop.cluster.metrics.containers_reserved",
		"gauge.hadoop.cluster.metrics.decommissioned_nodes",
		"gauge.hadoop.cluster.metrics.lost_nodes",
		"gauge.hadoop.cluster.metrics.rebooted_nodes",
		"gauge.hadoop.cluster.metrics.reserved_mb",
		"gauge.hadoop.cluster.metrics.reserved_virtual_cores",
		"gauge.hadoop.cluster.metrics.unhealthy_nodes",
		"gauge.hadoop.resource.manager.nodes.availMemoryMB",
		"gauge.hadoop.resource.manager.nodes.availableVirtualCores",
		"gauge.hadoop.resource.manager.nodes.numContainers",
		"gauge.hadoop.resource.manager.nodes.usedMemoryMB",
		"gauge.hadoop.resource.manager.nodes.usedVirtualCores",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.absoluteCapacity",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.absoluteMaxCapacity",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.absoluteUsedCapacity",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.allocatedContainers",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.capacity",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.maxApplications",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.maxApplicationsPerUser",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.maxCapacity",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.numActiveApplications",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.numApplications",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.numContainers",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.numPendingApplications",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.pendingContainers",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.reservedContainers",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.usedCapacity",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.userLimit",
		"gauge.hadoop.resource.manager.scheduler.leaf.queue.userLimitFactor",
		"gauge.hadoop.resource.manager.scheduler.root.queue.capacity",
		"gauge.hadoop.resource.manager.scheduler.root.queue.maxCapacity",
		"gauge.hadoop.resource.manager.scheduler.root.queue.usedCapacity",
	}, "all_metrics_config.yaml")
}

func checkMetricsPresence(t *testing.T, metricNames []string, configFile string) {
	f := otlpreceiver.NewFactory()
	port := testutils.GetAvailablePort(t)
	c := f.CreateDefaultConfig().(*otlpreceiver.Config)
	c.GRPC.NetAddr.Endpoint = fmt.Sprintf("localhost:%d", port)
	sink := &consumertest.MetricsSink{}
	receiver, err := f.CreateMetrics(context.Background(), receivertest.NewNopSettings(f.Type()), c, sink)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, receiver.Shutdown(context.Background()))
	})
	logger, _ := zap.NewDevelopment()

	dockerHost := "0.0.0.0"
	if runtime.GOOS == "darwin" {
		dockerHost = "host.docker.internal"
	}
	p, err := testutils.NewCollectorContainer().
		WithImage(testutils.GetCollectorImageOrSkipTest(t)).
		WithConfigPath(filepath.Join("testdata", configFile)).
		WithLogger(logger).
		WithEnv(map[string]string{"OTLP_ENDPOINT": fmt.Sprintf("%s:%d", dockerHost, port)}).
		Build()
	require.NoError(t, err)
	require.NoError(t, p.Start())
	t.Cleanup(func() {
		require.NoError(t, p.Shutdown())
	})

	missingMetrics := make(map[string]any, len(metricNames))
	for _, m := range metricNames {
		missingMetrics[m] = struct{}{}
	}

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		for i := 0; i < len(sink.AllMetrics()); i++ {
			m := sink.AllMetrics()[i]
			for j := 0; j < m.ResourceMetrics().Len(); j++ {
				rm := m.ResourceMetrics().At(j)
				for k := 0; k < rm.ScopeMetrics().Len(); k++ {
					sm := rm.ScopeMetrics().At(k)
					for l := 0; l < sm.Metrics().Len(); l++ {
						delete(missingMetrics, sm.Metrics().At(l).Name())
					}
				}
			}
		}
		msg := "Missing metrics:\n"
		for k := range missingMetrics {
			msg += fmt.Sprintf("- %q\n", k)
		}
		assert.Len(tt, missingMetrics, 0, msg)
	}, 1*time.Minute, 1*time.Second)
}
