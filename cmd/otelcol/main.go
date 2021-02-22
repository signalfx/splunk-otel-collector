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

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/version"
)

const (
	ballastEnvVarName     = "SPLUNK_BALLAST_SIZE_MIB"
	configEnvVarName      = "SPLUNK_CONFIG"
	memLimitMiBEnvVarName = "SPLUNK_MEMORY_LIMIT_MIB"
	memTotalEnvVarName    = "SPLUNK_MEMORY_TOTAL_MIB"
	realmEnvVarName       = "SPLUNK_REALM"
	tokenEnvVarName       = "SPLUNK_ACCESS_TOKEN"

	defaultDockerSAPMConfig        = "/etc/otel/collector/splunk_config_linux.yaml"
	defaultDockerOTLPConfig        = "/etc/otel/collector/otlp_config_linux.yaml"
	defaultLocalSAPMConfig         = "cmd/otelcol/config/collector/splunk_config_linux.yaml"
	defaultLocalOTLPConfig         = "cmd/otelcol/config/collector/otlp_config_linux.yaml"
	defaultMemoryBallastPercentage = 33
	defaultMemoryTotalMiB          = 512
)

func main() {
	// TODO: Use same format as the collector
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	args := os.Args[1:]
	if !contains(args, "-h") && !contains(args, "--help") {
		checkSetEnvVars()
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

func checkSetEnvVars() {
	// Check if the total memory is specified via the env var.
	memTotalEnvVarVal := os.Getenv(memTotalEnvVarName)
	memTotalSizeMiB := defaultMemoryTotalMiB
	if memTotalEnvVarVal != "" {
		// Check if it is a numeric value.
		val, err := strconv.Atoi(memTotalEnvVarVal)
		if err != nil {
			log.Fatalf("Expected a number in %s env variable but got %s", memTotalEnvVarName, memTotalEnvVarVal)
		}
		if 99 > val {
			log.Fatalf("Expected a number greater than 99 for %s env variable but got %s", memTotalEnvVarName, memTotalEnvVarVal)
		}
		memTotalSizeMiB = val
	}
	log.Printf("Set memory limit to %d MiB", memTotalSizeMiB)
	os.Setenv(memLimitMiBEnvVarName, strconv.Itoa(memTotalSizeMiB))

	// Check runtime parameters
	// Runtime parameters take priority over environment variables
	// Runtime parameters are not validated
	args := os.Args[1:]
	if !contains(args, "--config") {
		useConfigFromEnvVar()
	} else {
		log.Printf("Config CLI argument found, please ensure memory_limiter settings are correct")
	}
	if !contains(args, "--mem-ballast-size-mib") {
		useMemorySizeFromEnvVar(memTotalSizeMiB)
	} else {
		log.Printf("Ballast CLI argument found, ignoring %s if set", ballastEnvVarName)
	}
}

func useConfigFromEnvVar() {
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

	// Inject the command line flag that controls the configuration.
	os.Args = append(os.Args, "--config="+config)
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
		if 33 > val {
			log.Fatalf("Expected a number greater than 33 for %s env variable but got %s", ballastEnvVarName, ballastSize)
		}

		// Inject the command line flag that controls the ballast size.
		os.Args = append(os.Args, "--mem-ballast-size-mib="+ballastSize)
	} else {
		ballastSize = strconv.Itoa(memTotalSizeMiB * defaultMemoryBallastPercentage / 100)
		// Inject the command line flag that controls the ballast size.
		os.Args = append(os.Args, "--mem-ballast-size-mib="+ballastSize)
		os.Setenv(ballastEnvVarName, ballastSize)
	}
	log.Printf("Set ballast to %s MiB", ballastSize)
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
