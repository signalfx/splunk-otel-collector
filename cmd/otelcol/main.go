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
	realmEnvVarName       = "SPLUNK_REALM"
	tokenEnvVarName       = "SPLUNK_ACCESS_TOKEN"

	defaultSAPMLinuxConfig       = "/etc/otel/collector/splunk_config_linux.yaml"
	defaultSAPMNonLinuxConfig    = "/etc/otel/collector/splunk_config_non_linux.yaml"
	defaultOTLPLinuxConfig       = "/etc/otel/collector/otlp_config_linux.yaml"
	defaultOTLPNonLinuxConfig    = "/etc/otel/collector/otlp_config_non_linux.yaml"
	defaultMemoryLimitPercentage = 90
	defaultMemorySpikePercentage = 25
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Check runtime parameters
	// Note that runtime parameters take priority over environment variables
	args := os.Args[1:]
	if len(args) == 0 {
		useBallastSizeFromEnvVar()
		useConfigFromEnvVar()
	}

	// If GOOS is not linux then a custom configuration needs to be supplied
	// A non-linux default config is built-in and requires both
	// SPLUNK_MEMORY_LIMIT_MIB and SPLUNK_MEMORY_SPIKE_MIB to be set
	if runtime.GOOS != "linux" {
		config := os.Getenv(configEnvVarName)
		if config == defaultSAPMLinuxConfig || config == defaultOTLPLinuxConfig {
			log.Fatalf("For non-linux systems the %s must be specified. Consider using splunk_config_non-linux.yaml.", configEnvVarName)
		} else {
			useMemorySettingsMiBFromEnvVar()
		}
	} else {
		useMemorySettingsPercentageFromEnvVar()
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

func useBallastSizeFromEnvVar() {
	// Check if the ballast is specified via the env var.
	ballastSize := os.Getenv(ballastEnvVarName)
	if ballastSize != "" {
		// Check if it is a numeric value.
		_, err := strconv.Atoi(ballastSize)
		if err != nil {
			log.Fatalf("Expected a number in %s env variable but got %s", ballastEnvVarName, ballastSize)
		}

		// Inject the command line flag that controls the ballast size.
		os.Args = append(os.Args, "--mem-ballast-size-mib="+ballastSize)
	}
}

func useConfigFromEnvVar() {
	// Check if the config is specified via the env var.
	config := os.Getenv(configEnvVarName)
	if config == "" {
		if runtime.GOOS == "linux" {
			config = defaultSAPMLinuxConfig
		} else {
			config = defaultSAPMNonLinuxConfig
		}
	}

	// Check if file exists.
	_, err := os.Stat(config)
	if os.IsNotExist(err) {
		log.Fatalf("Unable to find the configuration file (%s) ensure %s environment variable is set properly", config, configEnvVarName)
	}

	switch config {
	case
		defaultSAPMLinuxConfig,
		defaultSAPMNonLinuxConfig,
		defaultOTLPLinuxConfig,
		defaultOTLPNonLinuxConfig:
		// The following environment variables are required.
		// If any are missing stop here.
		requiredEnvVars := []string{ballastEnvVarName, realmEnvVarName, tokenEnvVarName}
		for _, v := range requiredEnvVars {
			if len(os.Getenv(v)) == 0 {
				log.Printf("Usage: %s=12345 %s=us0 %s=684 %s", tokenEnvVarName, realmEnvVarName, ballastEnvVarName, os.Args[0])
				log.Fatalf("ERROR: Missing environment variable %s", v)
			}
		}
	}

	// Inject the command line flag that controls the configuration.
	os.Args = append(os.Args, "--config="+config)
}

func checkMemorySettingsMiBFromEnvVar(envVar string) int {
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
		if 0 > val {
			log.Fatalf("Expected a number greater than 0 for %s env variable but got %s", envVar, envVarVal)
		}
		envVarResult = val
	} else {
		log.Printf("Usage: %s=12345 %s=us0 %s=684 %s=1024 %s=256 %s", tokenEnvVarName, realmEnvVarName, ballastEnvVarName, memLimitMiBEnvVarName, memSpikeMiBEnvVarName, os.Args[0])
		log.Fatalf("ERROR: Missing environment variable %s", envVar)
	}
	return envVarResult
}

func useMemorySettingsMiBFromEnvVar() {
	memLimit := checkMemorySettingsMiBFromEnvVar(memLimitMiBEnvVarName)
	memSpike := checkMemorySettingsMiBFromEnvVar(memSpikeMiBEnvVarName)
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
