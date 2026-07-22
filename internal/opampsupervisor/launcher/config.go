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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

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
	mergeAppendFeatureGate      = "confmap.enableMergeAppendOption"
	opampSplunkExtension        = "opamp/splunk_o11y"

	managedAgentComment = "The launcher applies agent.executable, agent.config_files, and agent.args at runtime from the " +
		"installed collector path and current collector command-line arguments. Edit other supervisor settings in this file."
)

var initialSupervisorConfigTemplate = template.Must(template.New("initial-supervisor-config").Funcs(template.FuncMap{
	"indent": indentYAML,
	"yaml":   marshalYAMLFragment,
}).Parse(`server:
{{ yaml .Server | indent 2 -}}
storage:
{{ yaml .Storage | indent 2 -}}
# {{ .ManagedAgentComment }}
agent:
{{ yaml .Agent | indent 2 -}}
capabilities:
{{ yaml .Capabilities | indent 2 -}}
telemetry:
{{ yaml .Telemetry | indent 2 -}}
`))

// Paths contains package installation, supervisor config, and state paths.
type Paths struct {
	CollectorExecutable         string
	SupervisorExecutable        string
	SupervisorConfig            string
	RuntimeSupervisorConfig     string
	GeneratedCollectorConfigDir string
	StorageDirectory            string
	DefaultAgentConfig          string
	ConfigApplyTimeout          string
	UseHUPConfigReload          bool
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
	Telemetry    supervisorTelemetry    `yaml:"telemetry"`
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

type supervisorTelemetry struct {
	Logs supervisorTelemetryLogs `yaml:"logs"`
}

type supervisorTelemetryLogs struct {
	Encoding string `yaml:"encoding"`
}

type supervisorAgent struct {
	Env                map[string]string          `yaml:"env,omitempty"`
	Executable         string                     `yaml:"executable,omitempty"`
	ConfigApplyTimeout string                     `yaml:"config_apply_timeout"`
	ConfigFiles        []string                   `yaml:"config_files,omitempty"`
	Args               []string                   `yaml:"args,omitempty"`
	Description        supervisorAgentDescription `yaml:"description"`
	PassthroughLogs    bool                       `yaml:"passthrough_logs"`
	UseHUPConfigReload bool                       `yaml:"use_hup_config_reload,omitempty"`
	ValidateConfig     bool                       `yaml:"validate_config"`
}

type supervisorAgentDescription struct {
	IncludeResourceAttributes bool `yaml:"include_resource_attributes"`
}

type supervisorInputs struct {
	configFiles []string
	agentArgs   []string
}

type supervisorManagedAgentFields struct {
	Executable  string
	ConfigFiles []string
	Args        []string
}

type collectorConfigInput struct {
	Config map[string]any
	Path   string
}

// From internal/settings parseURI, which mirrors OpenTelemetry Collector confmap URI parsing.
var configURIRegexp = regexp.MustCompile(`(?s:^(?P<Scheme>[A-Za-z][A-Za-z0-9+.-]+):(?P<OpaqueValue>.*)$)`)

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

	inputs, err := supervisorInputsFromArgs(args, env)
	if err != nil {
		return Command{}, err
	}

	if prepErr := prepareSupervisor(inputs, env, paths); prepErr != nil {
		return Command{}, prepErr
	}

	return Command{
		Path: paths.SupervisorExecutable,
		Args: []string{"--config", paths.RuntimeSupervisorConfig},
		Env:  supervisorCommandEnv(environ, env, paths, inputs.configFiles),
	}, nil
}

// prepareSupervisor refreshes the launcher-managed collector config on every
// supervisor start and writes the runtime supervisor config with the current
// launcher-managed agent fields.
func prepareSupervisor(inputs supervisorInputs, env map[string]string, paths Paths) error {
	collectorConfigs, err := loadCollectorConfigFiles(inputs.configFiles)
	if err != nil {
		return err
	}

	configFiles, err := prepareManagedCollectorConfigs(collectorConfigs, paths.GeneratedCollectorConfigDir)
	if err != nil {
		return err
	}

	_, statErr := os.Stat(paths.SupervisorConfig)
	if statErr != nil {
		if !errors.Is(statErr, fs.ErrNotExist) {
			return fmt.Errorf("stat supervisor config %q: %w", paths.SupervisorConfig, statErr)
		}
		initialConfig, initErr := initialSupervisorConfig(paths, collectorConfigs, env)
		if initErr != nil {
			return initErr
		}
		if writeErr := writeInitialSupervisorConfig(paths.SupervisorConfig, initialConfig); writeErr != nil {
			return writeErr
		}
	}

	sourceConfig, err := loadSupervisorConfigFile(paths.SupervisorConfig)
	if err != nil {
		return err
	}
	managedFields := supervisorManagedAgentFields{
		Executable:  paths.CollectorExecutable,
		ConfigFiles: configFiles,
		Args:        inputs.agentArgs,
	}
	runtimeConfig := renderRuntimeConfig(sourceConfig, managedFields)
	return writeYAML(paths.RuntimeSupervisorConfig, runtimeConfig)
}

// supervisorCommandEnv preserves the collector's package default listen
// interface behavior when supervisor starts the collector from generated config.
func supervisorCommandEnv(environ []string, env map[string]string, paths Paths, collectorConfigPaths []string) []string {
	if _, ok := env[ListenInterfaceEnvVar]; ok {
		return environ
	}
	if len(collectorConfigPaths) == 0 {
		return environ
	}
	listenInterface := defaultListenInterfaceForConfigs(collectorConfigPaths, paths.DefaultAgentConfig)
	return append(environ, ListenInterfaceEnvVar+"="+listenInterface)
}

func defaultListenInterfaceForConfigs(configPaths []string, defaultAgentConfig string) string {
	for _, configPath := range configPaths {
		if filepath.Clean(configPath) == filepath.Clean(defaultAgentConfig) {
			return defaultAgentListenInterface
		}
	}
	return defaultListenInterface
}

// supervisorInputsFromArgs processes configs from flags/env values
// and removes the merge append feature gate when in supervisor mode.
func supervisorInputsFromArgs(args []string, env map[string]string) (supervisorInputs, error) {
	inputs := supervisorInputs{agentArgs: make([]string, 0, len(args))}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--config":
			if i+1 >= len(args) {
				return supervisorInputs{}, errors.New("missing value for --config")
			}
			path, err := normalizeLocalConfigPath(args[i+1])
			if err != nil {
				return supervisorInputs{}, err
			}
			inputs.configFiles = append(inputs.configFiles, path)
			i++
		case "--feature-gates":
			if i+1 >= len(args) {
				inputs.agentArgs = append(inputs.agentArgs, arg)
				continue
			}
			if filtered, ok := filterFeatureGateValue(args[i+1]); ok {
				inputs.agentArgs = append(inputs.agentArgs, arg, filtered)
			}
			i++
		default:
			if value, ok := strings.CutPrefix(arg, "--config="); ok {
				path, err := normalizeLocalConfigPath(value)
				if err != nil {
					return supervisorInputs{}, err
				}
				inputs.configFiles = append(inputs.configFiles, path)
				continue
			}
			if value, ok := strings.CutPrefix(arg, "--feature-gates="); ok {
				filtered, ok := filterFeatureGateValue(value)
				if ok {
					inputs.agentArgs = append(inputs.agentArgs, "--feature-gates="+filtered)
				}
				continue
			}
			inputs.agentArgs = append(inputs.agentArgs, arg)
		}
	}

	if len(inputs.configFiles) == 0 {
		collectorConfigPath := strings.TrimSpace(env[CollectorConfigEnvVar])
		if collectorConfigPath == "" {
			return supervisorInputs{},
				fmt.Errorf("supervisor mode requires at least one --config flag or %s", CollectorConfigEnvVar)
		}
		path, err := normalizeLocalConfigPath(collectorConfigPath)
		if err != nil {
			return supervisorInputs{}, err
		}
		inputs.configFiles = append(inputs.configFiles, path)
	}

	return inputs, nil
}

// normalizeLocalConfigPath accepts plain local paths and "file:" URIs, returning
// the filesystem path that the launcher can read and pass to supervisor.
func normalizeLocalConfigPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("collector config path is empty")
	}

	scheme, location, isURI := parseConfigURI(value)
	if !isURI || isWindowsDrivePath(scheme) {
		return value, nil
	}
	if scheme != "file" {
		return "",
			fmt.Errorf("unsupported config provider URI %q in supervisor mode; only local file paths and file: URIs are supported", value)
	}

	path := strings.TrimSpace(location)
	if path == "" {
		return "", errors.New("collector config path is empty")
	}
	return filepath.Clean(path), nil
}

func parseConfigURI(value string) (scheme, location string, isURI bool) {
	submatches := configURIRegexp.FindStringSubmatch(value)
	if len(submatches) != 3 {
		return "", "", false
	}
	return submatches[1], submatches[2], true
}

func isWindowsDrivePath(scheme string) bool {
	return len(scheme) == 1 && ((scheme[0] >= 'A' && scheme[0] <= 'Z') || (scheme[0] >= 'a' && scheme[0] <= 'z'))
}

func filterFeatureGateValue(value string) (string, bool) {
	gates := strings.Split(value, ",")
	filtered := make([]string, 0, len(gates))
	for _, gate := range gates {
		gate = strings.TrimSpace(gate)
		if gate == "" {
			continue
		}
		gateName := strings.TrimPrefix(strings.TrimPrefix(gate, "+"), "-")
		if gateName == mergeAppendFeatureGate {
			continue
		}
		filtered = append(filtered, gate)
	}
	return strings.Join(filtered, ","), len(filtered) > 0
}

// sanitizeCollectorConfig removes the splunk_o11y opamp extension definitions and service references
func sanitizeCollectorConfig(config map[string]any) (map[string]any, bool) {
	out := cloneYAMLMap(config)
	removedExtension := removeSplunkOpAMPFromExtensions(out)
	removedServiceExtension := removeSplunkOpAMPFromServiceExtensions(out)
	return out, removedExtension || removedServiceExtension
}

func removeSplunkOpAMPFromExtensions(config map[string]any) bool {
	extensions, ok := asMap(config["extensions"])
	if !ok {
		return false
	}
	if _, ok := extensions[opampSplunkExtension]; !ok {
		return false
	}
	delete(extensions, opampSplunkExtension)
	config["extensions"] = extensions
	return true
}

func removeSplunkOpAMPFromServiceExtensions(config map[string]any) bool {
	service, ok := asMap(config["service"])
	if !ok {
		return false
	}
	serviceExtensions, ok := asSlice(service["extensions"])
	if !ok {
		return false
	}

	filtered := make([]any, 0, len(serviceExtensions))
	changed := false
	for _, extension := range serviceExtensions {
		if name, ok := extension.(string); ok && name == opampSplunkExtension {
			changed = true
			continue
		}
		filtered = append(filtered, extension)
	}
	if !changed {
		return false
	}
	service["extensions"] = filtered
	config["service"] = service
	return true
}

// loadCollectorConfigFiles reads in collector configs so OpAMP server settings
// can be copied into the generated supervisor config when present.
func loadCollectorConfigFiles(paths []string) ([]collectorConfigInput, error) {
	configs := make([]collectorConfigInput, 0, len(paths))
	for _, path := range paths {
		config, err := loadCollectorConfigFile(path)
		if err != nil {
			return nil, err
		}
		configs = append(configs, collectorConfigInput{Path: path, Config: config})
	}
	return configs, nil
}

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

// prepareManagedCollectorConfigs passes unchanged configs through and writes
// managed copies only for configs that needed "opamp/splunk_o11y" extension removal.
func prepareManagedCollectorConfigs(configs []collectorConfigInput, managedDir string) ([]string, error) {
	configFiles := make([]string, 0, len(configs))
	usedNames := map[string]struct{}{}
	for _, input := range configs {
		sanitized, changed := sanitizeCollectorConfig(input.Config)
		if !changed {
			configFiles = append(configFiles, input.Path)
			continue
		}

		managedPath := managedCollectorConfigPath(managedDir, input.Path, usedNames)
		if err := writeYAML(managedPath, sanitized); err != nil {
			return nil, err
		}
		configFiles = append(configFiles, managedPath)
	}
	return configFiles, nil
}

// managedCollectorConfigPath creates managed filenames and de-dupes
// names using case-insensitive keys for Windows-compatible behavior.
func managedCollectorConfigPath(dir, sourcePath string, usedNames map[string]struct{}) string {
	ext := filepath.Ext(sourcePath)
	if ext == "" {
		ext = ".yaml"
	}
	base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	if base == "" {
		base = "config"
	}
	name := "managed_collector_" + base
	for suffix := 1; ; suffix++ {
		candidate := name + ext
		if suffix > 1 {
			candidate = fmt.Sprintf("%s_%d%s", name, suffix, ext)
		}
		key := strings.ToLower(candidate)
		if _, ok := usedNames[key]; !ok {
			usedNames[key] = struct{}{}
			return filepath.Join(dir, candidate)
		}
	}
}

func supervisorServerFromConfigs(configs []collectorConfigInput, env map[string]string) (supervisorServer, error) {
	for i := len(configs) - 1; i >= 0; i-- {
		if server, ok := configuredOpAMPServer(configs[i].Config); ok {
			return server, nil
		}
	}

	return supervisorServerFromEnv(env)
}

func supervisorServerFromEnv(env map[string]string) (supervisorServer, error) {
	endpoint := derivedOpAMPEndpoint(env)
	if endpoint == "" {
		return supervisorServer{},
			fmt.Errorf("could not derive OpAMP endpoint: neither %s nor %s is set", IngestURLEnvVar, GatewayURLEnvVar)
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

func initialSupervisorConfig(paths Paths, collectorConfigs []collectorConfigInput, env map[string]string) (SupervisorConfig, error) {
	server, err := supervisorServerFromConfigs(collectorConfigs, env)
	if err != nil {
		return SupervisorConfig{}, err
	}
	return SupervisorConfig{
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
		Telemetry: supervisorTelemetry{
			Logs: supervisorTelemetryLogs{Encoding: "console"},
		},
		Agent: supervisorAgent{
			Description:        supervisorAgentDescription{IncludeResourceAttributes: true},
			ConfigApplyTimeout: paths.ConfigApplyTimeout,
			PassthroughLogs:    true,
			UseHUPConfigReload: paths.UseHUPConfigReload,
			ValidateConfig:     true,
		},
	}, nil
}

func writeInitialSupervisorConfig(path string, config SupervisorConfig) error {
	var buf bytes.Buffer
	data := map[string]any{
		"Server":              config.Server,
		"Storage":             config.Storage,
		"Agent":               config.Agent,
		"Capabilities":        config.Capabilities,
		"Telemetry":           config.Telemetry,
		"ManagedAgentComment": managedAgentComment,
	}
	if err := initialSupervisorConfigTemplate.Execute(&buf, data); err != nil {
		return fmt.Errorf("marshal %q: %w", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}

func loadSupervisorConfigFile(path string) (map[string]any, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read supervisor config %q: %w", path, err)
	}
	var decoded any
	if err := yaml.Unmarshal(bytes, &decoded); err != nil {
		return nil, fmt.Errorf("parse supervisor config %q: %w", path, err)
	}
	config, ok := asMap(decoded)
	if !ok {
		return nil, fmt.Errorf("supervisor config %q must be a YAML mapping", path)
	}
	if _, ok := asMap(config["agent"]); !ok {
		return nil, fmt.Errorf("supervisor config %q must contain agent mapping", path)
	}
	return config, nil
}

func renderRuntimeConfig(sourceConfig map[string]any, managedFields supervisorManagedAgentFields) map[string]any {
	out := cloneYAMLMap(sourceConfig)
	agent, _ := asMap(out["agent"])
	agent["executable"] = managedFields.Executable
	agent["config_files"] = slices.Clone(managedFields.ConfigFiles)
	if len(managedFields.Args) == 0 {
		delete(agent, "args")
	} else {
		agent["args"] = slices.Clone(managedFields.Args)
	}
	out["agent"] = agent
	return out
}

func marshalYAMLFragment(value any) (string, error) {
	bytes, err := yaml.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func indentYAML(spaces int, value string) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(strings.TrimRight(value, "\n"), "\n")
	for i := range lines {
		if lines[i] != "" {
			lines[i] = prefix + lines[i]
		}
	}
	return strings.Join(lines, "\n") + "\n"
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
