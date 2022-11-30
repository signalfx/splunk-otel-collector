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
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var configFromArgsPattern = regexp.MustCompile("--config($|[^d-]+)")

var _ Collector = (*CollectorContainer)(nil)
var _ testcontainers.LogConsumer = (*collectorLogConsumer)(nil)

type CollectorContainer struct {
	contextArchive io.Reader
	Logger         *zap.Logger
	logConsumer    collectorLogConsumer
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
	return CollectorContainer{Args: []string{}, Container: NewContainer()}
}

// quay.io/signalfx/splunk-otel-collector:latest by default
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
	collector.Container = collector.Container.WithEnv(env)
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

func (collector CollectorContainer) Build() (Collector, error) {
	if collector.Image == "" {
		collector.Image = "quay.io/signalfx/splunk-otel-collector:latest"
	}
	if collector.Logger == nil {
		collector.Logger = zap.NewNop()
	}
	if collector.LogLevel == "" {
		collector.LogLevel = "info"
	}

	collector.logConsumer = newCollectorLogConsumer(collector.Logger)

	var err error
	collector.contextArchive, err = collector.buildContextArchive()
	if err != nil {
		return nil, err
	}
	collector.Container = collector.Container.WithContextArchive(
		collector.contextArchive,
	)

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

	collector.Container = *(collector.Container.Build())

	return &collector, nil
}

func (collector *CollectorContainer) Start() error {
	if collector.contextArchive == nil {
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
	if collector.contextArchive == nil {
		return fmt.Errorf("cannot Shutdown a CollectorContainer that hasn't been successfully built")
	}
	defer collector.Container.Terminate(context.Background())
	if err := collector.Container.Stop(context.Background(), nil); err != nil {
		return err
	}
	return collector.Container.StopLogProducer()
}

func (collector *CollectorContainer) buildContextArchive() (io.Reader, error) {
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
			Mode:     0777,
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
		var configSetByArgs bool
		for _, c := range collector.Args {
			if configFromArgsPattern.Match([]byte(c)) {
				configSetByArgs = true
			}
		}
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
		Mode:     0777,
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
	msg := string(log.Content)
	if log.LogType == testcontainers.StderrLog {
		l.logger.Info(msg)
	} else {
		l.logger.Debug(msg)
	}
}

func (collector *CollectorContainer) InitialConfig(t testing.TB, port uint16) map[string]any {
	return collector.execConfigRequest(t, fmt.Sprintf("http://localhost:%d/debug/configz/initial", port))
}

func (collector *CollectorContainer) EffectiveConfig(t testing.TB, port uint16) map[string]any {
	return collector.execConfigRequest(t, fmt.Sprintf("http://localhost:%d/debug/configz/effective", port))
}

func (collector *CollectorContainer) execConfigRequest(t testing.TB, uri string) map[string]any {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	n, r, err := collector.Container.Exec(ctx, []string{"curl", "-s", uri})
	require.NoError(t, err)
	require.Zero(t, n)
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	// strip control character from curl output
	require.True(t, len(out) >= 8, "invalid config server output")
	initial := strings.TrimSpace(string(out[8 : len(out)-1]))

	actual := map[string]any{}
	require.NoError(t, yaml.Unmarshal([]byte(initial), &actual))
	return confmap.NewFromStringMap(actual).ToStringMap()
}
