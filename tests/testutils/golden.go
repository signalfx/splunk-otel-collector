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
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
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
func RunMetricsCollectionTest(t *testing.T, configFile, expectedFilePath string,
	options ...MetricsCollectionTestOption,
) {
	opts := &metricCollectionTestOpts{}
	for _, opt := range options {
		opt(opts)
	}

	f := otlpreceiver.NewFactory()
	port := GetAvailablePort(t)
	c := f.CreateDefaultConfig().(*otlpreceiver.Config)
	c.GRPC = configoptional.Some(configgrpc.ServerConfig{
		NetAddr: confignet.AddrConfig{
			Endpoint:  fmt.Sprintf("localhost:%d", port),
			Transport: "tcp",
		},
	})
	c.HTTP = configoptional.None[otlpreceiver.HTTPConfig]()
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

	coverDest := os.Getenv("CONTAINER_COVER_DEST")
	coverSrc := os.Getenv("CONTAINER_COVER_SRC")

	cc := NewCollectorContainer().
		WithImage(GetCollectorImageOrSkipTest(t)).
		WithConfigPath(filepath.Join("testdata", configFile)).
		WithLogger(logger).
		WithEnv(map[string]string{
			"GOCOVERDIR":    coverDest,
			"OTLP_ENDPOINT": fmt.Sprintf("%s:%d", dockerHost, port),
		}).
		WithEnv(opts.collectorEnvVars)
	for k, v := range opts.fileMounts {
		cc.(*CollectorContainer).Container = cc.(*CollectorContainer).Container.WithFile(testcontainers.ContainerFile{
			HostFilePath:      k,
			ContainerFilePath: v,
			FileMode:          0o644,
		})
	}

	if coverSrc != "" && coverDest != "" {
		cc = cc.WithMount(coverSrc, coverDest)
	}

	p, err := cc.Build()
	require.NoError(t, err)
	require.NoError(t, p.Start())
	t.Cleanup(func() {
		require.NoError(t, p.Shutdown())
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

	// for dev purposes - set UPDATE_EXPECTED to update expected file after metrics have been collected
	if os.Getenv("UPDATE_EXPECTED") == "true" {
		allMetrics := sink.AllMetrics()
		if len(allMetrics) == 0 {
			t.Fatalf("Did not receive any metrics to write to golden file")
		}
		actual := allMetrics[len(allMetrics)-1]
		outputPath := filepath.Join("testdata", expectedFilePath)
		dir := filepath.Dir(outputPath)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, golden.WriteMetrics(t, outputPath, actual))
	}
}

func MaybeUpdateExpectedMetricsResults(t *testing.T, file string, metrics *pmetric.Metrics) {
	if shouldUpdateExpectedResults() {
		require.NoError(t, golden.WriteMetrics(t, file, *metrics))
		t.Logf("Wrote updated expected metric results to %s", file)
	}
}

var shouldUpdateExpectedResults = func() bool {
	return os.Getenv("UPDATE_EXPECTED_RESULTS") == "true"
}
