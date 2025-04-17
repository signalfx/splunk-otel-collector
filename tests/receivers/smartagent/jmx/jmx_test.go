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

func TestJmxReceiverProvidesAllMetrics(t *testing.T) {
	metricNames := []string{
		"cassandra.status",
		"cassandra.state",
		"cassandra.load",
		"cassandra.ownership",
	}

	checkMetricsPresence(t, metricNames, "all_metrics_config.yaml")
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
	mountDir, err := filepath.Abs(filepath.Join("testdata", "script.groovy"))
	require.NoError(t, err)
	p, err := testutils.NewCollectorContainer().
		WithImage(GetCollectorImageOrSkipTest(t)).
		WithConfigPath(filepath.Join("testdata", configFile)).
		WithLogger(logger).
		WithEnv(map[string]string{
			"OTLP_ENDPOINT": fmt.Sprintf("%s:%d", dockerHost, port),
			"HOST":          dockerHost,
			"GOCOVERDIR":    "/etc/otel/collector/coverage",
		}).
		WithMount(mountDir, "/opt/script.groovy")
	require.NoError(t, err)

	var path string
	testDirName := "tests"
	if path, err = filepath.Abs("."); err == nil {
		// Coverage should all be under the top-level `tests/coverage` dir that's mounted
		// to the container. This string parsing logic is to ensure different sub-directory
		// tests all put their coverage in the top-level directory.
		index := strings.Index(path, testDirName)
		mountPath := filepath.Join(path[0:index+len(testDirName)], "coverage")
		p = p.WithMount(mountPath, "/etc/otel/collector/coverage")
		fmt.Printf("PWD: %s, Container mount, source: %s, destination: %s\n", path, mountPath, "/etc/otel/collector/coverage")
		if fileStat, err := os.Stat(filepath.Join(path[0:index+len(testDirName)], "coverage")); err == nil {
			fmt.Printf("Coverage dir from source stat succeeded, is dir? %v, mode: %v\n", fileStat.IsDir(), fileStat.Mode())
		} else {
			fmt.Printf("coverdir stat err: %v\n", err)
		}
	} else {
		fmt.Printf("Container mount err: %v\n", err)
	}

	p, err = p.Build()
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
