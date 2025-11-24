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

package windows_install_script

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	// Old version to install first, this version by default is not installed as machine-wide.
	oldCollectorVersion = "0.94.0"
	// Service name
	serviceName = "splunk-otel-collector"
	// Service display name
	serviceDisplayName = "Splunk OpenTelemetry Collector"
)

func TestUpgradeFromNonMachineWideVersion(t *testing.T) {
	t.Setenv("VERIFY_ACCESS_TOKEN", "false")

	requireNoPendingFileOperations(t)

	scm, err := mgr.Connect()
	require.NoError(t, err)
	defer scm.Disconnect()

	t.Logf(" *** Installing old collector version %s", oldCollectorVersion)
	installCollector(t, oldCollectorVersion, "")
	verifyServiceExists(t, scm)
	verifyServiceState(t, scm, svc.Running)
	legacySvcVersion := getCurrentServiceVersion(t)
	require.Equal(t, oldCollectorVersion, legacySvcVersion)

	msiInstallerPath := getFilePathFromEnvVar(t, "MSI_COLLECTOR_PATH")
	t.Logf(" *** Installing collector from %q", msiInstallerPath)
	installCollector(t, "", msiInstallerPath)
	verifyServiceExists(t, scm)
	verifyServiceState(t, scm, svc.Running)
	latestSvcVersion := getCurrentServiceVersion(t)
	require.NotEqual(t, oldCollectorVersion, latestSvcVersion)
	requireNoPendingFileOperations(t)
}

func installCollector(t *testing.T, version, msiPath string) {
	require.False(t, version == "" && msiPath == "", "Either version or msiPath must be provided")
	require.False(t, version != "" && msiPath != "", "Only one of version or msiPath should be provided")
	args := []string{
		"-ExecutionPolicy", "Bypass",
		"-File", getFilePathFromEnvVar(t, "INSTALL_SCRIPT_PATH"),
		"-access_token", "fake-token",
	}

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
