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
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/overwritepropertiesconverter"

	"github.com/signalfx/splunk-otel-collector/internal/configconverter"
)

const (
	DefaultUndeclaredFlag = -1

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

	DiscoveryModeScheme = "splunk.discovery"
	ConfigDScheme       = "splunk.configd"
)

type Settings interface {
	// ResolverURIs returns the collector config provider resolver uris for the collector service
	ResolverURIs() []string
	// ConfMapConverters returns the collector config provider resolver confmap.Converters for the collector service
	ConfMapConverters() []confmap.Converter
	// ServiceArgs are the sanitized, adjusted args to be used in updating os.Args[1:] for the collector service
	ServiceArgs() []string
	// IsDryRun returns whether --dry-run mode was requested
	IsDryRun() bool
}

func New(args []string) (Settings, error) {
	f, err := newFlags(args)
	if err != nil {
		return nil, err
	}

	// immediate exit paths, no further setup required
	if f.helpFlag || f.versionFlag {
		return f, nil
	}

	if err = checkRuntimeParams(f); err != nil {
		return nil, err
	}

	if err = setDefaultEnvVars(); err != nil {
		return nil, err
	}

	return f, nil
}

var _ Settings = (*flags)(nil)

type flags struct {
	configPaths     *stringArrayFlagValue
	setProperties   *stringArrayFlagValue
	configDir       *stringPointerFlagValue
	serviceArgs     []string
	helpFlag        bool
	versionFlag     bool
	noConvertConfig bool
	configD         bool
	discoveryMode   bool
	dryRun          bool
}

func (f *flags) ResolverURIs() []string {
	var configPaths []string
	if configPaths = f.configPaths.value; len(configPaths) == 0 {
		if configEnvVal := os.Getenv(ConfigEnvVar); len(configEnvVal) != 0 {
			configPaths = []string{"file:" + configEnvVal}
		}
	}

	configDir := getConfigDir(f)

	if f.dryRun {
		removeFlag(&f.serviceArgs, "--dry-run")
	}

	if f.configD {
		removeFlag(&f.serviceArgs, "--configd")
		configPaths = append(configPaths, fmt.Sprintf("%s:%s", ConfigDScheme, configDir))
	}

	if f.discoveryMode {
		removeFlag(&f.serviceArgs, "--discovery")
		// discovery uri must come last to successfully merge w/ other config content
		configPaths = append(configPaths, fmt.Sprintf("%s:%s", DiscoveryModeScheme, f.configDir))
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

func getConfigDir(f *flags) string {
	configDir := DefaultConfigDir
	if envConfigDir, ok := os.LookupEnv(ConfigDirEnvVar); ok {
		configDir = envConfigDir
	}

	if f.configDir.value != nil {
		removeFlag(&f.serviceArgs, "--config-dir")
		configDir = f.configDir.String()
	}

	return configDir
}

func (f *flags) ConfMapConverters() []confmap.Converter {
	confMapConverters := []confmap.Converter{
		// nolint: staticcheck
		overwritepropertiesconverter.New(f.setProperties.value), // support until there's an actual replacement
	}

	if f.noConvertConfig {
		// the collector complains about this flag if we don't remove it. Unfortunately,
		// this must be done manually since the flag library has no functionality to remove
		// args
		removeFlag(&f.serviceArgs, "--no-convert-config")
	} else {
		confMapConverters = append(
			confMapConverters,
			configconverter.RemoveBallastKey{},
			configconverter.MoveOTLPInsecureKey{},
			configconverter.MoveHecTLS{},
			configconverter.RenameK8sTagger{},
		)
	}
	return confMapConverters
}

func (f *flags) ServiceArgs() []string {
	return f.serviceArgs
}

func (f *flags) IsDryRun() bool {
	return f.dryRun
}

func newFlags(args []string) (*flags, error) {
	flagSet := flag.FlagSet{}
	// we don't want to be responsible for tracking all supported collector service
	// flags, so allow any we don't use and defer parsing to the actual service
	flagSet.ParseErrorsWhitelist.UnknownFlags = true

	var cpArgs []string
	cpArgs = append(cpArgs, args...)

	settings := &flags{
		configPaths:   new(stringArrayFlagValue),
		setProperties: new(stringArrayFlagValue),
		serviceArgs:   cpArgs,
		configDir:     new(stringPointerFlagValue),
	}

	// This is an internal flag parser, it shouldn't give any output to user.
	flagSet.SetOutput(io.Discard)

	flagSet.BoolVarP(&settings.helpFlag, "help", "h", false, "")
	flagSet.BoolVarP(&settings.versionFlag, "version", "v", false, "")

	flagSet.BoolVar(&settings.noConvertConfig, "no-convert-config", false, "")

	flagSet.BoolVar(&settings.configD, "configd", false, "")
	flagSet.Var(settings.configDir, "config-dir", "")
	flagSet.BoolVar(&settings.discoveryMode, "discovery", false, "")
	flagSet.BoolVar(&settings.dryRun, "dry-run", false, "")

	flagSet.Var(settings.configPaths, "config", "")
	flagSet.Var(settings.setProperties, "set", "")

	if err := flagSet.Parse(cpArgs); err != nil {
		return nil, err
	}

	return settings, nil
}

func removeFlag(flags *[]string, flag string) {
	var out []string
	for _, s := range *flags {
		if s != flag {
			out = append(out, s)
		}
	}
	*flags = out
}

func checkRuntimeParams(settings *flags) error {
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
func checkConfig(settings *flags) error {
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

func checkInputConfigs(settings *flags) error {
	configPathVar := os.Getenv(ConfigEnvVar)
	configYaml := os.Getenv(ConfigYamlEnvVar)

	for _, filePath := range settings.configPaths.value {
		if _, err := os.Stat(filePath); err != nil {
			return fmt.Errorf("unable to find the configuration file %s, ensure flag '--config' is set properly: %w", filePath, err)
		}
	}

	if configPathVar != "" && !settings.configPaths.contains(configPathVar) {
		log.Printf("Both environment variable %v and flag '--config' were specified. Using the flag values and ignoring the environment variable value %s in this session", ConfigEnvVar, configPathVar)
	}

	if configYaml != "" {
		log.Printf("Both environment variable %s and flag '--config' were specified. Using the flag values and ignoring the environment variable in this session", ConfigYamlEnvVar)
	}

	return confirmRequiredEnvVarsForDefaultConfigs(settings.configPaths.value)
}

func checkConfigPathEnvVar(settings *flags) error {
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
