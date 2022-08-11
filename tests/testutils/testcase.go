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
)

// A Testcase is a central helper utility to provide Container, OTLPMetricsReceiverSink, ResourceMetrics,
// SplunkOtelCollector, and ObservedLogs to integration tests with minimal boilerplate.  It also embeds *testing.T
// for easy testing and testify usage.
type Testcase struct {
	*testing.T
	Logger                  *zap.Logger
	ObservedLogs            *observer.ObservedLogs
	OTLPMetricsReceiverSink *OTLPMetricsReceiverSink
	OTLPEndpoint            string
	ID                      string
}

// NewTestcase is the recommended constructor that will automatically configure an OTLPMetricsReceiverSink
// with available endpoint and ObservedLogs.
func NewTestcase(t *testing.T) *Testcase {
	tc := Testcase{T: t}
	var logCore zapcore.Core
	logCore, tc.ObservedLogs = observer.New(zap.DebugLevel)
	tc.Logger = zap.New(logCore)

	tc.OTLPEndpoint = getAvailableLocalAddress(t)

	var err error
	tc.OTLPMetricsReceiverSink, err = NewOTLPMetricsReceiverSink().WithEndpoint(tc.OTLPEndpoint).WithLogger(tc.Logger).Build()
	require.NoError(tc, err)
	require.NoError(tc, tc.OTLPMetricsReceiverSink.Start())

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

// Loads and validates a ResourceMetrics instance, assuming it's located in ./testdata/resource_metrics
func (t *Testcase) ResourceMetrics(filename string) *ResourceMetrics {
	expectedResourceMetrics, err := LoadResourceMetrics(
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
func (t *Testcase) SplunkOtelCollector(configFilename string) (collector Collector, shutdown func()) {
	return t.splunkOtelCollector(configFilename, nil)
}

// SplunkOtelCollectorWithEnv works as Testcase.SplunkOtelCollector but also passes an environment variable map
func (t *Testcase) SplunkOtelCollectorWithEnv(configFilename string, env map[string]string) (collector Collector, shutdown func()) {
	return t.splunkOtelCollector(configFilename, env)
}

func (t *Testcase) splunkOtelCollector(configFilename string, env map[string]string) (collector Collector, shutdown func()) {
	if image := os.Getenv("SPLUNK_OTEL_COLLECTOR_IMAGE"); strings.TrimSpace(image) != "" {
		cc := NewCollectorContainer().WithImage(image)
		collector = &cc
	} else {
		cp := NewCollectorProcess()
		collector = &cp
	}

	otlpEndpointForContainer := t.OTLPEndpoint
	if runtime.GOOS == "darwin" {
		port := strings.Split(otlpEndpointForContainer, ":")[1]
		otlpEndpointForContainer = fmt.Sprintf("host.docker.internal:%s", port)
	}

	envVars := map[string]string{
		"OTLP_ENDPOINT":  otlpEndpointForContainer,
		"SPLUNK_TEST_ID": t.ID,
	}

	for k, v := range env {
		envVars[k] = v
	}

	var err error
	collector = collector.WithConfigPath(
		path.Join(".", "testdata", configFilename),
	).WithEnv(envVars).WithLogLevel("debug").WithLogger(t.Logger)

	splunkEnv := map[string]string{}
	for _, s := range os.Environ() {
		split := strings.Split(s, "=")
		if strings.HasPrefix(strings.ToUpper(split[0]), "SPLUNK_") {
			splunkEnv[split[0]] = split[1]

		}
	}
	collector = collector.WithEnv(splunkEnv)

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

// Validating shutdown helper for the Testcase's OTLPMetricsReceiverSink
func (t *Testcase) ShutdownOTLPMetricsReceiverSink() {
	require.NoError(t, t.OTLPMetricsReceiverSink.Shutdown())
}

// AssertAllMetricsReceived is a central helper, designed to avoid most boilerplate.  Using the desired
// ResourceMetrics and Collector Config filenames and a slice of Container builders, AssertAllMetricsReceived
// creates a Testcase, builds and starts all Container and CollectorProcess instances, and asserts that all
// expected ResourceMetrics are received before running validated cleanup functionality.
func AssertAllMetricsReceived(t *testing.T, resourceMetricsFilename, collectorConfigFilename string, containers []Container) {
	tc := NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPMetricsReceiverSink()

	expectedResourceMetrics := tc.ResourceMetrics(resourceMetricsFilename)

	_, stop := tc.Containers(containers...)
	defer stop()

	_, shutdown := tc.SplunkOtelCollector(collectorConfigFilename)
	defer shutdown()

	require.NoError(t, tc.OTLPMetricsReceiverSink.AssertAllMetricsReceived(t, *expectedResourceMetrics, 30*time.Second))
}
