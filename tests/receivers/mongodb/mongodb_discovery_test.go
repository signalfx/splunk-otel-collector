// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration

package tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type otelContainer struct {
	testcontainers.Container
}

// LogConsumer represents any object that can
// handle a Log, it is up to the LogConsumer instance
// what to do with the log
type LogConsumer interface {
	Accept(Log)
}

// Log represents a message that was created by a process,
// LogType is either "STDOUT" or "STDERR",
// Content is the byte contents of the message itself
type Log struct {
	LogType string
	Content []byte
}

// StdoutLogConsumer is a LogConsumer that prints the log to stdout
type StdoutLogConsumer struct{}

// Accept prints the log to stdout
func (lc *StdoutLogConsumer) Accept(l Log) {
	fmt.Print(string(l.Content))
}

func fullSuccessfulDiscovery(ctx context.Context) (*otelContainer, error) {

	otelConfigPath, err1 := filepath.Abs(filepath.Join(".", "testdata", "otel-local-config.yaml"))
	r, err2 := os.Open(otelConfigPath)

	currPath, err3 := filepath.Abs(filepath.Join(".", "testdata"))

	req := testcontainers.ContainerRequest{
		Image: "otelcol:latest",
		Env: map[string]string{
			"SPLUNK_REALM":        "us2",
			"SPLUNK_ACCESS_TOKEN": "12345",
			"DOCKER_HOST":         "unix:///var/run/docker.sock",
		},
		Entrypoint: []string{"/otelcol", "--discovery", "--config", "/home/mongodb/otel-local-config.yaml", "--configd", "--config-dir", "/home/mongodb/testdata/configd", "--dry-run"},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            r,
				HostFilePath:      otelConfigPath,
				ContainerFilePath: "/home/mongodb/otel-local-config.yaml",
				FileMode:          0o777,
			},
			{
				HostFilePath:      currPath,
				ContainerFilePath: "/home/mongodb/",
				FileMode:          0o777,
			},
		},
		WaitingFor: wait.ForLog(`Successfully discovered "mongodb" using "docker_observer".*`).AsRegexp(),
		//WaitingFor: wait.ForLog(`Usage of otelcol:`).AsRegexp(),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		err = errors.Join(err, err1, err2, err3)
		return nil, err
	}

	return &otelContainer{Container: container}, nil
}

func TestIntegrationMongoDBAutoDiscovery(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("Integration tests are only run on linux architecture: https://github.com/signalfx/splunk-otel-collector/blob/main/.github/workflows/integration-test.yml#L35")
	}

	ctx := context.Background()

	otelC, err := fullSuccessfulDiscovery(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := otelC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

}
