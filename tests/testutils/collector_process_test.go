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

//go:build testutils

package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestCollectorProcessBuilders(t *testing.T) {
	builder := NewCollectorProcess()
	require.NotNil(t, builder)

	assert.Empty(t, builder.Path)
	assert.Empty(t, builder.ConfigPath)
	assert.Nil(t, builder.Args)
	assert.Nil(t, builder.Logger)
	assert.Empty(t, builder.LogLevel)

	withPath := builder.WithPath("somepath")
	assert.Equal(t, "somepath", withPath.Path)
	assert.Empty(t, builder.Path)

	withConfigPath, ok := builder.WithConfigPath("someconfigpath").(*CollectorProcess)
	require.True(t, ok)
	assert.Equal(t, "someconfigpath", withConfigPath.ConfigPath)
	assert.Empty(t, builder.ConfigPath)

	withArgs, ok := builder.WithArgs("arg_one", "arg_two", "arg_three").(*CollectorProcess)
	require.True(t, ok)
	assert.Equal(t, []string{"arg_one", "arg_two", "arg_three"}, withArgs.Args)
	assert.Empty(t, builder.Args)

	logger := zap.NewNop()
	withLogger, ok := builder.WithLogger(logger).(*CollectorProcess)
	require.True(t, ok)
	assert.Same(t, logger, withLogger.Logger)
	assert.Nil(t, builder.Logger)

	withLogLevel, ok := builder.WithLogLevel("someloglevel").(*CollectorProcess)
	require.True(t, ok)
	assert.Equal(t, "someloglevel", withLogLevel.LogLevel)
	assert.Empty(t, builder.LogLevel)
}

func TestCollectorProcessBuildDefaults(t *testing.T) {
	// specifying Path to avoid built otelcol requirement
	builder := NewCollectorProcess().WithPath("somepath").WithConfigPath("someconfigpath")

	c, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, c)

	collector, ok := c.(*CollectorProcess)
	require.True(t, ok)

	assert.Equal(t, "somepath", collector.Path)
	assert.Equal(t, "someconfigpath", collector.ConfigPath)
	assert.NotNil(t, collector.Logger)
	assert.Equal(t, "info", collector.LogLevel)
	assert.Equal(t, []string{"--set=service.telemetry.logs.level=info", "--config", "someconfigpath", "--set=service.telemetry.metrics.level=none"}, collector.Args)
}

func TestConfigPathNotRequiredUponBuildWithoutArgs(t *testing.T) {
	builder := NewCollectorProcess().WithPath("somepath")
	c, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, c)

	collector, ok := c.(*CollectorProcess)
	require.True(t, ok)
	require.Nil(t, collector.Args)
}

func TestStartAndShutdownInvalidWithoutBuilding(t *testing.T) {
	builder := NewCollectorProcess()

	err := builder.Start()
	require.Error(t, err)
	require.EqualError(t, err, "cannot Start a CollectorProcess that hasn't been successfully built")

	err = builder.Shutdown()
	require.Error(t, err)
	require.EqualError(t, err, "cannot Shutdown a CollectorProcess that hasn't been successfully built")
}

func TestCollectorProcessWithInvalidPaths(t *testing.T) {
	logCore, logObserver := observer.New(zap.DebugLevel)
	logger := zap.New(logCore)
	collector, err := NewCollectorProcess().WithPath("nototel").WithConfigPath("notaconfig").WithLogger(logger).Build()
	require.NotNil(t, collector)
	require.NoError(t, err)

	err = collector.Start()
	require.NoError(t, err)

	cp, ok := collector.(*CollectorProcess)
	require.True(t, ok)

	require.Equal(t, -1, cp.Process.Pid())

	assert.Eventually(t, func() bool {
		for _, entry := range logObserver.All() {
			if field, ok := entry.ContextMap()["error"]; ok {
				return field.(string) == `exec: "nototel": executable file not found in $PATH`
			}
		}
		return false
	}, 5*time.Second, 1*time.Millisecond)

	err = collector.Shutdown()
	require.NoError(t, err)
}
