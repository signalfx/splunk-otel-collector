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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathsForExecutable(t *testing.T) {
	bundleDir := "/opt/splunk-otel-collector"
	tests := []struct {
		name       string
		executable string
		want       Paths
	}{
		{
			name:       "installed package",
			executable: "/usr/bin/otelcollauncher",
			want: Paths{
				CollectorExecutable:         "/usr/bin/otelcol",
				SupervisorExecutable:        "/usr/bin/opampsupervisor",
				SupervisorConfig:            "/etc/otel/collector/supervisor/supervisor_config.yaml",
				RuntimeSupervisorConfig:     "/etc/otel/collector/supervisor/supervisor_runtime_config.yaml",
				GeneratedCollectorConfigDir: "/etc/otel/collector/supervisor",
				StorageDirectory:            "/var/lib/otelcol/supervisor",
				DefaultAgentConfig:          "/etc/otel/collector/agent_config.yaml",
				BootstrapTimeout:            "1m",
				ConfigApplyTimeout:          "2m",
				UseHUPConfigReload:          true,
			},
		},
		{
			name:       "tar bundle",
			executable: filepath.Join(bundleDir, "bin", "otelcollauncher"),
			want: Paths{
				CollectorExecutable:         filepath.Join(bundleDir, "bin", "otelcol"),
				SupervisorExecutable:        filepath.Join(bundleDir, "bin", "opampsupervisor"),
				SupervisorConfig:            filepath.Join(bundleDir, "config", "supervisor", "supervisor_config.yaml"),
				RuntimeSupervisorConfig:     filepath.Join(bundleDir, "config", "supervisor", "supervisor_runtime_config.yaml"),
				GeneratedCollectorConfigDir: filepath.Join(bundleDir, "config", "supervisor"),
				StorageDirectory:            "/var/lib/otelcol/supervisor",
				DefaultAgentConfig:          filepath.Join(bundleDir, "config", "agent_config.yaml"),
				BootstrapTimeout:            "1m",
				ConfigApplyTimeout:          "2m",
				UseHUPConfigReload:          true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, pathsForExecutable(test.executable))
		})
	}
}
