// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discoverytest

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

const (
	receiverTypeAttr         = "discovery.receiver.type"
	messageAttr              = "discovery.message"
	otelEntityAttributesAttr = "otel.entity.attributes"
)

func Run(t *testing.T, receiverName string, configFilePath string, logMessageToAssert string) {
	factory := otlpreceiver.NewFactory()
	port := 16745
	cfg := factory.CreateDefaultConfig().(*otlpreceiver.Config)
	cfg.GRPC.NetAddr.Endpoint = fmt.Sprintf("localhost:%d", port)
	endpoint := cfg.GRPC.NetAddr.Endpoint
	sink := &consumertest.LogsSink{}
	receiver, err := factory.CreateLogsReceiver(context.Background(), receivertest.NewNopSettings(), cfg, sink)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, receiver.Shutdown(context.Background()))
	})

	dockerGID, err := getDockerGID()
	require.NoError(t, err)

	req := testcontainers.ContainerRequest{
		Image: "otelcol:latest",
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{"/var/run/docker.sock:/var/run/docker.sock"}
			hc.NetworkMode = network.NetworkHost
			hc.GroupAdd = []string{dockerGID}
		},
		Env: map[string]string{
			"SPLUNK_REALM":                "us2",
			"SPLUNK_ACCESS_TOKEN":         "12345",
			"SPLUNK_DISCOVERY_LOG_LEVEL":  "info",
			"OTLP_ENDPOINT":               endpoint,
			"SPLUNK_OTEL_COLLECTOR_IMAGE": "otelcol:latest",
		},
		Entrypoint: []string{"/otelcol", "--config", "/home/otel-local-config.yaml"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      configFilePath,
				ContainerFilePath: "/home/otel-local-config.yaml",
				FileMode:          0o777,
			},
		},
	}

	c, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	seenMessageAttr := 0
	seenReceiverTypeAttr := 0
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(sink.AllLogs()) == 0 {
			assert.Fail(tt, "No logs collected")
			return
		}
		for i := 0; i < len(sink.AllLogs()); i++ {
			plogs := sink.AllLogs()[i]
			lrs := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords()
			for j := 0; j < lrs.Len(); j++ {
				lr := lrs.At(j)
				attrMap, ok := lr.Attributes().Get(otelEntityAttributesAttr)
				if ok {
					m := attrMap.Map()
					discoveryMsg, ok := m.Get(messageAttr)
					if ok {
						seenMessageAttr++
						assert.Equal(tt, logMessageToAssert, discoveryMsg.AsString())
					}
					discoveryType, ok := m.Get(receiverTypeAttr)
					if ok {
						seenReceiverTypeAttr++
						assert.Equal(tt, receiverName, discoveryType.AsString())
					}
				}
			}
		}
		assert.Greater(tt, seenMessageAttr, 0, "Did not see message '%s'", logMessageToAssert)
		assert.Greater(tt, seenReceiverTypeAttr, 0, "Did not see expected type '%s'", receiverName)
	}, 60*time.Second, 1*time.Second, "Did not get '%s' discovery in time", receiverName)

	t.Cleanup(func() {
		require.NoError(t, c.Terminate(context.Background()))
	})

}

func getDockerGID() (string, error) {
	finfo, err := os.Stat("/var/run/docker.sock")
	if err != nil {
		return "", err
	}
	fsys := finfo.Sys()
	stat, ok := fsys.(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("OS error occurred while trying to get GID ")
	}
	dockerGID := fmt.Sprintf("%d", stat.Gid)
	return dockerGID, nil
}
