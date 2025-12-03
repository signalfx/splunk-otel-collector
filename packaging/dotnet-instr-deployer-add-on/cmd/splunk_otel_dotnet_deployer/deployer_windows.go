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

//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/splunk/splunk_otel_dotnet_deployer/internal/modularinput"
)

const (
	wrapperScript = "installer-wrapper.ps1"
)

func runDeployer(input *modularinput.Input, stdin, stdout, stderr *os.File) error {
	// Launch the wrapper script using PowerShell capturing the output
	// and error streams and sending them to the stdout and stderr streams
	// of the deployer.

	// The wrapper script is expected to be in the same directory as the deployer.
	// Get the path to the deployer executable.
	deployerPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get the path to the deployer executable: %w", err)
	}

	scriptDir := filepath.Dir(deployerPath)
	scriptPath := filepath.Join(scriptDir, wrapperScript)
	if _, err := os.Stat(scriptPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wrapper script not found at %s", scriptPath)
		}
		return fmt.Errorf("failed to check if wrapper script exists: %w", err)
	}

	// Launch the wrapper script using PowerShell.
	args := []string{
		"-ExecutionPolicy", "ByPass",
		"-File", scriptPath,
	}
	args = append(args, scriptArgs(input)...)
	log.Printf("Running: powershell.exe args: %v\n", args)
	cmd := exec.Command("powershell.exe", args...)

	var stdBuf, errBuf strings.Builder
	cmd.Stdin = stdin
	cmd.Stdout = &stdBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()

	log.Printf("Wrapper stdout:\n%s\n", stdBuf.String())
	log.Printf("Wrapper stderr:\n%s\n", errBuf.String())

	return err
}

func scriptArgs(input *modularinput.Input) []string {
	var args []string
	var configStanza modularinput.Stanza
	for _, stanza := range input.Configuration.Stanza {
		if stanza.Name == modularinputName+"://"+modularinputName {
			configStanza = stanza
			break
		}
	}

	for _, param := range configStanza.Param {
		switch param.Name {
		case "uninstall":
			if param.Value == "true" {
				args = append(args, "-uninstall")
			}
		}
	}
	return args
}
