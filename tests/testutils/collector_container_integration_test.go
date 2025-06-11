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
//go:build testutils && testutilsintegration

package testutils

import (
	"context"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestcontainersContainerMethods(t *testing.T) {
	alpine := NewContainer().WithImage("alpine").WithEntrypoint("sh", "-c").WithCmd(
		"echo rdy > /tmp/something && tail -f /tmp/something",
	).WithExposedPorts("12345:12345").WithName("my-alpine").WithNetworkLabels(
		"bridge", "network_a", "network_b",
	).WillWaitForLogs("rdy").Build()

	defer func() {
		require.NoError(t, alpine.Stop(context.Background(), nil))
		require.NoError(t, alpine.Terminate(context.Background()))
		err := alpine.Terminate(context.Background())
		require.Error(t, err)
		require.Contains(t, err.Error(), "No such container")
	}()

	err := alpine.Start(context.Background())
	log := ""
	if err != nil {
		if logs, e := alpine.Logs(context.Background()); logs != nil && e == nil {
			buf := new(strings.Builder)
			io.Copy(buf, logs)
			log = buf.String()
		}
	}
	require.NoError(t, err, fmt.Sprintf("failed to start container: %q", log))

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
	require.NoError(t, err)
	assert.NotNil(t, rc)

	b := make([]byte, 15)
	rc.Read(b)
	assert.Contains(t, string(b), "rdy")
	assert.NotContains(t, string(b), "sleep inf") // confirm this isn't a bash error logging command

	lc := logConsumer{}
	alpine.FollowOutput(&lc)

	err = alpine.StartLogProducer(context.Background())
	require.NoError(t, err)

	ec, _, err := alpine.Exec(context.Background(), []string{"sh", "-c", "echo 'some message' >> /tmp/something"})
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

	ips, err := alpine.ContainerIPs(context.Background())
	assert.NotEmpty(t, ips)
	require.NoError(t, err)

	err = alpine.CopyFileToContainer(
		context.Background(), path.Join(".", "testdata", "file_to_transfer"),
		"/tmp/afile", 655,
	)
	require.NoError(t, err)

	sc, stdout, stderr := alpine.AssertExec(t, 5*time.Second, "sh", "-c", "echo stdout > /dev/stdout && echo stderr > /dev/stderr")
	require.Equal(t, "stdout\n", stdout)
	require.Equal(t, "stderr\n", stderr)
	require.Zero(t, sc)
}
