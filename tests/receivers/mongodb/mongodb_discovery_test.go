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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type otelContainer struct {
	testcontainers.Container
}

func mongoDBAutoDiscoveryHelper(ctx context.Context, configFile string, logMessageToAssert string) (*otelContainer, error) {
	finfo, err := os.Stat("/var/run/docker.sock")
	if err != nil {
		return nil, err
	}
	fsys := finfo.Sys()
	stat, ok := fsys.(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("OS error occurred while trying to get GID ")
	}
	dockerGID := fmt.Sprintf("%d", stat.Gid)
	otelConfigPath, err := filepath.Abs(filepath.Join(".", "testdata", configFile))
	if err != nil {
		return nil, err
	}
	r, err := os.Open(otelConfigPath)
	if err != nil {
		return nil, err
	}

	currPath, err := filepath.Abs(filepath.Join(".", "testdata"))
	if err != nil {
		return nil, err
	}
	req := testcontainers.ContainerRequest{
		Image: "otelcol:latest",
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{"/var/run/docker.sock:/var/run/docker.sock"}
			hc.NetworkMode = network.NetworkHost
			hc.GroupAdd = []string{dockerGID}
		},
		Env: map[string]string{
			"SPLUNK_REALM":               "us2",
			"SPLUNK_ACCESS_TOKEN":        "12345",
			"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
		},
		Entrypoint: []string{"/otelcol", "--config", "/home/otel-local-config.yaml"},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            r,
				HostFilePath:      otelConfigPath,
				ContainerFilePath: "/home/otel-local-config.yaml",
				FileMode:          0o777,
			},
			{
				HostFilePath:      currPath,
				ContainerFilePath: "/home/",
				FileMode:          0o777,
			},
		},
		NetworkMode: "host",
		WaitingFor:  wait.ForLog(logMessageToAssert).AsRegexp(),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &otelContainer{Container: container}, nil
}

func TestIntegrationMongoDBAutoDiscovery(t *testing.T) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		t.Skip("Integration tests are only run on linux architecture: https://github.com/signalfx/splunk-otel-collector/blob/main/.github/workflows/integration-test.yml#L35")
	}

	ctx := context.Background()
	successfulDiscoveryMsg := `mongodb receiver is working!`
	partialDiscoveryMsg := `Please ensure your user credentials are correctly specified with`

	tests := map[string]struct {
		ctx                context.Context
		configFileName     string
		logMessageToAssert string
		expected           error
	}{
		"Fully Successful Discovery test": {
			ctx:                ctx,
			configFileName:     "docker_observer_without_ssl_mongodb_config.yaml",
			logMessageToAssert: successfulDiscoveryMsg,
			expected:           nil,
		},
		"Partial Discovery test": {
			ctx:                ctx,
			configFileName:     "docker_observer_without_ssl_with_wrong_authentication_mongodb_config.yaml",
			logMessageToAssert: partialDiscoveryMsg,
			expected:           nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			otelC, err := mongoDBAutoDiscoveryHelper(test.ctx, test.configFileName, test.logMessageToAssert)

			if err != test.expected {
				t.Fatalf(" Expected %v, got %v", test.expected, err)
			}
			// Clean up the container after the test is complete
			t.Cleanup(func() {
				if err := otelC.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})
		})
	}

}
