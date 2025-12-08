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

//go:build !windows

package discoverytest

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	k8stest "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/xk8stest"
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

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const (
	receiverTypeAttr         = "discovery.receiver.type"
	messageAttr              = "discovery.message"
	otelEntityAttributesAttr = "otel.entity.attributes"
)

func setupReceiver(t *testing.T, endpoint string) *consumertest.LogsSink {
	f := otlpreceiver.NewFactory()
	cfg := f.CreateDefaultConfig().(*otlpreceiver.Config)
	cfg.GRPC = configoptional.Some(configgrpc.ServerConfig{
		NetAddr: confignet.AddrConfig{
			Endpoint:  endpoint,
			Transport: "tcp",
		},
	})
	sink := &consumertest.LogsSink{}
	receiver, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, sink)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, receiver.Shutdown(context.Background()))
	})
	return sink
}

func Run(t *testing.T, receiverName, configFilePath, logMessageToAssert string) {
	port := 16745
	endpoint := fmt.Sprintf("localhost:%d", port)
	sink := setupReceiver(t, endpoint)

	dockerGID, err := getDockerGID()
	require.NoError(t, err)

	coverDest := os.Getenv("CONTAINER_COVER_DEST")
	coverSrc := os.Getenv("CONTAINER_COVER_SRC")
	var coverDirBind string
	if coverSrc != "" && coverDest != "" {
		coverDirBind = fmt.Sprintf("%s:%s", coverSrc, coverDest)
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

// RunWithK8s create a collector in an existing k8s cluster (set env KUBECONFIG for access to cluster)
// and assert that all expectedEntityAttrs for discovered entities are received by the collector in k8s.
func RunWithK8s(t *testing.T, expectedEntityAttrs []map[string]string, setDiscoveryArgs []string) {
	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig == "" {
		t.Fatal("KUBECONFIG environment variable not set")
	}

	skipTearDown := false
	if os.Getenv("SKIP_TEARDOWN") == "true" {
		skipTearDown = true
	}

	port := int(testutils.GetAvailablePort(t))
	endpoint := fmt.Sprintf("0.0.0.0:%d", port)
	sink := setupReceiver(t, endpoint)

	k8sClient, err := k8stest.NewK8sClient(kubeConfig)
	require.NoError(t, err)

	dockerHost := getHostEndpoint(t)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get current file directory")
	}
	currentDir := path.Dir(filename)

	objs, err := k8stest.CreateObjects(k8sClient, filepath.Join(currentDir, "k8s"))
	require.NoError(t, err)
	t.Cleanup(func() {
		if skipTearDown {
			return
		}
		for _, obj := range objs {
			require.NoErrorf(t, k8stest.DeleteObject(k8sClient, obj), "failed to delete object %s", obj.GetName())
		}
	})

	var extraDiscoveryArgs string
	for _, arg := range setDiscoveryArgs {
		extraDiscoveryArgs += fmt.Sprintf("            - --set=%s\n", arg)
	}

	collectorObjs := k8stest.CreateCollectorObjects(t, k8sClient, "test", filepath.Join(currentDir, "k8s", "collector"), map[string]string{"ExtraDiscoveryArgs": extraDiscoveryArgs}, fmt.Sprintf("%s:%d", dockerHost, port))
	t.Cleanup(func() {
		if skipTearDown {
			return
		}
		for _, obj := range collectorObjs {
			require.NoErrorf(t, k8stest.DeleteObject(k8sClient, obj), "failed to delete object %s", obj.GetName())
		}
	})

	collectLogsWithAttrs(t, sink, expectedEntityAttrs)
}

func collectLogsWithAttrs(t *testing.T, sink *consumertest.LogsSink, expectedEntityAttrs []map[string]string) {
	seenLogs := make(map[string]bool)

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
					for _, expectedLog := range expectedEntityAttrs {
						matches := true
						for key, value := range expectedLog {
							attrValue, ok := m.Get(key)
							if !ok || attrValue.AsString() != value {
								matches = false
								break
							}
						}
						if matches {
							seenLogs[fmt.Sprintf("%v", expectedLog)] = true
						}
					}
				}
			}
		}
		for _, expectedLog := range expectedEntityAttrs {
			assert.True(tt, seenLogs[fmt.Sprintf("%v", expectedLog)], "Did not see expected log: %v", expectedLog)
		}
		if len(seenLogs) == len(expectedEntityAttrs) {
			t.Log("Successfully matched all expected logs")
		}
	}, 10*time.Minute, 10*time.Second, "Did not get expected logs in time")
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

// getHostEndpoint returns the docker host endpoint.
func getHostEndpoint(t *testing.T) string {
	if host, ok := os.LookupEnv("HOST_ENDPOINT"); ok {
		return host
	}
	if runtime.GOOS == "darwin" {
		return "host.docker.internal"
	}

	client, err := docker.NewClientWithOpts(docker.FromEnv)
	require.NoError(t, err)
	client.NegotiateAPIVersion(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	network, err := client.NetworkInspect(ctx, "kind", network.InspectOptions{})
	require.NoError(t, err)
	for _, ipam := range network.IPAM.Config {
		if ipam.Gateway != "" {
			return ipam.Gateway
		}
	}
	require.Fail(t, "failed to find host endpoint")
	return ""
}
