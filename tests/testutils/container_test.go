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
	"io"
	"path"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type noopReader struct{}

func (nr noopReader) Read(b []byte) (int, error) {
	return 0, nil
}

var _ io.Reader = noopReader{}

func TestDockerBuilderMethods(t *testing.T) {
	builder := NewContainer()
	withImage := builder.WithImage("some-image")
	assert.Equal(t, "some-image", withImage.Image)
	assert.NotSame(t, builder, withImage)
	assert.Empty(t, builder.Image)

	withDockerfile := builder.WithDockerfile("some-dockerfile")
	assert.Equal(t, "some-dockerfile", withDockerfile.Dockerfile.Dockerfile)
	assert.NotSame(t, builder, withDockerfile)
	assert.Empty(t, builder.Dockerfile.Dockerfile)

	withContext := builder.WithContext("some-context")
	assert.Equal(t, "some-context", withContext.Dockerfile.Context)
	assert.NotSame(t, builder, withContext)
	assert.Empty(t, builder.Dockerfile.Context)

	contextArchive := noopReader{}
	withContextArchive := builder.WithContextArchive(contextArchive)
	assert.Equal(t, contextArchive, withContextArchive.Dockerfile.ContextArchive)
	assert.NotSame(t, builder, withContextArchive)
	assert.Nil(t, builder.Dockerfile.ContextArchive)

	withCmd := builder.WithCmd("bash", "-c", "'sleep inf'")
	assert.Equal(t, []string{"bash", "-c", "'sleep inf'"}, withCmd.Cmd)
	assert.NotSame(t, builder, withCmd)
	assert.Empty(t, builder.Cmd)

	withName := builder.WithName("some-name")
	assert.Equal(t, "some-name", withName.ContainerName)
	assert.NotSame(t, builder, withName)
	assert.Empty(t, builder.ContainerName)

	withNetworks := builder.WithNetworks("network_one", "network_two")
	assert.Equal(t, []string{"network_one", "network_two"}, withNetworks.ContainerNetworks)
	assert.NotSame(t, builder, withNetworks)
	assert.Nil(t, builder.ContainerNetworks)
}

func TestEnvironmentBuilderMethods(t *testing.T) {
	builder := NewContainer()
	env := map[string]string{"one": "1", "two": "2"}
	withEnv := builder.WithEnv(env)
	assert.Equal(t, env, withEnv.Env)
	assert.NotSame(t, builder, withEnv)
	assert.Empty(t, builder.Env)

	envTwo := map[string]string{"three": "3", "four": "4"}
	additionalWithEnv := withEnv.WithEnv(envTwo)
	expectedEnv := map[string]string{"one": "1", "two": "2", "three": "3", "four": "4"}
	assert.Equal(t, expectedEnv, additionalWithEnv.Env)
	assert.NotSame(t, withEnv, additionalWithEnv)
	assert.Equal(t, env, withEnv.Env)
	assert.NotSame(t, builder, additionalWithEnv)
	assert.Empty(t, builder.Env)

	env = map[string]string{"some": "envvar"}
	withEnvVar := builder.WithEnvVar("some", "envvar")
	assert.Equal(t, env, withEnvVar.Env)
	assert.NotSame(t, builder, withEnvVar)
	assert.Empty(t, builder.Env)

	additionalWithEnvVar := withEnvVar.WithEnvVar("another", "envvar")
	expectedEnv = map[string]string{"some": "envvar", "another": "envvar"}
	assert.Equal(t, expectedEnv, additionalWithEnvVar.Env)
	assert.NotSame(t, withEnvVar, additionalWithEnvVar)
	assert.Equal(t, env, withEnvVar.Env)
	assert.NotSame(t, builder, additionalWithEnvVar)
	assert.Empty(t, builder.Env)
}

func TestExposedPortsBuilderMethod(t *testing.T) {
	builder := NewContainer()
	withExposedPorts := builder.WithExposedPorts("123", "234", "345")
	assert.Equal(t, []string{"123", "234", "345"}, withExposedPorts.ExposedPorts)
	assert.NotSame(t, builder, withExposedPorts)
	assert.Empty(t, builder.ExposedPorts)

	additionalWithExposedPorts := withExposedPorts.WithExposedPorts("456", "567")
	assert.Equal(t, []string{"123", "234", "345", "456", "567"}, additionalWithExposedPorts.ExposedPorts)
	assert.NotSame(t, withExposedPorts, additionalWithExposedPorts)
	assert.Equal(t, []string{"123", "234", "345"}, withExposedPorts.ExposedPorts)
	assert.NotSame(t, builder, withExposedPorts)
	assert.Empty(t, builder.ExposedPorts)
}

func TestWaitingForPortsBuilderMethod(t *testing.T) {
	builder := NewContainer()
	waitForPorts := builder.WillWaitForPorts("123", "234")
	assert.Len(t, waitForPorts.WaitingFor, 2)
	strategy, ok := waitForPorts.WaitingFor[0].(*wait.HostPortStrategy)
	require.True(t, ok)
	assert.EqualValues(t, "123", strategy.Port)
	strategy, ok = waitForPorts.WaitingFor[1].(*wait.HostPortStrategy)
	require.True(t, ok)
	assert.EqualValues(t, "234", strategy.Port)
	assert.Len(t, builder.WaitingFor, 0)

	additionalWaitForPorts := waitForPorts.WillWaitForPorts("345", "456")
	assert.Len(t, additionalWaitForPorts.WaitingFor, 4)
	strategy, ok = additionalWaitForPorts.WaitingFor[0].(*wait.HostPortStrategy)
	require.True(t, ok)
	assert.EqualValues(t, "123", strategy.Port)
	strategy, ok = additionalWaitForPorts.WaitingFor[1].(*wait.HostPortStrategy)
	require.True(t, ok)
	assert.EqualValues(t, "234", strategy.Port)
	strategy, ok = additionalWaitForPorts.WaitingFor[2].(*wait.HostPortStrategy)
	require.True(t, ok)
	assert.EqualValues(t, "345", strategy.Port)
	strategy, ok = additionalWaitForPorts.WaitingFor[3].(*wait.HostPortStrategy)
	require.True(t, ok)
	assert.EqualValues(t, "456", strategy.Port)

	assert.NotSame(t, waitForPorts, additionalWaitForPorts)
	assert.Len(t, waitForPorts.WaitingFor, 2)
	assert.NotSame(t, builder, additionalWaitForPorts)
	assert.Len(t, builder.WaitingFor, 0)
}

func TestWaitingForLogsBuilderMethod(t *testing.T) {
	builder := NewContainer()
	waitForLogs := builder.WillWaitForLogs("statement 1", "statement 2")
	assert.Len(t, waitForLogs.WaitingFor, 2)
	strategy, ok := waitForLogs.WaitingFor[0].(*wait.LogStrategy)
	require.True(t, ok)
	assert.Equal(t, "statement 1", strategy.Log)
	strategy, ok = waitForLogs.WaitingFor[1].(*wait.LogStrategy)
	require.True(t, ok)
	assert.Equal(t, "statement 2", strategy.Log)
	assert.Len(t, builder.WaitingFor, 0)

	additionalWaitForLogs := waitForLogs.WillWaitForLogs("statement 3", "statement 4")
	assert.Len(t, additionalWaitForLogs.WaitingFor, 4)
	strategy, ok = additionalWaitForLogs.WaitingFor[0].(*wait.LogStrategy)
	require.True(t, ok)
	assert.Equal(t, "statement 1", strategy.Log)
	strategy, ok = additionalWaitForLogs.WaitingFor[1].(*wait.LogStrategy)
	require.True(t, ok)
	assert.Equal(t, "statement 2", strategy.Log)
	strategy, ok = additionalWaitForLogs.WaitingFor[2].(*wait.LogStrategy)
	require.True(t, ok)
	assert.Equal(t, "statement 3", strategy.Log)
	strategy, ok = additionalWaitForLogs.WaitingFor[3].(*wait.LogStrategy)
	require.True(t, ok)
	assert.Equal(t, "statement 4", strategy.Log)

	assert.NotSame(t, waitForLogs, additionalWaitForLogs)
	assert.Len(t, waitForLogs.WaitingFor, 2)
	assert.NotSame(t, builder, additionalWaitForLogs)
	assert.Len(t, builder.WaitingFor, 0)
}

func TestBuildMethod(t *testing.T) {
	builder := NewContainer().WithImage("some-image")
	container := builder.Build()
	assert.NotSame(t, *container, builder)
	assert.NotNil(t, container.req)
	assert.Equal(t, "some-image", container.req.Image)
	assert.Nil(t, builder.req)
}

func TestTestcontainersContainerMethodsRequireBuilding(t *testing.T) {
	builder := NewContainer()
	err := builder.Start(context.Background())
	require.Error(t, err)
	assert.Equal(t, "cannot start a container that hasn't been built", err.Error())

	err = builder.Stop(context.Background())
	require.Error(t, err)
	assert.Equal(t, "cannot invoke Stop() on unstarted container", err.Error())

	assert.Empty(t, builder.GetContainerID())

	endpoint, err := builder.Endpoint(context.Background(), "endpoint")
	assert.Empty(t, endpoint)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke Endpoint() on unstarted container", err.Error())

	endpoint, err = builder.PortEndpoint(context.Background(), "", "")
	assert.Empty(t, endpoint)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke PortEndpoint() on unstarted container", err.Error())

	host, err := builder.Host(context.Background())
	assert.Empty(t, host)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke Host() on unstarted container", err.Error())

	mappedPort, err := builder.MappedPort(context.Background(), "0")
	assert.Empty(t, mappedPort)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke MappedPort() on unstarted container", err.Error())

	port, err := builder.Ports(context.Background())
	assert.Empty(t, port)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke Ports() on unstarted container", err.Error())

	sid := builder.SessionID()
	assert.Empty(t, sid)

	err = builder.Terminate(context.Background())
	require.Error(t, err)
	assert.Equal(t, "cannot invoke Terminate() on unstarted container", err.Error())

	rc, err := builder.Logs(context.Background())
	assert.Nil(t, rc)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke Logs() on unstarted container", err.Error())

	// doesn't panic
	builder.FollowOutput(nil)

	err = builder.StartLogProducer(context.Background())
	require.Error(t, err)
	assert.Equal(t, "cannot invoke StartLogProducer() on unstarted container", err.Error())

	err = builder.StopLogProducer()
	require.Error(t, err)
	assert.Equal(t, "cannot invoke StopLogProducer() on unstarted container", err.Error())

	name, err := builder.Name(context.Background())
	assert.Empty(t, name)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke Name() on unstarted container", err.Error())

	networks, err := builder.Networks(context.Background())
	assert.Empty(t, networks)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke Networks() on unstarted container", err.Error())

	aliases, err := builder.NetworkAliases(context.Background())
	assert.Empty(t, aliases)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke NetworkAliases() on unstarted container", err.Error())

	ec, err := builder.Exec(context.Background(), []string{})
	assert.Zero(t, ec)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke Exec() on unstarted container", err.Error())

	cip, err := builder.ContainerIP(context.Background())
	assert.Empty(t, cip)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke ContainerIP() on unstarted container", err.Error())

	err = builder.CopyFileToContainer(context.Background(), "", "", 0)
	require.Error(t, err)
	assert.Equal(t, "cannot invoke CopyFileToContainer() on unstarted container", err.Error())
}

type logConsumer struct {
	sync.Mutex
	statements []string
}

func (lc *logConsumer) Accept(l testcontainers.Log) {
	trimmed := strings.TrimSpace(string(l.Content))
	lc.Lock()
	defer lc.Unlock()
	lc.statements = append(lc.statements, trimmed)
}

var _ testcontainers.LogConsumer = (*logConsumer)(nil)

func TestTestcontainersContainerMethods(t *testing.T) {
	alpine := NewContainer().WithImage("alpine").WithCmd(
		"sh", "-c", "echo rdy > /tmp/something & tail -f /tmp/something",
	).WithExposedPorts("12345:12345").WithName("my-alpine").WithNetworks(
		"bridge", "network_a", "network_b",
	).WillWaitForLogs("rdy").Build()

	defer func() {
		err := alpine.Stop(context.Background())
		require.NoError(t, err)

		err = alpine.Terminate(context.Background())
		require.Error(t, err)
		require.Contains(t, err.Error(), "No such container")
	}()

	err := alpine.Start(context.Background())
	require.NoError(t, err)

	assert.NotEmpty(t, alpine.GetContainerID())

	endpoint, err := alpine.Endpoint(context.Background(), "localhost")
	assert.Equal(t, "localhost://localhost:12345", endpoint)
	require.NoError(t, err)

	endpoint, err = alpine.PortEndpoint(context.Background(), "12345", "localhost")
	assert.Equal(t, "localhost://localhost:12345", endpoint)
	require.NoError(t, err)

	host, err := alpine.Host(context.Background())
	assert.Equal(t, "localhost", host)
	require.NoError(t, err)

	port, err := alpine.MappedPort(context.Background(), "12345")
	assert.EqualValues(t, "12345/tcp", port)
	require.NoError(t, err)

	portMap, err := alpine.Ports(context.Background())

	assert.True(t, func() bool {
		expectedPorts := []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "12345"}}
		expectedPortMap := nat.PortMap(map[nat.Port][]nat.PortBinding{
			"12345/tcp": expectedPorts,
		})

		if assert.ObjectsAreEqual(expectedPortMap, portMap) {
			return true
		}

		expectedPorts = append(expectedPorts, nat.PortBinding{HostIP: "::", HostPort: "12345"})
		expectedPortMap = nat.PortMap(map[nat.Port][]nat.PortBinding{
			"12345/tcp": expectedPorts,
		})
		return assert.ObjectsAreEqual(expectedPortMap, portMap)
	}())
	require.NoError(t, err)

	sid := alpine.SessionID()
	assert.NotEmpty(t, sid)

	rc, err := alpine.Logs(context.Background())
	assert.NotNil(t, rc)
	b := make([]byte, 15)
	rc.Read(b)
	assert.Contains(t, string(b), "rdy")
	assert.NotContains(t, string(b), "sleep inf") // confirm this isn't a bash error logging command

	err = alpine.StartLogProducer(context.Background())
	require.NoError(t, err)

	lc := logConsumer{}
	alpine.FollowOutput(&lc)

	ec, err := alpine.Exec(context.Background(), []string{"sh", "-c", "echo 'some message' >> /tmp/something"})
	assert.Equal(t, 0, ec)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		lc.Lock()
		defer lc.Unlock()
		for _, statement := range lc.statements {
			if statement == "some message" {
				return true
			}
		}
		return false
	}, 10*time.Second, 1*time.Millisecond)

	require.Contains(t, lc.statements, "some message")

	err = alpine.StopLogProducer()
	require.NoError(t, err)

	name, err := alpine.Name(context.Background())
	assert.Equal(t, "/my-alpine", name)
	require.NoError(t, err)

	networks, err := alpine.Networks(context.Background())
	sort.Strings(networks)
	assert.Equal(t, []string{"bridge", "network_a", "network_b"}, networks)
	require.NoError(t, err)

	aliases, err := alpine.NetworkAliases(context.Background())
	assert.NotEmpty(t, aliases)
	require.NoError(t, err)

	cip, err := alpine.ContainerIP(context.Background())
	assert.NotEmpty(t, cip)
	require.NoError(t, err)

	err = alpine.CopyFileToContainer(
		context.Background(), path.Join(".", "testdata", "file_to_transfer"),
		"/tmp/afile", 655,
	)
	require.NoError(t, err)
}
