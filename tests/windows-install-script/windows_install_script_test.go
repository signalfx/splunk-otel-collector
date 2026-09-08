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

//go:build windows && msi

package windows_install_script

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	// Old version to install first, minimum version supported by the installation script.
	oldCollectorVersion    = "0.128.0"
	lastPreLauncherVersion = "0.158.0"
	fipsMode               = "fips140=on"
	serviceRegistryPath    = `SYSTEM\CurrentControlSet\Services\splunk-otel-collector`
	preservationMarker     = "# Preserve this customization across package upgrade."
	// Service name
	serviceName = "splunk-otel-collector"
	// Service display name
	serviceDisplayName = "Splunk OpenTelemetry Collector"
)

func TestUpgradeAndUninstallFromNonMachineWideVersion(t *testing.T) {
	t.Setenv("VERIFY_ACCESS_TOKEN", "false")
	// Force the installer to set two Zero Code resource attributes so the upgrade
	// verifies that it migrates one without removing the other.
	zeroCodeArgs := []string{
		"-with_dotnet_instrumentation", "1",
		"-deployment_env", "test",
	}

	requireNoPendingFileOperations(t)

	scm, err := mgr.Connect()
	require.NoError(t, err)
	defer scm.Disconnect()

	t.Logf(" *** Installing old collector version %s", oldCollectorVersion)
	installCollector(t, getTestDataFilePath(t, "install-before-platform-indexes.ps1"), oldCollectorVersion, "", zeroCodeArgs...)
	verifyServiceExists(t, scm)
	verifyServiceState(t, scm, svc.Running)
	verifyZeroConfigResourceAttributes(t, 1, "deployment.environment=test")
	legacySvcVersion := getCurrentServiceVersion(t)
	require.Equal(t, oldCollectorVersion, legacySvcVersion)

	// Uninstall the .NET instrumentation, so the upgrade is not blocked by the install.ps1 script
	uninstallDotnetInstrumentation(t)

	msiInstallerPath := getFilePathFromEnvVar(t, "MSI_COLLECTOR_PATH")
	t.Logf(" *** Installing collector from %q", msiInstallerPath)
	installCollector(t, getFilePathFromEnvVar(t, "INSTALL_SCRIPT_PATH"), "", msiInstallerPath, zeroCodeArgs...)
	verifyServiceExists(t, scm)
	verifyServiceState(t, scm, svc.Running)
	verifyZeroConfigResourceAttributes(t, 1, "deployment.environment.name=test")
	latestSvcVersion := getCurrentServiceVersion(t)
	require.NotEqual(t, oldCollectorVersion, latestSvcVersion)
	requireNoPendingFileOperations(t)

	uninstallCollector(t)
	verifyZeroConfigResourceAttributes(t, 0, "deployment.environment.name=test")
}

func TestUpgradeAndRollbackSupervisorMode(t *testing.T) {
	t.Setenv("VERIFY_ACCESS_TOKEN", "false")
	requireNoPendingFileOperations(t)

	scm, err := mgr.Connect()
	require.NoError(t, err)
	defer scm.Disconnect()

	installScriptPath := getFilePathFromEnvVar(t, "INSTALL_SCRIPT_PATH")
	t.Logf(" *** Installing pre-launcher collector version %s", lastPreLauncherVersion)
	installCollector(
		t,
		installScriptPath,
		lastPreLauncherVersion,
		"",
		"-realm", "test",
		"-godebug", fipsMode,
	)
	verifyServiceExists(t, scm)
	verifyServiceState(t, scm, svc.Running)

	configPath := filepath.Join(os.Getenv("PROGRAMDATA"), "Splunk", "OpenTelemetry Collector", "agent_config.yaml")
	configFile, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	_, err = configFile.WriteString("\n" + preservationMarker + "\n")
	require.NoError(t, err)
	require.NoError(t, configFile.Close())
	configHash := hashFile(t, configPath)

	fipsEnvironment := "GODEBUG=" + fipsMode
	verifyServiceEnvironmentContains(t, fipsEnvironment)

	msiInstallerPath := getFilePathFromEnvVar(t, "MSI_COLLECTOR_PATH")
	t.Logf(" *** Upgrading collector from %q", msiInstallerPath)
	installCollector(
		t,
		installScriptPath,
		"",
		msiInstallerPath,
		"-realm", "test",
		"-godebug", fipsMode,
		"-preserve_prev_default_config", "1",
	)

	verifyLauncherServiceEntrypoint(t, scm)
	installDir := filepath.Join(os.Getenv("PROGRAMFILES"), "Splunk", "OpenTelemetry Collector")
	for _, executableName := range []string{"otelcol.exe", "otelcollauncher.exe", "opampsupervisor.exe"} {
		require.FileExists(t, filepath.Join(installDir, executableName))
	}
	verifyServiceEnvironmentContains(t, fipsEnvironment)
	require.Equal(t, configHash, hashFile(t, configPath), "Collector configuration changed during package upgrade")
	verifyServiceRunningWithProcess(t, scm, "otelcol.exe")

	setServiceEnvironmentVariable(t, "SPLUNK_OPAMP_SUPERVISOR_ENABLED", "true")
	restartService(t, scm)
	verifyServiceEnvironmentContains(t, fipsEnvironment, "SPLUNK_OPAMP_SUPERVISOR_ENABLED=true")
	verifyServiceRunningWithProcess(t, scm, "opampsupervisor.exe")
	verifySupervisorConfigFiles(t)

	setServiceEnvironmentVariable(t, "SPLUNK_OPAMP_SUPERVISOR_ENABLED", "false")
	restartService(t, scm)
	verifyServiceEnvironmentContains(t, fipsEnvironment, "SPLUNK_OPAMP_SUPERVISOR_ENABLED=false")
	verifyServiceRunningWithProcess(t, scm, "otelcol.exe")
	require.Equal(t, configHash, hashFile(t, configPath), "Collector configuration changed while switching supervisor modes")
	requireNoPendingFileOperations(t)

	uninstallCollector(t)
}

func installCollector(t *testing.T, installScriptPath, version, msiPath string, extraArgs ...string) {
	t.Helper()
	require.False(t, version == "" && msiPath == "", "Either version or msiPath must be provided")
	require.False(t, version != "" && msiPath != "", "Only one of version or msiPath should be provided")
	args := []string{
		"-ExecutionPolicy", "Bypass",
		"-Command", "& " + installScriptPath,
		"-access_token", "fake-token",
	}
	args = append(args, extraArgs...)

	if version != "" {
		args = append(args, "-collector_version", version)
	} else if msiPath != "" {
		args = append(args, "-msi_path", msiPath)
	} else {
		require.Fail(t, "Either version or msiPath must be provided")
	}

	cmd := exec.Command("powershell.exe", args...)

	output, err := cmd.CombinedOutput()
	t.Logf("Install output: %s", string(output))
	require.NoError(t, err, "Failed to install collector (version:%q msiPath:%q)", version, msiPath)
}

func getTestDataFilePath(t *testing.T, fileName string) string {
	filePath := filepath.Join("testdata", fileName)
	_, err := os.Stat(filePath)
	require.NoError(t, err, "File %s does not exist", filePath)
	return "." + string(os.PathSeparator) + filePath
}

func uninstallCollector(t *testing.T) {
	args := []string{
		"-ExecutionPolicy", "Bypass",
		"-Command", "& " + getFilePathFromEnvVar(t, "INSTALL_SCRIPT_PATH"),
		"-uninstall_collector", // Uninstall the collector and clean up any Splunk .NET zero-code resource attribute from OTEL_RESOURCE_ATTRIBUTES.
	}

	cmd := exec.Command("powershell.exe", args...)

	output, err := cmd.CombinedOutput()
	t.Logf("Uninstall output: %s", string(output))
	require.NoError(t, err, "Failed to uninstall collector")
}

func verifyServiceExists(t *testing.T, scm *mgr.Mgr) {
	service, err := scm.OpenService(serviceName)
	require.NoError(t, err)
	service.Close()
}

func verifyServiceState(t *testing.T, scm *mgr.Mgr, desiredState svc.State) {
	service, err := scm.OpenService(serviceName)
	require.NoError(t, err)
	defer service.Close()

	// Wait for the service to reach the running state
	require.Eventually(t, func() bool {
		status, err := service.Query()
		require.NoError(t, err)
		return status.State == desiredState
	}, 10*time.Second, 500*time.Millisecond, "Service failed to reach the desired state")
}

func verifyLauncherServiceEntrypoint(t *testing.T, scm *mgr.Mgr) {
	t.Helper()
	service, err := scm.OpenService(serviceName)
	require.NoError(t, err)
	defer service.Close()

	serviceConfig, err := service.Config()
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(serviceConfig.BinaryPathName), "otelcollauncher.exe")
}

func restartService(t *testing.T, scm *mgr.Mgr) {
	t.Helper()
	service, err := scm.OpenService(serviceName)
	require.NoError(t, err)
	defer service.Close()

	_, err = service.Control(svc.Stop)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		status, queryErr := service.Query()
		return queryErr == nil && status.State == svc.Stopped
	}, 30*time.Second, time.Second, "Service failed to stop")

	require.NoError(t, service.Start())
	require.Eventually(t, func() bool {
		status, queryErr := service.Query()
		return queryErr == nil && status.State == svc.Running
	}, 30*time.Second, time.Second, "Service failed to restart")
}

func hashFile(t *testing.T, filePath string) string {
	t.Helper()
	contents, err := os.ReadFile(filePath)
	require.NoError(t, err)
	return fmt.Sprintf("%x", sha256.Sum256(contents))
}

func setServiceEnvironmentVariable(t *testing.T, name, value string) {
	t.Helper()
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, serviceRegistryPath, registry.QUERY_VALUE|registry.SET_VALUE)
	require.NoError(t, err)
	defer key.Close()

	environment, _, err := key.GetStringsValue("Environment")
	require.NoError(t, err)
	updatedEnvironment := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], name) {
			continue
		}
		updatedEnvironment = append(updatedEnvironment, entry)
	}
	updatedEnvironment = append(updatedEnvironment, name+"="+value)
	require.NoError(t, key.SetStringsValue("Environment", updatedEnvironment))
}

func verifyServiceEnvironmentContains(t *testing.T, expectedEntries ...string) {
	t.Helper()
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, serviceRegistryPath, registry.QUERY_VALUE)
	require.NoError(t, err)
	defer key.Close()

	environment, _, err := key.GetStringsValue("Environment")
	require.NoError(t, err)
	for _, expectedEntry := range expectedEntries {
		require.Contains(t, environment, expectedEntry)
	}
}

func verifyServiceRunningWithProcess(t *testing.T, scm *mgr.Mgr, expectedProcessName string) {
	t.Helper()
	require.Eventually(t, func() bool {
		service, err := scm.OpenService(serviceName)
		if err != nil {
			return false
		}
		defer service.Close()

		status, err := service.Query()
		if err != nil || status.State != svc.Running || status.ProcessId == 0 {
			return false
		}

		serviceProcess, err := process.NewProcess(int32(status.ProcessId))
		if err != nil {
			return false
		}
		children, err := serviceProcess.Children()
		if err != nil {
			return false
		}

		for _, child := range children {
			name, nameErr := child.Name()
			if nameErr != nil {
				return false
			}
			if strings.EqualFold(name, expectedProcessName) {
				return true
			}
		}
		return false
	}, 30*time.Second, time.Second, "Service failed to run with %s", expectedProcessName)
}

func verifySupervisorConfigFiles(t *testing.T) {
	t.Helper()
	programData := os.Getenv("PROGRAMDATA")
	require.NotEmpty(t, programData, "PROGRAMDATA environment variable is not set")
	supervisorDir := filepath.Join(programData, "Splunk", "OpenTelemetry Collector", "supervisor")
	require.FileExists(t, filepath.Join(supervisorDir, "supervisor_config.yaml"))
	require.FileExists(t, filepath.Join(supervisorDir, "supervisor_runtime_config.yaml"))
}

func getCurrentServiceVersion(t *testing.T) string {
	// Read the service version from the registry, need to find the GUID registry key
	// given the service name.
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\Uninstall`, registry.ALL_ACCESS)
	require.NoError(t, err)
	defer key.Close()

	// Enumerate all subkeys to find the one that matches the service name
	subKeys, err := key.ReadSubKeyNames(0)
	require.NoError(t, err)

	for _, subKey := range subKeys {
		subKeyPath := fmt.Sprintf(`Software\Microsoft\Windows\CurrentVersion\Uninstall\%s`, subKey)
		subKeyHandle, err := registry.OpenKey(registry.LOCAL_MACHINE, subKeyPath, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		defer subKeyHandle.Close()

		displayName, _, err := subKeyHandle.GetStringValue("DisplayName")
		if err == nil && strings.Contains(displayName, serviceDisplayName) {
			// Found the subkey for the service, now get the version
			version, _, err := subKeyHandle.GetStringValue("DisplayVersion")
			require.NoError(t, err)
			return version
		}
	}

	require.Fail(t, "Failed to find service version in registry")
	return ""
}

func verifyZeroConfigResourceAttributes(t *testing.T, expectedCount int, expectedAttribute string) {
	envKey, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE,
	)
	require.NoError(t, err)
	defer envKey.Close()

	otelResourceAttributes, _, err := envKey.GetStringValue("OTEL_RESOURCE_ATTRIBUTES")
	require.NoError(t, err, "OTEL_RESOURCE_ATTRIBUTES machine-wide environment variable not found")
	require.NotEmpty(t, otelResourceAttributes, "OTEL_RESOURCE_ATTRIBUTES should not be empty")
	t.Logf("OTEL_RESOURCE_ATTRIBUTES: %q", otelResourceAttributes)

	const expectedPrefix = "splunk.zc.method=splunk-otel-dotnet-"
	count := 0
	for _, attribute := range strings.Split(otelResourceAttributes, ",") {
		trimmed := strings.TrimSpace(attribute)
		if len(trimmed) > len(expectedPrefix) && strings.HasPrefix(trimmed, expectedPrefix) {
			count++
		}
	}

	require.Equal(t, expectedCount, count,
		"OTEL_RESOURCE_ATTRIBUTES didn't contain the expected count of zero-code attributes, got %d in %q",
		count, otelResourceAttributes,
	)
	require.Contains(t, strings.Split(otelResourceAttributes, ","), expectedAttribute)
}

func requireNoPendingFileOperations(t *testing.T) {
	// Check for pending file rename operations
	pendingFileRenameKey, err := registry.OpenKey(
		registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager`, registry.QUERY_VALUE)
	require.NoError(t, err)
	defer pendingFileRenameKey.Close()
	pendingFileRenameEntries, _, err := pendingFileRenameKey.GetStringsValue("PendingFileRenameOperations")
	if err != nil {
		require.ErrorIs(t, err, registry.ErrNotExist)
	}

	for _, fileName := range pendingFileRenameEntries {
		if strings.Contains(strings.ToLower(fileName), "splunk") {
			require.Fail(t, "Found pending file rename: %s", fileName)
		}
	}
}

func getFilePathFromEnvVar(t *testing.T, envVar string) string {
	filePath := os.Getenv(envVar)
	require.NotEmpty(t, filePath, "%s environment variable is not set", envVar)
	_, err := os.Stat(filePath)
	require.NoError(t, err, "File %s does not exist", filePath)
	if strings.Contains(filePath, " ") {
		filePath = "\"" + filePath + "\""
	}
	return filePath
}

func uninstallDotnetInstrumentation(t *testing.T) {
	args := []string{
		"-ExecutionPolicy", "Bypass",
		"-File", filepath.Join("testdata", "Uninstall-SplunkDotnetInstrumentation.ps1"),
	}

	cmd := exec.Command("powershell.exe", args...)

	output, err := cmd.CombinedOutput()
	t.Logf("Uninstall Splunk .NET instrumentation output:\n%s", string(output))
	require.NoError(t, err, "Failed to uninstall .NET instrumentation")
}
