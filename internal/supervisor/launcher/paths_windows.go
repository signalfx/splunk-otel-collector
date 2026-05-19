// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package launcher

import (
	"os"
	"path/filepath"
)

// DefaultPaths returns the installed collector, supervisor, and state
// locations used by the package-managed service on Windows packages.
func DefaultPaths() Paths {
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = `\Program Files`
	}
	programData := os.Getenv("ProgramData")
	if programData == "" {
		programData = `\ProgramData`
	}
	installDir := filepath.Join(programFiles, "Splunk", "OpenTelemetry Collector")
	stateDir := filepath.Join(programData, "Splunk", "OpenTelemetry Collector", "supervisor")
	return Paths{
		CollectorExecutable:  filepath.Join(installDir, "otelcol.exe"),
		SupervisorExecutable: filepath.Join(installDir, "opampsupervisor.exe"),
		SupervisorConfig:     filepath.Join(programData, "Splunk", "OpenTelemetry Collector", "supervisor.yaml"),
		StorageDirectory:     stateDir,
		UseHUPConfigReload:   false,
	}
}
