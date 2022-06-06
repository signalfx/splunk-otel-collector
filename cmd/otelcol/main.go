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
	"fmt"
	"log"
	"os"
	"strconv"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/mapconverter/overwritepropertiesmapconverter"
	"go.opentelemetry.io/collector/config/mapprovider/envmapprovider"
	"go.opentelemetry.io/collector/config/mapprovider/filemapprovider"
	"go.opentelemetry.io/collector/service"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/configconverter"
	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
	"github.com/signalfx/splunk-otel-collector/internal/configsources"
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

	// Core flag parser will handle errors, we don't have to handle them here.
	inputFlags, err := parseFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %v\nUse \"--help\" to show valid usage", err)
	}

	if !inputFlags.help && !inputFlags.version {
		checkRuntimeParams(inputFlags)
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

	configMapConverters := []config.MapConverter{
		overwritepropertiesmapconverter.New(inputFlags.sets.values),
	}

	if inputFlags.noConvertConfig {
		// the collector complains about this flag if we don't remove it. Unfortunately,
		// this must be done manually since the flag library has no functionality to remove
		// args
		removeFlag(&os.Args, "--no-convert-config")
	} else {
		configMapConverters = append(
			configMapConverters,
			configconverter.RemoveBallastKey{},
			configconverter.MoveOTLPInsecureKey{},
			configconverter.MoveHecTLS{},
			configconverter.RenameK8sTagger{},
		)
	}

	emp := envmapprovider.New()
	fmp := filemapprovider.New()
	serviceConfigProvider, err := service.NewConfigProvider(
		service.ConfigProviderSettings{
			Locations: configLocations(inputFlags),
			MapProviders: map[string]config.MapProvider{
				emp.Scheme(): configprovider.NewConfigSourceConfigMapProvider(
					emp,
					zap.NewNop(), // The service logger is not available yet, setting it to NoP.
					info,
					configsources.Get()...,
				),
				fmp.Scheme(): configprovider.NewConfigSourceConfigMapProvider(
					fmp,
					zap.NewNop(), // The service logger is not available yet, setting it to NoP.
					info,
					configsources.Get()...,
				),
			},
			MapConverters: configMapConverters,
		})
	if err != nil {
		log.Fatal(err)
	}

	serviceParams := service.CollectorSettings{
		BuildInfo:      info,
		Factories:      factories,
		ConfigProvider: serviceConfigProvider,
	}

	if err := run(serviceParams); err != nil {
		log.Fatal(err)
	}
}

// Check runtime parameters
// Runtime parameters take priority over environment variables
// Config and ballast flags are checked
// Config and all memory env vars are checked
func checkRuntimeParams(inputFlags flags) {
	checkConfig(inputFlags)

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

	ballastSize := setMemoryBallast(inputFlags, memTotalSize)
	memLimit := setMemoryLimit(memTotalSize)

	// Validate memoryLimit and memoryBallast are sane
	if 2*ballastSize > memLimit {
		log.Fatalf("Memory limit (%d) is less than 2x ballast (%d). Increase memory limit or decrease ballast size.", memLimit, ballastSize)
	}
}

func checkInputConfigs(inputFlags flags) {
	configPathVar := os.Getenv(configEnvVarName)
	configYaml := os.Getenv(configYamlEnvVarName)

	for _, filePath := range inputFlags.configs.values {
		if _, err := os.Stat(filePath); err != nil {
			log.Fatalf("Unable to find the configuration file (%s) ensure flag '--config' is set properly", filePath)
		}
	}

	if configPathVar != "" && !inputFlags.configs.contains(configPathVar) {
		log.Printf("Both environment variable %v and flag '--config' were specified. Using the flag values and ignoring the environment variable value %s in this session", configEnvVarName, configPathVar)
	}

	if configYaml != "" {
		log.Printf("Both environment variable %s and flag '--config' were specified. Using the flag values and ignoring the environment variable in this session", configYamlEnvVarName)
	}

	checkRequiredEnvVars(inputFlags.configs.values)
}

func checkConfigPathEnvVar(inputFlags flags) {
	configPathVar := os.Getenv(configEnvVarName)
	configYaml := os.Getenv(configYamlEnvVarName)

	if _, err := os.Stat(configPathVar); err != nil {
		log.Fatalf("Unable to find the configuration file (%s) ensure %s environment variable is set properly", configPathVar, configEnvVarName)
	}

	if configYaml != "" {
		log.Printf("Both %s and %s were specified. Using %s environment variable value %s for this session", configEnvVarName, configYamlEnvVarName, configEnvVarName, configPathVar)
	}

	if !inputFlags.configs.contains(configPathVar) {
		_ = inputFlags.configs.Set(configPathVar)
	}

	checkRequiredEnvVars(inputFlags.configs.values)
}

// Config priority queue (highest to lowest): '--config' flag, SPLUNK_CONFIG env var,
// SPLUNK_CONFIG_YAML env var, default config path.
func checkConfig(inputFlags flags) {
	configPathVar := os.Getenv(configEnvVarName)
	configYaml := os.Getenv(configYamlEnvVarName)

	switch {
	case len(inputFlags.configs.values) != 0:
		checkInputConfigs(inputFlags)
		log.Printf("Set config to %v", inputFlags.configs.String())
	case configPathVar != "":
		checkConfigPathEnvVar(inputFlags)
		log.Printf("Set config to %v", configPathVar)
	case configYaml != "":
		log.Printf("Using environment variable %s for configuration", configYamlEnvVarName)
	default:
		defaultConfigPath := getExistingDefaultConfigPath()
		inputFlags.configs.Set(defaultConfigPath)
		checkRequiredEnvVars(inputFlags.configs.values)
		log.Printf("Set config to %v", defaultConfigPath)
	}
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

func checkRequiredEnvVars(paths []string) {
	// Check environment variables required by default configuration.
	for _, path := range paths {
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
}

// Validate and set the memory ballast
func setMemoryBallast(inputFlags flags, memTotalSizeMiB int) int {
	// Check if deprecated memory ballast flag was passed, if so, ensure the env variable for memory ballast is set.
	// Then set memory ballast and limit properly
	if inputFlags.memBallastSizeMib != defaultUndeclaredFlag {
		if os.Getenv(ballastEnvVarName) != "" {
			log.Fatalf("Both %v and '--mem-ballast-size-mib' were specified, but only one is allowed", ballastEnvVarName)
		}
		_ = os.Setenv(ballastEnvVarName, strconv.Itoa(inputFlags.memBallastSizeMib))
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

// configLocations returns a config location based on provided environment variables and --config argument.
func configLocations(inputFlags flags) []string {
	var configPaths []string
	if configPaths = inputFlags.configs.values; len(configPaths) == 0 {
		if configEnvVal := os.Getenv(configEnvVarName); len(configEnvVal) != 0 {
			configPaths = []string{"file:" + configEnvVal}
		}
	}

	configYaml := os.Getenv(configYamlEnvVarName)

	switch {
	case len(configPaths) == 0 && configYaml != "":
		return []string{"env:" + configYamlEnvVarName}
	case len(configPaths) == 0:
		return []string{""}
	default:
		return configPaths
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
