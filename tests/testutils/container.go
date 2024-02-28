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
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultContainerTimeout = 5 * time.Minute
)

// Container is a combination builder and testcontainers.Container wrapper
// for convenient creation and management of docker images and containers.
type Container struct {
	req                  *testcontainers.ContainerRequest
	container            *testcontainers.Container
	startupTimeout       *time.Duration
	Env                  map[string]string
	Labels               map[string]string
	Dockerfile           testcontainers.FromDockerfile
	User                 string
	Image                string
	ContainerName        string
	ContainerNetworkMode string
	Entrypoint           []string
	Cmd                  []string
	ContainerNetworks    []string
	ExposedPorts         []string
	Binds                []string
	WaitingFor           []wait.Strategy
	Mounts               []testcontainers.ContainerMount
	HostConfigModifiers  []func(*dockerContainer.HostConfig)
	Privileged           bool
}

var _ testcontainers.Container = (*Container)(nil)

// To be used as a builder whose Build() method provides the actual instance capable of being started, and that
// implements a testcontainers.Container.
func NewContainer() Container {
	return Container{
		Env:                 map[string]string{},
		HostConfigModifiers: []func(*dockerContainer.HostConfig){},
	}
}

func (container Container) WithImage(image string) Container {
	container.Image = image
	return container
}

func (container Container) WithDockerfile(dockerfile string) Container {
	container.Dockerfile.Dockerfile = dockerfile
	return container
}

func (container Container) WithContext(path string) Container {
	container.Dockerfile.Context = path
	return container
}

func (container Container) WithBuildArgs(args map[string]*string) Container {
	container.Dockerfile.BuildArgs = args
	return container
}

func (container Container) WithContextArchive(contextArchive io.Reader) Container {
	container.Dockerfile.ContextArchive = contextArchive
	return container
}

func (container Container) WithEntrypoint(entrypoint ...string) Container {
	container.Entrypoint = entrypoint
	return container
}

func (container Container) WithCmd(cmd ...string) Container {
	container.Cmd = cmd
	return container
}

func (container Container) WithStartupTimeout(startupTimeout time.Duration) Container {
	container.startupTimeout = &startupTimeout
	return container
}

func (container Container) WithHostConfigModifier(cm func(*dockerContainer.HostConfig)) Container {
	// copy current modifiers since we are in a builder
	var hcm []func(*dockerContainer.HostConfig)
	hcm = append(hcm, container.HostConfigModifiers...)
	hcm = append(hcm, cm)
	container.HostConfigModifiers = hcm
	return container
}

func copyMap(m map[string]string) map[string]string {
	returned := map[string]string{}
	for k, v := range m {
		returned[k] = v
	}
	return returned
}

func (container Container) WithEnv(env map[string]string) Container {
	builder := container
	builder.Env = copyMap(builder.Env)
	for k, v := range env {
		builder.Env[k] = v
	}
	return builder
}

func (container Container) WithEnvVar(key, value string) Container {
	builder := container
	builder.Env = copyMap(builder.Env)
	builder.Env[key] = value
	return builder
}

func (container Container) WithExposedPorts(ports ...string) Container {
	container.ExposedPorts = append(container.ExposedPorts, ports...)
	return container
}

func (container Container) WithMount(mount testcontainers.ContainerMount) Container {
	container.Mounts = append(container.Mounts, mount)
	return container
}

func (container Container) WithName(name string) Container {
	container.ContainerName = name
	return container
}

func (container Container) WithNetworks(networks ...string) Container {
	container.ContainerNetworks = append(container.ContainerNetworks, networks...)
	return container
}

func (container Container) WithNetworkMode(mode string) Container {
	container.ContainerNetworkMode = mode
	return container
}

func (container Container) WillWaitForPorts(ports ...string) Container {
	for _, port := range ports {
		container.WaitingFor = append(container.WaitingFor, wait.ForListeningPort(nat.Port(port)))
	}
	return container
}

func (container Container) WillWaitForLogs(logStatements ...string) Container {
	for _, logStatement := range logStatements {
		container.WaitingFor = append(container.WaitingFor, wait.ForLog(logStatement))
	}
	return container
}

func (container Container) WillWaitForHealth(waitTime time.Duration) Container {
	container.WaitingFor = append(container.WaitingFor, wait.NewHealthStrategy().WithStartupTimeout(waitTime))
	return container
}

func (container Container) WithUser(user string) Container {
	container.User = user
	return container
}

func (container Container) WithPriviledged(privileged bool) Container {
	container.Privileged = privileged
	return container
}

func (container Container) WithBinds(binds ...string) Container {
	container.Binds = append(container.Binds, binds...)
	return container
}

func (container Container) WithLabels(labels map[string]string) Container {
	builder := container
	builder.Labels = copyMap(builder.Labels)
	for k, v := range labels {
		builder.Labels[k] = v
	}
	return builder
}

func (container Container) WithLabel(key, value string) Container {
	builder := container
	builder.Labels = copyMap(builder.Labels)
	builder.Labels[key] = value
	return builder
}

func (container Container) Build() *Container {
	networkMode := dockerContainer.NetworkMode("default")
	if container.ContainerNetworkMode != "" {
		networkMode = dockerContainer.NetworkMode(container.ContainerNetworkMode)
	}
	var startupTimeout time.Duration
	if container.startupTimeout == nil {
		startupTimeout = defaultContainerTimeout
	} else {
		startupTimeout = *container.startupTimeout
	}

	var hostConfigModifier func(config *dockerContainer.HostConfig)
	if len(container.HostConfigModifiers) != 0 {
		hostConfigModifier = func(config *dockerContainer.HostConfig) {
			for _, cm := range container.HostConfigModifiers {
				cm(config)
			}
		}
	}

	container.req = &testcontainers.ContainerRequest{
		Binds:              container.Binds,
		User:               container.User,
		Image:              container.Image,
		FromDockerfile:     container.Dockerfile,
		Cmd:                container.Cmd,
		Entrypoint:         container.Entrypoint,
		Env:                container.Env,
		ExposedPorts:       container.ExposedPorts,
		Name:               container.ContainerName,
		Networks:           container.ContainerNetworks,
		Mounts:             container.Mounts,
		NetworkMode:        networkMode,
		Labels:             container.Labels,
		Privileged:         container.Privileged,
		HostConfigModifier: hostConfigModifier,
		WaitingFor:         wait.ForAll(container.WaitingFor...).WithDeadline(startupTimeout),
	}
	return &container
}

func (container *Container) Start(ctx context.Context) error {
	if container.req == nil {
		return fmt.Errorf("cannot start a container that hasn't been built")
	}
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: *container.req,
		Started:          true,
	}

	err := container.createNetworksIfNecessary(req)
	if err != nil {
		return nil
	}

	started, err := testcontainers.GenericContainer(ctx, req)
	container.container = &started
	return err
}

func (container *Container) assertStarted(operation string) error {
	if container.container == nil || (*container.container) == nil {
		return fmt.Errorf("cannot invoke %s() on unstarted container", operation)
	}
	return nil
}

func (container *Container) Stop(ctx context.Context, timeout *time.Duration) error {
	if err := container.assertStarted("Stop"); err != nil {
		return err
	}
	return (*container.container).Stop(ctx, timeout)
}

func (container *Container) GetContainerID() string {
	if err := container.assertStarted("GetContainerID"); err != nil {
		return ""
	}
	return (*container.container).GetContainerID()
}

func (container *Container) Endpoint(ctx context.Context, s string) (string, error) {
	if err := container.assertStarted("Endpoint"); err != nil {
		return "", err
	}
	return (*container.container).Endpoint(ctx, s)
}

func (container *Container) PortEndpoint(ctx context.Context, port nat.Port, s string) (string, error) {
	if err := container.assertStarted("PortEndpoint"); err != nil {
		return "", err
	}
	return (*container.container).PortEndpoint(ctx, port, s)
}

func (container *Container) Host(ctx context.Context) (string, error) {
	if err := container.assertStarted("Host"); err != nil {
		return "", err
	}
	return (*container.container).Host(ctx)
}

func (container *Container) MappedPort(ctx context.Context, port nat.Port) (nat.Port, error) {
	if err := container.assertStarted("MappedPort"); err != nil {
		return "", err
	}
	return (*container.container).MappedPort(ctx, port)
}

func (container *Container) Ports(ctx context.Context) (nat.PortMap, error) {
	if err := container.assertStarted("Ports"); err != nil {
		return nil, err
	}
	return (*container.container).Ports(ctx)
}

func (container *Container) SessionID() string {
	if err := container.assertStarted("SessionID"); err != nil {
		return ""
	}
	return (*container.container).SessionID()
}

func (container *Container) Terminate(ctx context.Context) error {
	if err := container.assertStarted("Terminate"); err != nil {
		return err
	}
	return (*container.container).Terminate(ctx)
}

func (container *Container) Logs(ctx context.Context) (io.ReadCloser, error) {
	if err := container.assertStarted("Logs"); err != nil {
		return nil, err
	}
	return (*container.container).Logs(ctx)
}

func (container *Container) FollowOutput(consumer testcontainers.LogConsumer) {
	if err := container.assertStarted("FollowOutput"); err == nil {
		(*container.container).FollowOutput(consumer)
	}
}

func (container *Container) StartLogProducer(ctx context.Context) error {
	if err := container.assertStarted("StartLogProducer"); err != nil {
		return err
	}
	return (*container.container).StartLogProducer(ctx)
}

func (container *Container) StopLogProducer() error {
	if err := container.assertStarted("StopLogProducer"); err != nil {
		return err
	}
	return (*container.container).StopLogProducer()
}

func (container *Container) Name(ctx context.Context) (string, error) {
	if err := container.assertStarted("Name"); err != nil {
		return "", err
	}
	return (*container.container).Name(ctx)
}

func (container *Container) Networks(ctx context.Context) ([]string, error) {
	if err := container.assertStarted("Networks"); err != nil {
		return nil, err
	}
	return (*container.container).Networks(ctx)
}

func (container *Container) NetworkAliases(ctx context.Context) (map[string][]string, error) {
	if err := container.assertStarted("NetworkAliases"); err != nil {
		return nil, err
	}
	return (*container.container).NetworkAliases(ctx)
}

func (container *Container) Exec(ctx context.Context, cmd []string, options ...exec.ProcessOption) (int, io.Reader, error) {
	if err := container.assertStarted("Exec"); err != nil {
		return 0, nil, err
	}
	return (*container.container).Exec(ctx, cmd, options...)
}

func (container *Container) ContainerIP(ctx context.Context) (string, error) {
	if err := container.assertStarted("ContainerIP"); err != nil {
		return "", err
	}
	return (*container.container).ContainerIP(ctx)
}

func (container *Container) ContainerIPs(ctx context.Context) ([]string, error) {
	if err := container.assertStarted("ContainerIPs"); err != nil {
		return nil, err
	}
	return (*container.container).ContainerIPs(ctx)
}

func (container *Container) CopyDirToContainer(ctx context.Context, hostDirPath string, containerParentPath string, fileMode int64) error {
	if err := container.assertStarted("CopyDirToContainer"); err != nil {
		return err
	}
	return (*container.container).CopyDirToContainer(ctx, hostDirPath, containerParentPath, fileMode)
}

func (container *Container) CopyFileToContainer(ctx context.Context, hostFilePath string, containerFilePath string, fileMode int64) error {
	if err := container.assertStarted("CopyFileToContainer"); err != nil {
		return err
	}
	return (*container.container).CopyFileToContainer(ctx, hostFilePath, containerFilePath, fileMode)
}

func (container *Container) IsRunning() bool {
	return (*container.container).IsRunning()
}

func (container *Container) State(ctx context.Context) (*types.ContainerState, error) {
	if err := container.assertStarted("State"); err != nil {
		return nil, err
	}
	return (*container.container).State(ctx)
}

func (container *Container) CopyToContainer(ctx context.Context, fileContent []byte, containerFilePath string, fileMode int64) error {
	if err := container.assertStarted("CopyToContainer"); err != nil {
		return err
	}
	return (*container.container).CopyToContainer(ctx, fileContent, containerFilePath, fileMode)
}

func (container *Container) CopyFileFromContainer(ctx context.Context, filePath string) (io.ReadCloser, error) {
	if err := container.assertStarted("CopyFileFromContainer"); err != nil {
		return nil, err
	}
	return (*container.container).CopyFileFromContainer(ctx, filePath)
}

// AssertExec will assert that the exec'ed command completes within the specified timeout, returning
// the return code and demuxed stdout and stderr
func (container *Container) AssertExec(t testing.TB, timeout time.Duration, cmd ...string) (rc int, stdout, stderr string) {
	var err error
	var reader io.Reader
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	rc, reader, err = container.Exec(ctx, cmd)
	assert.NoError(t, err)
	require.NotNil(t, reader)
	var sout, serr bytes.Buffer
	_, err = stdcopy.StdCopy(&sout, &serr, reader)
	require.NoError(t, err)
	return rc, sout.String(), serr.String()
}

// Will create any networks that don't already exist on system.
// Teardown/cleanup is handled by the testcontainers reaper.
func (container *Container) createNetworksIfNecessary(req testcontainers.GenericContainerRequest) error {
	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return err
	}
	for _, networkName := range container.ContainerNetworks {
		query := testcontainers.NetworkRequest{
			Name: networkName,
		}
		networkResource, err := provider.GetNetwork(context.Background(), query)
		if err != nil && !errdefs.IsNotFound(err) {
			return err
		}
		if networkResource.Name != networkName {
			create := testcontainers.NetworkRequest{
				Driver:     "bridge",
				Name:       networkName,
				Attachable: true,
			}
			_, err := provider.CreateNetwork(context.Background(), create)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
