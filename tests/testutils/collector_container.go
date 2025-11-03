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
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	dockerContainer "github.com/docker/docker/api/types/container"
	dockerMount "github.com/docker/docker/api/types/mount"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"
)

const collectorImageEnvVar = "SPLUNK_OTEL_COLLECTOR_IMAGE"

var (
	_ Collector                  = (*CollectorContainer)(nil)
	_ testcontainers.LogConsumer = (*collectorLogConsumer)(nil)
)

type CollectorContainer struct {
	contextArchive io.ReadSeeker
	Logger         *zap.Logger
	logConsumer    collectorLogConsumer
	Mounts         map[string]string
	Image          string
	ConfigPath     string
	LogLevel       string
	Args           []string
	Ports          []string
	Container      Container
	Fail           bool
}

// To be used as a builder whose Build() method provides the actual instance capable of launching the process.
func NewCollectorContainer() CollectorContainer {
	return CollectorContainer{Args: []string{}, Container: NewContainer(), Mounts: map[string]string{}}
}

// otelcol:latest by default
func (collector CollectorContainer) WithImage(image string) CollectorContainer {
	collector.Image = image
	return collector
}

func (collector CollectorContainer) WithExposedPorts(ports ...string) CollectorContainer {
	collector.Ports = append(collector.Ports, ports...)
	return collector
}

// Will use bundled config by default
func (collector CollectorContainer) WithConfigPath(path string) Collector {
	collector.ConfigPath = path
	return &collector
}

// []string{} by default
func (collector CollectorContainer) WithArgs(args ...string) Collector {
	collector.Args = args
	return &collector
}

// empty by default
func (collector CollectorContainer) WithEnv(env map[string]string) Collector {
	for k, v := range env {
		collector.Container.Env[k] = v
	}
	return &collector
}

// Nop logger by default
func (collector CollectorContainer) WithLogger(logger *zap.Logger) Collector {
	collector.Logger = logger
	return &collector
}

// "info" by default, but currently a noop
func (collector CollectorContainer) WithLogLevel(level string) Collector {
	collector.LogLevel = level
	return &collector
}

func (collector CollectorContainer) WillFail(fail bool) Collector {
	collector.Fail = fail
	return &collector
}

func (collector CollectorContainer) WithMount(path, mountPoint string) Collector {
	collector.Mounts[path] = mountPoint
	return &collector
}

func (collector CollectorContainer) Build() (Collector, error) {
	if collector.Image == "" && collector.Container.Dockerfile.Context == "" {
		collector.Image = "otelcol:latest"
	}

	if collector.Logger == nil {
		collector.Logger = zap.NewNop()
	}
	if collector.LogLevel == "" {
		collector.LogLevel = "info"
	}

	collector.logConsumer = newCollectorLogConsumer(collector.Logger)

	if collector.Container.Dockerfile.Context == "" {
		var err error
		collector.contextArchive, err = collector.buildContextArchive()
		if err != nil {
			return nil, err
		}
		collector.Container = collector.Container.WithContextArchive(
			collector.contextArchive,
		)
	}

	if collector.Container.ContainerNetworkMode == "" {
		collector.Container = collector.Container.WithNetworkMode("host")
	}

	collector.Container = collector.Container.WithExposedPorts(collector.Ports...)

	if len(collector.Container.WaitingFor) == 0 {
		if collector.Fail {
			collector.Container = collector.Container.WillWaitForLogs("")
		} else {
			collector.Container = collector.Container.WillWaitForLogs("Everything is ready. Begin running and processing data.")
		}
	}

	if len(collector.Args) > 0 {
		collector.Container = collector.Container.WithCmd(collector.Args...)
	}

	if len(collector.Mounts) > 0 {
		collector.Container = collector.Container.WithHostConfigModifier(func(hostConfig *dockerContainer.HostConfig) {
			for path, mountPoint := range collector.Mounts {
				hostConfig.Mounts = append(hostConfig.Mounts, dockerMount.Mount{Source: path, Target: mountPoint, Type: dockerMount.TypeBind})
			}
		})
	}

	collector.Container = *(collector.Container.Build())

	return &collector, nil
}

func (collector *CollectorContainer) Start() error {
	if collector.Container.req == nil {
		return fmt.Errorf("cannot Start a CollectorContainer that hasn't been successfully built")
	}

	err := collector.Container.Start(context.Background())
	if err != nil {
		return err
	}
	collector.Container.FollowOutput(collector.logConsumer)
	return collector.Container.StartLogProducer(context.Background())
}

func (collector *CollectorContainer) Shutdown() error {
	if collector.Container.req == nil {
		return fmt.Errorf("cannot Shutdown a CollectorContainer that hasn't been successfully built")
	}
	defer collector.Container.Terminate(context.Background())
	if err := collector.Container.Stop(context.Background(), nil); err != nil {
		return err
	}
	return collector.Container.StopLogProducer()
}

func (collector *CollectorContainer) buildContextArchive() (io.ReadSeeker, error) {
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	dockerfile := fmt.Sprintf("FROM %s\n", collector.Image)
	if collector.ConfigPath != "" {
		config, err := os.ReadFile(collector.ConfigPath)
		if err != nil {
			return nil, err
		}
		header := tar.Header{
			Name:     "config.yaml",
			Mode:     0o777,
			Size:     int64(len(config)),
			Typeflag: tar.TypeReg,
			Format:   tar.FormatGNU,
		}
		if err := tarWriter.WriteHeader(&header); err != nil {
			return nil, err
		}
		if _, err := tarWriter.Write(config); err != nil {
			return nil, err
		}

		dockerfile += "COPY config.yaml /etc/config.yaml\n"

		// We need to tell the Collector to use the provided config
		// but only if not already done so in the test
		configSetByArgs := configIsSetByArgs(collector.Args)
		_, configSetByEnvVar := collector.Container.Env["SPLUNK_CONFIG"]
		if !configSetByArgs && !configSetByEnvVar {
			// only specify w/ args if none are used in the test
			if len(collector.Args) == 0 {
				collector.Args = append(collector.Args, "--config", "/etc/config.yaml")
			} else {
				// fallback to env var
				collector.Container.Env["SPLUNK_CONFIG"] = "/etc/config.yaml"
			}
		}
	}

	header := tar.Header{
		Name:     "Dockerfile",
		Mode:     0o777,
		Size:     int64(len(dockerfile)),
		Typeflag: tar.TypeReg,
		Format:   tar.FormatGNU,
	}
	if err := tarWriter.WriteHeader(&header); err != nil {
		return nil, err
	}
	if _, err := tarWriter.Write([]byte(dockerfile)); err != nil {
		return nil, err
	}
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	reader := bytes.NewReader(buf.Bytes())
	return reader, nil
}

type collectorLogConsumer struct {
	logger *zap.Logger
}

func newCollectorLogConsumer(logger *zap.Logger) collectorLogConsumer {
	return collectorLogConsumer{logger: logger}
}

func (l collectorLogConsumer) Accept(log testcontainers.Log) {
	msg := strings.TrimSpace(string(log.Content))
	if log.LogType == testcontainers.StderrLog {
		l.logger.Info(msg)
	} else {
		l.logger.Debug(msg)
	}
}

func (collector *CollectorContainer) InitialConfig(t testing.TB) map[string]any {
	return collector.execConfigRequest(t, "http://localhost:55679/debug/expvarz", "initial")
}

func (collector *CollectorContainer) EffectiveConfig(t testing.TB) map[string]any {
	return collector.execConfigRequest(t, "http://localhost:55679/debug/expvarz", "effective")
}

func (collector *CollectorContainer) execConfigRequest(t testing.TB, uri, configType string) map[string]any {
	// Wait until the splunk-otel-collector is up: relying on the entrypoint of the image
	// can have the request happening before the collector is ready.
	var body []byte
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", uri, nil)
		require.NoError(t, err)
		resp, err := httpClient.Do(req)
		require.NoError(tt, err)

		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		require.NoError(tt, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
	}, 30*time.Second, 100*time.Millisecond)

	return expvarzPageToMap(t, body, configType)
}

func GetCollectorImage() string {
	return strings.TrimSpace(os.Getenv(collectorImageEnvVar))
}

func CollectorImageIsSet() bool {
	return GetCollectorImage() != ""
}

func SkipIfNotContainerTest(t testing.TB) {
	_ = GetCollectorImageOrSkipTest(t)
}

func GetCollectorImageOrSkipTest(t testing.TB) string {
	image := GetCollectorImage()
	if image == "" {
		t.Skipf("skipping container-only test (set SPLUNK_OTEL_COLLECTOR_IMAGE env var).")
	}
	return image
}

func CollectorImageIsForArm(t testing.TB) bool {
	image := GetCollectorImage()
	if image == "" {
		return false
	}
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	require.NoError(t, err)
	client.NegotiateAPIVersion(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var buf bytes.Buffer
	inspect, err := client.ImageInspect(ctx, image, docker.ImageInspectWithRawResponse(&buf))
	require.NoError(t, err)
	return inspect.Architecture == "arm64"
}
