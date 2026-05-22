// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package launcher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSupervisorEnabled(t *testing.T) {
	assert.False(t, SupervisorEnabled(map[string]string{}))
	assert.False(t, SupervisorEnabled(map[string]string{SupervisorEnabledEnvVar: "false"}))
	assert.True(t, SupervisorEnabled(map[string]string{SupervisorEnabledEnvVar: "true"}))
	assert.True(t, SupervisorEnabled(map[string]string{SupervisorEnabledEnvVar: " TRUE "}))
	assert.False(t, SupervisorEnabled(map[string]string{SupervisorEnabledEnvVar: "1"}))
	assert.False(t, SupervisorEnabled(map[string]string{SupervisorEnabledEnvVar: "YES"}))
}

func TestFilterSupervisorAgentArgs(t *testing.T) {
	args := filterSupervisorAgentArgs([]string{
		"--feature-gates=+splunk.opamp.enabled,-other.gate",
		"--set=processors.batch.timeout=2s",
		"--feature-gates",
		"splunk.opamp.enabled",
	})

	assert.Equal(t, []string{
		"--feature-gates=-other.gate",
		"--set=processors.batch.timeout=2s",
	}, args)
}

func TestFilterSupervisorAgentArgsPreservesMalformedFeatureGateArg(t *testing.T) {
	args := filterSupervisorAgentArgs([]string{"--feature-gates"})
	assert.Equal(t, []string{"--feature-gates"}, args)
}

func TestPrepareSupervisorUsesConfiguredOpAMPEndpointAndOriginalCollectorConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
extensions:
  health_check: {}
  opamp/splunk_o11y:
    server:
      http:
        endpoint: "https://custom.example/v1/opamp"
        polling_interval: 15s
        headers:
          X-SF-Token: "${SPLUNK_ACCESS_TOKEN}"
        tls:
          insecure_skip_verify: true
service:
  extensions: [health_check, opamp/splunk_o11y]
`), 0o600))

	paths := testPaths(dir)
	err := PrepareSupervisor(
		[]string{"--feature-gates=+splunk.opamp.enabled,+other.gate"},
		map[string]string{CollectorConfigEnvVar: configPath, IngestURLEnvVar: "https://ingest.example"},
		paths,
	)
	require.NoError(t, err)

	var supervisorConfig SupervisorConfig
	readYAML(t, paths.SupervisorConfig, &supervisorConfig)
	assert.Equal(t, "https://custom.example/v1/opamp", supervisorConfig.Server.Endpoint)
	assert.Equal(t, map[string]any{"X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"}, supervisorConfig.Server.Headers)
	assert.Equal(t, map[string]any{"insecure_skip_verify": true}, supervisorConfig.Server.TLS)
	assert.Contains(t, readFile(t, paths.SupervisorConfig), "X-SF-Token: ${SPLUNK_ACCESS_TOKEN}")
	assertMinimalCapabilities(t, supervisorConfig.Capabilities)
	assertMinimalCapabilitiesYAML(t, paths.SupervisorConfig)
	assert.Equal(t, []string{collectorConfigEnvRef}, supervisorConfig.Agent.ConfigFiles)
	assert.Equal(t, []string{"--feature-gates=+other.gate"}, supervisorConfig.Agent.Args)
	assert.Contains(t, readFile(t, paths.SupervisorConfig), collectorConfigEnvRef)
	assert.NotContains(t, readFile(t, paths.SupervisorConfig), configPath)
	assert.True(t, supervisorConfig.Agent.PassthroughLogs)
	assert.True(t, supervisorConfig.Agent.UseHUPConfigReload)
	assert.True(t, supervisorConfig.Agent.ValidateConfig)

	assertDirExists(t, paths.StorageDirectory)
	_, err = os.Stat(filepath.Join(paths.StorageDirectory, "collector_config.yaml"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestPrepareSupervisorDerivesEndpointFromIngestURL(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
extensions:
  health_check: {}
service:
  extensions: [health_check]
`), 0o600))

	paths := testPaths(dir)
	err := PrepareSupervisor(nil, map[string]string{
		CollectorConfigEnvVar: configPath,
		IngestURLEnvVar:       "ingest.example/",
	}, paths)
	require.NoError(t, err)

	var supervisorConfig SupervisorConfig
	readYAML(t, paths.SupervisorConfig, &supervisorConfig)
	assert.Equal(t, "${SPLUNK_INGEST_URL}/v1/opamp", supervisorConfig.Server.Endpoint)
	assert.Equal(t, map[string]any{"X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"}, supervisorConfig.Server.Headers)
	assert.Contains(t, readFile(t, paths.SupervisorConfig), "X-SF-Token: ${SPLUNK_ACCESS_TOKEN}")
}

func TestPrepareSupervisorDerivesEndpointFromGatewayURL(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))

	paths := testPaths(dir)
	err := PrepareSupervisor(nil, map[string]string{
		CollectorConfigEnvVar: configPath,
		GatewayURLEnvVar:      "http://gateway.example",
		IngestURLEnvVar:       "https://ingest.example",
	}, paths)
	require.NoError(t, err)

	var supervisorConfig SupervisorConfig
	readYAML(t, paths.SupervisorConfig, &supervisorConfig)
	assert.Equal(t, "${SPLUNK_GATEWAY_URL}:4320/v1/opamp", supervisorConfig.Server.Endpoint)
}

func TestPrepareSupervisorOmitsHUPConfigReloadWhenDisabled(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))

	paths := testPaths(dir)
	paths.UseHUPConfigReload = false
	err := PrepareSupervisor(nil, map[string]string{
		CollectorConfigEnvVar: configPath,
		IngestURLEnvVar:       "https://ingest.example",
	}, paths)
	require.NoError(t, err)

	var supervisorConfig SupervisorConfig
	readYAML(t, paths.SupervisorConfig, &supervisorConfig)
	assert.True(t, supervisorConfig.Agent.PassthroughLogs)
	assert.False(t, supervisorConfig.Agent.UseHUPConfigReload)
	assert.NotContains(t, readFile(t, paths.SupervisorConfig), "use_hup_config_reload")
}

func TestPrepareSupervisorPreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	require.NoError(t, os.MkdirAll(paths.StorageDirectory, 0o755))
	require.NoError(t, os.WriteFile(paths.SupervisorConfig, []byte("user: edited\n"), 0o600))

	err := PrepareSupervisor(
		[]string{"--feature-gates=+splunk.opamp.enabled"},
		map[string]string{IngestURLEnvVar: "https://ingest.example"},
		paths,
	)
	require.NoError(t, err)

	assert.Equal(t, "user: edited\n", readFile(t, paths.SupervisorConfig))
	assertDirExists(t, paths.StorageDirectory)
}

func TestPrepareSupervisorRequiresConfigPath(t *testing.T) {
	err := PrepareSupervisor(nil, map[string]string{IngestURLEnvVar: "https://ingest.example"}, testPaths(t.TempDir()))
	require.Error(t, err)
	assert.Contains(t, err.Error(), CollectorConfigEnvVar)
}

func TestPrepareCommandDirectMode(t *testing.T) {
	paths := testPaths(t.TempDir())
	env := []string{"A=B"}
	cmd, err := PrepareCommand([]string{"--discovery"}, env, paths)
	require.NoError(t, err)
	assert.Equal(t, "otelcol", cmd.Path)
	assert.Equal(t, []string{"--discovery"}, cmd.Args)
	assert.Equal(t, env, cmd.Env)

	_, err = os.Stat(paths.StorageDirectory)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestPrepareCommandSupervisorMode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
extensions:
  health_check: {}
service:
  extensions: [health_check]
`), 0o600))

	paths := testPaths(dir)
	paths.CollectorExecutable = filepath.Join(dir, "missing-otelcol")

	env := []string{
		SupervisorEnabledEnvVar + "=true",
		CollectorConfigEnvVar + "=" + configPath,
		IngestURLEnvVar + "=https://ingest.example",
	}
	cmd, err := PrepareCommand(
		[]string{"--feature-gates=+splunk.opamp.enabled,+other.gate"},
		env,
		paths,
	)
	require.NoError(t, err)

	assert.Equal(t, paths.SupervisorExecutable, cmd.Path)
	assert.Equal(t, []string{"--config", paths.SupervisorConfig}, cmd.Args)
	assert.Equal(t, env, cmd.Env)

	var supervisorConfig SupervisorConfig
	readYAML(t, paths.SupervisorConfig, &supervisorConfig)
	assert.Equal(t, []string{"--feature-gates=+other.gate"}, supervisorConfig.Agent.Args)
	assert.Equal(t, []string{collectorConfigEnvRef}, supervisorConfig.Agent.ConfigFiles)
	assert.Equal(t, paths.CollectorExecutable, supervisorConfig.Agent.Executable)
	assertDirExists(t, paths.StorageDirectory)
}

func assertMinimalCapabilities(t *testing.T, capabilities supervisorCapabilities) {
	t.Helper()

	assert.True(t, capabilities.ReportsEffectiveConfig)
	assert.False(t, capabilities.ReportsOwnMetrics)
	assert.True(t, capabilities.ReportsHealth)
	assert.False(t, capabilities.ReportsHeartbeat)
	assert.True(t, capabilities.ReportsRemoteConfig)
	assert.True(t, capabilities.ReportsAvailableComponents)
	assert.True(t, capabilities.AcceptsRemoteConfig)
}

func assertMinimalCapabilitiesYAML(t *testing.T, path string) {
	t.Helper()

	var raw map[string]any
	readYAML(t, path, &raw)
	capabilities := raw["capabilities"].(map[string]any)
	assert.Equal(t, map[string]any{
		"reports_effective_config":     true,
		"reports_health":               true,
		"reports_available_components": true,
		"accepts_remote_config":        true,
		"reports_remote_config":        true,
		"reports_own_metrics":          false,
		"reports_heartbeat":            false,
	}, capabilities)
}

func testPaths(dir string) Paths {
	return Paths{
		CollectorExecutable:  "otelcol",
		SupervisorExecutable: "opampsupervisor",
		StorageDirectory:     filepath.Join(dir, "supervisor"),
		SupervisorConfig:     filepath.Join(dir, "supervisor", "supervisor_config.yaml"),
		UseHUPConfigReload:   true,
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func readYAML(t *testing.T, path string, out any) {
	t.Helper()
	bytes := []byte(readFile(t, path))
	require.NoError(t, yaml.Unmarshal(bytes, out))
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	bytes, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(bytes)
}
