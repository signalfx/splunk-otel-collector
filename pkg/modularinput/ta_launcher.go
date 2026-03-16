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
	"maps"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

// modularInputMode represents the mode in which the modular input is running
type modularInputMode int

const (
	// notModularInput indicates the executable is not running as a modular input
	notModularInput modularInputMode = iota
	// executionMode indicates the executable is running as a modular input with no other arguments
	executionMode
	// introspectionMode indicates the executable is running as a modular input with --scheme argument
	introspectionMode
	// validationMode indicates the executable is running as a modular input with --validate-arguments argument
	validationMode
)

var (
	ErrQueryMode = errors.New("modular input called in query mode")

	// Function variables to facilitate testing
	setEnvFn                           = os.Setenv
	isParentProcessSplunkdFn           = isParentProcessSplunkd
	stdoutWriter             io.Writer = os.Stdout
)

// HandleLaunchAsTA handles the launch of the collector as a Splunk TA modular input.
// It checks if the collector is running in modular input mode and processes the input XML
// to set environment variables from the configuration stanza.
// Returns an error if the launch fails, ErrQueryMode if running in query mode,
// or nil if not running in modular input mode or on success.
func HandleLaunchAsTA(args []string, stdin io.Reader, configStanzaPrefix, scheme string) error {
	mode := isModularInputMode(args)
	if mode == notModularInput {
		return nil
	}

	if mode == introspectionMode {
		// The caller is just expected to exit when receiving ErrQueryMode
		if _, err := fmt.Fprintln(stdoutWriter, scheme); err != nil {
			return fmt.Errorf("failed to write scheme to stdout: %w", err)
		}
		return ErrQueryMode
	}

	if mode == validationMode {
		// The caller is just expected to exit when receiving ErrQueryMode
		return ErrQueryMode
	}

	input, err := ReadXML(stdin)
	if err != nil {
		return fmt.Errorf("launch as TA failed to read modular input XML from stdin: %w", err)
	}

	var configStanza Stanza
	for _, stanza := range input.Configuration.Stanza {
		if strings.HasPrefix(stanza.Name, configStanzaPrefix) {
			configStanza = stanza
			break
		}
	}

	// First pass: build a map of parameters starting with "splunk_"
	modularInputEnvVars := make(map[string]string)
	for _, param := range configStanza.Param {
		paramName := strings.ToLower(param.Name)
		if !strings.HasPrefix(paramName, "splunk_") {
			continue
		}

		// Process special parameters
		switch {
		// TODO: to be refactored: the caller will specify which parameters should be parsed as env var pairs instead of using a naming convention
		case strings.HasSuffix(paramName, "_env_vars"):
			var pairs map[string]string
			pairs, err = parseEnvVarPairs(param.Value)
			if err != nil {
				return fmt.Errorf("launch as TA failed to parse env vars from parameter '%s': %w", param.Name, err)
			}
			maps.Copy(modularInputEnvVars, pairs)
		default:
			modularInputEnvVars[strings.ToUpper(param.Name)] = param.Value
		}
	}

	// Second pass: set environment variables in dependency order
	err = setEnvVarsInOrder(modularInputEnvVars, setEnvFn)
	if err != nil {
		return err
	}

	return nil
}

func isModularInputMode(args []string) modularInputMode {
	// SPLUNK_HOME must be defined if this is running as a modular input.
	_, hasSplunkHome := os.LookupEnv("SPLUNK_HOME")
	if !hasSplunkHome {
		return notModularInput
	}

	// TA v1 is a special case of the collector being launched as a modular input
	// with TA specific behavior being handled by scripts. Use the SPLUNK_OTEL_TA_HOME
	// environment variable to determine if this is running as a TA v1 modular input.
	_, isTAv1Launch := os.LookupEnv("SPLUNK_OTEL_TA_HOME")
	if isTAv1Launch {
		// TA v1, let the scripts handle the TA specific behavior
		return notModularInput
	}

	// Check if the parent process is splunkd
	if !isParentProcessSplunkdFn() {
		return notModularInput
	}

	// This is running as a modular input
	if isArgScheme(args) {
		return introspectionMode
	}

	if isArgValidate(args) {
		return validationMode
	}

	return executionMode
}

// setEnvVarsInOrder sets environment variables in dependency order.
// Variables that don't reference other environment variables are set first,
// followed by those that do reference other environment variables.
func setEnvVarsInOrder(envVars map[string]string, setEnvFunc func(string, string) error) error {
	// Pattern to match environment variable references: $VAR, ${VAR} (case insensitive)
	envVarRefPattern := regexp.MustCompile(`(?i)\$\{?[A-Z_][A-Z0-9_]*\}?`)

	// Separate variables into those with and without dependencies
	var noDeps []string
	var withDeps []string

	for envVarName, envVarValue := range envVars {
		if envVarRefPattern.MatchString(envVarValue) {
			withDeps = append(withDeps, envVarName)
		} else {
			noDeps = append(noDeps, envVarName)
		}
	}

	// Sort for deterministic ordering
	sort.Strings(noDeps)
	sort.Strings(withDeps)

	// Set variables without dependencies first
	for _, envVarName := range noDeps {
		envVarValue := envVars[envVarName]
		err := setEnvFunc(envVarName, envVarValue)
		if err != nil {
			return fmt.Errorf("launch as TA failed to set environment variable '%s': %w", envVarName, err)
		}
	}

	// Set variables with dependencies, expanding environment variable references
	for _, envVarName := range withDeps {
		envVarValue := os.ExpandEnv(envVars[envVarName])
		err := setEnvFunc(envVarName, envVarValue)
		if err != nil {
			return fmt.Errorf("launch as TA failed to set environment variable '%s': %w", envVarName, err)
		}
	}

	return nil
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

func isArgScheme(args []string) bool {
	return len(args) == 2 && args[1] == "--scheme"
}

func isArgValidate(args []string) bool {
	return len(args) == 2 && args[1] == "--validate-arguments"
}

// parseEnvVarPairs parses a comma-separated list of key=value pairs into a map of
// environment variable names to their string values.
// Commas in keys and values must be percent-encoded as %2C so they are not treated as pair
// separators. '=' characters in keys must be percent-encoded as %3D so the first literal '='
// in each pair can be used as the key/value separator. '=' characters in values may be
// percent-encoded as %3D but do not need to be, because only the first '=' is treated as the
 // separator. This function expects standard percent-encoding: literal '%' characters must be
 // written as %25, and other characters may also be percent-encoded (e.g., non-ASCII characters).
 // Example input: "KEY1=value1,KEY2=value=2%2Cextra" → {"KEY1": "value1", "KEY2": "value=2,extra"}
func parseEnvVarPairs(s string) (map[string]string, error) {
	result := make(map[string]string)
	if s == "" {
		return result, nil
	}
	for pair := range strings.SplitSeq(s, ",") {
		pair = strings.TrimSpace(pair)
		rawKey, rawVal, found := strings.Cut(pair, "=")
		if !found {
			return nil, fmt.Errorf("invalid key=value pair %q: missing '='", pair)
		}
		rawKey = strings.TrimSpace(rawKey)
		rawVal = strings.TrimSpace(rawVal)

		key, err := url.PathUnescape(rawKey)
		if err != nil {
			return nil, fmt.Errorf("invalid percent-encoding in key %q: %w", rawKey, err)
		}
		val, err := url.PathUnescape(rawVal)
		if err != nil {
			return nil, fmt.Errorf("invalid percent-encoding in value %q: %w", rawVal, err)
		}

		if key == "" {
			return nil, fmt.Errorf("invalid key=value pair %q: key must not be empty", pair)
		}
		result[key] = val
	}
	return result, nil
}
