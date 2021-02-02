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
	"runtime"
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/version"
)

const (
	ballastEnvVarName     = "SPLUNK_BALLAST_SIZE_MIB"
	configEnvVarName      = "SPLUNK_CONFIG"
	memLimitEnvVarName    = "SPLUNK_MEMORY_LIMIT_PERCENTAGE"
	memLimitMiBEnvVarName = "SPLUNK_MEMORY_LIMIT_MIB"
	memSpikeEnvVarName    = "SPLUNK_MEMORY_SPIKE_PERCENTAGE"
	memSpikeMiBEnvVarName = "SPLUNK_MEMORY_SPIKE_MIB"
	memTotalEnvVarName    = "SPLUNK_MEMORY_TOTAL_MIB"
	realmEnvVarName       = "SPLUNK_REALM"
	tokenEnvVarName       = "SPLUNK_ACCESS_TOKEN"

	defaultDockerSAPMLinuxConfig    = "/etc/otel/collector/splunk_config_linux.yaml"
	defaultDockerSAPMNonLinuxConfig = "/etc/otel/collector/splunk_config_non_linux.yaml"
	defaultDockerOTLPLinuxConfig    = "/etc/otel/collector/otlp_config_linux.yaml"
	defaultDockerOTLPNonLinuxConfig = "/etc/otel/collector/otlp_config_non_linux.yaml"
	defaultLocalSAPMLinuxConfig     = "cmd/otelcol/config/collector/splunk_config_linux.yaml"
	defaultLocalSAPMNonLinuxConfig  = "cmd/otelcol/config/collector/splunk_config_non_linux.yaml"
	defaultLocalOTLPLinuxConfig     = "cmd/otelcol/config/collector/otlp_config_linux.yaml"
	defaultLocalOTLPNonLinuxConfig  = "cmd/otelcol/config/collector/otlp_config_non_linux.yaml"
	defaultMemoryBallastPercentage  = 50
	defaultMemoryLimitPercentage    = 90
	defaultMemoryLimitMaxMiB        = 2048
	defaultMemorySpikePercentage    = 25
	defaultMemorySpikeMaxMiB        = 2048
)

func main() {
	// TODO: Use same format as the collector
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Check if the total memory is specified via the env var.
	memTotalEnvVarVal := os.Getenv(memTotalEnvVarName)
	memTotalSizeMiB := 0
	if memTotalEnvVarVal != "" {
		// Check if it is a numeric value.
		val, err := strconv.Atoi(memTotalEnvVarVal)
		if err != nil {
			log.Fatalf("Expected a number in %s env variable but got %s", memTotalEnvVarName, memTotalEnvVarVal)
		}
		if 10 > val {
			log.Fatalf("Expected a number greater than 10 for %s env variable but got %s", memTotalEnvVarName, memTotalEnvVarVal)
		}
		memTotalSizeMiB = val
	}

	// Check runtime parameters
	// Runtime parameters take priority over environment variables
	// Runtime parameters are not validated
	args := os.Args[1:]
	if !contains(args, "--mem-ballast-size-mib") {
		useMemorySizeFromEnvVar(memTotalSizeMiB)
	} else {
		log.Printf("Ballast CLI argument found, ignoring %s if set", ballastEnvVarName)
	}
	if !contains(args, "--config") {
		useConfigFromEnvVar()
	} else {
		log.Printf("Config CLI argument found, please ensure memory_limiter settings are correct")
	}
	if runtime.GOOS == "linux" {
		useMemorySettingsPercentageFromEnvVar()
	} else {
		useMemorySettingsMiBFromEnvVar(memTotalSizeMiB)
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

	if err := run(service.Parameters{ApplicationStartInfo: info, Factories: factories}); err != nil {
		log.Fatal(err)
	}
}

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

func useMemorySizeFromEnvVar(memTotalSizeMiB int) {
	// Check if the ballast is specified via the env var.
	ballastSize := os.Getenv(ballastEnvVarName)
	if ballastSize != "" {
		// Check if it is a numeric value.
		val, err := strconv.Atoi(ballastSize)
		if err != nil {
			log.Fatalf("Expected a number in %s env variable but got %s", ballastEnvVarName, ballastSize)
		}
		if 0 > val {
			log.Fatalf("Expected a number greater than 0 for %s env variable but got %s", ballastEnvVarName, ballastSize)
		}

		// Inject the command line flag that controls the ballast size.
		os.Args = append(os.Args, "--mem-ballast-size-mib="+ballastSize)
	} else if memTotalSizeMiB > 0 {
		halfMem := strconv.Itoa(memTotalSizeMiB * defaultMemoryBallastPercentage / 100)
		log.Printf("Set ballast to %s MiB", halfMem)
		// Inject the command line flag that controls the ballast size.
		os.Args = append(os.Args, "--mem-ballast-size-mib="+halfMem)
		os.Setenv(ballastEnvVarName, halfMem)
	}
}

func useConfigFromEnvVar() {
	// Check if the config is specified via the env var.
	config := os.Getenv(configEnvVarName)
	// If not attempt to use a default config; supports Docker and local
	if config == "" {
		if runtime.GOOS == "linux" {
			_, err := os.Stat(defaultDockerSAPMLinuxConfig)
			if err == nil {
				config = defaultDockerSAPMLinuxConfig
			}
			_, err = os.Stat(defaultLocalSAPMLinuxConfig)
			if err == nil {
				config = defaultLocalSAPMLinuxConfig
			}
			if config == "" {
				log.Fatalf("Unable to find the default configuration file, ensure %s environment variable is set properly", configEnvVarName)
			}
		} else {
			_, err := os.Stat(defaultDockerSAPMNonLinuxConfig)
			if err == nil {
				config = defaultDockerSAPMNonLinuxConfig
			}
			_, err = os.Stat(defaultLocalSAPMNonLinuxConfig)
			if err == nil {
				config = defaultLocalSAPMNonLinuxConfig
			}
			if config == "" {
				log.Fatalf("Unable to find the default configuration file, ensure %s environment variable is set properly", configEnvVarName)
			}
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
		defaultDockerSAPMLinuxConfig,
		defaultDockerSAPMNonLinuxConfig,
		defaultDockerOTLPLinuxConfig,
		defaultDockerOTLPNonLinuxConfig,
		defaultLocalSAPMLinuxConfig,
		defaultLocalSAPMNonLinuxConfig,
		defaultLocalOTLPLinuxConfig,
		defaultLocalOTLPNonLinuxConfig:
		// The following environment variables are required.
		// If any are missing stop here.
		requiredEnvVars := []string{realmEnvVarName, tokenEnvVarName}
		for _, v := range requiredEnvVars {
			if len(os.Getenv(v)) == 0 {
				log.Printf("Usage: %s=12345 %s=us0 %s=1024 %s", tokenEnvVarName, realmEnvVarName, memTotalEnvVarName, os.Args[0])
				log.Fatalf("ERROR: Missing environment variable %s", v)
			}
		}
		// Needed for backwards compatibility
		if len(os.Getenv(memTotalEnvVarName)) == 0 && len(os.Getenv(ballastEnvVarName)) == 0 {
			log.Printf("Usage: %s=12345 %s=us0 %s=1024 %s", tokenEnvVarName, realmEnvVarName, memTotalEnvVarName, os.Args[0])
			log.Fatalf("ERROR: Missing environment variable %s", memTotalEnvVarName)
		}
	}

	// Inject the command line flag that controls the configuration.
	os.Args = append(os.Args, "--config="+config)
}

func checkMemorySettingsMiBFromEnvVar(envVar string, memTotalSizeMiB int) int {
	// Check if the memory limit is specified via the env var
	// Ensure memory limit is valid
	var envVarResult int = 0
	envVarVal := os.Getenv(envVar)
	switch {
	case envVarVal != "":
		// Check if it is a numeric value.
		val, err := strconv.Atoi(envVarVal)
		if err != nil {
			log.Fatalf("Expected a number in %s env variable but got %s", envVar, envVarVal)
		}
		if 0 > val {
			log.Fatalf("Expected a number greater than 0 for %s env variable but got %s", envVar, envVarVal)
		}
		envVarResult = val
	case memTotalSizeMiB > 0:
		break
	default:
		log.Printf("Usage: %s=12345 %s=us0 %s=684 %s=1024 %s=256 %s", tokenEnvVarName, realmEnvVarName, ballastEnvVarName, memLimitMiBEnvVarName, memSpikeMiBEnvVarName, os.Args[0])
		log.Fatalf("ERROR: Missing environment variable %s", envVar)
	}
	return envVarResult
}

func useMemorySettingsMiBFromEnvVar(memTotalSizeMiB int) {
    // Check if memory limit is specified via environment variable
	memLimit := checkMemorySettingsMiBFromEnvVar(memLimitMiBEnvVarName, memTotalSizeMiB)
    // Use if set, otherwise memory total size must be specified
	if memLimit == 0 {
		if memTotalSizeMiB == 0 {
			panic("PANIC: Both memory limit MiB and memory total size are set to zero. This should never happen.")
		}
        // If not set, compute based on memory total size specified
        // and default memory limit percentage const
		memLimitMiB := memTotalSizeMiB * defaultMemoryLimitPercentage / 100
        // The memory limit should be set to defaultMemoryLimitPercentage of total memory
        // while reserving a maximum of defaultMemoryLimitMaxMiB of memory.
		if (memTotalSizeMiB - memLimitMiB) < defaultMemoryLimitMaxMiB {
			memLimit = memLimitMiB
		} else {
			memLimit = (memTotalSizeMiB - defaultMemoryLimitMaxMiB)
		}
		log.Printf("Set memory limit to %d MiB", memLimit)
	}
    // Check if memory spike is specified via environment variable
	memSpike := checkMemorySettingsMiBFromEnvVar(memSpikeMiBEnvVarName, memTotalSizeMiB)
    // Use if set, otherwise memory total size must be specified
	if memSpike == 0 {
		if memTotalSizeMiB == 0 {
			panic("PANIC: Both memory limit MiB and memory total size are set to zero. This should never happen.")
		}
        // If not set, compute based on memory total size specified
        // and default memory spike percentage const
		memSpikeMiB := memTotalSizeMiB * defaultMemorySpikePercentage / 100
        // The memory spike should be set to defaultMemorySpikePercentage of total memory
        // while specifying a maximum of defaultMemorySpikeMaxMiB of memory.
		if memSpikeMiB < defaultMemorySpikeMaxMiB {
			memSpike = memSpikeMiB
		} else {
			memSpike = defaultMemorySpikeMaxMiB
		}
		log.Printf("Set memory spike limit to %d MiB", memSpike)
	}
	setMemorySettingsToEnvVar(memLimit, memLimitMiBEnvVarName, memSpike, memSpikeMiBEnvVarName)
}

func checkMemorySettingsPercentageFromEnvVar(envVar string, defaultVal int) int {
	// Check if the memory limit is specified via the env var
	// Ensure memory limit is valid
	var envVarResult int
	envVarVal := os.Getenv(envVar)
	if envVarVal != "" {
		// Check if it is a numeric value.
		val, err := strconv.Atoi(envVarVal)
		if err != nil {
			log.Fatalf("Expected a number in %s env variable but got %s", envVar, envVarVal)
		}
		if 0 > val || val > 100 {
			log.Fatalf("Expected a number in the range 0-100 for %s env variable but got %s", envVar, envVarVal)
		}
		envVarResult = val
	} else {
		envVarResult = defaultVal
	}
	return envVarResult
}

func useMemorySettingsPercentageFromEnvVar() {
	memLimit := checkMemorySettingsPercentageFromEnvVar(memLimitEnvVarName, defaultMemoryLimitPercentage)
	memSpike := checkMemorySettingsPercentageFromEnvVar(memSpikeEnvVarName, defaultMemorySpikePercentage)
	setMemorySettingsToEnvVar(memLimit, memLimitEnvVarName, memSpike, memSpikeEnvVarName)
}

func setMemorySettingsToEnvVar(limit int, limitName string, spike int, spikeName string) {
	// Ensure spike and limit are valid
	if spike >= limit {
		log.Fatalf("%s env variable must be less than %s env variable but got %d and %d respectively", spikeName, limitName, spike, limit)
	}

	// Set memory environment variables
	os.Setenv(limitName, strconv.Itoa(limit))
	os.Setenv(spikeName, strconv.Itoa(spike))
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
