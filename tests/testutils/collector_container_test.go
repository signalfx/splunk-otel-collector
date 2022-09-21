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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCollectorContainerBuilders(t *testing.T) {
	builder := NewCollectorContainer()
	require.NotNil(t, builder)

	assert.Empty(t, builder.Image)
	assert.Empty(t, builder.ConfigPath)

	withImage := builder.WithImage("someimage")
	assert.Equal(t, "someimage", withImage.Image)
	assert.Empty(t, builder.Image)

	withConfigPath, ok := builder.WithConfigPath("someconfigpath").(*CollectorContainer)
	require.True(t, ok)
	assert.Equal(t, "someconfigpath", withConfigPath.ConfigPath)
	assert.Empty(t, builder.ConfigPath)

	withArgs, ok := builder.WithArgs("arg_one", "arg_two", "arg_three").(*CollectorContainer)
	require.True(t, ok)
	assert.Equal(t, []string{"arg_one", "arg_two", "arg_three"}, withArgs.Args)
	assert.Empty(t, builder.Args)

	willFail, ok := builder.WillFail(true).(*CollectorContainer)
	require.True(t, ok)
	assert.True(t, willFail.Fail)
	assert.False(t, builder.Fail)

	wontFail, ok := willFail.WillFail(false).(*CollectorContainer)
	require.True(t, ok)
	assert.False(t, wontFail.Fail)
	assert.True(t, willFail.Fail)

	logger := zap.NewNop()
	withLogger, ok := builder.WithLogger(logger).(*CollectorContainer)
	require.True(t, ok)
	assert.Same(t, logger, withLogger.Logger)
	assert.Nil(t, builder.Logger)

	withLogLevel, ok := builder.WithLogLevel("someloglevel").(*CollectorContainer)
	require.True(t, ok)
	assert.Equal(t, "someloglevel", withLogLevel.LogLevel)
	assert.Empty(t, builder.LogLevel)
}

func TestContainerConfigPathNotRequiredUponBuildWithArgs(t *testing.T) {
	withArgs := NewCollectorContainer().WithArgs("arg_one", "arg_two")

	collector, err := withArgs.Build()
	require.NoError(t, err)
	require.NotNil(t, collector)
}

func TestCollectorContainerBuildDefaults(t *testing.T) {
	builder := NewCollectorContainer()

	c, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, c)

	collector, ok := c.(*CollectorContainer)
	require.True(t, ok)

	assert.Equal(t, "quay.io/signalfx/splunk-otel-collector:latest", collector.Image)
	assert.Equal(t, "", collector.ConfigPath)
	assert.NotNil(t, collector.Logger)
	assert.Equal(t, "info", collector.LogLevel)
	assert.Equal(t, []string{}, collector.Args)
}

func TestStartAndShutdownInvalidWithoutBuildingContainer(t *testing.T) {
	builder := NewCollectorContainer()

	err := builder.Start()
	require.Error(t, err)
	require.EqualError(t, err, "cannot Start a CollectorContainer that hasn't been successfully built")

	err = builder.Shutdown()
	require.Error(t, err)
	require.EqualError(t, err, "cannot Shutdown a CollectorContainer that hasn't been successfully built")
}

func TestCollectorContainerWithInvalidImage(t *testing.T) {
	collector, err := NewCollectorContainer().WithImage("&%notanimage%&:latest").Build()
	require.NotNil(t, collector)
	require.NoError(t, err)

	err = collector.Start()
	require.Contains(t, err.Error(), "Error: No such image")

	err = collector.Shutdown()
	require.EqualError(t, err, "cannot invoke Stop() on unstarted container")
}

func TestCollectorContainerWithInvalidConfigPath(t *testing.T) {
	collector, err := NewCollectorContainer().WithConfigPath("notaconfig").Build()
	require.Nil(t, collector)
	require.EqualError(t, err, "open notaconfig: no such file or directory")
}

func TestCollectorContainerLogging(t *testing.T) {
	builder := NewCollectorContainer()

	c, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, c)

	collector, ok := c.(*CollectorContainer)
	require.True(t, ok)

	assert.Equal(t, "quay.io/signalfx/splunk-otel-collector:latest", collector.Image)
	assert.Equal(t, "", collector.ConfigPath)
	assert.NotNil(t, collector.Logger)
	assert.Equal(t, "info", collector.LogLevel)
	assert.Equal(t, []string{}, collector.Args)
}
