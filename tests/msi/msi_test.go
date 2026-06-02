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
		{
			name: "platform-logs",
			collectorMSIProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN":         "fakeToken",
				"SPLUNK_PLATFORM_URL":         "http://localhost:8088/services/collector",
				"SPLUNK_PLATFORM_TOKEN":       "platformToken",
				"SPLUNK_PLATFORM_LOGS_INDEX":  "otel_logs",
				"SPLUNK_SETUP_COLLECTOR_MODE": "agent",
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

func TestMSILaunchConditions(t *testing.T) {
	msiInstallerPath := getInstallerPath(t)

	tests := []struct {
		name                   string
		collectorMSIProperties map[string]string
		expectedLogMessage     string
	}{
		{
			name: "platform-url-requires-platform-token",
			collectorMSIProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN": "fakeToken",
				"SPLUNK_PLATFORM_URL": "http://localhost:8088/services/collector",
			},
			expectedLogMessage: "SPLUNK_PLATFORM_TOKEN is required when SPLUNK_PLATFORM_URL is set.",
		},
		{
			name: "platform-token-requires-platform-url",
			collectorMSIProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN":   "fakeToken",
				"SPLUNK_PLATFORM_TOKEN": "platformToken",
			},
			expectedLogMessage: "SPLUNK_PLATFORM_URL is required when SPLUNK_PLATFORM_TOKEN is set.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runMsiInstallFailureTest(t, msiTest{
				name:                   tt.name,
				collectorMSIProperties: tt.collectorMSIProperties,
			}, msiInstallerPath, tt.expectedLogMessage)
		})
	}
}

func TestExpectedCollectorServiceArgs(t *testing.T) {
	t.Setenv("PROGRAMDATA", `C:\ProgramData`)

	collectorConfigDir := filepath.Join(os.Getenv("PROGRAMDATA"), "Splunk", "OpenTelemetry Collector")
	defaultConfigArg := "--config " + quotedIfRequired(filepath.Join(collectorConfigDir, "agent_config.yaml"))
	logsConfigArg := "--config " + quotedIfRequired(filepath.Join(collectorConfigDir, "splunk_logs_config_windows.yaml"))
	mergeAppendFeatureGateArg := "--feature-gates=confmap.enableMergeAppendOption"

	tests := []struct {
		name          string
		msiProperties map[string]string
		expectedArgs  string
	}{
		{
			name: "default-config-only",
			msiProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN": "fakeToken",
			},
			expectedArgs: defaultConfigArg,
		},
		{
			name: "logs-config-only",
			msiProperties: map[string]string{
				"SPLUNK_PLATFORM_URL":   "http://localhost:8088/services/collector",
				"SPLUNK_PLATFORM_TOKEN": "platformToken",
			},
			expectedArgs: logsConfigArg,
		},
		{
			name: "default-and-logs-configs",
			msiProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN":   "fakeToken",
				"SPLUNK_PLATFORM_URL":   "http://localhost:8088/services/collector",
				"SPLUNK_PLATFORM_TOKEN": "platformToken",
			},
			expectedArgs: strings.Join([]string{defaultConfigArg, logsConfigArg, mergeAppendFeatureGateArg}, " "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualArgs := expectedCollectorServiceArgs(t, tt.msiProperties)
			assert.Equal(t, tt.expectedArgs, actualArgs)
			assert.Equal(
				t,
				strings.Count(actualArgs, "--config") == 2,
				strings.Contains(actualArgs, mergeAppendFeatureGateArg),
				"merge append feature gate must be present only when two --config entries are present",
			)
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
	t.Logf("Install command: %s", installCmd.SysProcAttr.CmdLine)
	err := installCmd.Run()
	if err != nil {
		// Log file is in UTF16 use the proper encoding to get a nice rendering on GH logs
		logFile, errLog := os.Open(installLogFile)
		if errLog != nil {
			t.Logf("Failed to open install log file: %v", errLog)
		} else {
			defer logFile.Close()
			scanner := bufio.NewScanner(transform.NewReader(
				logFile,
				unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))
			for scanner.Scan() {
				t.Log(scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				t.Logf("Error reading install log file: %v", err)
			}
		}
		require.Failf(t, "MSI installation failed", "Failed to install the MSI: %v\nArgs: %v", err, args)
	}

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
		startServiceIfStopped(t, service)
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

func startServiceIfStopped(t *testing.T, service *mgr.Service) {
	status, err := service.Query()
	require.NoError(t, err)
	if status.State == svc.Running {
		return
	}

	err = service.Start()
	require.NoError(t, err)
}

func runMsiInstallFailureTest(t *testing.T, test msiTest, msiInstallerPath, expectedLogMessage string) {
	allMSIProperties := make(map[string]string)
	for key, value := range test.genericMSIProperties {
		allMSIProperties[key] = value
	}
	for key, value := range test.collectorMSIProperties {
		allMSIProperties[key] = value
	}

	installLogFile := filepath.Join(os.TempDir(), "install.log")
	args := []string{"/i", msiInstallerPath, "/qn", "/l*v", installLogFile}
	for key, value := range allMSIProperties {
		if strings.Contains(value, "\"") || strings.Contains(value, " ") {
			value = strings.ReplaceAll(value, "\"", "\"\"")
			value = "\"" + value + "\""
		}
		args = append(args, key+"="+value)
	}

	installCmd := exec.Command("msiexec")
	cmdLine := strings.Join(args, " ")
	installCmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "msiexec " + cmdLine}
	t.Logf("Install command: %s", installCmd.SysProcAttr.CmdLine)

	err := installCmd.Run()
	require.Error(t, err, "MSI installation should fail")

	installLog := readInstallLog(t, installLogFile)
	assert.Contains(t, installLog, expectedLogMessage)
}

func assertServiceConfiguration(t *testing.T, msiProperties map[string]string, svcConfig mgr.Config) {
	programDataDir := os.Getenv("PROGRAMDATA")
	require.NotEmpty(t, programDataDir, "PROGRAMDATA environment variable is not set")
	programFilesDir := os.Getenv("PROGRAMFILES")
	require.NotEmpty(t, programFilesDir, "PROGRAMFILES environment variable is not set")

	installRealm := optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_REALM", "us0")
	ingestURL := optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_INGEST_URL", "https://ingest."+installRealm+".observability.splunkcloud.com")
	installMode, ok := msiProperties["SPLUNK_SETUP_COLLECTOR_MODE"]
	if !ok {
		installMode = "agent"
	}

	configFileName := installMode + "_config.yaml"
	configFileFullName := filepath.Join(programDataDir, "Splunk", "OpenTelemetry Collector", configFileName)
	assert.FileExists(t, configFileFullName)
	assert.NoFileExists(t, filepath.Join(programFilesDir, "Splunk", "OpenTelemetry Collector", configFileName))
	if msiProperties["SPLUNK_PLATFORM_URL"] != "" {
		logsConfigFileName := "splunk_logs_config_windows.yaml"
		assert.FileExists(t, filepath.Join(programDataDir, "Splunk", "OpenTelemetry Collector", logsConfigFileName))
		assert.NoFileExists(t, filepath.Join(programFilesDir, "Splunk", "OpenTelemetry Collector", logsConfigFileName))
	}

	expectedEnvVars := map[string]string{
		"SPLUNK_ACCESS_TOKEN": msiProperties["SPLUNK_ACCESS_TOKEN"], // Required install property for a successful start of the service
		"SPLUNK_REALM":        installRealm,
		"SPLUNK_API_URL":      optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_API_URL", "https://api."+installRealm+".observability.splunkcloud.com"),
		"SPLUNK_INGEST_URL":   ingestURL,
		"SPLUNK_HEC_URL":      ingestURL + "/v1/log",
		"SPLUNK_HEC_TOKEN":    optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_HEC_TOKEN", msiProperties["SPLUNK_ACCESS_TOKEN"]),
	}
	if memoryTotalMib, ok := msiProperties["SPLUNK_MEMORY_TOTAL_MIB"]; ok {
		expectedEnvVars["SPLUNK_MEMORY_TOTAL_MIB"] = memoryTotalMib
	}
	if goDebug, ok := msiProperties["GODEBUG"]; ok {
		expectedEnvVars["GODEBUG"] = goDebug
	}
	for _, key := range []string{
		"GOMEMLIMIT",
		"SPLUNK_GATEWAY_URL",
		"SPLUNK_LISTEN_INTERFACE",
		"SPLUNK_MEMORY_LIMIT_MIB",
		"SPLUNK_PLATFORM_URL",
		"SPLUNK_PLATFORM_TOKEN",
		"SPLUNK_PLATFORM_LOGS_INDEX",
	} {
		if value, ok := msiProperties[key]; ok {
			expectedEnvVars[key] = value
		}
	}

	// Verify the environment variables set for the service
	svcEnvVars := getServiceEnvVars(t, "splunk-otel-collector")
	assert.Equal(t, expectedEnvVars, svcEnvVars)

	assert.Equal(t, expectedServiceCommand(t, expectedCollectorServiceArgs(t, msiProperties)), svcConfig.BinaryPathName)
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
		assert.NotContains(t, envVars, parts[0], "Duplicate service environment variable %q", parts[0])
		envVars[parts[0]] = parts[1]
	}

	return envVars
}

func readInstallLog(t *testing.T, installLogFile string) string {
	logFile, err := os.Open(installLogFile)
	require.NoError(t, err, "Failed to open install log file")
	defer logFile.Close()

	var logLines []string
	scanner := bufio.NewScanner(transform.NewReader(
		logFile,
		unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()))
	for scanner.Scan() {
		logLines = append(logLines, scanner.Text())
	}
	require.NoError(t, scanner.Err(), "Error reading install log file")

	return strings.Join(logLines, "\n")
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

func expectedCollectorServiceArgs(t *testing.T, msiProperties map[string]string) string {
	programDataDir := os.Getenv("PROGRAMDATA")
	require.NotEmpty(t, programDataDir, "PROGRAMDATA environment variable is not set")

	collectorServiceArgs := optionalInstallPropertyOrDefault(msiProperties, "COLLECTOR_SVC_ARGS", "")
	collectorServiceArgs = strings.Trim(collectorServiceArgs, "\"")
	collectorServiceArgs = strings.ReplaceAll(collectorServiceArgs, "\"\"", "\"")

	if _, ok := msiProperties["SPLUNK_CONFIG"]; ok {
		return collectorServiceArgs
	}

	if msiProperties["SPLUNK_ACCESS_TOKEN"] != "" {
		installMode := optionalInstallPropertyOrDefault(msiProperties, "SPLUNK_SETUP_COLLECTOR_MODE", "agent")
		configFileFullName := filepath.Join(programDataDir, "Splunk", "OpenTelemetry Collector", installMode+"_config.yaml")
		collectorServiceArgs = appendServiceArg(collectorServiceArgs, "--config "+quotedIfRequired(configFileFullName))
	}

	if msiProperties["SPLUNK_PLATFORM_URL"] != "" {
		logsConfigFileFullName := filepath.Join(programDataDir, "Splunk", "OpenTelemetry Collector", "splunk_logs_config_windows.yaml")
		collectorServiceArgs = appendServiceArg(collectorServiceArgs, "--config "+quotedIfRequired(logsConfigFileFullName))
	}

	if msiProperties["SPLUNK_ACCESS_TOKEN"] != "" && msiProperties["SPLUNK_PLATFORM_URL"] != "" {
		collectorServiceArgs = appendServiceArg(collectorServiceArgs, "--feature-gates=confmap.enableMergeAppendOption")
	}

	return collectorServiceArgs
}

func appendServiceArg(collectorServiceArgs, arg string) string {
	if collectorServiceArgs == "" {
		return arg
	}
	return collectorServiceArgs + " " + arg
}

func expectedServiceCommand(t *testing.T, collectorServiceArgs string) string {
	programFilesDir := os.Getenv("PROGRAMFILES")
	require.NotEmpty(t, programFilesDir, "PROGRAMFILES environment variable is not set")

	collectorDir := filepath.Join(programFilesDir, "Splunk", "OpenTelemetry Collector")
	collectorExe := filepath.Join(collectorDir, "otelcol") + ".exe"

	if collectorServiceArgs == "" {
		return quotedIfRequired(collectorExe)
	}

	return quotedIfRequired(collectorExe) + " " + collectorServiceArgs
}

func quotedIfRequired(s string) string {
	if strings.Contains(s, "\"") || strings.Contains(s, " ") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}
