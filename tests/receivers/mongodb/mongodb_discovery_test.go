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

//go:build integration

package tests

import (
	"context"
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

func mongoDBAutoDiscoveryHelper(t *testing.T, ctx context.Context, configFile string, logMessageToAssert string) (*otelContainer, error) {
	factory := otlpreceiver.NewFactory()
	port := 16745
	c := factory.CreateDefaultConfig().(*otlpreceiver.Config)
	c.GRPC.NetAddr.Endpoint = fmt.Sprintf("localhost:%d", port)
	endpoint := c.GRPC.NetAddr.Endpoint
	sink := &consumertest.LogsSink{}
	receiver, err := factory.CreateLogsReceiver(context.Background(), receivertest.NewNopCreateSettings(), c, sink)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, receiver.Shutdown(context.Background()))
	})

	dockerGID, err := getDockerGID()
	require.NoError(t, err)

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

	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(sink.AllLogs()) == 0 {
			assert.Fail(tt, "No logs collected")
			return
		}
		countAtleastOneGoodLogAttr := 0
		for i := 0; i < len(sink.AllLogs()); i++ {
			plogs := sink.AllLogs()[i]
			lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			attrMap, ok := lr.Attributes().Get(OtelEntityAttributesAttr)
			if ok {
				countAtleastOneGoodLogAttr++
				m := attrMap.Map()
				discoveryMsg, ok := m.Get(MessageAttr)
				if ok {
					assert.Equal(t, logMessageToAssert, discoveryMsg.AsString())
				}
				discoveryType, ok := m.Get(ReceiverTypeAttr)
				if ok {
					assert.Equal(t, "mongodb", discoveryType.AsString())
				}
			}
		}
		assert.True(t, countAtleastOneGoodLogAttr > 0)
	}, 30*time.Second, 1*time.Second)

	return &otelContainer{Container: container}, nil
}
func TestIntegrationMongoDBAutoDiscovery(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("Integration tests are only run on linux architecture: https://github.com/signalfx/splunk-otel-collector/blob/main/.github/workflows/integration-test.yml#L35")
	}

	successfulDiscoveryMsg := `mongodb receiver is working!`
	partialDiscoveryMsg := "Please ensure your user credentials are correctly specified with `--set {{ configProperty \"username\" \"<username>\" }}` and `--set {{ configProperty \"password\" \"<password>\" }}` or `{{ configPropertyEnvVar \"username\" \"<username>\" }}` and `{{ configPropertyEnvVar \"password\" \"<password>\" }}` environment variables."
	ctx := context.Background()

	tests := map[string]struct {
		ctx                context.Context
		configFileName     string
		logMessageToAssert string
		expected           error
	}{

		"Partial Discovery test": {
			ctx:                ctx,
			configFileName:     "docker_observer_without_ssl_with_wrong_authentication_mongodb_config.yaml",
			logMessageToAssert: partialDiscoveryMsg,
			expected:           nil,
		},
		"Successful Discovery test": {
			ctx:                ctx,
			configFileName:     "docker_observer_without_ssl_mongodb_config.yaml",
			logMessageToAssert: successfulDiscoveryMsg,
			expected:           nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			container, err := mongoDBAutoDiscoveryHelper(t, test.ctx, test.configFileName, test.logMessageToAssert)

			if err != test.expected {
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
