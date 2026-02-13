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

package modularinput

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

var (
	ErrQueryMode = errors.New("modular input called in query mode")

	// Function variables to facilitate testing
	setEnvFn                 = os.Setenv
	isParentProcessSplunkdFn = isParentProcessSplunkd
)

// HandleLaunchAsTA handles the launch of the collector as a Splunk TA modular input.
// It checks if the collector is running in modular input mode and processes the input XML
// to set environment variables from the configuration stanza.
// Returns an error if the launch fails, ErrQueryMode if running in query mode,
// or nil if not running in modular input mode or on success.
func HandleLaunchAsTA(args []string, stdin io.Reader, configStanzaName string) error {
	isModularInput, isQueryMode := isModularInputMode(args)
	if !isModularInput {
		return nil
	}

	if isQueryMode {
		// Query modes (scheme/validate) are empty no-ops for now.
		// Do not write anything to stdout, just signal it to the caller
		// with a specific error.
		return ErrQueryMode
	}

	input, err := ReadXML(stdin)
	if err != nil {
		return fmt.Errorf("launch as TA failed to read modular input XML from stdin: %w", err)
	}

	var configStanza Stanza
	for _, stanza := range input.Configuration.Stanza {
		if stanza.Name == configStanzaName {
			configStanza = stanza
			break
		}
	}

	for _, param := range configStanza.Param {
		envVarName := strings.ToUpper(param.Name)
		envVarValue := os.ExpandEnv(param.Value)
		err := setEnvFn(envVarName, envVarValue)
		if err != nil {
			return fmt.Errorf("launch as TA failed to set environment variable '%s': %w", envVarName, err)
		}
	}

	return nil
}

func isModularInputMode(args []string) (isModularInput, isQueryMode bool) {
	// SPLUNK_HOME must be defined if this is running as a modular input.
	_, hasSplunkHome := os.LookupEnv("SPLUNK_HOME")
	if !hasSplunkHome {
		return false, false
	}

	// TA v1 is a special case of the collector being launched as a modular input
	// with TA specific behavior being handled by scripts. Use the SPLUNK_OTEL_TA_HOME
	// environment variable to determine if this is running as a TA v1 modular input.
	_, isTAv1Launch := os.LookupEnv("SPLUNK_OTEL_TA_HOME")
	if isTAv1Launch {
		// TA v1, let the scripts handle the TA specific behavior
		return false, false
	}

	// Check if the parent process is splunkd
	if !isParentProcessSplunkdFn() {
		return false, false
	}

	// This is running as a modular input
	if len(args) == 2 && (args[1] == "--scheme" || args[1] == "--validate-arguments") {
		return true, true
	}

	return true, false
}

// isParentProcessSplunkd checks if the parent process name is splunkd (Linux) or splunkd.exe (Windows)
// error cases are a good indication that the parent is not splunkd, so log errors and return false
// in those cases.
func isParentProcessSplunkd() bool {
	// Get parent process ID
	ppid := os.Getppid()
	parentProc, err := process.NewProcess(int32(ppid)) //nolint:gosec // disable G115
	if err != nil {
		log.Printf("ERROR unable to get parent process: %v\n", err)
		return false
	}

	// Get parent process name
	parentName, err := parentProc.Name()
	if err != nil {
		log.Printf("ERROR unable to get parent process name: %v\n", err)
		return false
	}

	// Check if parent process is splunkd (Linux) or splunkd.exe (Windows)
	return parentName == "splunkd" || parentName == "splunkd.exe"
}
