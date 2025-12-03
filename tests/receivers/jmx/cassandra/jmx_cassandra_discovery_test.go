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

//go:build discovery_integration_jmx

package tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

type otelContainer struct {
	testcontainers.Container
}

const (
	ReceiverTypeAttr         = "discovery.receiver.type"
	MessageAttr              = "discovery.message"
	OtelEntityAttributesAttr = "otel.entity.attributes"
)

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

func jmxCassandraAutoDiscoveryHelper(t *testing.T, ctx context.Context, configFile, logMessageToAssert string) (*otelContainer, error) {
	factory := otlpreceiver.NewFactory()
	port := 16745
	c := factory.CreateDefaultConfig().(*otlpreceiver.Config)
	endpoint := fmt.Sprintf("localhost:%d", port)
	c.GRPC = configoptional.Some(configgrpc.ServerConfig{
		NetAddr: confignet.AddrConfig{
			Endpoint:  endpoint,
			Transport: "tcp",
		},
	})
	sink := &consumertest.LogsSink{}
	receiver, err := factory.CreateLogs(context.Background(), receivertest.NewNopSettings(factory.Type()), c, sink)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, receiver.Shutdown(context.Background()))
	})

	dockerGID, err := getDockerGID()
	require.NoError(t, err)

	coverDest := os.Getenv("CONTAINER_COVER_DEST")
	coverSrc := os.Getenv("CONTAINER_COVER_SRC")
	var coverDirBind string
	if coverSrc != "" && coverDest != "" {
		coverDirBind = fmt.Sprintf("%s:%s", coverSrc, coverDest)
	}

	otelConfigPath, err := filepath.Abs(filepath.Join(".", "testdata", configFile))
	if err != nil {
		return nil, err
	}
	r, err := os.Open(otelConfigPath)
	if err != nil {
		return nil, err
	}

	currPath, err := filepath.Abs(filepath.Join(".", "testdata"))
	if err != nil {
		return nil, err
	}
	req := testcontainers.ContainerRequest{
		Image: "otelcol:latest",
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{"/var/run/docker.sock:/var/run/docker.sock", coverDirBind}
			hc.NetworkMode = network.NetworkHost
			hc.GroupAdd = []string{dockerGID}
		},
		Env: map[string]string{
			"GOCOVERDIR":                  coverDest,
			"SPLUNK_REALM":                "us2",
			"SPLUNK_ACCESS_TOKEN":         "12345",
			"SPLUNK_DISCOVERY_LOG_LEVEL":  "info",
			"OTLP_ENDPOINT":               endpoint,
			"SPLUNK_OTEL_COLLECTOR_IMAGE": "otelcol:latest",
			"USERNAME":                    "hello",
			"PASSWORD":                    "world",
		},
		Entrypoint: []string{"/otelcol", "--config", "/home/otel-local-config.yaml"},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            r,
				HostFilePath:      otelConfigPath,
				ContainerFilePath: "/home/otel-local-config.yaml",
				FileMode:          0o777,
			},
			{
				HostFilePath:      currPath,
				ContainerFilePath: "/home/",
				FileMode:          0o777,
			},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	seenMessageAttr := 0
	seenReceiverTypeAttr := 0
	expectedReceiver := "jmx"
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		allLogs := sink.AllLogs()
		if len(allLogs) == 0 {
			assert.Fail(tt, "No logs collected")
			return
		}
		for i := 0; i < len(allLogs); i++ {
			plogs := allLogs[i]
			lrs := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords()
			for j := 0; j < lrs.Len(); j++ {
				lr := lrs.At(j)
				attrMap, ok := lr.Attributes().Get(OtelEntityAttributesAttr)
				if ok {
					m := attrMap.Map()
					discoveryMsg, ok := m.Get(MessageAttr)
					if ok {
						assert.Equal(tt, logMessageToAssert, discoveryMsg.AsString())
						seenMessageAttr++
					}
					discoveryType, ok := m.Get(ReceiverTypeAttr)
					if ok {
						assert.Equal(tt, expectedReceiver, discoveryType.AsString())
						seenReceiverTypeAttr++
					}
				}
			}
		}
		assert.Greater(tt, seenMessageAttr, 0, "Did not see message '%s'", logMessageToAssert)
		assert.Greater(tt, seenReceiverTypeAttr, 0, "Did not see expected type '%s'", expectedReceiver)
	}, 60*time.Second, 1*time.Second, "Did not get '%s' discovery in time", expectedReceiver)

	return &otelContainer{Container: container}, nil
}

func TestIntegrationCassandraJmxAutoDiscovery(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("Integration tests are only run on linux architecture: https://github.com/signalfx/splunk-otel-collector/blob/main/.github/workflows/integration-test.yml#L35")
	}

	successfulDiscoveryMsg := `jmx/cassandra receiver is working!`
	ctx := context.Background()

	tests := map[string]struct {
		ctx                context.Context
		configFileName     string
		logMessageToAssert string
		expected           error
	}{
		"Successful Discovery test": {
			ctx:                ctx,
			configFileName:     "docker_observer_without_ssl_jmx_cassandra_config.yaml",
			logMessageToAssert: successfulDiscoveryMsg,
			expected:           nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			container, err := jmxCassandraAutoDiscoveryHelper(t, test.ctx, test.configFileName, test.logMessageToAssert)

			if !errors.Is(err, test.expected) {
				t.Fatalf(" Expected %v, got %v", test.expected, err)
			}
			t.Cleanup(func() {
				if err := container.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})
		})
	}
}
