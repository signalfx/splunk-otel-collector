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

package supportbundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const productDisplayName = "Splunk OpenTelemetry Collector"

// collectMSIInfo gathers MSI install/uninstall registry data for the
// collector product using golang.org/x/sys/windows/registry, in place of the
// two `Get-ItemProperty ... | Where-Object { $_.DisplayName -eq ... }`
// pipelines the original script ran.
func collectMSIInfo(logger *logger, tempDir string) {
	logger.Info("Getting MSI information...")

	msiDir := filepath.Join(tempDir, "msi")
	if err := os.MkdirAll(msiDir, 0o755); err != nil {
		logger.Warnf("failed to create %s: %v", msiDir, err)
		return
	}
	writeOrWarn(logger, filepath.Join(msiDir, "user-installs.txt"), collectUserInstalls())
	writeOrWarn(logger, filepath.Join(msiDir, "uninstall-info.txt"), collectUninstallInfo())
}

// collectUserInstalls walks
// HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Installer\UserData\*\Products\*\InstallProperties
// for the key whose DisplayName matches the collector product.
func collectUserInstalls() string {
	const userDataPath = `SOFTWARE\Microsoft\Windows\CurrentVersion\Installer\UserData`

	root, err := registry.OpenKey(registry.LOCAL_MACHINE, userDataPath, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return fmt.Sprintf("ERROR: %v\n", err)
	}
	defer root.Close()

	sids, err := root.ReadSubKeyNames(-1)
	if err != nil {
		return fmt.Sprintf("ERROR: %v\n", err)
	}

	var sb strings.Builder
	for _, sid := range sids {
		productsPath := userDataPath + `\` + sid + `\Products`
		productsKey, err := registry.OpenKey(registry.LOCAL_MACHINE, productsPath, registry.ENUMERATE_SUB_KEYS)
		if err != nil {
			continue
		}
		products, err := productsKey.ReadSubKeyNames(-1)
		productsKey.Close()
		if err != nil {
			continue
		}
		for _, product := range products {
			writeMatchingProduct(&sb, productsPath+`\`+product+`\InstallProperties`)
		}
	}
	return sb.String()
}

// collectUninstallInfo walks
// HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\* for the key
// whose DisplayName matches the collector product.
func collectUninstallInfo() string {
	const base = `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`

	root, err := registry.OpenKey(registry.LOCAL_MACHINE, base, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return fmt.Sprintf("ERROR: failed to open registry key %q: %v\n", base, err)
	}
	defer root.Close()

	names, err := root.ReadSubKeyNames(-1)
	if err != nil {
		return fmt.Sprintf("ERROR: failed to enumerate subkeys under registry key %q: %v\n", base, err)
	}

	var sb strings.Builder
	for _, name := range names {
		writeMatchingProduct(&sb, base+`\`+name)
	}
	return sb.String()
}

// writeMatchingProduct appends a summary of the registry key at path to sb,
// if and only if its DisplayName value matches productDisplayName.
func writeMatchingProduct(sb *strings.Builder, path string) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer key.Close()

	displayName, _, err := key.GetStringValue("DisplayName")
	if err != nil || displayName != productDisplayName {
		return
	}

	fmt.Fprintf(sb, "Key: %s\n", path)
	for _, name := range []string{"DisplayName", "DisplayVersion", "InstallDate", "InstallSource", "InstallLocation", "Publisher"} {
		if v, _, err := key.GetStringValue(name); err == nil {
			fmt.Fprintf(sb, "  %s: %s\n", name, v)
		}
	}
	sb.WriteString("\n")
}

// collectServiceEnvironment reads the Environment multi-string value of the
// collector service's registry key, equivalent to the original script's
// getServiceEnvironment, masking any TOKEN-named entries.
func collectServiceEnvironment(logger *logger, tempDir string) {
	logger.Info("Getting splunk-otel-collector service environment variables...")
	dest := filepath.Join(tempDir, configDirName, "service_environment.txt")

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\`+serviceName, registry.QUERY_VALUE)
	if err != nil {
		logger.Warnf("Could not find environment variables for %s service: %v", serviceName, err)
		return
	}
	defer key.Close()

	env, _, err := key.GetStringsValue("Environment")
	if err != nil {
		logger.Warnf("Could not read environment variables for %s service: %v", serviceName, err)
		return
	}

	content := strings.Join(maskTokenEnv(env), "\n")
	if content != "" {
		content += "\n"
	}
	writeOrWarn(logger, dest, content)
}

func maskTokenEnv(entries []string) []string {
	masked := make([]string, len(entries))
	for i, entry := range entries {
		key, value, found := strings.Cut(entry, "=")
		if found && value != "" && strings.Contains(strings.ToUpper(key), "TOKEN") {
			entry = key + "=********"
		}
		masked[i] = entry
	}
	return masked
}
