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
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const networkName = "cassandra"

var cassandra = testutils.NewContainer().WithContext(
	path.Join(".", "testdata", "server"),
).WithEnv(map[string]string{
	"LOCAL_JMX": "no",
}).WithExposedPorts("7199:7199").
	WithStartupTimeout(3 * time.Minute).
	WithHostConfigModifier(func(cm *container.HostConfig) {
		cm.NetworkMode = "bridge"
	}).
	WithName("cassandra").
	WithNetworks("cassandra").
	WillWaitForPorts("7199").
	WillWaitForLogs("JMX is enabled to receive remote connections on port").
	WillWaitForLogs("Startup complete")

func TestJmxReceiverProvidesAllMetrics(t *testing.T) {
	t.Skip("Issues with test-containers networking, need to wait for -contrib to update the docker api version for us to update testcontainers-go locally")
	testutils.SkipIfNotContainerTest(t)

	tc := testutils.Testcase{TB: t}
	_, stopDependentContainers := tc.Containers(cassandra)
	defer stopDependentContainers()

	endpoint, err := GetDockerNetworkGateway(t, networkName)
	require.NoError(t, err)
	require.NotEmpty(t, endpoint)
	sinkBuilder := GetSinkAndLogs(t, endpoint)
	sink, err := sinkBuilder.Build()
	require.NoError(t, err)
	tc.OTLPEndpoint = sink.Endpoint
	tc.OTLPEndpointForCollector = sink.Endpoint
	require.Contains(t, tc.OTLPEndpointForCollector, ":")
	require.NoError(t, sink.Start())
	defer func() { require.NoError(t, sink.Shutdown()) }()

	_, shutdown := tc.SplunkOtelCollector("all_metrics_config.yaml",
		func(collector testutils.Collector) testutils.Collector {
			collector = collector.WithEnv(map[string]string{"OTLP_ENDPOINT": tc.OTLPEndpointForCollector})
			collector = collector.WithLogLevel("debug")
			collectorContainer := collector.(*testutils.CollectorContainer)
			collectorContainer.Container = collectorContainer.Container.WithHostConfigModifier(func(chc *container.HostConfig) {
				chc.NetworkMode = "bridge"
			}).WithName("otelcol-jmxtest").WithNetworks(networkName)
			p, err := filepath.Abs(filepath.Join(".", "testdata", "script.groovy"))
			require.NoError(t, err)
			return collector.WithMount(p, "/opt/script.groovy")
		})
	defer shutdown()

	resourceMetrics := tc.ResourceMetrics("all.yaml")
	require.NoError(t, sink.AssertAllMetricsReceived(t, *resourceMetrics, time.Minute))

}

func GetDockerNetworkGateway(t testing.TB, dockerNetwork string) (string, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	require.NoError(t, err)
	client.NegotiateAPIVersion(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	network, err := client.NetworkInspect(ctx, dockerNetwork, types.NetworkInspectOptions{})
	require.NoError(t, err)
	for _, ipam := range network.IPAM.Config {
		if ipam.Gateway != "" {
			return ipam.Gateway, nil
		}
	}
	return "", errors.New("Could not find gateway for network " + dockerNetwork)
}

func GetSinkAndLogs(t testing.TB, sinkHost string) testutils.OTLPReceiverSink {
	otlpPort := testutils.GetAvailablePort(t)
	endpoint := fmt.Sprintf("%s:%d", sinkHost, otlpPort)
	sink := testutils.NewOTLPReceiverSink().WithEndpoint(endpoint)
	return sink
}
