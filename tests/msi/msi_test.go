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
	"bufio"
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
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Test structure for MSI installation tests
type msiTest struct {
	name                   string
	collectorMSIProperties map[string]string
	genericMSIProperties   map[string]string
	skipUninstall          bool
	skipSvcStart           bool
	skipSvcStop            bool
}

func TestMSI(t *testing.T) {
	msiInstallerPath := getInstallerPath(t)

	tests := []msiTest{
		{
			name: "default",
			collectorMSIProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN": "fakeToken",
			},
		},
		{
			name: "default-plus-cli-args",
			collectorMSIProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN": "fakeToken",
				"COLLECTOR_SVC_ARGS":  "--discovery --set=processors.batch.timeout=10s",
			},
		},
		{
			name: "gateway",
			collectorMSIProperties: map[string]string{
				"SPLUNK_SETUP_COLLECTOR_MODE": "gateway",
				"SPLUNK_ACCESS_TOKEN":         "testing123",
			},
		},
		{
			name: "realm",
			collectorMSIProperties: map[string]string{
				"SPLUNK_REALM":        "my-realm",
				"SPLUNK_ACCESS_TOKEN": "testing",
			},
		},
		{
			name: "ingest-url",
			collectorMSIProperties: map[string]string{
				"SPLUNK_INGEST_URL":   "https://fake.ingest.url",
				"SPLUNK_ACCESS_TOKEN": "testing",
			},
		},
		{
			name: "optional-params",
			collectorMSIProperties: map[string]string{
				"SPLUNK_MEMORY_TOTAL_MIB": "256",
				"SPLUNK_ACCESS_TOKEN":     "fakeToken",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runMsiTest(t, tt, msiInstallerPath)
		})
	}
}

// TestCollectorReconfiguration tests the MSI use to change the settings of an installation of the same version.
func TestCollectorReconfiguration(t *testing.T) {
	msiInstallerPath := getInstallerPath(t)

	tests := []msiTest{
		{
			name: "first-install",
			collectorMSIProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN": "1stInstall",
				"SPLUNK_REALM":        "1st",
			},
			skipUninstall: true,
			skipSvcStop:   true,
		},
		{
			name: "second-install",
			collectorMSIProperties: map[string]string{
				"SPLUNK_SETUP_COLLECTOR_MODE": "gateway",
				"SPLUNK_ACCESS_TOKEN":         "2ndInstall",
				"SPLUNK_REALM":                "2nd",
				"SPLUNK_MEMORY_TOTAL_MIB":     "256",
				"GODEBUG":                     "fips140=on",
			},
			skipSvcStart: true,
			genericMSIProperties: map[string]string{
				"REINSTALL": "SplunkCollectorConfiguration", // MSI property to reinstall the configuration
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runMsiTest(t, tt, msiInstallerPath)
		})
	}
}

func runMsiTest(t *testing.T, test msiTest, msiInstallerPath string) {
	allMSIProperties := make(map[string]string)
	for key, value := range test.genericMSIProperties {
		allMSIProperties[key] = value
	}
	for key, value := range test.collectorMSIProperties {
		allMSIProperties[key] = value
	}

	// Build the MSI installation arguments and include the MSI properties map.
	installLogFile := filepath.Join(os.TempDir(), "install.log")
	args := []string{"/i", msiInstallerPath, "/qn", "/l*v", installLogFile}
	for key, value := range allMSIProperties {
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
	logFile, _ := os.Open(installLogFile)
	scanner := bufio.NewScanner(transform.NewReader(logFile, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))
	for scanner.Scan() {
		t.Log(scanner.Text())
	}
	t.Logf("Install command: %s", installCmd.SysProcAttr.CmdLine)
	require.NoError(t, err, "Failed to install the MSI: %v\nArgs: %v", err, args)

	if !test.skipUninstall {
		defer func() {
			// Uninstall the MSI
			uninstallCmd := exec.Command("msiexec")
			uninstallCmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "msiexec /x " + msiInstallerPath + " /qn"}
			errUninstallCmd := uninstallCmd.Run()
			t.Logf("Uninstall command: %s", uninstallCmd.SysProcAttr.CmdLine)
			require.NoError(t, errUninstallCmd, "Failed to uninstall the MSI: %v", errUninstallCmd)
		}()
	}

	// Verify the service
	scm, err := mgr.Connect()
	require.NoError(t, err)
	defer scm.Disconnect()

	service, err := scm.OpenService("splunk-otel-collector")
	require.NoError(t, err)
	defer service.Close()

	if !test.skipSvcStart {
		err = service.Start()
		require.NoError(t, err)
	}
	if !test.skipSvcStop {
		defer func() {
			_, err = service.Control(svc.Stop)
			require.NoError(t, err)

			require.Eventually(t, func() bool {
				status, err := service.Query()
				require.NoError(t, err)
				return status.State == svc.Stopped
			}, 10*time.Second, 500*time.Millisecond, "Failed to stop the service")
		}()
	}

	// Wait for the service to reach the running state
	require.Eventually(t, func() bool {
		status, err := service.Query()
		require.NoError(t, err)
		return status.State == svc.Running
	}, 10*time.Second, 500*time.Millisecond, "Failed to start the service")

	svcConfig, err := service.Config()
	require.NoError(t, err, "Failed to get service configuration")

	assertServiceConfiguration(t, test.collectorMSIProperties, svcConfig)
}

func assertServiceConfiguration(t *testing.T, msiProperties map[string]string, svcConfig mgr.Config) {
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
		"SPLUNK_HEC_URL":      ingestURL + "/v1/log",
		"SPLUNK_HEC_TOKEN":    optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_HEC_TOKEN", msiProperties["SPLUNK_ACCESS_TOKEN"]),
		"SPLUNK_BUNDLE_DIR":   filepath.Join(programFilesDir, "Splunk", "OpenTelemetry Collector", "agent-bundle"),
	}
	if memoryTotalMib, ok := msiProperties["SPLUNK_MEMORY_TOTAL_MIB"]; ok {
		expectedEnvVars["SPLUNK_MEMORY_TOTAL_MIB"] = memoryTotalMib
	}
	if goDebug, ok := msiProperties["GODEBUG"]; ok {
		expectedEnvVars["GODEBUG"] = goDebug
	}

	// Verify the environment variables set for the service
	svcEnvVars := getServiceEnvVars(t, "splunk-otel-collector")
	assert.Equal(t, expectedEnvVars, svcEnvVars)

	if svcArgs, ok := msiProperties["COLLECTOR_SVC_ARGS"]; ok {
		assert.Equal(t, expectedServiceCommand(t, svcArgs), svcConfig.BinaryPathName)
	}
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

func getInstallerPath(t *testing.T) string {
	msiInstallerPath := os.Getenv("MSI_COLLECTOR_PATH")
	require.NotEmpty(t, msiInstallerPath, "MSI_COLLECTOR_PATH environment variable is not set")
	_, err := os.Stat(msiInstallerPath)
	require.NoError(t, err)
	if strings.Contains(msiInstallerPath, " ") {
		msiInstallerPath = "\"" + msiInstallerPath + "\""
	}
	return msiInstallerPath
}

func expectedServiceCommand(t *testing.T, collectorServiceArgs string) string {
	programFilesDir := os.Getenv("PROGRAMFILES")
	require.NotEmpty(t, programFilesDir, "PROGRAMFILES environment variable is not set")

	collectorDir := filepath.Join(programFilesDir, "Splunk", "OpenTelemetry Collector")
	collectorExe := filepath.Join(collectorDir, "otelcol") + ".exe"

	if collectorServiceArgs == "" {
		return quotedIfRequired(collectorExe)
	}

	// Remove any quotation added for the msiexec command line
	collectorServiceArgs = strings.Trim(collectorServiceArgs, "\"")
	collectorServiceArgs = strings.ReplaceAll(collectorServiceArgs, "\"\"", "\"")

	return quotedIfRequired(collectorExe) + " " + collectorServiceArgs
}

func quotedIfRequired(s string) string {
	if strings.Contains(s, "\"") || strings.Contains(s, " ") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}
