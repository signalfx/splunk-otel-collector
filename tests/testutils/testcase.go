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

package testutils

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/tests/testutils/telemetry"
)

type TestOption int

const (
	OTLPReceiverSinkAllInterfaces TestOption = iota
	OTLPReceiverSinkBindToBridgeGateway
)

type CollectorBuilder func(Collector) Collector

func HasTestOption(opt TestOption, opts []TestOption) bool {
	for _, o := range opts {
		if o == opt {
			return true
		}
	}
	return false
}

// A Testcase is a central helper utility to provide Container, OTLPReceiverSink, ResourceMetrics,
// SplunkOtelCollector, and ObservedLogs to integration tests with minimal boilerplate.  It also embeds testing.TB
// for easy testing and testify usage.
type Testcase struct {
	testing.TB
	Logger                              *zap.Logger
	ObservedLogs                        *observer.ObservedLogs
	OTLPReceiverSink                    *OTLPReceiverSink
	OTLPEndpoint                        string
	OTLPEndpointForCollector            string
	ID                                  string
	OTLPReceiverShouldBindAllInterfaces bool
}

// NewTestcase is the recommended constructor that will automatically configure an OTLPReceiverSink
// with available endpoint and ObservedLogs.
func NewTestcase(t testing.TB, opts ...TestOption) *Testcase {
	tc := Testcase{TB: t}
	var logCore zapcore.Core
	logCore, tc.ObservedLogs = observer.New(zap.DebugLevel)
	tc.Logger = zap.New(logCore)

	tc.setOTLPEndpoint(opts)
	var err error
	tc.OTLPReceiverSink, err = NewOTLPReceiverSink().WithEndpoint(tc.OTLPEndpoint).WithLogger(tc.Logger).Build()
	require.NoError(tc, err)
	require.NoError(tc, tc.OTLPReceiverSink.Start())

	id, err := uuid.NewRandom()
	require.NoError(tc, err)
	tc.ID = id.String()
	return &tc
}

func (t *Testcase) setOTLPEndpoint(opts []TestOption) {
	otlpPort := GetAvailablePort(t)
	otlpHost := "localhost"
	switch {
	case HasTestOption(OTLPReceiverSinkAllInterfaces, opts):
		otlpHost = "0.0.0.0"
	case HasTestOption(OTLPReceiverSinkBindToBridgeGateway, opts):
		client, err := docker.NewClientWithOpts(docker.FromEnv)
		require.NoError(t, err)
		client.NegotiateAPIVersion(context.Background())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		network, err := client.NetworkInspect(ctx, "bridge", types.NetworkInspectOptions{})
		require.NoError(t, err)
		for _, ipam := range network.IPAM.Config {
			otlpHost = ipam.Gateway
		}
		require.NotEmpty(t, otlpHost, "no bridge network gateway detected. Host IP is inaccessible.")
	}
	t.OTLPEndpoint = fmt.Sprintf("%s:%d", otlpHost, otlpPort)
	t.OTLPEndpointForCollector = t.OTLPEndpoint
}

// Loads and validates a ResourceLogs instance, assuming it's located in ./testdata/resource_metrics
func (t *Testcase) ResourceLogs(filename string) *telemetry.ResourceLogs {
	expectedResourceLogs, err := telemetry.LoadResourceLogs(
		path.Join(".", "testdata", "resource_logs", filename),
	)
	require.NoError(t, err)
	require.NotNil(t, expectedResourceLogs)
	return expectedResourceLogs
}

// Loads and validates a ResourceMetrics instance, assuming it's located in ./testdata/resource_metrics
func (t *Testcase) ResourceMetrics(filename string) *telemetry.ResourceMetrics {
	expectedResourceMetrics, err := telemetry.LoadResourceMetrics(
		path.Join(".", "testdata", "resource_metrics", filename),
	)
	require.NoError(t, err)
	require.NotNil(t, expectedResourceMetrics)
	return expectedResourceMetrics
}

// Builds and starts all provided Container builder instances, returning them and a validating stop function.
func (t *Testcase) Containers(builders ...Container) (containers []*Container, stop func()) {
	for _, builder := range builders {
		containers = append(containers, builder.Build())
	}

	for _, container := range containers {
		assert.NoError(t, container.Start(context.Background()))
	}

	stop = func() {
		for _, container := range containers {
			assert.NoError(t, container.Stop(context.Background(), nil))
			assert.NoError(t, container.Terminate(context.Background()))
		}
	}

	return
}

// SplunkOtelCollector builds and starts a collector container or process using the desired config filename
// (assuming it's in the ./testdata directory) returning it and a validating shutdown function.
func (t *Testcase) SplunkOtelCollector(configFilename string, builders ...CollectorBuilder) (collector Collector, shutdown func()) {
	return t.splunkOtelCollector(configFilename, builders...)
}

// SplunkOtelCollectorContainer is the same as SplunkOtelCollector but returns *CollectorContainer.
// If SPLUNK_OTEL_COLLECTOR_IMAGE isn't set, tests that call this will be skipped.
func (t *Testcase) SplunkOtelCollectorContainer(configFilename string, builders ...CollectorBuilder) (collector *CollectorContainer, shutdown func()) {
	cc := NewCollectorContainer().WithImage(GetCollectorImageOrSkipTest(t))
	if runtime.GOOS == "darwin" {
		port := strings.Split(t.OTLPEndpointForCollector, ":")[1]
		t.OTLPEndpointForCollector = fmt.Sprintf("host.docker.internal:%s", port)
	}

	var c Collector
	c, shutdown = t.newCollector(&cc, configFilename, builders...)
	return c.(*CollectorContainer), shutdown
}

// SplunkOtelCollectorProcess is the same as SplunkOtelCollector but returns *CollectorProcess.
func (t *Testcase) SplunkOtelCollectorProcess(configFilename string, builders ...CollectorBuilder) (collector *CollectorProcess, shutdown func()) {
	cp := NewCollectorProcess()

	var c Collector
	c, shutdown = t.newCollector(&cp, configFilename, builders...)
	return c.(*CollectorProcess), shutdown
}

func (t *Testcase) splunkOtelCollector(configFilename string, builders ...CollectorBuilder) (collector Collector, shutdown func()) {
	if image := os.Getenv("SPLUNK_OTEL_COLLECTOR_IMAGE"); strings.TrimSpace(image) != "" {
		return t.SplunkOtelCollectorContainer(configFilename, builders...)
	}
	return t.SplunkOtelCollectorProcess(configFilename, builders...)
}

func (t *Testcase) newCollector(initial Collector, configFilename string, builders ...CollectorBuilder) (collector Collector, shutdown func()) {
	collector = initial
	envVars := map[string]string{
		"OTLP_ENDPOINT":  t.OTLPEndpointForCollector,
		"SPLUNK_TEST_ID": t.ID,
	}

	if configFilename != "" {
		collector = collector.WithConfigPath(
			path.Join(".", "testdata", configFilename),
		)
	}

	collector = collector.WithEnv(envVars).WithLogLevel("debug").WithLogger(t.Logger)

	for _, builder := range builders {
		collector = builder(collector)
	}

	splunkEnv := map[string]string{}
	for _, s := range os.Environ() {
		split := strings.Split(s, "=")
		if strings.HasPrefix(strings.ToUpper(split[0]), "SPLUNK_") {
			splunkEnv[split[0]] = split[1]

		}
	}
	collector = collector.WithEnv(splunkEnv)

	var err error
	collector, err = collector.Build()
	require.NoError(t, err)
	require.NotNil(t, collector)
	require.NoError(t, collector.Start())

	return collector, func() { require.NoError(t, collector.Shutdown()) }
}

// PrintLogsOnFailure will print all ObserverLogs messages if the test has failed.  It's intended to be
// deferred after Testcase creation.
// There is a bug in testcontainers-go so it's not certain these are complete:
// https://github.com/testcontainers/testcontainers-go/pull/323
func (t *Testcase) PrintLogsOnFailure() {
	if !t.Failed() {
		return
	}
	fmt.Printf("Logs: \n")
	for _, statement := range t.ObservedLogs.All() {
		fmt.Printf("%v\n", statement)
	}
}

// Validating shutdown helper for the Testcase's OTLPReceiverSink
func (t *Testcase) ShutdownOTLPReceiverSink() {
	require.NoError(t, t.OTLPReceiverSink.Shutdown())
}

// AssertAllLogsReceived is a central helper, designed to avoid most boilerplate. Using the desired
// ResourceLogs and Collector Config filenames, a slice of Container builders, and a slice of CollectorBuilder
// AssertAllLogsReceived creates a Testcase, builds and starts all Container and CollectorBuilder-determined Collector
// instances, and asserts that all expected ResourceLogs are received before running validated cleanup functionality.
func AssertAllLogsReceived(
	t testing.TB, resourceLogsFilename, collectorConfigFilename string,
	containers []Container, builders []CollectorBuilder,
) {
	tc := NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	expectedResourceLogs := tc.ResourceLogs(resourceLogsFilename)

	_, stop := tc.Containers(containers...)
	defer stop()

	_, shutdown := tc.SplunkOtelCollector(collectorConfigFilename, builders...)
	defer shutdown()

	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedResourceLogs, 30*time.Second))
}

// AssertAllMetricsReceived is a central helper, designed to avoid most boilerplate. Using the desired
// ResourceMetrics and Collector Config filenames, a slice of Container builders, and a slice of CollectorBuilder
// AssertAllMetricsReceived creates a Testcase, builds and starts all Container and CollectorBuilder-determined Collector
// instances, and asserts that all expected ResourceMetrics are received before running validated cleanup functionality.
func AssertAllMetricsReceived(
	t testing.TB, resourceMetricsFilename, collectorConfigFilename string,
	containers []Container, builders []CollectorBuilder,
) {
	tc := NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	expectedResourceMetrics := tc.ResourceMetrics(resourceMetricsFilename)

	_, stop := tc.Containers(containers...)
	defer stop()

	_, shutdown := tc.SplunkOtelCollector(collectorConfigFilename, builders...)
	defer shutdown()

	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}

// WaitForKeyboard is a helper for adding breakpoints during test creation
func WaitForKeyboard(t testing.TB) {
	tty, err := os.Open("/dev/tty")
	require.NoError(t, err)
	reader := bufio.NewReader(tty)
	fmt.Print("Press ENTER to continue.\n")
	_, _ = reader.ReadString('\n')
}
