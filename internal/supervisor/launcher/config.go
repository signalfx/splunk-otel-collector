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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	SupervisorEnabledEnvVar = "SPLUNK_OTEL_SUPERVISOR_ENABLED"

	CollectorConfigEnvVar = "SPLUNK_CONFIG"
	GatewayURLEnvVar      = "SPLUNK_GATEWAY_URL"
	IngestURLEnvVar       = "SPLUNK_INGEST_URL"

	opampSplunkExtension = "opamp/splunk_o11y"
	opampFeatureGate     = "splunk.opamp.enabled"
)

// Paths contains package installation paths and supervisor state paths.
type Paths struct {
	CollectorExecutable  string
	SupervisorExecutable string
	SupervisorConfig     string
	StorageDirectory     string
	UseHUPConfigReload   bool
}

// Command is the executable, arguments, and environment the launcher should run.
type Command struct {
	Path string
	Args []string
	Env  []string
}

// SupervisorConfig is the subset of the upstream opampsupervisor config this
// launcher renders for package-managed supervisor mode.
type SupervisorConfig struct {
	Server       supervisorServer       `yaml:"server"`
	Storage      supervisorStorage      `yaml:"storage"`
	Agent        supervisorAgent        `yaml:"agent"`
	Capabilities supervisorCapabilities `yaml:"capabilities"`
}

type supervisorServer struct {
	TLS      any            `yaml:"tls,omitempty"`
	Headers  map[string]any `yaml:"headers,omitempty"`
	Endpoint string         `yaml:"endpoint"`
}

type supervisorCapabilities struct {
	ReportsEffectiveConfig     bool `yaml:"reports_effective_config"`
	ReportsHealth              bool `yaml:"reports_health"`
	ReportsAvailableComponents bool `yaml:"reports_available_components"`
	AcceptsRemoteConfig        bool `yaml:"accepts_remote_config"`
	ReportsRemoteConfig        bool `yaml:"reports_remote_config"`
	ReportsOwnMetrics          bool `yaml:"reports_own_metrics"`
	ReportsHeartbeat           bool `yaml:"reports_heartbeat"`
}

type supervisorStorage struct {
	Directory string `yaml:"directory"`
}

type supervisorAgent struct {
	Env                map[string]string `yaml:"env,omitempty"`
	Executable         string            `yaml:"executable"`
	ConfigFiles        []string          `yaml:"config_files"`
	Args               []string          `yaml:"args,omitempty"`
	PassthroughLogs    bool              `yaml:"passthrough_logs"`
	UseHUPConfigReload bool              `yaml:"use_hup_config_reload,omitempty"`
	ValidateConfig     bool              `yaml:"validate_config"`
}

// SupervisorEnabled reports whether the persisted service-scoped supervisor
// switch is enabled in the provided environment.
func SupervisorEnabled(env map[string]string) bool {
	return strings.EqualFold(strings.TrimSpace(env[SupervisorEnabledEnvVar]), "true")
}

// PrepareCommand builds the command the launcher should run. In direct mode this
// is the collector, while supervisor mode first renders supervisor config and
// validates the local collector config before returning the opampsupervisor
// command.
func PrepareCommand(args, environ []string, paths Paths) (Command, error) {
	env := environToMap(environ)
	if !SupervisorEnabled(env) {
		return Command{
			Path: paths.CollectorExecutable,
			Args: args,
			Env:  environ,
		}, nil
	}

	agentArgs, configFiles, err := PrepareSupervisor(args, env, paths)
	if err != nil {
		return Command{}, err
	}

	if err := validateCollectorConfig(paths.CollectorExecutable, agentArgs, configFiles, environ); err != nil {
		return Command{}, err
	}

	return Command{
		Path: paths.SupervisorExecutable,
		Args: []string{"--config", paths.SupervisorConfig},
		Env:  environ,
	}, nil
}

// PrepareSupervisor renders supervisor config from the package-managed
// SPLUNK_CONFIG path and current collector args.
func PrepareSupervisor(args []string, env map[string]string, paths Paths) ([]string, []string, error) {
	agentArgs := filterSupervisorAgentArgs(args)

	configPath := strings.TrimSpace(env[CollectorConfigEnvVar])
	if configPath == "" {
		return nil, nil, fmt.Errorf("%s must be set in supervisor mode", CollectorConfigEnvVar)
	}

	config, err := loadConfigFile(configPath)
	if err != nil {
		return nil, nil, err
	}

	server, err := supervisorServerFromConfig(config, env)
	if err != nil {
		return nil, nil, err
	}

	if err := prepareSupervisorStorageDir(paths); err != nil {
		return nil, nil, err
	}

	supervisorConfig := SupervisorConfig{
		Server: server,
		Capabilities: supervisorCapabilities{
			ReportsEffectiveConfig:     true,
			ReportsHealth:              true,
			ReportsRemoteConfig:        true,
			ReportsAvailableComponents: true,
			AcceptsRemoteConfig:        true,
			ReportsOwnMetrics:          false,
			ReportsHeartbeat:           false,
		},
		Storage: supervisorStorage{Directory: paths.StorageDirectory},
		Agent: supervisorAgent{
			Executable:         paths.CollectorExecutable,
			ConfigFiles:        []string{configPath},
			Args:               agentArgs,
			PassthroughLogs:    true,
			UseHUPConfigReload: paths.UseHUPConfigReload,
			ValidateConfig:     true,
		},
	}
	if err := writeYAML(paths.SupervisorConfig, supervisorConfig, 0o600); err != nil {
		return nil, nil, err
	}

	return agentArgs, []string{configPath}, nil
}

// loadConfigFile reads a collector config into a generic map so direct OpAMP
// settings can be copied into the generated supervisor config when present.
func loadConfigFile(path string) (map[string]any, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read collector config %q: %w", path, err)
	}
	var config map[string]any
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("parse collector config %q: %w", path, err)
	}
	if config == nil {
		config = map[string]any{}
	}
	return config, nil
}

// supervisorServerFromConfig chooses the backend OpAMP endpoint. A direct
// OpAMP endpoint already present in SPLUNK_CONFIG wins over package fallback
// values.
func supervisorServerFromConfig(config map[string]any, env map[string]string) (supervisorServer, error) {
	if server, ok := configuredOpAMPServer(config); ok {
		return server, nil
	}

	endpoint := derivedOpAMPEndpoint(env)
	if endpoint == "" {
		return supervisorServer{}, fmt.Errorf("could not derive OpAMP endpoint: neither %s nor %s is set", IngestURLEnvVar, GatewayURLEnvVar)
	}
	return supervisorServer{
		Endpoint: endpoint,
		Headers: map[string]any{
			"X-SF-Token": "${SPLUNK_ACCESS_TOKEN}",
		},
	}, nil
}

// configuredOpAMPServer extracts compatible server settings from the direct
// Splunk OpAMP extension block, if that block exists.
func configuredOpAMPServer(config map[string]any) (supervisorServer, bool) {
	extensions, ok := asMap(config["extensions"])
	if !ok {
		return supervisorServer{}, false
	}
	opamp, ok := asMap(extensions[opampSplunkExtension])
	if !ok {
		return supervisorServer{}, false
	}
	server, ok := asMap(opamp["server"])
	if !ok {
		return supervisorServer{}, false
	}
	httpServer, ok := asMap(server["http"])
	if !ok {
		return supervisorServer{}, false
	}
	endpoint, ok := httpServer["endpoint"].(string)
	if !ok || strings.TrimSpace(endpoint) == "" {
		return supervisorServer{}, false
	}

	out := supervisorServer{Endpoint: endpoint}
	if headers, ok := asMap(httpServer["headers"]); ok {
		out.Headers = headers
	}
	if tls, ok := httpServer["tls"]; ok {
		out.TLS = tls
	}
	return out, true
}

// derivedOpAMPEndpoint builds a package-default OpAMP endpoint when direct
// OpAMP config is not already present.
func derivedOpAMPEndpoint(env map[string]string) string {
	if strings.TrimSpace(env[GatewayURLEnvVar]) != "" {
		return "${" + GatewayURLEnvVar + "}:4320/v1/opamp"
	}
	if strings.TrimSpace(env[IngestURLEnvVar]) != "" {
		return "${" + IngestURLEnvVar + "}/v1/opamp"
	}
	return ""
}

// prepareSupervisorStorageDir creates the launcher-owned supervisor state
// directory before supervisor persistence files are used.
func prepareSupervisorStorageDir(paths Paths) error {
	if err := os.MkdirAll(paths.StorageDirectory, 0o700); err != nil {
		return fmt.Errorf("create supervisor storage directory %q: %w", paths.StorageDirectory, err)
	}
	return nil
}

// filterSupervisorAgentArgs removes only the direct Splunk OpAMP feature gate
// and preserves other collector args for the child agent.
func filterSupervisorAgentArgs(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--feature-gates" {
			if i+1 < len(args) {
				filtered, keep := filterFeatureGateValue(args[i+1])
				if keep {
					out = append(out, arg, filtered)
				}
				i++
			} else {
				out = append(out, arg)
			}
			continue
		}
		if strings.HasPrefix(arg, "--feature-gates=") {
			filtered, keep := filterFeatureGateValue(strings.TrimPrefix(arg, "--feature-gates="))
			if keep {
				out = append(out, "--feature-gates="+filtered)
			}
			continue
		}

		out = append(out, arg)
	}
	return out
}

// filterFeatureGateValue drops only the direct Splunk OpAMP feature gate so
// other collector feature gates continue to flow to the child collector.
func filterFeatureGateValue(value string) (string, bool) {
	var kept []string
	for _, item := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		gate := strings.TrimLeft(trimmed, "+-")
		if gate == opampFeatureGate {
			continue
		}
		kept = append(kept, trimmed)
	}
	return strings.Join(kept, ","), len(kept) > 0
}

// writeYAML writes generated config files with their parent directories created
// for package-managed state paths.
func writeYAML(path string, value any, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %q: %w", path, err)
	}
	bytes, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal %q: %w", path, err)
	}
	if err := os.WriteFile(path, bytes, perm); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}

// asMap normalizes YAML-decoded map shapes used by different decoders and test
// inputs.
func asMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, val := range typed {
			keyString, ok := key.(string)
			if !ok {
				return nil, false
			}
			out[keyString] = val
		}
		return out, true
	default:
		return nil, false
	}
}

// environToMap converts an os.Environ-style slice into lookup form.
func environToMap(environ []string) map[string]string {
	out := make(map[string]string, len(environ))
	for _, entry := range environ {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		out[name] = value
	}
	return out
}
