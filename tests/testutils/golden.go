// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package testutils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
)

type metricCollectionTestOpts struct {
	collectorEnvVars      map[string]string
	fileMounts            map[string]string
	compareMetricsOptions []pmetrictest.CompareMetricsOption
}

type MetricsCollectionTestOption func(*metricCollectionTestOpts)

func WithCompareMetricsOptions(options ...pmetrictest.CompareMetricsOption) MetricsCollectionTestOption {
	return func(opts *metricCollectionTestOpts) {
		opts.compareMetricsOptions = append(opts.compareMetricsOptions, options...)
	}
}

func WithCollectorEnvVars(envVars map[string]string) MetricsCollectionTestOption {
	return func(opts *metricCollectionTestOpts) {
		if opts.collectorEnvVars == nil {
			opts.collectorEnvVars = envVars
			return
		}
		for k, v := range envVars {
			opts.collectorEnvVars[k] = v
		}
	}
}

func WithFileMounts(mounts map[string]string) MetricsCollectionTestOption {
	return func(opts *metricCollectionTestOpts) {
		if opts.fileMounts == nil {
			opts.fileMounts = mounts
			return
		}
		for k, v := range mounts {
			opts.fileMounts[k] = v
		}
	}
}

// RunMetricsCollectionTest runs a test that collects metrics using a collector container with provided configFile and
// compares the result with the expected metrics defined in the file expectedFilePath.
func RunMetricsCollectionTest(t *testing.T, configFile string, expectedFilePath string,
	options ...MetricsCollectionTestOption) {
	opts := &metricCollectionTestOpts{}
	for _, opt := range options {
		opt(opts)
	}

	f := otlpreceiver.NewFactory()
	port := GetAvailablePort(t)
	c := f.CreateDefaultConfig().(*otlpreceiver.Config)
	c.GRPC.NetAddr.Endpoint = fmt.Sprintf("localhost:%d", port)
	c.HTTP = nil
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
	cc := NewCollectorContainer().
		WithImage("otelcol:latest").
		WithConfigPath(filepath.Join("testdata", configFile)).
		WithLogger(logger).
		WithEnv(map[string]string{"OTLP_ENDPOINT": fmt.Sprintf("%s:%d", dockerHost, port)}).
		WithEnv(opts.collectorEnvVars).
		WithEnv(map[string]string{"GOCOVERDIR": "/etc/otel/collector/coverage"})

	var path string
	testDirName := "tests"
	if path, err = filepath.Abs("."); err == nil {
		// Coverage should all be under the top-level `tests/coverage` dir that's mounted
		// to the container. This string parsing logic is to ensure different sub-directory
		// tests all put their coverage in the top-level directory.
		index := strings.Index(path, testDirName)
		mountPath := filepath.Join(path[0:index+len(testDirName)], "coverage")
		cc = cc.WithMount(mountPath, "/etc/otel/collector/coverage")
		fmt.Printf("PWD: %s, Container mount, source: %s, destination: %s\n", path, mountPath, "/etc/otel/collector/coverage")
		if fileStat, err := os.Stat(filepath.Join(path[0:index+len(testDirName)], "coverage")); err == nil {
			fmt.Printf("Coverage dir from source stat succeeded, is dir? %v, mode: %v\n", fileStat.IsDir(), fileStat.Mode())
		} else {
			fmt.Printf("coverdir stat err: %v\n", err)
		}
	} else {
		fmt.Printf("Container mount err: %v\n", err)
	}

	for k, v := range opts.fileMounts {
		cc.(*CollectorContainer).Container = cc.(*CollectorContainer).Container.WithFile(testcontainers.ContainerFile{
			HostFilePath:      k,
			ContainerFilePath: v,
			FileMode:          0644,
		})
	}
	p, err := cc.Build()
	require.NoError(t, err)
	require.NoError(t, p.Start())
	t.Cleanup(func() {
		require.NoError(t, p.Shutdown())
		index := strings.Index(path, testDirName)
		cmd := exec.Command("ls", "-la", filepath.Join(path[0:index+len(testDirName)], "coverage"))
		output, er := cmd.CombinedOutput()
		if er == nil {
			fmt.Printf("After shutdown, ls -al %s: %s", filepath.Join(path[0:index+len(testDirName)], "coverage"), string(output))
		} else {
			fmt.Printf("Ran into an error with ls: %v", er)
		}
	})

	expected, err := golden.ReadMetrics(filepath.Join("testdata", expectedFilePath))
	require.NoError(t, err)

	index := 0
	assert.EventuallyWithT(t, func(tt *assert.CollectT) {
		err := fmt.Errorf("no matching metrics found, %d collected", index)
		newIndex := len(sink.AllMetrics())
		for i := index; i < newIndex; i++ {
			m := sink.AllMetrics()[i]
			err = pmetrictest.CompareMetrics(expected, m,
				opts.compareMetricsOptions...)
			if err == nil {
				return
			}
		}
		index = newIndex
		assert.NoError(tt, err)
	}, 30*time.Second, 1*time.Second)
}
