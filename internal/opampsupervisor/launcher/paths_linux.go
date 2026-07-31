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

//go:build linux

package launcher

import (
	"os"
	"path/filepath"
)

// DefaultPaths returns collector, supervisor, and state locations for Linux
// packages. Tar bundles use paths relative to the launcher executable.
func DefaultPaths() Paths {
	launcherExecutable, err := os.Executable()
	if err != nil {
		launcherExecutable = "/usr/bin/otelcollauncher"
	}
	return pathsForExecutable(launcherExecutable)
}

func pathsForExecutable(launcherExecutable string) Paths {
	binDir := filepath.Dir(launcherExecutable)
	configDir := "/etc/otel/collector"
	if binDir != "/usr/bin" {
		configDir = filepath.Join(filepath.Dir(binDir), "config")
	}
	supervisorConfigDir := filepath.Join(configDir, "supervisor")
	return Paths{
		CollectorExecutable:         filepath.Join(binDir, "otelcol"),
		SupervisorExecutable:        filepath.Join(binDir, "opampsupervisor"),
		SupervisorConfig:            filepath.Join(supervisorConfigDir, "supervisor_config.yaml"),
		RuntimeSupervisorConfig:     filepath.Join(supervisorConfigDir, "supervisor_runtime_config.yaml"),
		GeneratedCollectorConfigDir: supervisorConfigDir,
		StorageDirectory:            "/var/lib/otelcol/supervisor",
		DefaultAgentConfig:          filepath.Join(configDir, "agent_config.yaml"),
		ConfigApplyTimeout:          "1m",
		UseHUPConfigReload:          true,
	}
}
