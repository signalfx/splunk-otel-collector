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

package msi

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func TestMSI(t *testing.T) {
	msiInstallerPath := os.Getenv("MSI_COLLECTOR_PATH")
	require.NotEmpty(t, msiInstallerPath, "MSI_COLLECTOR_PATH environment variable is not set")
	_, err := os.Stat(msiInstallerPath)
	require.NoError(t, err)
	if strings.Contains(msiInstallerPath, " ") {
		msiInstallerPath = "\"" + msiInstallerPath + "\""
	}

	tests := []struct {
		name          string
		msiProperties map[string]string
	}{
		{
			name: "default",
			msiProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN": "fakeToken",
			},
		},
		{
			name: "gateway",
			msiProperties: map[string]string{
				"SPLUNK_SETUP_COLLECTOR_MODE": "gateway",
				"SPLUNK_ACCESS_TOKEN":         "testing123",
			},
		},
		{
			name: "realm",
			msiProperties: map[string]string{
				"SPLUNK_REALM":        "my-realm",
				"SPLUNK_ACCESS_TOKEN": "testing",
			},
		},
		{
			name: "ingest-url",
			msiProperties: map[string]string{
				"SPLUNK_INGEST_URL":   "https://fake.ingest.url",
				"SPLUNK_ACCESS_TOKEN": "testing",
			},
		},
		{
			name: "optional-params",
			msiProperties: map[string]string{
				"SPLUNK_MEMORY_TOTAL_MIB": "256",
				"SPLUNK_ACCESS_TOKEN":     "fakeToken",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installLogFile := filepath.Join(os.TempDir(), "install.log")
			// Build the MSI installation arguments and include the msiProperties map.
			args := []string{"/i", msiInstallerPath, "/qn", "/l*v", installLogFile}
			for key, value := range tt.msiProperties {
				// Escape the key and value if they contain spaces or quotes
				// See https://learn.microsoft.com/en-us/windows/win32/msi/command-line-options
				if strings.Contains(value, "\"") || strings.Contains(value, " ") {
					value = strings.ReplaceAll(value, "\"", "\"\"")
					value = "\"" + value + "\""
				}
				args = append(args, key+"="+value)
			}

			// Run the MSI installer
			installCmd := exec.Command("msiexec")
			// msiexec is one of the noticeable exceptions about how to format the parameters,
			// see https://pkg.go.dev/os/exec#Command, so we need to join the args manually.
			cmdLine := strings.Join(args, " ")
			installCmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "msiexec " + cmdLine}
			err := installCmd.Run()
			if err != nil {
				logText, _ := os.ReadFile(installLogFile)
				t.Log(string(logText))
			}
			t.Logf("Install command: %s", installCmd.SysProcAttr.CmdLine)
			require.NoError(t, err, "Failed to install the MSI: %v\nArgs: %v", err, args)

			defer func() {
				// Uninstall the MSI
				uninstallCmd := exec.Command("msiexec")
				uninstallCmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "msiexec /x " + msiInstallerPath + " /qn"}
				err := uninstallCmd.Run()
				t.Logf("Uninstall command: %s", uninstallCmd.SysProcAttr.CmdLine)
				require.NoError(t, err, "Failed to uninstall the MSI: %v", err)
			}()

			// Verify the service
			scm, err := mgr.Connect()
			require.NoError(t, err)
			defer scm.Disconnect()

			service, err := scm.OpenService("splunk-otel-collector")
			require.NoError(t, err)
			defer service.Close()

			err = service.Start()
			require.NoError(t, err)
			defer func() {
				_, err = service.Control(svc.Stop)
				require.NoError(t, err)

				require.Eventually(t, func() bool {
					status, err := service.Query()
					require.NoError(t, err)
					return status.State == svc.Stopped
				}, 10*time.Second, 500*time.Millisecond, "Failed to stop the service")
			}()

			// Wait for the service to reach the running state
			require.Eventually(t, func() bool {
				status, err := service.Query()
				require.NoError(t, err)
				return status.State == svc.Running
			}, 10*time.Second, 500*time.Millisecond, "Failed to start the service")

			assertServiceConfiguration(t, tt.msiProperties)
		})
	}
}

func assertServiceConfiguration(t *testing.T, msiProperties map[string]string) {
	programDataDir := os.Getenv("PROGRAMDATA")
	require.NotEmpty(t, programDataDir, "PROGRAMDATA environment variable is not set")
	programFilesDir := os.Getenv("PROGRAMFILES")
	require.NotEmpty(t, programFilesDir, "PROGRAMFILES environment variable is not set")

	installRealm := optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_REALM", "us0")
	ingestURL := optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_INGEST_URL", "https://ingest."+installRealm+".signalfx.com")
	installMode, ok := msiProperties["SPLUNK_SETUP_COLLECTOR_MODE"]
	if !ok {
		installMode = "agent"
	}

	configFileName := installMode + "_config.yaml"
	configFileFullName := filepath.Join(programDataDir, "Splunk", "OpenTelemetry Collector", configFileName)
	assert.FileExists(t, configFileFullName)
	assert.NoFileExists(t, filepath.Join(programFilesDir, "Splunk", "OpenTelemetry Collector", configFileName))

	expectedEnvVars := map[string]string{
		"SPLUNK_CONFIG":       configFileFullName,
		"SPLUNK_ACCESS_TOKEN": msiProperties["SPLUNK_ACCESS_TOKEN"], // Required install property for a successful start of the service
		"SPLUNK_REALM":        installRealm,
		"SPLUNK_API_URL":      optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_API_URL", "https://api."+installRealm+".signalfx.com"),
		"SPLUNK_INGEST_URL":   ingestURL,
		"SPLUNK_TRACE_URL":    ingestURL + "/v2/trace",
		"SPLUNK_HEC_URL":      ingestURL + "/v1/log",
		"SPLUNK_HEC_TOKEN":    optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_HEC_TOKEN", msiProperties["SPLUNK_ACCESS_TOKEN"]),
		"SPLUNK_BUNDLE_DIR":   filepath.Join(programFilesDir, "Splunk", "OpenTelemetry Collector", "agent-bundle"),
	}
	if memoryTotalMib, ok := msiProperties["SPLUNK_MEMORY_TOTAL_MIB"]; ok {
		expectedEnvVars["SPLUNK_MEMORY_TOTAL_MIB"] = memoryTotalMib
	}

	// Verify the environment variables set for the service
	svcConfig := getServiceEnvVars(t, "splunk-otel-collector")
	assert.Equal(t, expectedEnvVars, svcConfig)
}

func optionalInstallPropertyOrDefault(msiProperties map[string]string, key, defaultValue string) string {
	value, ok := msiProperties[key]
	if !ok {
		return defaultValue
	}
	return value
}

func getServiceEnvVars(t *testing.T, serviceName string) map[string]string {
	// Read the Environment set in the registry for the service.
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Services\\"+serviceName, registry.QUERY_VALUE)
	require.NoError(t, err)
	defer key.Close()

	svcEnv, _, err := key.GetStringsValue("Environment")
	require.NoError(t, err)

	envVars := make(map[string]string)
	for _, envVar := range svcEnv {
		parts := strings.SplitN(envVar, "=", 2)
		require.Len(t, parts, 2)
		envVars[parts[0]] = parts[1]
	}

	return envVars
}
