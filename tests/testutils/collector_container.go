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

	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"
)

var _ Collector = (*CollectorContainer)(nil)
var _ testcontainers.LogConsumer = (*collectorLogConsumer)(nil)

type CollectorContainer struct {
	Image          string
	ConfigPath     string
	Args           []string
	Ports          []string
	Logger         *zap.Logger
	LogLevel       string
	Container      Container
	contextArchive io.Reader
	logConsumer    collectorLogConsumer
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

// []string{} by default, but currently a noop
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
	).WithNetworkMode("host").WillWaitForLogs("Everything is ready. Begin running and processing data.")

	collector.Container = collector.Container.WithExposedPorts(collector.Ports...)

	collector.Container = collector.Container.WithCmd("--config", "/etc/config.yaml", "--log-level", "debug")

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
	defer collector.Container.StopLogProducer()
	return collector.Container.Terminate(context.Background())
}

func (collector CollectorContainer) buildContextArchive() (io.Reader, error) {
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

		dockerfile += `
COPY config.yaml /etc/config.yaml
# ENV SPLUNK_CONFIG=/etc/config.yaml

ENV SPLUNK_ACCESS_TOKEN=12345
ENV SPLUNK_REALM=us0
`
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
