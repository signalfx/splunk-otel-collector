// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

// Program otelcol is the OpenTelemetry Collector that collects stats
// and traces and exports to a configured backend.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmapprovider"
	"go.opentelemetry.io/collector/service"

	"github.com/signalfx/splunk-otel-collector/internal/collectorconfig"
	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/version"
)

// The list of environment variables must be the same as what is used in the yaml configs.
const (
	ballastEnvVarName         = "SPLUNK_BALLAST_SIZE_MIB"
	configEnvVarName          = "SPLUNK_CONFIG"
	configYamlEnvVarName      = "SPLUNK_CONFIG_YAML"
	configServerEnabledEnvVar = "SPLUNK_DEBUG_CONFIG_SERVER"
	memLimitMiBEnvVarName     = "SPLUNK_MEMORY_LIMIT_MIB"
	memTotalEnvVarName        = "SPLUNK_MEMORY_TOTAL_MIB"
	realmEnvVarName           = "SPLUNK_REALM"
	tokenEnvVarName           = "SPLUNK_ACCESS_TOKEN"

	defaultDockerSAPMConfig        = "/etc/otel/collector/gateway_config.yaml"
	defaultDockerOTLPConfig        = "/etc/otel/collector/otlp_config_linux.yaml"
	defaultLocalSAPMConfig         = "cmd/otelcol/config/collector/gateway_config.yaml"
	defaultLocalOTLPConfig         = "cmd/otelcol/config/collector/otlp_config_linux.yaml"
	defaultMemoryBallastPercentage = 33
	defaultMemoryLimitPercentage   = 90
	defaultMemoryTotalMiB          = 512
)

func main() {
	// TODO: Use same format as the collector
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if !hasFlag("-h") && !hasFlag("--help") && !hasFlag("-v") && !hasFlag("--version") {
		checkRuntimeParams()
		setDefaultEnvVars()
	}

	// Allow dumping configuration locally by default
	// Used by support bundle script
	_ = os.Setenv(configServerEnabledEnvVar, "true")

	factories, err := components.Get()
	if err != nil {
		log.Fatalf("failed to build default components: %v", err)
	}

	info := component.BuildInfo{
		Command: "otelcol",
		Version: version.Version,
	}

	serviceParams := service.CollectorSettings{
		BuildInfo:         info,
		Factories:         factories,
		ConfigMapProvider: newConfigMapProvider(info),
	}

	if err := run(serviceParams); err != nil {
		log.Fatal(err)
	}
}

func newConfigMapProvider(info component.BuildInfo) configmapprovider.Provider {
	return collectorconfig.NewConfigMapProvider(
		info,
		hasNoConvertFlag(),
		getConfigPath(),
		os.Getenv(configYamlEnvVarName),
		getSetProperties(),
	)
}

func hasNoConvertFlag() bool {
	const noConvertConfigFlag = "--no-convert-config"
	if hasFlag(noConvertConfigFlag) {
		// the collector complains about this flag if we don't remove it
		removeFlag(&os.Args, noConvertConfigFlag)
		return true
	}
	return false
}

func getConfigPath() string {
	ok, configPath := getKeyValue(os.Args[1:], "--config")
	if !ok {
		return os.Getenv(configEnvVarName)
	}
	return configPath
}

// required to support --set functionality no longer directly parsed by the core config loader.
// taken from https://github.com/open-telemetry/opentelemetry-collector/blob/48a2e01652fa679c89259866210473fc0d42ca95/service/flags.go#L39
type stringArrayValue struct {
	values []string
}

func (s *stringArrayValue) Set(val string) error {
	s.values = append(s.values, val)
	return nil
}

func (s *stringArrayValue) String() string {
	return "[" + strings.Join(s.values, ",") + "]"
}

func getSetProperties() []string {
	properties := &stringArrayValue{}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	flagSet.Var(properties, "set", "")
	// we are only interested in the --set option so ignore errors
	_ = flagSet.Parse(os.Args[1:])
	return properties.values
}

func hasFlag(flag string) bool {
	return contains(os.Args[1:], flag)
}

// Check whether a string exists in an array of CLI arguments
// Support key/value with and without an equal sign
func contains(arr []string, str string) bool {
	for _, a := range arr {
		// Command line argument may be of form
		// --key value OR --key=value
		if a == str {
			return true
		} else if strings.Contains(a, str+"=") {
			return true
		}
	}
	return false
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

// Get the value of a key in an array
// Support key/value with and with an equal sign
func getKeyValue(args []string, arg string) (exists bool, value string) {
	argEq := arg + "="
	for i := range args {
		switch {
		case strings.HasPrefix(args[i], argEq):
			return true, strings.SplitN(args[i], "=", 2)[1]
		case args[i] == arg:
			exists = true
			if i < (len(args) - 1) {
				value = args[i+1]
			}
			return
		}
	}
	return
}

// Check runtime parameters
// Runtime parameters take priority over environment variables
// Config and ballast flags are checked
// Config and all memory env vars are checked
func checkRuntimeParams() {
	checkConfig()

	// Set default total memory
	memTotalSize := defaultMemoryTotalMiB
	// Check if the total memory is specified via the env var
	// If so, validate and change total memory
	if os.Getenv(memTotalEnvVarName) != "" {
		// Check if it is a numeric value.
		memTotalSize = envVarAsInt(memTotalEnvVarName)
		// Ensure number is above some threshold
		if 99 > memTotalSize {
			log.Fatalf("Expected a number greater than 99 for %s env variable but got %d", memTotalEnvVarName, memTotalSize)
		}
	}

	ballastSize := setMemoryBallast(memTotalSize)
	memLimit := setMemoryLimit(memTotalSize)

	// Validate memoryLimit and memoryBallast are sane
	if 2*ballastSize > memLimit {
		log.Fatalf("Memory limit (%d) is less than 2x ballast (%d). Increase memory limit or decrease ballast size.", memLimit, ballastSize)
	}
}

// Sets flag '--config' to specified env var SPLUNK_CONFIG, if the flag not specified.
// Logs a message and returns if env var SPLUNK_CONFIG_YAML specified, and '--config' and SPLUNK_CONFIG not specified.
// Sets '--config' to default config file path if '--config', SPLUNK_CONFIG and SPLUNK_CONFIG_YAML not specified.
func checkConfig() {
	configPathFlagExists, configPathFlag := getKeyValue(os.Args[1:], "--config")
	configPathVar := os.Getenv(configEnvVarName)
	configYaml := os.Getenv(configYamlEnvVarName)

	if configPathFlagExists && configPathFlag == "" {
		log.Fatal("Command line flag --config specified but empty")
	}

	if configPathFlag != "" {
		if _, err := os.Stat(configPathFlag); err != nil {
			log.Fatalf("Unable to find the configuration file (%s) ensure flag '--config' is set properly", configPathFlag)
		}

		if configPathVar != "" && configPathVar != configPathFlag {
			log.Printf("Both environment variable %v and flag '--config' were specified. Using the flag value %s and ignoring the environment variable value %s in this session", configEnvVarName, configPathFlag, configPathVar)
		}

		if configYaml != "" {
			log.Printf("Both environment variable %s and flag '--config' were specified. Using the flag value %s and ignoring the environment variable in this session", configYamlEnvVarName, configPathFlag)
		}

		checkRequiredEnvVars(configPathFlag)

		log.Printf("Set config to %v", configPathFlag)
		return
	}

	if configPathVar != "" {
		if _, err := os.Stat(configPathVar); err != nil {
			log.Fatalf("Unable to find the configuration file (%s) ensure %s environment variable is set properly", configPathVar, configEnvVarName)
		}

		os.Args = append(os.Args, "--config="+configPathVar)

		if configYaml != "" {
			log.Printf("Both %s and %s were specified. Using %s environment variable value %s for this session", configEnvVarName, configYamlEnvVarName, configEnvVarName, configPathVar)
		}

		checkRequiredEnvVars(configPathVar)

		log.Printf("Set config to %v", configPathVar)
		return
	}

	if configYaml != "" {
		log.Printf("Using environment variable %s for configuration", configYamlEnvVarName)
		return
	}

	defaultConfigPath := getExistingDefaultConfigPath()
	checkRequiredEnvVars(defaultConfigPath)
	os.Args = append(os.Args, "--config="+defaultConfigPath)
	log.Printf("Set config to %v", defaultConfigPath)
}

func getExistingDefaultConfigPath() (path string) {
	if _, err := os.Stat(defaultLocalSAPMConfig); err == nil {
		return defaultLocalSAPMConfig
	}
	if _, err := os.Stat(defaultDockerSAPMConfig); err == nil {
		return defaultDockerSAPMConfig
	}
	log.Fatalf("Unable to find the default configuration file (%s) or (%s)", defaultLocalSAPMConfig, defaultDockerSAPMConfig)
	return
}

func checkRequiredEnvVars(path string) {
	// Check environment variables required by default configuration.
	switch path {
	case
		defaultDockerSAPMConfig,
		defaultDockerOTLPConfig,
		defaultLocalSAPMConfig,
		defaultLocalOTLPConfig:
		requiredEnvVars := []string{realmEnvVarName, tokenEnvVarName}
		for _, v := range requiredEnvVars {
			if len(os.Getenv(v)) == 0 {
				log.Printf("Usage: %s=12345 %s=us0 %s", tokenEnvVarName, realmEnvVarName, os.Args[0])
				log.Fatalf("ERROR: Missing required environment variable %s with default config path %s", v, path)
			}
		}
	}
}

// Validate and set the memory ballast
func setMemoryBallast(memTotalSizeMiB int) int {
	// Check if deprecated memory ballast flag was passed, if so, ensure the env variable for memory ballast is set.
	// Then set memory ballast and limit properly
	_, ballastSizeFlag := getKeyValue(os.Args[1:], "--mem-ballast-size-mib")
	if ballastSizeFlag != "" {
		if os.Getenv(ballastEnvVarName) != "" {
			log.Fatalf("Both %v and '--mem-ballast-size-mib' were specified, but only one is allowed", ballastEnvVarName)
		}
		_ = os.Setenv(ballastEnvVarName, ballastSizeFlag)
	}

	ballastSize := memTotalSizeMiB * defaultMemoryBallastPercentage / 100
	// Check if the memory ballast is specified via the env var, if so, validate and set properly.
	if os.Getenv(ballastEnvVarName) != "" {
		ballastSize = envVarAsInt(ballastEnvVarName)
		if 33 > ballastSize {
			log.Fatalf("Expected a number greater than 33 for %s env variable but got %d", ballastEnvVarName, ballastSize)
		}
	}

	_ = os.Setenv(ballastEnvVarName, strconv.Itoa(ballastSize))
	log.Printf("Set ballast to %d MiB", ballastSize)
	return ballastSize
}

// Validate and set the memory limit
func setMemoryLimit(memTotalSizeMiB int) int {
	memLimit := memTotalSizeMiB * defaultMemoryLimitPercentage / 100

	// Check if the memory limit is specified via the env var, if so, validate and set properly.
	if os.Getenv(memLimitMiBEnvVarName) != "" {
		memLimit = envVarAsInt(memLimitMiBEnvVarName)
	}

	_ = os.Setenv(memLimitMiBEnvVarName, strconv.Itoa(memLimit))
	log.Printf("Set memory limit to %d MiB", memLimit)
	return memLimit
}

// Sets environment variables expected by agent_config.yaml if missing
func setDefaultEnvVars() {
	realm, realmOk := os.LookupEnv("SPLUNK_REALM")
	if realmOk {
		testArgs := [][]string{
			{"SPLUNK_API_URL", "https://api." + realm + ".signalfx.com"},
			{"SPLUNK_INGEST_URL", "https://ingest." + realm + ".signalfx.com"},
			{"SPLUNK_TRACE_URL", "https://ingest." + realm + ".signalfx.com/v2/trace"},
			{"SPLUNK_HEC_URL", "https://ingest." + realm + ".signalfx.com/v1/log"},
		}
		for _, v := range testArgs {
			_, ok := os.LookupEnv(v[0])
			if !ok {
				_ = os.Setenv(v[0], v[1])
			}
		}
	}
	token, tokenOk := os.LookupEnv("SPLUNK_ACCESS_TOKEN")
	if tokenOk {
		_, ok := os.LookupEnv("SPLUNK_HEC_TOKEN")
		if !ok {
			_ = os.Setenv("SPLUNK_HEC_TOKEN", token)
		}
	}
}

func runInteractive(settings service.CollectorSettings) error {
	cmd := service.NewCommand(settings)
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("application run finished with error: %w", err)
	}

	return nil
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
