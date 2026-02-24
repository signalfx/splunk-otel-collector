// Copyright Splunk, Inc.
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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	schemeFile := flag.String("scheme", "", "Path to the XML scheme file (required)")
	globalSettings := flag.String("global-settings", "", "Path to the global settings file (required)")
	inputName := flag.String("name", "", "Name of the modular input (required)")
	assetsDir := flag.String("assets", "", "Path to the assets directory (required)")
	flag.Parse()

	if *schemeFile == "" || *globalSettings == "" || *inputName == "" || *assetsDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*schemeFile, *globalSettings, *inputName, *assetsDir); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Successfully generated inputs.conf and inputs.conf.spec")
}

func run(schemeFile, globalSettingsFile, inputName, assetsDir string) error {
	// Parse the scheme XML
	scheme, err := parseSchemeXML(schemeFile)
	if err != nil {
		return fmt.Errorf("failed to parse scheme XML: %w", err)
	}

	// Read global settings
	globalSettings, err := os.ReadFile(globalSettingsFile)
	if err != nil {
		return fmt.Errorf("failed to read global settings file: %w", err)
	}

	// Generate inputs.conf
	inputsConf, err := generateInputsConf(scheme, string(globalSettings), inputName)
	if err != nil {
		return fmt.Errorf("failed to generate inputs.conf: %w", err)
	}

	// Generate inputs.conf.spec
	inputsConfSpec, err := generateInputsConfSpec(scheme, inputName)
	if err != nil {
		return fmt.Errorf("failed to generate inputs.conf.spec: %w", err)
	}

	// Write inputs.conf
	defaultDir := assetsDir + "/default"
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		return fmt.Errorf("failed to create default directory: %w", err)
	}
	if err := os.WriteFile(defaultDir+"/inputs.conf", []byte(inputsConf), 0644); err != nil {
		return fmt.Errorf("failed to write inputs.conf: %w", err)
	}

	// Write inputs.conf.spec
	readmeDir := assetsDir + "/README"
	if err := os.MkdirAll(readmeDir, 0755); err != nil {
		return fmt.Errorf("failed to create README directory: %w", err)
	}
	if err := os.WriteFile(readmeDir+"/inputs.conf.spec", []byte(inputsConfSpec), 0644); err != nil {
		return fmt.Errorf("failed to write inputs.conf.spec: %w", err)
	}

	return nil
}
