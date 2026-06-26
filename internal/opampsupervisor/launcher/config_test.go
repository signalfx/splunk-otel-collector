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
	"errors"
	"os"
	"path/filepath"
	"testing"

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
	assert.Equal(t, []string{"--config", paths.SupervisorConfig}, cmd.Args)
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

func TestPrepareCommandSupervisorModeErrors(t *testing.T) {
	_, err := PrepareCommand(nil, []string{SupervisorEnabledEnvVar + "=true"}, testPaths(t, t.TempDir()))
	require.Error(t, err)
	assert.Contains(t, err.Error(), CollectorConfigEnvVar)
}

func TestPrepareSupervisorUsesConfiguredOpAMPEndpointAndOriginalCollectorConfig(t *testing.T) {
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
	err := PrepareSupervisor(
		[]string{
			"--feature-gates=+splunk.opamp.enabled,+other.gate",
			"--set=processors.batch.timeout=2s",
		},
		map[string]string{CollectorConfigEnvVar: configPath, IngestURLEnvVar: "https://ingest.example"},
		paths,
	)
	require.NoError(t, err)

	var supervisorConfig SupervisorConfig
	readYAML(t, paths.SupervisorConfig, &supervisorConfig)
	supervisorYAML := readFile(t, paths.SupervisorConfig)

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
	assert.Equal(t, []string{paths.GeneratedCollectorConfig}, supervisorConfig.Agent.ConfigFiles)
	assert.Equal(t, []string{
		"--feature-gates=+splunk.opamp.enabled,+other.gate",
		"--set=processors.batch.timeout=2s",
	}, supervisorConfig.Agent.Args)
	assert.Contains(t, supervisorYAML, paths.GeneratedCollectorConfig)
	assert.NotContains(t, supervisorYAML, configPath)
	assert.Nil(t, supervisorConfig.Agent.Env)
	assert.NotContains(t, supervisorYAML, "\nenv:")
	assert.True(t, supervisorConfig.Agent.PassthroughLogs)
	assert.True(t, supervisorConfig.Agent.UseHUPConfigReload)
	assert.True(t, supervisorConfig.Agent.ValidateConfig)

	var generatedConfig map[string]any
	readYAML(t, paths.GeneratedCollectorConfig, &generatedConfig)
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

func TestPrepareCommandPreservesExistingConfigAndRecomputesEnv(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(t, dir)
	require.NoError(t, os.WriteFile(paths.SupervisorConfig, []byte("user: edited\n"), 0o600))
	require.NoError(t, os.WriteFile(paths.GeneratedCollectorConfig, []byte("old: generated\n"), 0o600))

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
	cmd, err := PrepareCommand([]string{"--feature-gates=+splunk.opamp.enabled"}, env, paths)
	require.NoError(t, err)

	assert.Equal(t, "user: edited\n", readFile(t, paths.SupervisorConfig))
	assert.NotContains(t, readFile(t, paths.GeneratedCollectorConfig), opampSplunkExtension)
	assert.Contains(t, readFile(t, paths.GeneratedCollectorConfig), "health_check")
	assert.Equal(t, append(append([]string{}, env...), ListenInterfaceEnvVar+"="+defaultListenInterface), cmd.Env)
}

func TestPrepareSupervisorReturnsErrors(t *testing.T) {
	tests := map[string]struct {
		setup   func(t *testing.T, dir string, paths Paths) map[string]string
		wantErr string
	}{
		"Missing config path": {
			setup: func(_ *testing.T, _ string, _ Paths) map[string]string {
				return map[string]string{IngestURLEnvVar: "https://ingest.example"}
			},
			wantErr: CollectorConfigEnvVar,
		},
		"Collector config read": {
			setup: func(_ *testing.T, dir string, _ Paths) map[string]string {
				return map[string]string{
					CollectorConfigEnvVar: filepath.Join(dir, "missing.yaml"),
					IngestURLEnvVar:       "https://ingest.example",
				}
			},
			wantErr: "read collector config",
		},
		"Collector config parse": {
			setup: func(t *testing.T, dir string, _ Paths) map[string]string {
				configPath := filepath.Join(dir, "collector.yaml")
				require.NoError(t, os.WriteFile(configPath, []byte("extensions: ["), 0o600))
				return map[string]string{
					CollectorConfigEnvVar: configPath,
					IngestURLEnvVar:       "https://ingest.example",
				}
			},
			wantErr: "parse collector config",
		},
		"Missing endpoint inputs": {
			setup: func(t *testing.T, dir string, _ Paths) map[string]string {
				configPath := filepath.Join(dir, "collector.yaml")
				require.NoError(t, os.WriteFile(configPath, []byte(`service: {}`), 0o600))
				return map[string]string{CollectorConfigEnvVar: configPath}
			},
			wantErr: "could not derive OpAMP endpoint",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			paths := testPaths(t, dir)
			err := PrepareSupervisor(nil, tt.setup(t, dir, paths), paths)
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
		configPath string
		env        []string
		want       []string
	}{
		"linux agent config": {
			configPath: defaultAgentConfig,
			want: []string{
				CollectorConfigEnvVar + "=" + defaultAgentConfig,
				ListenInterfaceEnvVar + "=" + defaultAgentListenInterface,
			},
		},
		"linux agent config cleaned path": {
			configPath: "/etc/otel/collector/./agent_config.yaml",
			want: []string{
				CollectorConfigEnvVar + "=/etc/otel/collector/./agent_config.yaml",
				ListenInterfaceEnvVar + "=" + defaultAgentListenInterface,
			},
		},
		"linux gateway config": {
			configPath: "/etc/otel/collector/gateway_config.yaml",
			want: []string{
				CollectorConfigEnvVar + "=/etc/otel/collector/gateway_config.yaml",
				ListenInterfaceEnvVar + "=" + defaultListenInterface,
			},
		},
		"custom config": {
			configPath: "/etc/otel/collector/custom.yaml",
			want: []string{
				CollectorConfigEnvVar + "=/etc/otel/collector/custom.yaml",
				ListenInterfaceEnvVar + "=" + defaultListenInterface,
			},
		},
		"existing listen interface": {
			configPath: defaultAgentConfig,
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
			configPath: defaultAgentConfig,
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
				environ = []string{CollectorConfigEnvVar + "=" + tt.configPath}
			}
			assert.Equal(t, tt.want, supervisorCommandEnv(environ, environToMap(environ), Paths{
				DefaultAgentConfig: defaultAgentConfig,
			}))
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
			server, err := supervisorServerFromConfig(map[string]any{}, tt.env)
			require.NoError(t, err)

			assert.Equal(t, tt.wantEndpoint, server.Endpoint)
			assert.Equal(t, map[string]any{"X-SF-Token": "${SPLUNK_ACCESS_TOKEN}"}, server.Headers)
		})
	}
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
		CollectorExecutable:      "otelcol",
		SupervisorExecutable:     "opampsupervisor",
		StorageDirectory:         filepath.Join(dir, "supervisor"),
		SupervisorConfig:         supervisorConfig,
		GeneratedCollectorConfig: filepath.Join(filepath.Dir(supervisorConfig), "collector_config.yaml"),
		DefaultAgentConfig:       filepath.Join(dir, "agent_config.yaml"),
		ConfigApplyTimeout:       "1m",
		UseHUPConfigReload:       true,
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
