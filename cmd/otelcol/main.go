// Copyright 2020 Splunk, Inc.
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
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
	"github.com/signalfx/splunk-otel-collector/internal/configsources"
	"github.com/signalfx/splunk-otel-collector/internal/version"
)

const (
	ballastEnvVarName     = "SPLUNK_BALLAST_SIZE_MIB"
	configEnvVarName      = "SPLUNK_CONFIG"
	memLimitMiBEnvVarName = "SPLUNK_MEMORY_LIMIT_MIB"
	memTotalEnvVarName    = "SPLUNK_MEMORY_TOTAL_MIB"
	realmEnvVarName       = "SPLUNK_REALM"
	tokenEnvVarName       = "SPLUNK_ACCESS_TOKEN"

	defaultDockerSAPMConfig        = "/etc/otel/collector/gateway_config.yaml"
	defaultDockerOTLPConfig        = "/etc/otel/collector/otlp_config_linux.yaml"
	defaultLocalSAPMConfig         = "cmd/otelcol/config/collector/gateway_config.yaml"
	defaultLocalOTLPConfig         = "cmd/otelcol/config/collector/otlp_config_linux.yaml"
	defaultMemoryBallastPercentage = 33
	defaultMemoryLimitPercentage   = 90
	defaultMemoryLimitMaxMiB       = 2048
	defaultMemoryTotalMiB          = 512
)

func main() {
	// TODO: Use same format as the collector
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	args := os.Args[1:]
	if !contains(args, "-h") && !contains(args, "--help") {
		checkRuntimeParams()
	}

	factories, err := components.Get()
	if err != nil {
		log.Fatalf("failed to build default components: %v", err)
	}

	info := component.ApplicationStartInfo{
		ExeName:  "otelcol",
		LongName: "OpenTelemetry Collector",
		Version:  version.Version,
		GitHash:  version.GitHash,
	}

	parserProvider := configprovider.NewConfigSourceParserProvider(
		zap.NewNop(), // The service logger is not available yet, setting it to NoP.
		info,
		configsources.Get()...,
	)
	serviceParams := service.Parameters{
		ApplicationStartInfo: info,
		Factories:            factories,
		ParserProvider:       parserProvider,
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
func getKeyValue(args []string, argName string) string {
	val := ""
	for i, arg := range args {
		switch {
		case strings.HasPrefix(arg, argName+"="):
			s := strings.Split(arg, "=")
			val = s[1]
		case arg == argName:
			i++
			val = args[i]
		}
	}
	return val
}

// Check runtime parameters
// Runtime parameters take priority over environment variables
// Config and ballast flags are checked
// Config and all memory env vars are checked
func checkRuntimeParams() {
	args := os.Args[1:]
	config := ""

	// Check if config flag was passed
	// If so, ensure config env var is not set
	// Then set config properly
	cliConfig := getKeyValue(args, "--config")
	if cliConfig != "" {
		config = os.Getenv(configEnvVarName)
		if config != "" {
			log.Fatalf("Both %v and '--config' were specified, but only one is allowed", configEnvVarName)
		}
		os.Setenv(configEnvVarName, cliConfig)
	}
	setConfig()

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
	ballastSize := getKeyValue(args, "--mem-ballast-size-mib")
	if ballastSize != "" {
		config = os.Getenv(ballastEnvVarName)
		if config != "" {
			log.Fatalf("Both %v and '--config' were specified, but only one is allowed", ballastEnvVarName)
		}
		os.Setenv(ballastEnvVarName, ballastSize)
	}
	setMemoryBallast(memTotalSizeMiB)
	setMemoryLimit(memTotalSizeMiB)
}

// Validate and set the configuration
func setConfig() {
	// Check if the config is specified via the env var.
	config := os.Getenv(configEnvVarName)
	// If not attempt to use a default config; supports Docker and local
	if config == "" {
		_, err := os.Stat(defaultDockerSAPMConfig)
		if err == nil {
			config = defaultDockerSAPMConfig
		}
		_, err = os.Stat(defaultLocalSAPMConfig)
		if err == nil {
			config = defaultLocalSAPMConfig
		}
		if config == "" {
			log.Fatalf("Unable to find the default configuration file, ensure %s environment variable is set properly", configEnvVarName)
		}
	} else {
		// Check if file exists.
		_, err := os.Stat(config)
		if err != nil {
			log.Fatalf("Unable to find the configuration file (%s) ensure %s environment variable is set properly", config, configEnvVarName)
		}
	}

	switch config {
	case
		defaultDockerSAPMConfig,
		defaultDockerOTLPConfig,
		defaultLocalSAPMConfig,
		defaultLocalOTLPConfig:
		// The following environment variables are required.
		// If any are missing stop here.
		requiredEnvVars := []string{realmEnvVarName, tokenEnvVarName}
		for _, v := range requiredEnvVars {
			if len(os.Getenv(v)) == 0 {
				log.Printf("Usage: %s=12345 %s=us0 %s", tokenEnvVarName, realmEnvVarName, os.Args[0])
				log.Fatalf("ERROR: Missing environment variable %s", v)
			}
		}
	}

	args := os.Args[1:]
	if !contains(args, "--config") {
		// Inject the command line flag that controls the configuration.
		os.Args = append(os.Args, "--config="+config)
	}
	log.Printf("Set config to %v", config)
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
		// The memory limit should be set to defaultMemoryLimitPercentage of total memory
		// while reserving a maximum of defaultMemoryLimitMaxMiB of memory.
		if (memTotalSizeMiB - memLimit) > defaultMemoryLimitMaxMiB {
			memLimit = defaultMemoryLimitMaxMiB
		}
	} else {
		memLimit, _ = strconv.Atoi(memoryLimit)
	}

	// Validate memoryLimit is sane
	args := os.Args[1:]
	b := getKeyValue(args, "--mem-ballast-size-mib")
	ballastSize, _ := strconv.Atoi(b)
	if (ballastSize * 2) >= memLimit {
		log.Fatalf("Memory limit (%v) is less than 2x ballast (%v). Increase memory limit or decrease ballast size.", memLimit, ballastSize)
	}

	os.Setenv(memLimitMiBEnvVarName, strconv.Itoa(memLimit))
	log.Printf("Set memory limit to %d MiB", memLimit)
}

func runInteractive(params service.Parameters) error {
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
