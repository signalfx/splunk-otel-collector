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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPathsWindowsFallbacksAlignWithInstaller(t *testing.T) {
	t.Setenv("ProgramFiles", "")
	t.Setenv("ProgramData", "")

	paths := DefaultPaths()

	installDir := filepath.Join(`\Program Files`, "Splunk", "OpenTelemetry Collector")
	stateDir := filepath.Join(`\ProgramData`, "Splunk", "OpenTelemetry Collector", "supervisor")

	assert.Equal(t, filepath.Join(installDir, "otelcol.exe"), paths.CollectorExecutable)
	assert.Equal(t, filepath.Join(installDir, "opampsupervisor.exe"), paths.SupervisorExecutable)
	assert.Equal(t, filepath.Join(stateDir, "supervisor_config.yaml"), paths.SupervisorConfig)
	assert.Equal(t, filepath.Join(stateDir, "supervisor_runtime_config.yaml"), paths.RuntimeSupervisorConfig)
	assert.Equal(t, stateDir, paths.GeneratedCollectorConfigDir)
	assert.Equal(t, stateDir, paths.StorageDirectory)
	assert.Equal(t, filepath.Join(`\ProgramData`, "Splunk", "OpenTelemetry Collector", "agent_config.yaml"), paths.DefaultAgentConfig)
	assert.Equal(t, "2m", paths.ConfigApplyTimeout)
	assert.False(t, paths.UseHUPConfigReload)
}

func TestDefaultListenInterfaceForWindowsAgentConfig(t *testing.T) {
	t.Setenv("ProgramData", "")

	programData := `\ProgramData`
	agentConfig := filepath.Join(programData, "Splunk", "OpenTelemetry Collector", "agent_config.yaml")
	gatewayConfig := filepath.Join(programData, "Splunk", "OpenTelemetry Collector", "gateway_config.yaml")

	assert.Equal(t, defaultAgentListenInterface, defaultListenInterfaceForConfigs([]string{agentConfig}, agentConfig))
	assert.Equal(t, defaultAgentListenInterface, defaultListenInterfaceForConfigs([]string{gatewayConfig, agentConfig}, agentConfig))
	assert.Equal(t, defaultListenInterface, defaultListenInterfaceForConfigs([]string{gatewayConfig}, agentConfig))
}
