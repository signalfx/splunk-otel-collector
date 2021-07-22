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
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"
	"go.opentelemetry.io/collector/service/parserprovider"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
	"github.com/signalfx/splunk-otel-collector/internal/configsources"
	"github.com/signalfx/splunk-otel-collector/internal/version"
)

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

	if !contains(os.Args[1:], "-h") && !contains(os.Args[1:], "--help") {
		checkRuntimeParams()
	}

	// Allow dumping configuration locally by default
	// Used by support bundle script
	os.Setenv(configServerEnabledEnvVar, "true")

	factories, err := components.Get()
	if err != nil {
		log.Fatalf("failed to build default components: %v", err)
	}

	info := component.BuildInfo{
		Command: "otelcol",
		Version: version.Version,
	}

	parserProvider := configprovider.NewConfigSourceParserProvider(
		newBaseParserProvider(),
		zap.NewNop(), // The service logger is not available yet, setting it to NoP.
		info,
		configsources.Get()...,
	)
	serviceParams := service.CollectorSettings{
		BuildInfo:      info,
		Factories:      factories,
		ParserProvider: parserProvider,
	}

	if err := run(serviceParams); err != nil {
		log.Fatal(err)
	}
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
	memTotalSizeMiB := defaultMemoryTotalMiB
	// Check if the total memory is specified via the env var
	memTotalEnvVarVal := os.Getenv(memTotalEnvVarName)
	// If so, validate and change total memory
	if memTotalEnvVarVal != "" {
		// Check if it is a numeric value.
		val, err := strconv.Atoi(memTotalEnvVarVal)
		if err != nil {
			log.Fatalf("Expected a number in %s env variable but got %s", memTotalEnvVarName, memTotalEnvVarVal)
		}
		// Ensure number is above some threshold
		if 99 > val {
			log.Fatalf("Expected a number greater than 99 for %s env variable but got %s", memTotalEnvVarName, memTotalEnvVarVal)
		}
		memTotalSizeMiB = val
	}

	// Check if memory ballast flag was passed
	// If so, ensure memory ballast env var is not set
	// Then set memory ballast and limit properly
	_, ballastSize := getKeyValue(os.Args[1:], "--mem-ballast-size-mib")
	if ballastSize != "" {
		if os.Getenv(ballastEnvVarName) != "" {
			log.Fatalf("Both %v and '--mem-ballast-size-mib' were specified, but only one is allowed", ballastEnvVarName)
		}
		os.Setenv(ballastEnvVarName, ballastSize)
	}
	setMemoryBallast(memTotalSizeMiB)
	setMemoryLimit(memTotalSizeMiB)
}

// Sets flag '--config' to specified env var SPLUNK_CONFIG, if the flag not specified.
// Logs a message and returns if env var SPLUNK_CONFIG_YAML specified, and '--config' and SPLUNK_CONFIG no specified.
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
func setMemoryBallast(memTotalSizeMiB int) {
	// Check if the memory ballast is specified via the env var
	ballastSize := os.Getenv(ballastEnvVarName)
	// If so, validate and set properly
	if ballastSize != "" {
		// Check if it is a numeric value.
		val, err := strconv.Atoi(ballastSize)
		if err != nil {
			log.Fatalf("Expected a number in %s env variable but got %s", ballastEnvVarName, ballastSize)
		}
		if 33 > val {
			log.Fatalf("Expected a number greater than 33 for %s env variable but got %s", ballastEnvVarName, ballastSize)
		}
	} else {
		ballastSize = strconv.Itoa(memTotalSizeMiB * defaultMemoryBallastPercentage / 100)
		os.Setenv(ballastEnvVarName, ballastSize)
	}

	args := os.Args[1:]
	if !contains(args, "--mem-ballast-size-mib") {
		// Inject the command line flag that controls the ballast size.
		os.Args = append(os.Args, "--mem-ballast-size-mib="+ballastSize)
	}
	log.Printf("Set ballast to %s MiB", ballastSize)
}

// Validate and set the memory limit
func setMemoryLimit(memTotalSizeMiB int) {
	memLimit := 0
	// Check if the memory limit is specified via the env var
	memoryLimit := os.Getenv(memLimitMiBEnvVarName)
	// If not, calculate it from memTotalSizeMiB
	if memoryLimit == "" {
		memLimit = memTotalSizeMiB * defaultMemoryLimitPercentage / 100
	} else {
		memLimit, _ = strconv.Atoi(memoryLimit)
	}

	// Validate memoryLimit is sane
	args := os.Args[1:]
	_, b := getKeyValue(args, "--mem-ballast-size-mib")
	ballastSize, _ := strconv.Atoi(b)
	if (ballastSize * 2) > memLimit {
		log.Fatalf("Memory limit (%v) is less than 2x ballast (%v). Increase memory limit or decrease ballast size.", memLimit, ballastSize)
	}

	os.Setenv(memLimitMiBEnvVarName, strconv.Itoa(memLimit))
	log.Printf("Set memory limit to %d MiB", memLimit)
}

// Returns a ParserProvider that reads configuration YAML from an environment variable when applicable.
func newBaseParserProvider() parserprovider.ParserProvider {
	_, configPathFlag := getKeyValue(os.Args[1:], "--config")
	configPathVar := os.Getenv(configEnvVarName)
	configYaml := os.Getenv(configYamlEnvVarName)

	if configPathFlag == "" && configPathVar == "" && configYaml != "" {
		return parserprovider.NewInMemory(bytes.NewBufferString(configYaml))
	}

	return parserprovider.Default()
}

func runInteractive(params service.CollectorSettings) error {
	app, err := service.New(params)
	if err != nil {
		return fmt.Errorf("failed to construct the application: %w", err)
	}

	err = app.Run()
	if err != nil {
		return fmt.Errorf("application run finished with error: %w", err)
	}

	return nil
}
