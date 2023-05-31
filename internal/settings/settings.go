// Copyright  Splunk, Inc.
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

package settings

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"

	"github.com/signalfx/splunk-otel-collector/internal/configconverter"
	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/discovery"
)

const (
	APIURLEnvVar              = "SPLUNK_API_URL"
	BallastEnvVar             = "SPLUNK_BALLAST_SIZE_MIB"
	ConfigEnvVar              = "SPLUNK_CONFIG"
	ConfigDirEnvVar           = "SPLUNK_CONFIG_DIR"
	ConfigServerEnabledEnvVar = "SPLUNK_DEBUG_CONFIG_SERVER"
	ConfigYamlEnvVar          = "SPLUNK_CONFIG_YAML"
	HecLogIngestURLEnvVar     = "SPLUNK_HEC_URL"
	// nolint:gosec
	HecTokenEnvVar    = "SPLUNK_HEC_TOKEN" // this isn't a hardcoded token
	IngestURLEnvVar   = "SPLUNK_INGEST_URL"
	MemLimitMiBEnvVar = "SPLUNK_MEMORY_LIMIT_MIB"
	MemTotalEnvVar    = "SPLUNK_MEMORY_TOTAL_MIB"
	RealmEnvVar       = "SPLUNK_REALM"
	// nolint:gosec
	TokenEnvVar          = "SPLUNK_ACCESS_TOKEN" // this isn't a hardcoded token
	TraceIngestURLEnvVar = "SPLUNK_TRACE_URL"

	DefaultGatewayConfig           = "/etc/otel/collector/gateway_config.yaml"
	DefaultOTLPLinuxConfig         = "/etc/otel/collector/otlp_config_linux.yaml"
	DefaultConfigDir               = "/etc/otel/collector/config.d"
	DefaultMemoryBallastPercentage = 33
	DefaultMemoryLimitPercentage   = 90
	DefaultMemoryTotalMiB          = 512
)

var (
	envProvider  = envprovider.New()
	fileProvider = fileprovider.New()
)

type Settings struct {
	discovery           discovery.Provider
	configPaths         *stringArrayFlagValue
	setOptionArguments  *stringArrayFlagValue
	configDir           *stringPointerFlagValue
	confMapProviders    map[string]confmap.Provider
	discoveryProperties []string
	setProperties       []string
	colCoreArgs         []string
	supportedURISchemes []string
	versionFlag         bool
	noConvertConfig     bool
	configD             bool
	discoveryMode       bool
	dryRun              bool
}

func New(args []string) (*Settings, error) {
	s, err := parseArgs(args)
	if err != nil {
		return nil, err
	}

	// immediate exit paths, no further setup required
	if s.versionFlag {
		return s, nil
	}

	if err = checkRuntimeParams(s); err != nil {
		return nil, err
	}

	if err = setDefaultEnvVars(); err != nil {
		return nil, err
	}

	return s, nil
}

// ResolverURIs returns config provider resolver URIs for the core collector service.
func (s *Settings) ResolverURIs() []string {
	var configPaths []string
	if configPaths = s.configPaths.value; len(configPaths) == 0 {
		if configEnvVal := os.Getenv(ConfigEnvVar); len(configEnvVal) != 0 {
			configPaths = []string{"file:" + configEnvVal}
		}
	}

	configDir := getConfigDir(s)

	for _, property := range s.discoveryProperties {
		configPaths = append(configPaths, fmt.Sprintf("%s:%s", s.discovery.PropertyScheme(), property))
	}

	if s.configD {
		configPaths = append(configPaths, fmt.Sprintf("%s:%s", s.discovery.ConfigDScheme(), configDir))
	}

	if s.discoveryMode {
		// discovery uri must come last to successfully merge w/ other config content
		configPaths = append(configPaths, fmt.Sprintf("%s:%s", s.discovery.DiscoveryModeScheme(), configDir))
	}

	configYaml := os.Getenv(ConfigYamlEnvVar)

	switch {
	case len(configPaths) == 0 && configYaml != "":
		return []string{"env:" + ConfigYamlEnvVar}
	case len(configPaths) == 0:
		return []string{""}
	default:
		return configPaths
	}
}

func getConfigDir(f *Settings) string {
	configDir := DefaultConfigDir
	if envConfigDir, ok := os.LookupEnv(ConfigDirEnvVar); ok {
		configDir = envConfigDir
	}

	if f.configDir.value != nil {
		configDir = f.configDir.String()
	}

	return configDir
}

// ConfMapProviders returns the confmap.Providers by their scheme for the collector core service.
func (s *Settings) ConfMapProviders() map[string]confmap.Provider {
	return s.confMapProviders
}

func loadConfMapProviders(s *Settings) error {
	var err error
	if s.discovery, err = discovery.New(); err != nil {
		return fmt.Errorf("failed to create discovery provider: %w", err)
	}

	s.confMapProviders = map[string]confmap.Provider{
		envProvider.Scheme():  envProvider,
		fileProvider.Scheme(): fileProvider,
	}

	for p := range s.confMapProviders {
		s.supportedURISchemes = append(s.supportedURISchemes, p)
	}
	sort.Strings(s.supportedURISchemes)

	// though supported, these schemes shouldn't be advertised for use w/ --config
	s.confMapProviders[s.discovery.PropertyScheme()] = s.discovery.PropertyProvider()
	s.confMapProviders[s.discovery.ConfigDScheme()] = s.discovery.ConfigDProvider()
	s.confMapProviders[s.discovery.DiscoveryModeScheme()] = s.discovery.DiscoveryModeProvider()
	return nil
}

// ConfMapConverters returns confmap.Converters for the collector core service.
func (s *Settings) ConfMapConverters() []confmap.Converter {
	confMapConverters := []confmap.Converter{
		configconverter.NewOverwritePropertiesConverter(s.setProperties),
		configconverter.Discovery{},
	}
	if !s.noConvertConfig {
		confMapConverters = append(
			confMapConverters,
			configconverter.RemoveBallastKey{},
			configconverter.MoveOTLPInsecureKey{},
			configconverter.MoveHecTLS{},
			configconverter.RenameK8sTagger{},
			configconverter.NormalizeGcp{},
		)
	}
	return confMapConverters
}

// ColCoreArgs returns list of arguments to be passed to the collector core service.
func (s *Settings) ColCoreArgs() []string {
	return s.colCoreArgs
}

// IsDryRun returns whether --dry-run mode was requested
func (s *Settings) IsDryRun() bool {
	return s.dryRun
}

// parseArgs returns new Settings instance from command line arguments.
func parseArgs(args []string) (*Settings, error) {
	flagSet := flag.NewFlagSet("otelcol", flag.ContinueOnError)

	settings := &Settings{
		configPaths:        new(stringArrayFlagValue),
		setOptionArguments: new(stringArrayFlagValue),
		configDir:          new(stringPointerFlagValue),
	}

	if err := loadConfMapProviders(settings); err != nil {
		return nil, fmt.Errorf("failed loading confmap.Providers: %w", err)
	}

	flagSet.Var(settings.configPaths, "config", "Locations to the config file(s), "+
		"note that only a single location can be set per flag entry e.g. --config=/path/to/first "+
		"--config=path/to/second.")
	flagSet.Var(settings.setOptionArguments, "set", "Set arbitrary component config property. "+
		"The component has to be defined in the config file and the flag has a higher precedence. "+
		"Array config properties are overridden and maps are joined. Example --set=processors.batch.timeout=2s")
	flagSet.BoolVar(&settings.dryRun, "dry-run", false, "Don't run the service, just show the configuration")
	flagSet.MarkHidden("dry-run")
	flagSet.BoolVar(&settings.noConvertConfig, "no-convert-config", false,
		"Do not translate old configurations to the new format automatically. "+
			"By default, old configurations are translated to the new format for backward compatibility.")

	// Deprecated "--metrics-addr" flag is a noop, but temporarily required to run the collector GKE/Autopilot.
	addressFlag := ""
	flagSet.StringVar(&addressFlag, "metrics-addr", "", "")
	flagSet.MarkHidden("metrics-addr")

	// Experimental flags
	flagSet.VarPF(settings.configDir, "config-dir", "", "").Hidden = true
	flagSet.BoolVar(&settings.configD, "configd", false, "")
	flagSet.MarkHidden("configd")
	flagSet.BoolVar(&settings.discoveryMode, "discovery", false, "")
	flagSet.MarkHidden("discovery")

	// OTel Collector Core flags
	colCoreFlags := []string{"version", "feature-gates"}
	flagSet.BoolVarP(&settings.versionFlag, colCoreFlags[0], "v", false, "Version of the collector.")
	flagSet.Var(new(stringArrayFlagValue), colCoreFlags[1],
		"Comma-delimited list of feature gate identifiers. Prefix with '-' to disable the feature. "+
			"'+' or no prefix will enable the feature.")

	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}

	settings.setProperties, settings.discoveryProperties = parseSetOptionArguments(settings.setOptionArguments.value)

	// Pass flags that are handled by the collector core service as raw command line arguments.
	settings.colCoreArgs = flagSetToArgs(colCoreFlags, flagSet)

	return settings, nil
}

func parseSetOptionArguments(arguments []string) (setProperties, discoveryProperties []string) {
	for _, arg := range arguments {
		if strings.HasPrefix(arg, "splunk.discovery") {
			discoveryProperties = append(discoveryProperties, arg)
		} else {
			setProperties = append(setProperties, arg)
		}
	}
	return
}

// flagSetToArgs takes a list of flag names and returns a list of corresponding command line arguments
// using values from the provided flagSet.
// The flagSet must be populated (flagSet.Parse is called), otherwise the returned list of arguments will be empty.
func flagSetToArgs(flagNames []string, flagSet *flag.FlagSet) []string {
	var out []string
	for _, flagName := range flagNames {
		flag := flagSet.Lookup(flagName)
		if flag.Changed {
			switch fv := flag.Value.(type) {
			case *stringArrayFlagValue:
				for _, val := range fv.value {
					out = append(out, "--"+flagName, val)
				}
			default:
				out = append(out, "--"+flagName, flag.Value.String())
			}
		}
	}
	return out
}

func checkRuntimeParams(settings *Settings) error {
	if err := checkConfig(settings); err != nil {
		return err
	}

	// Set default total memory
	memTotalSize := DefaultMemoryTotalMiB
	// Check if the total memory is specified via the env var
	// If so, validate and change total memory
	if os.Getenv(MemTotalEnvVar) != "" {
		// Check if it is a numeric value.
		memTotalSize = envVarAsInt(MemTotalEnvVar)
		// Ensure number is above some threshold
		if 99 > memTotalSize {
			return fmt.Errorf("expected a number greater than 99 for %s env variable but got %d", MemTotalEnvVar, memTotalSize)
		}
	}

	ballastSize := setMemoryBallast(memTotalSize)
	memLimit, err := setMemoryLimit(memTotalSize)
	if err != nil {
		return err
	}

	// Validate memoryLimit and memoryBallast are sane
	if 2*ballastSize > memLimit {
		return fmt.Errorf("memory limit (%d) is less than 2x ballast (%d). Increase memory limit or decrease ballast size", memLimit, ballastSize)
	}
	return nil
}

func setDefaultEnvVars() error {
	type ev struct{ e, v string }

	envVars := []ev{{e: ConfigServerEnabledEnvVar, v: "true"}}

	if realm, ok := os.LookupEnv(RealmEnvVar); ok {
		envVars = append(envVars,
			ev{e: APIURLEnvVar, v: fmt.Sprintf("https://api.%s.signalfx.com", realm)},
			ev{e: IngestURLEnvVar, v: fmt.Sprintf("https://ingest.%s.signalfx.com", realm)},
			ev{e: TraceIngestURLEnvVar, v: fmt.Sprintf("https://ingest.%s.signalfx.com/v2/trace", realm)},
			ev{e: HecLogIngestURLEnvVar, v: fmt.Sprintf("https://ingest.%s.signalfx.com/v1/log", realm)},
		)
	}

	if token, ok := os.LookupEnv(TokenEnvVar); ok {
		envVars = append(envVars, ev{e: HecTokenEnvVar, v: token})
	}

	for _, envVar := range envVars {
		if _, ok := os.LookupEnv(envVar.e); !ok {
			if err := os.Setenv(envVar.e, envVar.v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Config priority (highest to lowest):
// 1. '--config' flags (multiple supported),
// 2. SPLUNK_CONFIG env var,
// 3. SPLUNK_CONFIG_YAML env var,
// 4. default gateway config path.
func checkConfig(settings *Settings) error {
	configPathVar := os.Getenv(ConfigEnvVar)
	configYaml := os.Getenv(ConfigYamlEnvVar)

	switch {
	case len(settings.configPaths.value) != 0:
		if err := checkInputConfigs(settings); err != nil {
			return err
		}
		log.Printf("Set config to %v", settings.configPaths.String())
	case configPathVar != "":
		if err := checkConfigPathEnvVar(settings); err != nil {
			return err
		}
		log.Printf("Set config to %v", configPathVar)
	case configYaml != "":
		log.Printf("Using environment variable %s for configuration", ConfigYamlEnvVar)
	default:
		defaultConfigPath, err := getExistingDefaultConfigPath()
		if err != nil {
			return err
		}
		settings.configPaths.Set(defaultConfigPath)
		if err = confirmRequiredEnvVarsForDefaultConfigs(settings.configPaths.value); err != nil {
			return err
		}
		log.Printf("Set config to %v", defaultConfigPath)
	}
	return nil
}

func getExistingDefaultConfigPath() (path string, err error) {
	if _, err = os.Stat(DefaultGatewayConfig); err == nil {
		path = DefaultGatewayConfig
		return
	}
	err = fmt.Errorf("unable to find the default configuration file %s", DefaultGatewayConfig)
	return
}

func envVarAsInt(env string) int {
	envVal := os.Getenv(env)
	// Check if it is a numeric value.
	val, err := strconv.Atoi(envVal)
	if err != nil {
		log.Fatalf("Expected a number in %s env variable but got %s", env, envVal)
	}
	return val
}

// Validate and set the memory ballast
func setMemoryBallast(memTotalSizeMiB int) int {
	ballastSize := memTotalSizeMiB * DefaultMemoryBallastPercentage / 100
	// Check if the memory ballast is specified via the env var, if so, validate and set properly.
	if os.Getenv(BallastEnvVar) != "" {
		ballastSize = envVarAsInt(BallastEnvVar)
		if 33 > ballastSize {
			log.Fatalf("Expected a number greater than 33 for %s env variable but got %d", BallastEnvVar, ballastSize)
		}
	}

	_ = os.Setenv(BallastEnvVar, strconv.Itoa(ballastSize))
	log.Printf("Set ballast to %d MiB", ballastSize)
	return ballastSize
}

// Validate and set the memory limit
func setMemoryLimit(memTotalSizeMiB int) (int, error) {
	memLimit := memTotalSizeMiB * DefaultMemoryLimitPercentage / 100

	// Check if the memory limit is specified via the env var, if so, validate and set properly.
	if os.Getenv(MemLimitMiBEnvVar) != "" {
		memLimit = envVarAsInt(MemLimitMiBEnvVar)
	}

	if err := os.Setenv(MemLimitMiBEnvVar, strconv.Itoa(memLimit)); err != nil {
		return -1, err
	}
	log.Printf("Set memory limit to %d MiB", memLimit)
	return memLimit, nil
}

func checkInputConfigs(settings *Settings) error {
	configPathVar := os.Getenv(ConfigEnvVar)
	configYaml := os.Getenv(ConfigYamlEnvVar)

	var configFilePaths []string
	for _, filePath := range settings.configPaths.value {
		scheme, location, isURI := parseURI(filePath)
		if isURI {
			if _, ok := settings.confMapProviders[scheme]; !ok {
				log.Printf("%q is an unsupported config provider scheme for this Collector distribution (not in %v).", scheme, settings.supportedURISchemes)
				continue
			}
			if scheme != fileProvider.Scheme() {
				continue
			}
			filePath = location
		}
		if _, err := os.Stat(filePath); err != nil {
			return fmt.Errorf("unable to find the configuration file %s, ensure flag '--config' is set properly: %w", filePath, err)
		}
		configFilePaths = append(configFilePaths, filePath)
	}

	if len(configFilePaths) == 0 {
		return nil
	}

	if configPathVar != "" {
		differingVals := true
		for _, p := range configFilePaths {
			if p == configPathVar {
				differingVals = false
				break
			}
		}
		if differingVals {
			log.Printf("Both environment variable %v and flag '--config' were specified. Using the flag values and ignoring the environment variable value %s in this session", ConfigEnvVar, configPathVar)
		}
	}

	if configYaml != "" {
		log.Printf("Both environment variable %s and flag '--config' were specified. Using the flag values and ignoring the environment variable in this session", ConfigYamlEnvVar)
	}

	return confirmRequiredEnvVarsForDefaultConfigs(configFilePaths)
}

func checkConfigPathEnvVar(settings *Settings) error {
	configPathVar := os.Getenv(ConfigEnvVar)
	configYaml := os.Getenv(ConfigYamlEnvVar)

	if _, err := os.Stat(configPathVar); err != nil {
		return fmt.Errorf("unable to find the configuration file (%s), ensure %s environment variable is set properly: %w", configPathVar, ConfigEnvVar, err)
	}

	if configYaml != "" {
		log.Printf("Both %s and %s were specified. Using %s environment variable value %s for this session", ConfigEnvVar, ConfigYamlEnvVar, ConfigEnvVar, configPathVar)
	}

	if !settings.configPaths.contains(configPathVar) {
		_ = settings.configPaths.Set(configPathVar)
	}

	return confirmRequiredEnvVarsForDefaultConfigs(settings.configPaths.value)
}

func confirmRequiredEnvVarsForDefaultConfigs(paths []string) error {
	// Check environment variables required by default configuration.
	for _, path := range paths {
		switch path {
		case
			DefaultGatewayConfig,
			DefaultOTLPLinuxConfig:
			requiredEnvVars := []string{RealmEnvVar, TokenEnvVar}
			for _, v := range requiredEnvVars {
				if len(os.Getenv(v)) == 0 {
					log.Printf("Usage: %s=12345 %s=us0 %s", TokenEnvVar, RealmEnvVar, os.Args[0])
					return fmt.Errorf("ERROR: Missing required environment variable %s with default config path %s", v, path)
				}
			}
		}
	}
	return nil
}

var _ flag.Value = (*stringArrayFlagValue)(nil)

// based on https://github.com/open-telemetry/opentelemetry-collector/blob/48a2e01652fa679c89259866210473fc0d42ca95/service/flags.go#L39
type stringArrayFlagValue struct {
	value []string
}

func (s *stringArrayFlagValue) Type() string {
	return "string"
}

func (s *stringArrayFlagValue) Set(val string) error {
	s.value = append(s.value, val)
	return nil
}

func (s *stringArrayFlagValue) String() string {
	return "[" + strings.Join(s.value, ",") + "]"
}

func (s *stringArrayFlagValue) contains(input string) bool {
	for _, val := range s.value {
		if val == input {
			return true
		}
	}

	return false
}

var _ flag.Value = (*stringPointerFlagValue)(nil)

// based on https://github.com/open-telemetry/opentelemetry-collector/blob/48a2e01652fa679c89259866210473fc0d42ca95/service/flags.go#L39
type stringPointerFlagValue struct {
	value *string
}

func (s *stringPointerFlagValue) Type() string {
	return "string"
}

func (s *stringPointerFlagValue) Set(val string) error {
	s.value = &val
	return nil
}

func (s *stringPointerFlagValue) String() string {
	if s.value == nil {
		return ""
	}
	return *s.value
}

// From https://github.com/open-telemetry/opentelemetry-collector/blob/18a11ec09b3f4883d0360a41054ce8f4a8736ea8/confmap/expand.go
// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
var uriRegexp = regexp.MustCompile(`(?s:^(?P<Scheme>[A-Za-z][A-Za-z0-9+.-]+):(?P<OpaqueValue>.*)$)`)

func parseURI(uri string) (scheme string, location string, isURI bool) {
	submatches := uriRegexp.FindStringSubmatch(uri)
	if len(submatches) != 3 {
		return "", "", false
	}
	return submatches[1], submatches[2], true
}
