// Copyright 2021 Splunk, Inc.
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
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/testutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type Testcase struct {
	*testing.T
	Logger                  *zap.Logger
	ObservedLogs            *observer.ObservedLogs
	OTLPMetricsReceiverSink *OTLPMetricsReceiverSink
	OTLPEndpoint            string
}

func NewTestcase(t *testing.T) *Testcase {
	tc := Testcase{T: t}
	var logCore zapcore.Core
	logCore, tc.ObservedLogs = observer.New(zap.DebugLevel)
	tc.Logger = zap.New(logCore)

	tc.OTLPEndpoint = testutil.GetAvailableLocalAddress(t)

	var err error
	tc.OTLPMetricsReceiverSink, err = NewOTLPMetricsReceiverSink().WithEndpoint(tc.OTLPEndpoint).WithLogger(tc.Logger).Build()
	require.NoError(tc, err)
	require.NoError(tc, tc.OTLPMetricsReceiverSink.Start())
	return &tc
}

func (t *Testcase) ResourceMetrics(filename string) *ResourceMetrics {
	expectedResourceMetrics, err := LoadResourceMetrics(
		path.Join(".", "testdata", "resource_metrics", filename),
	)
	require.NoError(t, err)
	require.NotNil(t, expectedResourceMetrics)
	return expectedResourceMetrics
}

func (t *Testcase) Containers(builders ...Container) (containers []*Container, stop func()) {
	for _, builder := range builders {
		containers = append(containers, builder.Build())
	}

	for _, container := range containers {
		require.NoError(t, container.Start(context.Background()))
	}

	stop = func() {
		for _, container := range containers {
			require.NoError(t, container.Stop(context.Background()))
		}
	}

	return
}

func (t *Testcase) SplunkOtelCollector(configFilename string) (*CollectorProcess, func()) {
	collector, err := NewCollectorProcess().WithConfigPath(
		path.Join(".", "testdata", configFilename),
	).WithEnv(map[string]string{
		"OTLP_ENDPOINT": t.OTLPEndpoint,
	}).WithLogLevel("debug").WithLogger(t.Logger).Build()

	require.NoError(t, err)
	require.NotNil(t, collector)
	require.NoError(t, collector.Start())

	return collector, func() { require.NoError(t, collector.Shutdown()) }
}

func (t *Testcase) PrintLogsOnFailure() {
	if !t.Failed() {
		return
	}
	fmt.Printf("Logs: \n")
	for _, statement := range t.ObservedLogs.All() {
		fmt.Printf("%v\n", statement)
	}
}

func (t *Testcase) ShutdownOTLPMetricsReceiverSink() {
	require.NoError(t, t.OTLPMetricsReceiverSink.Shutdown())
}

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
