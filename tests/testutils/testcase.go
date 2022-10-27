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
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/tests/testutils/telemetry"
)

type CollectorBuilder func(Collector) Collector

// A Testcase is a central helper utility to provide Container, OTLPReceiverSink, ResourceMetrics,
// SplunkOtelCollector, and ObservedLogs to integration tests with minimal boilerplate.  It also embeds *testing.T
// for easy testing and testify usage.
type Testcase struct {
	*testing.T
	Logger           *zap.Logger
	ObservedLogs     *observer.ObservedLogs
	OTLPReceiverSink *OTLPReceiverSink
	OTLPEndpoint     string
	ID               string
}

// NewTestcase is the recommended constructor that will automatically configure an OTLPReceiverSink
// with available endpoint and ObservedLogs.
func NewTestcase(t *testing.T) *Testcase {
	tc := Testcase{T: t}
	var logCore zapcore.Core
	logCore, tc.ObservedLogs = observer.New(zap.DebugLevel)
	tc.Logger = zap.New(logCore)

	tc.OTLPEndpoint = getAvailableLocalAddress(t)

	var err error
	tc.OTLPReceiverSink, err = NewOTLPReceiverSink().WithEndpoint(tc.OTLPEndpoint).WithLogger(tc.Logger).Build()
	require.NoError(tc, err)
	require.NoError(tc, tc.OTLPReceiverSink.Start())

	id, err := uuid.NewRandom()
	require.NoError(tc, err)
	tc.ID = id.String()
	return &tc
}

// SkipIfNotContainer will skip the test if SPLUNK_OTEL_COLLECTOR_IMAGE env var is empty, otherwise it will
// return the configured image name
func (t *Testcase) SkipIfNotContainer() string {
	image := os.Getenv("SPLUNK_OTEL_COLLECTOR_IMAGE")
	if strings.TrimSpace(image) == "" {
		t.Skipf("skipping container-only test (set SPLUNK_OTEL_COLLECTOR_IMAGE env var).")
		return ""
	}
	return image
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

// SplunkOtelCollectorWithEnv works as Testcase.SplunkOtelCollector but also passes an environment variable map
func (t *Testcase) SplunkOtelCollectorWithEnv(configFilename string, env map[string]string) (collector Collector, shutdown func()) {
	withEnv := func(c Collector) Collector {
		return c.WithEnv(env)
	}
	return t.splunkOtelCollector(configFilename, withEnv)
}

// SplunkOtelCollectorWithBuilderFuncs works as Testcase.SplunkOtelCollector but evaluates an arbitrary number of
// CollectorBuilder functions before calling Build().
func (t *Testcase) SplunkOtelCollectorWithBuilders(configFilename string, builders ...CollectorBuilder) (collector Collector, shutdown func()) {
	return t.splunkOtelCollector(configFilename, builders...)
}

func (t *Testcase) splunkOtelCollector(configFilename string, builders ...CollectorBuilder) (collector Collector, shutdown func()) {
	useDocker := false
	if image := os.Getenv("SPLUNK_OTEL_COLLECTOR_IMAGE"); strings.TrimSpace(image) != "" {
		cc := NewCollectorContainer().WithImage(image)
		collector = &cc
		useDocker = true
	} else {
		cp := NewCollectorProcess()
		collector = &cp
	}

	otlpEndpointForContainer := t.OTLPEndpoint
	if runtime.GOOS == "darwin" && useDocker {
		port := strings.Split(otlpEndpointForContainer, ":")[1]
		otlpEndpointForContainer = fmt.Sprintf("host.docker.internal:%s", port)

	}

	envVars := map[string]string{
		"OTLP_ENDPOINT":  otlpEndpointForContainer,
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

// AssertAllLogsReceived is a central helper, designed to avoid most boilerplate.  Using the desired
// ResourceLogs and Collector Config filenames and a slice of Container builders, AssertAllLogsReceived
// creates a Testcase, builds and starts all Container and CollectorProcess instances, and asserts that all
// expected ResourceLogs are received before running validated cleanup functionality.
func AssertAllLogsReceived(t *testing.T, resourceLogsFilename, collectorConfigFilename string, containers []Container) {
	tc := NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	expectedResourceLogs := tc.ResourceLogs(resourceLogsFilename)

	_, stop := tc.Containers(containers...)
	defer stop()

	_, shutdown := tc.SplunkOtelCollector(collectorConfigFilename)
	defer shutdown()

	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedResourceLogs, 30*time.Second))
}

// AssertAllMetricsReceived is a central helper, designed to avoid most boilerplate.  Using the desired
// ResourceMetrics and Collector Config filenames and a slice of Container builders, AssertAllMetricsReceived
// creates a Testcase, builds and starts all Container and CollectorProcess instances, and asserts that all
// expected ResourceMetrics are received before running validated cleanup functionality.
func AssertAllMetricsReceived(t *testing.T, resourceMetricsFilename, collectorConfigFilename string, containers []Container) {
	tc := NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	expectedResourceMetrics := tc.ResourceMetrics(resourceMetricsFilename)

	_, stop := tc.Containers(containers...)
	defer stop()

	_, shutdown := tc.SplunkOtelCollector(collectorConfigFilename)
	defer shutdown()

	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
