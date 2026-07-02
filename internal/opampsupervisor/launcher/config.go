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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	SupervisorEnabledEnvVar = "SPLUNK_OPAMP_SUPERVISOR_ENABLED"

	CollectorConfigEnvVar = "SPLUNK_CONFIG"
	GatewayURLEnvVar      = "SPLUNK_GATEWAY_URL"
	IngestURLEnvVar       = "SPLUNK_INGEST_URL"
	ListenInterfaceEnvVar = "SPLUNK_LISTEN_INTERFACE"

	defaultAgentListenInterface = "127.0.0.1"
	defaultListenInterface      = "0.0.0.0"
	opampSplunkExtension        = "opamp/splunk_o11y"
)

// Paths contains package installation, supervisor config, and state paths.
type Paths struct {
	CollectorExecutable      string
	SupervisorExecutable     string
	SupervisorConfig         string
	GeneratedCollectorConfig string
	StorageDirectory         string
	DefaultAgentConfig       string
	ConfigApplyTimeout       string
	UseHUPConfigReload       bool
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
	Env                map[string]string          `yaml:"env,omitempty"`
	Executable         string                     `yaml:"executable"`
	ConfigApplyTimeout string                     `yaml:"config_apply_timeout"`
	ConfigFiles        []string                   `yaml:"config_files"`
	Args               []string                   `yaml:"args,omitempty"`
	Description        supervisorAgentDescription `yaml:"description"`
	PassthroughLogs    bool                       `yaml:"passthrough_logs"`
	UseHUPConfigReload bool                       `yaml:"use_hup_config_reload,omitempty"`
	ValidateConfig     bool                       `yaml:"validate_config"`
}

type supervisorAgentDescription struct {
	IncludeResourceAttributes bool `yaml:"include_resource_attributes"`
}

// SupervisorEnabled reports whether the persisted service-scoped supervisor
// switch is enabled in the provided environment.
func SupervisorEnabled(env map[string]string) bool {
	return strings.EqualFold(strings.TrimSpace(env[SupervisorEnabledEnvVar]), "true")
}

// PrepareCommand builds the command the launcher should run. In direct mode this
// is the collector, while supervisor mode ensures supervisor config exists
// before returning the opampsupervisor command.
func PrepareCommand(args, environ []string, paths Paths) (Command, error) {
	env := environToMap(environ)
	if !SupervisorEnabled(env) {
		return Command{
			Path: paths.CollectorExecutable,
			Args: args,
			Env:  environ,
		}, nil
	}

	if err := PrepareSupervisor(args, env, paths); err != nil {
		return Command{}, err
	}

	return Command{
		Path: paths.SupervisorExecutable,
		Args: []string{"--config", paths.SupervisorConfig},
		Env:  supervisorCommandEnv(environ, env, paths),
	}, nil
}

// PrepareSupervisor refreshes the launcher-managed collector config on every
// supervisor start and writes the initial supervisor config only when it does
// not already exist. Existing supervisor config is user-editable and preserved
// across launcher restarts.
func PrepareSupervisor(args []string, env map[string]string, paths Paths) error {
	collectorConfigPath := strings.TrimSpace(env[CollectorConfigEnvVar])
	if collectorConfigPath == "" {
		return fmt.Errorf("%s must be set in supervisor mode", CollectorConfigEnvVar)
	}

	collectorConfig, err := loadCollectorConfigFile(collectorConfigPath)
	if err != nil {
		return err
	}

	if writeErr := writeYAML(paths.GeneratedCollectorConfig, sanitizedCollectorConfig(collectorConfig)); writeErr != nil {
		return writeErr
	}

	_, err = os.Stat(paths.SupervisorConfig)
	if err == nil {
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat supervisor config %q: %w", paths.SupervisorConfig, err)
	}

	server, err := supervisorServerFromConfig(collectorConfig, env)
	if err != nil {
		return err
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
			Description:        supervisorAgentDescription{IncludeResourceAttributes: true},
			ConfigApplyTimeout: paths.ConfigApplyTimeout,
			ConfigFiles:        []string{paths.GeneratedCollectorConfig},
			Args:               args,
			PassthroughLogs:    true,
			UseHUPConfigReload: paths.UseHUPConfigReload,
			ValidateConfig:     true,
		},
	}
	if writeErr := writeYAML(paths.SupervisorConfig, supervisorConfig); writeErr != nil {
		return writeErr
	}

	return nil
}

// supervisorCommandEnv preserves the collector's package default listen
// interface behavior when supervisor starts the collector from generated config.
func supervisorCommandEnv(environ []string, env map[string]string, paths Paths) []string {
	if _, ok := env[ListenInterfaceEnvVar]; ok {
		return environ
	}
	collectorConfigPath := strings.TrimSpace(env[CollectorConfigEnvVar])
	if collectorConfigPath == "" {
		return environ
	}
	listenInterface := defaultListenInterfaceForConfig(collectorConfigPath, paths.DefaultAgentConfig)
	return append(environ, ListenInterfaceEnvVar+"="+listenInterface)
}

func defaultListenInterfaceForConfig(configPath, defaultAgentConfig string) string {
	if filepath.Clean(configPath) == filepath.Clean(defaultAgentConfig) {
		return defaultAgentListenInterface
	}
	return defaultListenInterface
}

// sanitizedCollectorConfig returns the collector config used as supervisor's
// local base config, with only the direct Splunk OpAMP extension removed.
func sanitizedCollectorConfig(config map[string]any) map[string]any {
	out := cloneYAMLMap(config)
	removeSplunkOpAMPFromExtensions(out)
	removeSplunkOpAMPFromServiceExtensions(out)
	return out
}

func removeSplunkOpAMPFromExtensions(config map[string]any) {
	extensions, ok := asMap(config["extensions"])
	if !ok {
		return
	}
	delete(extensions, opampSplunkExtension)
	config["extensions"] = extensions
}

func removeSplunkOpAMPFromServiceExtensions(config map[string]any) {
	service, ok := asMap(config["service"])
	if !ok {
		return
	}
	serviceExtensions, ok := asSlice(service["extensions"])
	if !ok {
		return
	}

	filtered := make([]any, 0, len(serviceExtensions))
	for _, extension := range serviceExtensions {
		if name, ok := extension.(string); ok && name == opampSplunkExtension {
			continue
		}
		filtered = append(filtered, extension)
	}
	service["extensions"] = filtered
	config["service"] = service
}

// loadCollectorConfigFile reads a collector config into a generic map so direct
// OpAMP settings can be copied into the generated supervisor config when
// present.
func loadCollectorConfigFile(path string) (map[string]any, error) {
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

// writeYAML writes generated config files to package-managed config paths.
func writeYAML(path string, value any) error {
	bytes, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal %q: %w", path, err)
	}
	if err := os.WriteFile(path, bytes, 0o600); err != nil {
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

func asSlice(value any) ([]any, bool) {
	switch typed := value.(type) {
	case []any:
		return typed, true
	case []string:
		out := make([]any, len(typed))
		for i, value := range typed {
			out[i] = value
		}
		return out, true
	default:
		return nil, false
	}
}

func cloneYAMLMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = cloneYAMLValue(value)
	}
	return out
}

func cloneYAMLValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneYAMLMap(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cloneYAMLValue(item)
		}
		return out
	default:
		return value
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
