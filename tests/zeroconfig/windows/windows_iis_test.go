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

//go:build windows && zeroconfig

package zeroconfig

import (
	"fmt"
	"net/http"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/ptracetest"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestWindowsIISInstrumentation(t *testing.T) {
	// WARNING:
	//
	// 1. Testcontainers for Go doesn't support Windows containers, see https://github.com/testcontainers/testcontainers-go/issues/948
	//    In light of that will issue docker commands via exec.Command.
	//
	// 2. The test uses the default configuration of the collector that uses signalfx to export metrics.
	//    To avoid building an (expected to be) short lived signalfx sink the test launches a Docker compose
	//    configuration that runs a splunk-otel-collector that receives the O11y signals from the instrumented container.
	//    This way the test can leverage the existing testutils OTLP sink.

	dockerComposeFile := path.Join(".", "testdata", "docker-compose.yaml")
	requireNoErrorExecCommand(t, "docker", "compose", "-f", dockerComposeFile, "up", "--detach")
	defer func() {
		requireNoErrorExecCommand(t, "docker", "compose", "-f", dockerComposeFile, "down")
	}()

	// A firewall rule must be in place for the OTLP Endpoint to be visible to the Docker compose containers.
	// Administrative rights are necessary to create the rule. The rul can be created via the following
	// PowerShell command:
	//
	// New-NetFirewallRule -DisplayName 'zc-iis-test' -Direction Inbound -LocalAddress 10.1.1.1 -LocalPort 4318 -Protocol TCP -Action Allow -Profile Any
	//
	// The command to remove the rule is:
	//
	// Remove-NetFirewallRule -DisplayName 'zc-iis-test'
	//
	// The OTLP also requires the Docker compose network to be created before it is initialized, since
	// the address is part of that network.
	//
	otlp, err := testutils.NewOTLPReceiverSink().WithEndpoint("10.1.1.1:4318").Build()
	require.NoError(t, err)
	require.NoError(t, otlp.Start())
	defer func() {
		require.Nil(t, otlp.Shutdown())
	}()

	// Wait until the splunk-otel-collector is up: relying on the entrypoint of the image
	// can have the request happening before the collector is ready.
	assert.Eventually(t, func() bool {
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:13133/health_check", nil)
		require.NoError(t, err)
		resp, err := httpClient.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 30*time.Second, 100*time.Millisecond)

	testExpectedTracesForHTTPGetRequest(t, otlp, "http://localhost:8000/aspnetcoreapp/api/values/6", filepath.Join("testdata", "expected", "aspnetcore.yaml"))

	testExpectedTracesForHTTPGetRequest(t, otlp, "http://localhost:8000/aspnetfxapp/api/values/4", filepath.Join("testdata", "expected", "aspnetfx.yaml"))
}

func requireNoErrorExecCommand(t *testing.T, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	var out strings.Builder
	cmd.Stdout = &out
	err := cmd.Run()
	require.NoError(t, err)
}

func assertHTTPGetRequestSuccess(c *assert.CollectT, url string) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(c, err)
	if err != nil {
		return
	}

	resp, err := httpClient.Do(req)
	assert.NoError(c, err)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	assert.Equal(c, http.StatusOK, resp.StatusCode)
}

func testExpectedTracesForHTTPGetRequest(t *testing.T, otlp *testutils.OTLPReceiverSink, url, expectedTracesFileName string) {
	expected, err := golden.ReadTraces(expectedTracesFileName)
	require.NoError(t, err)

	// Make only a single successful request to the server to avoid creating multiple traces.
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assertHTTPGetRequestSuccess(c, url)
	}, 3*time.Minute, 100*time.Millisecond, "Failed to connect to target")

	var index int
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		matchErr := fmt.Errorf("no matching traces found, %d collected", index)
		newIndex := len(otlp.AllTraces())
		for i := index; i < newIndex && matchErr != nil; i++ {
			matchErr = ptracetest.CompareTraces(expected, otlp.AllTraces()[i],
				ptracetest.IgnoreResourceAttributeValue("host.id"),
				ptracetest.IgnoreResourceAttributeValue("host.name"),
				ptracetest.IgnoreResourceAttributeValue("process.owner"),
				ptracetest.IgnoreResourceAttributeValue("process.pid"),
				ptracetest.IgnoreResourceAttributeValue("process.runtime.description"),
				ptracetest.IgnoreResourceAttributeValue("process.runtime.version"),
				ptracetest.IgnoreResourceAttributeValue("splunk.zc.method"),
				ptracetest.IgnoreResourceAttributeValue("telemetry.sdk.version"),
				ptracetest.IgnoreResourceAttributeValue("splunk.distro.version"),
				ptracetest.IgnoreResourceAttributeValue("telemetry.distro.version"),
				ptracetest.IgnoreResourceAttributeValue("os.description"),
				ptracetest.IgnoreScopeSpanInstrumentationScopeVersion(),
				ptracetest.IgnoreStartTimestamp(),
				ptracetest.IgnoreEndTimestamp(),
				ptracetest.IgnoreTraceID(),
				ptracetest.IgnoreSpanID(),
			)
		}
		index = newIndex
		assert.NoError(c, matchErr)
	}, 1*time.Minute, 10*time.Millisecond, "Failed to receive expected traces")
}
