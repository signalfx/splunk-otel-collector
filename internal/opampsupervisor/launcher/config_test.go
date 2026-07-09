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
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type badYAML struct{}

func (badYAML) MarshalYAML() (any, error) {
	return nil, errors.New("bad yaml")
}

func TestSupervisorEnabled(t *testing.T) {
	assert.False(t, SupervisorEnabled(map[string]string{}))
	assert.False(t, SupervisorEnabled(map[string]string{SupervisorEnabledEnvVar: "false"}))
	assert.True(t, SupervisorEnabled(map[string]string{SupervisorEnabledEnvVar: "true"}))
	assert.True(t, SupervisorEnabled(map[string]string{SupervisorEnabledEnvVar: " TRUE "}))
	assert.False(t, SupervisorEnabled(map[string]string{SupervisorEnabledEnvVar: "1"}))
}

func TestPrepareCommandDirectMode(t *testing.T) {
	paths := testPaths(t, t.TempDir())
	env := []string{"A=B"}
	args := []string{
		"--config", "/etc/otel/collector/agent_config.yaml",
		"--feature-gates=confmap.enableMergeAppendOption,+other.gate",
		"--discovery",
	}
	cmd, err := PrepareCommand(args, env, paths)
	require.NoError(t, err)
	assert.Equal(t, "otelcol", cmd.Path)
	assert.Equal(t, args, cmd.Args)
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

	paths := testPaths(t, dir)
	paths.CollectorExecutable = filepath.Join(dir, "test-otelcol")

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
	assert.Equal(t, []string{"--config", paths.RuntimeSupervisorConfig}, cmd.Args)
	assert.Equal(t, append(append([]string{}, env...), ListenInterfaceEnvVar+"="+defaultListenInterface), cmd.Env)
}

func TestPrepareCommandSupervisorModeDefaultAgentListenInterface(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(t, dir)
	require.NoError(t, os.WriteFile(paths.DefaultAgentConfig, []byte(`service: {}`), 0o600))

	env := []string{
		SupervisorEnabledEnvVar + "=true",
		CollectorConfigEnvVar + "=" + paths.DefaultAgentConfig,
		IngestURLEnvVar + "=https://ingest.example",
	}
	cmd, err := PrepareCommand(nil, env, paths)
	require.NoError(t, err)

	assert.Equal(t, append(append([]string{}, env...), ListenInterfaceEnvVar+"="+defaultAgentListenInterface), cmd.Env)
}

func TestPrepareCommandSupervisorModePreservesExplicitListenInterface(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))

	paths := testPaths(t, dir)
	env := []string{
		SupervisorEnabledEnvVar + "=true",
		CollectorConfigEnvVar + "=" + configPath,
		IngestURLEnvVar + "=https://ingest.example",
		ListenInterfaceEnvVar + "=127.0.0.2",
	}
	cmd, err := PrepareCommand(nil, env, paths)
	require.NoError(t, err)

	assert.Equal(t, env, cmd.Env)
}

func TestPrepareSupervisorConfigUsesConfiguredOpAMPEndpointAndOriginalCollectorConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
extensions:
  health_check: {}
  http_forwarder/opamp_splunk_o11y: {}
  opamp/custom: {}
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
  extensions: [health_check, opamp/splunk_o11y, http_forwarder/opamp_splunk_o11y, opamp/custom]
`), 0o600))

	paths := testPaths(t, dir)
	err := prepareSupervisor(
		supervisorInputs{
			configFiles: []string{configPath},
			agentArgs: []string{
				"--feature-gates=+splunk.opamp.enabled,+other.gate",
				"--set=processors.batch.timeout=2s",
			},
		},
		map[string]string{IngestURLEnvVar: "https://ingest.example"},
		paths,
	)
	require.NoError(t, err)

	var supervisorConfig SupervisorConfig
	readYAML(t, paths.SupervisorConfig, &supervisorConfig)
	supervisorYAML := readFile(t, paths.SupervisorConfig)
	var runtimeConfig SupervisorConfig
	readYAML(t, paths.RuntimeSupervisorConfig, &runtimeConfig)
	runtimeYAML := readFile(t, paths.RuntimeSupervisorConfig)

	assert.Equal(t, "https://custom.example/v1/opamp", supervisorConfig.Server.Endpoint)
	assert.Equal(t, map[string]any{"X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"}, supervisorConfig.Server.Headers)
	assert.Equal(t, map[string]any{"insecure_skip_verify": true}, supervisorConfig.Server.TLS)
	assert.Contains(t, supervisorYAML, "X-SF-Token: ${SPLUNK_ACCESS_TOKEN}")
	assertMinimalCapabilities(t, supervisorConfig.Capabilities)
	assertMinimalCapabilitiesYAML(t, paths.SupervisorConfig)
	assert.True(t, supervisorConfig.Agent.Description.IncludeResourceAttributes)
	assert.Equal(t, paths.ConfigApplyTimeout, supervisorConfig.Agent.ConfigApplyTimeout)
	assert.Contains(t, supervisorYAML, "include_resource_attributes: true")
	assert.Contains(t, supervisorYAML, "config_apply_timeout: "+paths.ConfigApplyTimeout)
	assert.Contains(t, supervisorYAML, managedAgentComment)
	assert.Contains(t, supervisorYAML, "# "+managedAgentComment+"\nagent:\n")
	assert.Empty(t, supervisorConfig.Agent.Executable)
	assert.Empty(t, supervisorConfig.Agent.ConfigFiles)
	assert.Empty(t, supervisorConfig.Agent.Args)
	assert.Equal(t, paths.CollectorExecutable, runtimeConfig.Agent.Executable)
	require.Len(t, runtimeConfig.Agent.ConfigFiles, 1)
	managedConfigPath := runtimeConfig.Agent.ConfigFiles[0]
	assert.Equal(t, []string{
		"--feature-gates=+splunk.opamp.enabled,+other.gate",
		"--set=processors.batch.timeout=2s",
	}, runtimeConfig.Agent.Args)
	assert.Contains(t, runtimeYAML, managedConfigPath)
	assert.NotContains(t, runtimeYAML, configPath)
	assert.Nil(t, supervisorConfig.Agent.Env)
	assert.NotContains(t, supervisorYAML, "\nenv:")
	assert.True(t, supervisorConfig.Agent.PassthroughLogs)
	assert.True(t, supervisorConfig.Agent.UseHUPConfigReload)
	assert.True(t, supervisorConfig.Agent.ValidateConfig)
	assert.Equal(t, "console", supervisorConfig.Telemetry.Logs.Encoding)
	assert.Equal(t, "console", runtimeConfig.Telemetry.Logs.Encoding)
	assert.Contains(t, supervisorYAML, "telemetry:")
	assert.Contains(t, supervisorYAML, "encoding: console")

	var generatedConfig map[string]any
	readYAML(t, managedConfigPath, &generatedConfig)
	extensions := generatedConfig["extensions"].(map[string]any)
	assert.Contains(t, extensions, "health_check")
	assert.Contains(t, extensions, "http_forwarder/opamp_splunk_o11y")
	assert.Contains(t, extensions, "opamp/custom")
	assert.NotContains(t, extensions, opampSplunkExtension)

	service := generatedConfig["service"].(map[string]any)
	assert.Equal(t, []any{
		"health_check",
		"http_forwarder/opamp_splunk_o11y",
		"opamp/custom",
	}, service["extensions"])
}

func TestPrepareSupervisorConfigUsesManagedLocalConfigs(t *testing.T) {
	dir := t.TempDir()
	baseConfigPath := filepath.Join(dir, "base.yaml")
	overlayConfigPath := filepath.Join(dir, "overlay.yaml")
	otherConfigPath := filepath.Join(dir, "other.yaml")
	require.NoError(t, os.WriteFile(baseConfigPath, []byte(`
extensions:
  health_check: {}
  opamp/splunk_o11y:
    server:
      http:
        endpoint: "https://custom.example/v1/opamp"
service:
  extensions: [health_check, opamp/splunk_o11y]
`), 0o600))
	require.NoError(t, os.WriteFile(overlayConfigPath, []byte(`
extensions:
  zpages: {}
  opamp/splunk_o11y:
    server:
      http:
        endpoint: "https://overlay.example/v1/opamp"
service:
  extensions: [zpages, opamp/splunk_o11y]
`), 0o600))
	require.NoError(t, os.WriteFile(otherConfigPath, []byte(`
extensions:
  health_check: {}
service:
  extensions: [health_check]
`), 0o600))

	paths := testPaths(t, dir)
	err := prepareSupervisor(
		supervisorInputs{
			configFiles: []string{baseConfigPath, overlayConfigPath, otherConfigPath},
			agentArgs: []string{
				"--feature-gates=+other.gate",
				"--set=processors.batch.timeout=2s",
			},
		},
		map[string]string{IngestURLEnvVar: "https://ingest.example"},
		paths,
	)
	require.NoError(t, err)

	var supervisorConfig SupervisorConfig
	readYAML(t, paths.SupervisorConfig, &supervisorConfig)
	assert.Equal(t, "https://overlay.example/v1/opamp", supervisorConfig.Server.Endpoint)

	var runtimeConfig SupervisorConfig
	readYAML(t, paths.RuntimeSupervisorConfig, &runtimeConfig)
	require.Len(t, runtimeConfig.Agent.ConfigFiles, 3)
	baseManagedPath := runtimeConfig.Agent.ConfigFiles[0]
	overlayManagedPath := runtimeConfig.Agent.ConfigFiles[1]
	assert.Equal(t, otherConfigPath, runtimeConfig.Agent.ConfigFiles[2])
	assert.Equal(t, []string{
		"--feature-gates=+other.gate",
		"--set=processors.batch.timeout=2s",
	}, runtimeConfig.Agent.Args)

	var baseManagedConfig map[string]any
	readYAML(t, baseManagedPath, &baseManagedConfig)
	extensions := baseManagedConfig["extensions"].(map[string]any)
	assert.Contains(t, extensions, "health_check")
	assert.NotContains(t, extensions, opampSplunkExtension)
	service := baseManagedConfig["service"].(map[string]any)
	assert.Equal(t, []any{"health_check"}, service["extensions"])

	var overlayManagedConfig map[string]any
	readYAML(t, overlayManagedPath, &overlayManagedConfig)
	extensions = overlayManagedConfig["extensions"].(map[string]any)
	assert.Contains(t, extensions, "zpages")
	assert.NotContains(t, extensions, opampSplunkExtension)
	service = overlayManagedConfig["service"].(map[string]any)
	assert.Equal(t, []any{"zpages"}, service["extensions"])
}

func TestPrepareCommandPreservesExistingConfigAndRecomputesEnv(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(t, dir)
	overlayPath := filepath.Join(dir, "overlay.yaml")
	require.NoError(t, os.WriteFile(overlayPath, []byte(`extensions: {}`), 0o600))
	sourceConfig := `
server:
  # user-owned server setting
  endpoint: https://user.example/v1/opamp
  headers:
    X-SF-Token: user-token
storage:
  directory: /custom/storage
capabilities:
  reports_effective_config: false
  reports_health: false
  reports_available_components: false
  accepts_remote_config: false
  reports_remote_config: false
  reports_own_metrics: true
  reports_heartbeat: true
agent:
  executable: old-otelcol
  config_apply_timeout: 9m
  config_files:
    - old.yaml
  args:
    - --old
  description:
    include_resource_attributes: false
  passthrough_logs: false
  validate_config: false
`
	require.NoError(t, os.WriteFile(paths.SupervisorConfig, []byte(sourceConfig), 0o600))
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
extensions:
  health_check: {}
  opamp/splunk_o11y: {}
service:
  extensions: [opamp/splunk_o11y, health_check]
`), 0o600))

	env := []string{
		SupervisorEnabledEnvVar + "=true",
		CollectorConfigEnvVar + "=" + configPath,
	}
	cmd, err := PrepareCommand([]string{
		"--config", configPath,
		"--config", overlayPath,
		"--feature-gates=+splunk.opamp.enabled",
	}, env, paths)
	require.NoError(t, err)

	var supervisorConfig SupervisorConfig
	readYAML(t, paths.SupervisorConfig, &supervisorConfig)
	var runtimeConfig SupervisorConfig
	readYAML(t, paths.RuntimeSupervisorConfig, &runtimeConfig)
	assert.Equal(t, "https://user.example/v1/opamp", supervisorConfig.Server.Endpoint)
	assert.Equal(t, map[string]any{"X-SF-Token": "user-token"}, supervisorConfig.Server.Headers)
	assert.Equal(t, "/custom/storage", supervisorConfig.Storage.Directory)
	assert.False(t, supervisorConfig.Capabilities.ReportsEffectiveConfig)
	assert.True(t, supervisorConfig.Capabilities.ReportsOwnMetrics)
	assert.Equal(t, "9m", supervisorConfig.Agent.ConfigApplyTimeout)
	assert.False(t, supervisorConfig.Agent.Description.IncludeResourceAttributes)
	assert.False(t, supervisorConfig.Agent.PassthroughLogs)
	assert.False(t, supervisorConfig.Agent.ValidateConfig)
	assert.Equal(t, "old-otelcol", supervisorConfig.Agent.Executable)
	assert.Equal(t, []string{"old.yaml"}, supervisorConfig.Agent.ConfigFiles)
	assert.Equal(t, []string{"--old"}, supervisorConfig.Agent.Args)
	assert.Equal(t, paths.CollectorExecutable, runtimeConfig.Agent.Executable)
	require.Len(t, runtimeConfig.Agent.ConfigFiles, 2)
	managedConfigPath := runtimeConfig.Agent.ConfigFiles[0]
	assert.Equal(t, overlayPath, runtimeConfig.Agent.ConfigFiles[1])
	assert.Equal(t, []string{"--feature-gates=+splunk.opamp.enabled"}, runtimeConfig.Agent.Args)
	supervisorYAML := readFile(t, paths.SupervisorConfig)
	assert.True(t, bytes.Equal([]byte(sourceConfig), []byte(supervisorYAML)))
	assert.Contains(t, supervisorYAML, "user-owned server setting")
	assert.NotContains(t, readFile(t, managedConfigPath), opampSplunkExtension)
	assert.Contains(t, readFile(t, managedConfigPath), "health_check")
	assert.Equal(t, append(append([]string{}, env...), ListenInterfaceEnvVar+"="+defaultListenInterface), cmd.Env)
}

func TestPrepareSupervisorConfigDoesNotRewriteExistingSourceConfig(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(t, dir)
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))

	args := []string{"--feature-gates=+splunk.opamp.enabled"}
	inputs := supervisorInputs{configFiles: []string{configPath}, agentArgs: args}
	env := map[string]string{IngestURLEnvVar: "https://ingest.example"}
	require.NoError(t, prepareSupervisor(inputs, env, paths))
	want := readFile(t, paths.SupervisorConfig)
	oldTime := time.Unix(1000, 0)
	require.NoError(t, os.Chtimes(paths.SupervisorConfig, oldTime, oldTime))

	require.NoError(t, prepareSupervisor(inputs, env, paths))

	assert.Equal(t, want, readFile(t, paths.SupervisorConfig))
	info, err := os.Stat(paths.SupervisorConfig)
	require.NoError(t, err)
	assert.True(t, info.ModTime().Equal(oldTime))
}

func TestPrepareSupervisorConfigRefreshesRuntimeConfigFromCurrentInputs(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(t, dir)
	baseConfigPath := filepath.Join(dir, "base.yaml")
	overlayConfigPath := filepath.Join(dir, "overlay.yaml")
	require.NoError(t, os.WriteFile(baseConfigPath, []byte(`service: {}`), 0o600))
	require.NoError(t, os.WriteFile(overlayConfigPath, []byte(`extensions: {}`), 0o600))
	sourceConfig := `
server:
  endpoint: https://user.example/v1/opamp
storage:
  directory: /custom/storage
agent:
  description:
    include_resource_attributes: false
  config_apply_timeout: 9m
  passthrough_logs: false
  validate_config: false
`
	require.NoError(t, os.WriteFile(paths.SupervisorConfig, []byte(sourceConfig), 0o600))

	err := prepareSupervisor(
		supervisorInputs{
			configFiles: []string{baseConfigPath, overlayConfigPath},
			agentArgs:   []string{"--feature-gates=+splunk.opamp.enabled"},
		},
		map[string]string{IngestURLEnvVar: "https://ingest.example"},
		paths,
	)
	require.NoError(t, err)

	assert.True(t, bytes.Equal([]byte(sourceConfig), []byte(readFile(t, paths.SupervisorConfig))))
	var runtimeConfig SupervisorConfig
	readYAML(t, paths.RuntimeSupervisorConfig, &runtimeConfig)
	assert.Equal(t, "https://user.example/v1/opamp", runtimeConfig.Server.Endpoint)
	assert.Equal(t, "/custom/storage", runtimeConfig.Storage.Directory)
	assert.False(t, runtimeConfig.Agent.Description.IncludeResourceAttributes)
	assert.Equal(t, "9m", runtimeConfig.Agent.ConfigApplyTimeout)
	assert.False(t, runtimeConfig.Agent.PassthroughLogs)
	assert.False(t, runtimeConfig.Agent.ValidateConfig)
	assert.Equal(t, paths.CollectorExecutable, runtimeConfig.Agent.Executable)
	assert.Equal(t, []string{baseConfigPath, overlayConfigPath}, runtimeConfig.Agent.ConfigFiles)
	assert.Equal(t, []string{"--feature-gates=+splunk.opamp.enabled"}, runtimeConfig.Agent.Args)
	assertNoManagedCollectorConfigs(t, paths)
}

func TestPrepareSupervisorRuntimeConfigRemovesStaleArgsWhenCurrentArgsAreEmpty(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(t, dir)
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))
	sourceConfig := `
server:
  endpoint: https://user.example/v1/opamp
storage:
  directory: /custom/storage
agent:
  executable: old-otelcol
  config_files:
    - old.yaml
  args:
    - --old
  description:
    include_resource_attributes: false
  config_apply_timeout: 9m
  passthrough_logs: false
  validate_config: false
`
	require.NoError(t, os.WriteFile(paths.SupervisorConfig, []byte(sourceConfig), 0o600))

	err := prepareSupervisor(supervisorInputs{configFiles: []string{configPath}}, map[string]string{}, paths)
	require.NoError(t, err)

	assert.True(t, bytes.Equal([]byte(sourceConfig), []byte(readFile(t, paths.SupervisorConfig))))
	var runtimeConfig SupervisorConfig
	readYAML(t, paths.RuntimeSupervisorConfig, &runtimeConfig)
	assert.Equal(t, paths.CollectorExecutable, runtimeConfig.Agent.Executable)
	assert.Equal(t, []string{configPath}, runtimeConfig.Agent.ConfigFiles)
	assert.Empty(t, runtimeConfig.Agent.Args)
	assert.NotContains(t, readFile(t, paths.RuntimeSupervisorConfig), "\n  args:")
	assertNoManagedCollectorConfigs(t, paths)
}

func TestPrepareSupervisorConfigReturnsErrors(t *testing.T) {
	tests := map[string]struct {
		setup   func(t *testing.T, dir string, paths Paths) (supervisorInputs, map[string]string)
		wantErr string
	}{
		"Collector config read": {
			setup: func(_ *testing.T, dir string, _ Paths) (supervisorInputs, map[string]string) {
				return supervisorInputs{configFiles: []string{filepath.Join(dir, "missing.yaml")}},
					map[string]string{IngestURLEnvVar: "https://ingest.example"}
			},
			wantErr: "read collector config",
		},
		"Collector config parse": {
			setup: func(t *testing.T, dir string, _ Paths) (supervisorInputs, map[string]string) {
				configPath := filepath.Join(dir, "collector.yaml")
				require.NoError(t, os.WriteFile(configPath, []byte("extensions: ["), 0o600))
				return supervisorInputs{configFiles: []string{configPath}},
					map[string]string{IngestURLEnvVar: "https://ingest.example"}
			},
			wantErr: "parse collector config",
		},
		"Missing endpoint inputs": {
			setup: func(t *testing.T, dir string, _ Paths) (supervisorInputs, map[string]string) {
				configPath := filepath.Join(dir, "collector.yaml")
				require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))
				return supervisorInputs{configFiles: []string{configPath}}, map[string]string{}
			},
			wantErr: "could not derive OpAMP endpoint",
		},
		"Existing supervisor config parse": {
			setup: func(t *testing.T, dir string, paths Paths) (supervisorInputs, map[string]string) {
				configPath := filepath.Join(dir, "collector.yaml")
				require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))
				require.NoError(t, os.WriteFile(paths.SupervisorConfig, []byte("agent: ["), 0o600))
				return supervisorInputs{configFiles: []string{configPath}},
					map[string]string{IngestURLEnvVar: "https://ingest.example"}
			},
			wantErr: "parse supervisor config",
		},
		"Existing supervisor config missing agent": {
			setup: func(t *testing.T, dir string, paths Paths) (supervisorInputs, map[string]string) {
				configPath := filepath.Join(dir, "collector.yaml")
				require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))
				require.NoError(t, os.WriteFile(paths.SupervisorConfig, []byte("server: {endpoint: https://user.example/v1/opamp}\n"), 0o600))
				return supervisorInputs{configFiles: []string{configPath}},
					map[string]string{IngestURLEnvVar: "https://ingest.example"}
			},
			wantErr: "must contain agent mapping",
		},
		"Existing supervisor config invalid agent": {
			setup: func(t *testing.T, dir string, paths Paths) (supervisorInputs, map[string]string) {
				configPath := filepath.Join(dir, "collector.yaml")
				require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))
				require.NoError(t, os.WriteFile(paths.SupervisorConfig, []byte("agent: invalid\n"), 0o600))
				return supervisorInputs{configFiles: []string{configPath}},
					map[string]string{IngestURLEnvVar: "https://ingest.example"}
			},
			wantErr: "must contain agent mapping",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			paths := testPaths(t, dir)
			inputs, env := tt.setup(t, dir, paths)
			err := prepareSupervisor(inputs, env, paths)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestLoadCollectorConfigFileEmptyConfigReturnsEmptyMap(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, nil, 0o600))

	config, err := loadCollectorConfigFile(configPath)
	require.NoError(t, err)
	assert.Empty(t, config)
}

func TestSupervisorCommandEnv(t *testing.T) {
	defaultAgentConfig := "/etc/otel/collector/agent_config.yaml"
	tests := map[string]struct {
		configPaths []string
		env         []string
		want        []string
	}{
		"default agent config first": {
			configPaths: []string{defaultAgentConfig, "/etc/otel/collector/custom.yaml"},
			want: []string{
				CollectorConfigEnvVar + "=" + defaultAgentConfig,
				ListenInterfaceEnvVar + "=" + defaultAgentListenInterface,
			},
		},
		"default agent config later": {
			configPaths: []string{"/etc/otel/collector/custom.yaml", defaultAgentConfig},
			want: []string{
				CollectorConfigEnvVar + "=/etc/otel/collector/custom.yaml",
				ListenInterfaceEnvVar + "=" + defaultAgentListenInterface,
			},
		},
		"default agent config cleaned path": {
			configPaths: []string{"/etc/otel/collector/./agent_config.yaml"},
			want: []string{
				CollectorConfigEnvVar + "=/etc/otel/collector/./agent_config.yaml",
				ListenInterfaceEnvVar + "=" + defaultAgentListenInterface,
			},
		},
		"gateway and custom configs": {
			configPaths: []string{"/etc/otel/collector/gateway_config.yaml", "/etc/otel/collector/custom.yaml"},
			want: []string{
				CollectorConfigEnvVar + "=/etc/otel/collector/gateway_config.yaml",
				ListenInterfaceEnvVar + "=" + defaultListenInterface,
			},
		},
		"existing listen interface": {
			configPaths: []string{defaultAgentConfig},
			env: []string{
				CollectorConfigEnvVar + "=" + defaultAgentConfig,
				ListenInterfaceEnvVar + "=127.0.0.2",
			},
			want: []string{
				CollectorConfigEnvVar + "=" + defaultAgentConfig,
				ListenInterfaceEnvVar + "=127.0.0.2",
			},
		},
		"existing empty listen interface": {
			configPaths: []string{defaultAgentConfig},
			env: []string{
				CollectorConfigEnvVar + "=" + defaultAgentConfig,
				ListenInterfaceEnvVar + "=",
			},
			want: []string{
				CollectorConfigEnvVar + "=" + defaultAgentConfig,
				ListenInterfaceEnvVar + "=",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			environ := tt.env
			if environ == nil {
				environ = []string{CollectorConfigEnvVar + "=" + tt.configPaths[0]}
			}
			assert.Equal(t, tt.want, supervisorCommandEnv(environ, environToMap(environ), Paths{
				DefaultAgentConfig: defaultAgentConfig,
			}, tt.configPaths))
		})
	}
}

func TestSupervisorInputsFromArgsFiltersFeatureGates(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))

	tests := map[string]struct {
		args     []string
		wantArgs []string
	}{
		"split form removes merge append and preserves other gates": {
			args: []string{
				"--feature-gates", "confmap.enableMergeAppendOption,+other.gate",
				"--set=processors.batch.timeout=2s",
			},
			wantArgs: []string{
				"--feature-gates", "+other.gate",
				"--set=processors.batch.timeout=2s",
			},
		},
		"equals form removes signed merge append and preserves other gates": {
			args:     []string{"--feature-gates=-confmap.enableMergeAppendOption,bare.gate,+other.gate"},
			wantArgs: []string{"--feature-gates=bare.gate,+other.gate"},
		},
		"only merge append gate removes feature gate arg": {
			args:     []string{"--feature-gates=confmap.enableMergeAppendOption", "--discovery"},
			wantArgs: []string{"--discovery"},
		},
		"malformed split form is preserved": {
			args:     []string{"--feature-gates"},
			wantArgs: []string{"--feature-gates"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			inputs, err := supervisorInputsFromArgs(
				tt.args,
				map[string]string{CollectorConfigEnvVar: configPath},
			)
			require.NoError(t, err)
			assert.Equal(t, []string{configPath}, inputs.configFiles)
			assert.Equal(t, tt.wantArgs, inputs.agentArgs)
		})
	}
}

func TestSupervisorInputsFromArgsConsumesLocalConfigArgs(t *testing.T) {
	dir := t.TempDir()
	baseConfigPath := filepath.Join(dir, "base.yaml")
	overlayConfigPath := filepath.Join(dir, "overlay.yaml")
	envConfigPath := filepath.Join(dir, "env.yaml")

	inputs, err := supervisorInputsFromArgs(
		[]string{
			"--config", baseConfigPath,
			"--config=file:" + overlayConfigPath,
			"--feature-gates=confmap.enableMergeAppendOption,+other.gate",
			"--set=processors.batch.timeout=2s",
		},
		map[string]string{CollectorConfigEnvVar: envConfigPath},
	)
	require.NoError(t, err)
	assert.Equal(t, []string{baseConfigPath, overlayConfigPath}, inputs.configFiles)
	assert.Equal(t, []string{
		"--feature-gates=+other.gate",
		"--set=processors.batch.timeout=2s",
	}, inputs.agentArgs)
}

func TestSupervisorInputsFromArgsReturnsErrors(t *testing.T) {
	tests := map[string]struct {
		env     map[string]string
		wantErr string
		args    []string
	}{
		"missing config path": {
			env:     map[string]string{IngestURLEnvVar: "https://ingest.example"},
			wantErr: "requires at least one --config flag or " + CollectorConfigEnvVar,
		},
		"missing config flag value": {
			args:    []string{"--config"},
			env:     map[string]string{IngestURLEnvVar: "https://ingest.example"},
			wantErr: "missing value for --config",
		},
		"empty config flag value": {
			args:    []string{"--config="},
			env:     map[string]string{IngestURLEnvVar: "https://ingest.example"},
			wantErr: "collector config path is empty",
		},
		"empty file URI config flag value": {
			args:    []string{"--config=file:"},
			env:     map[string]string{IngestURLEnvVar: "https://ingest.example"},
			wantErr: "collector config path is empty",
		},
		"unsupported config provider URI": {
			args:    []string{"--config", "env:SOME_CONFIG"},
			env:     map[string]string{IngestURLEnvVar: "https://ingest.example"},
			wantErr: "unsupported config provider URI",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := supervisorInputsFromArgs(tt.args, tt.env)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestNormalizeLocalConfigPath(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr string
	}{
		"plain local path": {
			input: "/etc/otel/collector/agent_config.yaml",
			want:  "/etc/otel/collector/agent_config.yaml",
		},
		"file URI absolute path": {
			input: "file:/etc/otel/collector/agent_config.yaml",
			want:  "/etc/otel/collector/agent_config.yaml",
		},
		"file URI cleaned path": {
			input: "file:/etc/otel/collector/./agent_config.yaml",
			want:  "/etc/otel/collector/agent_config.yaml",
		},
		"file URI relative path": {
			input: "file:agent_config.yaml",
			want:  "agent_config.yaml",
		},
		"file URI windows drive path": {
			input: `file:C:\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`,
			want:  `C:\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`,
		},
		"windows drive path": {
			input: `C:\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`,
			want:  `C:\ProgramData\Splunk\OpenTelemetry Collector\agent_config.yaml`,
		},
		"empty path": {
			input:   " ",
			wantErr: "collector config path is empty",
		},
		"empty file URI": {
			input:   "file:",
			wantErr: "collector config path is empty",
		},
		"unsupported provider": {
			input:   "env:SOME_CONFIG",
			wantErr: "unsupported config provider URI",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := normalizeLocalConfigPath(tt.input)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSupervisorServerFromConfigDerivesFallbackEndpoint(t *testing.T) {
	tests := map[string]struct {
		env          map[string]string
		wantEndpoint string
	}{
		"ingest url": {
			env:          map[string]string{IngestURLEnvVar: "ingest.example/"},
			wantEndpoint: "${SPLUNK_INGEST_URL}/v1/opamp",
		},
		"gateway url takes precedence": {
			env: map[string]string{
				GatewayURLEnvVar: "http://gateway.example",
				IngestURLEnvVar:  "https://ingest.example",
			},
			wantEndpoint: "${SPLUNK_GATEWAY_URL}:4320/v1/opamp",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server, err := supervisorServerFromConfigs(nil, tt.env)
			require.NoError(t, err)

			assert.Equal(t, tt.wantEndpoint, server.Endpoint)
			assert.Equal(t, map[string]any{"X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"}, server.Headers)
		})
	}
}

func TestSupervisorServerFromConfigsUsesLastConfiguredEndpoint(t *testing.T) {
	server, err := supervisorServerFromConfigs([]collectorConfigInput{
		{
			Path: "first.yaml",
			Config: map[string]any{
				"extensions": map[string]any{
					opampSplunkExtension: map[string]any{
						"server": map[string]any{
							"http": map[string]any{"endpoint": "https://first.example/v1/opamp"},
						},
					},
				},
			},
		},
		{
			Path: "second.yaml",
			Config: map[string]any{
				"extensions": map[string]any{
					opampSplunkExtension: map[string]any{
						"server": map[string]any{
							"http": map[string]any{
								"endpoint": "https://second.example/v1/opamp",
								"headers":  map[string]any{"X-SF-Token": "second-token"},
							},
						},
					},
				},
			},
		},
	}, map[string]string{IngestURLEnvVar: "https://ingest.example"})
	require.NoError(t, err)

	assert.Equal(t, "https://second.example/v1/opamp", server.Endpoint)
	assert.Equal(t, map[string]any{"X-SF-Token": "second-token"}, server.Headers)
}

func TestSupervisorServerFromConfigsIgnoresLaterIncompleteFragments(t *testing.T) {
	server, err := supervisorServerFromConfigs([]collectorConfigInput{
		{
			Path: "first.yaml",
			Config: map[string]any{
				"extensions": map[string]any{
					opampSplunkExtension: map[string]any{
						"server": map[string]any{
							"http": map[string]any{
								"endpoint": "https://first.example/v1/opamp",
								"headers":  map[string]any{"X-SF-Token": "first-token"},
							},
						},
					},
				},
			},
		},
		{
			Path: "second.yaml",
			Config: map[string]any{
				"extensions": map[string]any{
					opampSplunkExtension: map[string]any{
						"server": map[string]any{
							"http": map[string]any{
								"headers": map[string]any{"X-SF-Token": "second-token"},
							},
						},
					},
				},
			},
		},
	}, map[string]string{IngestURLEnvVar: "https://ingest.example"})
	require.NoError(t, err)

	assert.Equal(t, "https://first.example/v1/opamp", server.Endpoint)
	assert.Equal(t, map[string]any{"X-SF-Token": "first-token"}, server.Headers)
}

func TestConfiguredOpAMPServerErrors(t *testing.T) {
	tests := map[string]map[string]any{
		"Missing extensions": {},
		"Invalid extensions": {"extensions": "invalid"},
		"Missing splunk opamp": {
			"extensions": map[string]any{"health_check": map[string]any{}},
		},
		"Invalid splunk opamp": {
			"extensions": map[string]any{opampSplunkExtension: "invalid"},
		},
		"Missing server": {
			"extensions": map[string]any{opampSplunkExtension: map[string]any{}},
		},
		"Missing http": {
			"extensions": map[string]any{
				opampSplunkExtension: map[string]any{"server": map[string]any{}},
			},
		},
		"Missing endpoint": {
			"extensions": map[string]any{
				opampSplunkExtension: map[string]any{
					"server": map[string]any{"http": map[string]any{}},
				},
			},
		},
		"Blank endpoint": {
			"extensions": map[string]any{
				opampSplunkExtension: map[string]any{
					"server": map[string]any{"http": map[string]any{"endpoint": "  "}},
				},
			},
		},
	}

	for name, config := range tests {
		t.Run(name, func(t *testing.T) {
			server, ok := configuredOpAMPServer(config)
			assert.False(t, ok)
			assert.Empty(t, server)
		})
	}
}

func TestManagedCollectorConfigPath(t *testing.T) {
	dir := t.TempDir()
	usedNames := map[string]struct{}{}

	assert.Equal(t,
		filepath.Join(dir, "managed_collector_agent_config.yaml"),
		managedCollectorConfigPath(dir, "/etc/otel/collector/agent_config.yaml", usedNames),
	)
	assert.Equal(t,
		filepath.Join(dir, "managed_collector_agent_config_2.yaml"),
		managedCollectorConfigPath(dir, "/custom/agent_config.yaml", usedNames),
	)
	assert.Equal(t,
		filepath.Join(dir, "managed_collector_Agent_Config_3.yaml"),
		managedCollectorConfigPath(dir, "/custom/Agent_Config.yaml", usedNames),
	)
	assert.Equal(t,
		filepath.Join(dir, "managed_collector_splunk_logs_config_linux.yml"),
		managedCollectorConfigPath(dir, "/etc/otel/collector/splunk_logs_config_linux.yml", usedNames),
	)
	assert.Equal(t,
		filepath.Join(dir, "managed_collector_custom config.conf"),
		managedCollectorConfigPath(dir, "/etc/otel/collector/custom config.conf", usedNames),
	)
	assert.Equal(t,
		filepath.Join(dir, "managed_collector_no_extension.yaml"),
		managedCollectorConfigPath(dir, "/etc/otel/collector/no_extension", usedNames),
	)
	assert.Equal(t,
		filepath.Join(dir, "managed_collector_agent_config_2_2.yaml"),
		managedCollectorConfigPath(dir, "/custom/agent_config_2.yaml", usedNames),
	)
	assert.Equal(t,
		filepath.Join(dir, "managed_collector_config.yaml"),
		managedCollectorConfigPath(dir, "/custom/.yaml", usedNames),
	)
}

func TestAsMap(t *testing.T) {
	value, ok := asMap(map[any]any{"key": "value"})
	require.True(t, ok)
	assert.Equal(t, map[string]any{"key": "value"}, value)

	value, ok = asMap(map[any]any{1: "value"})
	assert.False(t, ok)
	assert.Nil(t, value)

	value, ok = asMap("invalid")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestAsSlice(t *testing.T) {
	value, ok := asSlice([]string{"health_check", opampSplunkExtension})
	require.True(t, ok)
	assert.Equal(t, []any{"health_check", opampSplunkExtension}, value)

	value, ok = asSlice("invalid")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestWriteYAMLErrors(t *testing.T) {
	tests := map[string]struct {
		setup   func(t *testing.T, dir string) (string, any)
		wantErr string
	}{
		"Parent path is file": {
			setup: func(t *testing.T, dir string) (string, any) {
				parentPath := filepath.Join(dir, "parent")
				require.NoError(t, os.WriteFile(parentPath, []byte("not a directory"), 0o600))
				return filepath.Join(parentPath, "config.yaml"), map[string]any{}
			},
			wantErr: "write",
		},
		"Parent path is missing": {
			setup: func(_ *testing.T, dir string) (string, any) {
				return filepath.Join(dir, "missing", "config.yaml"), map[string]any{}
			},
			wantErr: "write",
		},
		"Marshal failure": {
			setup: func(_ *testing.T, dir string) (string, any) {
				return filepath.Join(dir, "config.yaml"), badYAML{}
			},
			wantErr: "marshal",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			path, value := tt.setup(t, t.TempDir())
			err := writeYAML(path, value)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestWriteInitialSupervisorConfigErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "supervisor_config.yaml")
	err := writeInitialSupervisorConfig(path, SupervisorConfig{
		Server: supervisorServer{
			Endpoint: "https://example.com/v1/opamp",
			TLS:      badYAML{},
		},
		Storage: supervisorStorage{Directory: "/var/lib/otelcol/supervisor"},
		Agent: supervisorAgent{
			Description:        supervisorAgentDescription{IncludeResourceAttributes: true},
			ConfigApplyTimeout: "1m",
			PassthroughLogs:    true,
			ValidateConfig:     true,
		},
		Capabilities: supervisorCapabilities{ReportsEffectiveConfig: true},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal")
	assert.Contains(t, err.Error(), "bad yaml")
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

func testPaths(t *testing.T, dir string) Paths {
	t.Helper()

	supervisorConfig := filepath.Join(dir, "config", "supervisor_config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(supervisorConfig), 0o700))
	return Paths{
		CollectorExecutable:         "otelcol",
		SupervisorExecutable:        "opampsupervisor",
		StorageDirectory:            filepath.Join(dir, "supervisor"),
		SupervisorConfig:            supervisorConfig,
		RuntimeSupervisorConfig:     filepath.Join(filepath.Dir(supervisorConfig), "supervisor_runtime_config.yaml"),
		GeneratedCollectorConfigDir: filepath.Dir(supervisorConfig),
		DefaultAgentConfig:          filepath.Join(dir, "agent_config.yaml"),
		ConfigApplyTimeout:          "1m",
		UseHUPConfigReload:          true,
	}
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

func assertNoManagedCollectorConfigs(t *testing.T, paths Paths) {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(paths.GeneratedCollectorConfigDir, "managed_collector_*"))
	require.NoError(t, err)
	assert.Empty(t, matches)
}
