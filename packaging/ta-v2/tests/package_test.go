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

//go:build ta_v2

package tests

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	distDir                  = "../out/distribution"
	modularInputName         = "Splunk_TA_OTel_Collector"
	multiOSTgz               = "Splunk_TA_OTel_Collector.tgz"
	linuxTgz                 = "Splunk_TA_OTel_Collector_linux_x86_64.tgz"
	windowsTgz               = "Splunk_TA_OTel_Collector_windows_x86_64.tgz"
	linuxBinPath             = "Splunk_TA_OTel_Collector/linux_x86_64/"
	windowsBinPath           = "Splunk_TA_OTel_Collector/windows_x86_64/"
)

// getFileSize returns the size of a file in bytes
func getFileSize(t *testing.T, path string) int64 {
	info, err := os.Stat(path)
	require.NoError(t, err, "Failed to stat file: %s", path)
	return info.Size()
}

// getTarContents returns a list of all file paths in a tar.gz archive
func getTarContents(t *testing.T, tgzPath string) []string {
	file, err := os.Open(tgzPath)
	require.NoError(t, err, "Failed to open tgz file: %s", tgzPath)
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	require.NoError(t, err, "Failed to create gzip reader for: %s", tgzPath)
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var contents []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err, "Failed to read tar entry")

		// Skip "pax" names, paths ending in "/", and exclude files named:
		// "._*", ".DS_Store", and "__MACOSX/"
		if header.Typeflag == tar.TypeXGlobalHeader || header.Typeflag == tar.TypeXHeader {
			continue
		}
		if strings.HasSuffix(header.Name, "/") {
			continue
		}
		baseName := filepath.Base(header.Name)
		if strings.HasPrefix(baseName, "._") || baseName == ".DS_Store" {
			continue
		}
		if strings.Contains(header.Name, "__MACOSX/") {
			continue
		}

		contents = append(contents, header.Name)
	}

	return contents
}

// containsPath checks if any path in the list starts with the given prefix
func containsPath(paths []string, prefix string) bool {
	for _, path := range paths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func TestPackageSizes(t *testing.T) {
	multiOSPath := filepath.Join(distDir, multiOSTgz)
	linuxPath := filepath.Join(distDir, linuxTgz)
	windowsPath := filepath.Join(distDir, windowsTgz)

	// Check that all files exist
	require.FileExists(t, multiOSPath, "Multi-OS package not found")
	require.FileExists(t, linuxPath, "Linux package not found")
	require.FileExists(t, windowsPath, "Windows package not found")

	// Get file sizes
	multiOSSize := getFileSize(t, multiOSPath)
	linuxSize := getFileSize(t, linuxPath)
	windowsSize := getFileSize(t, windowsPath)

	t.Logf("Package sizes:")
	t.Logf("  Multi-OS: %d bytes", multiOSSize)
	t.Logf("  Linux:    %d bytes", linuxSize)
	t.Logf("  Windows:  %d bytes", windowsSize)

	// OS-specific packages should be smaller than multi-OS package
	assert.Less(t, linuxSize, multiOSSize,
		"Linux package (%d bytes) should be smaller than multi-OS package (%d bytes)",
		linuxSize, multiOSSize)

	assert.Less(t, windowsSize, multiOSSize,
		"Windows package (%d bytes) should be smaller than multi-OS package (%d bytes)",
		windowsSize, multiOSSize)
}

func TestLinuxPackageContents(t *testing.T) {
	linuxPath := filepath.Join(distDir, linuxTgz)
	require.FileExists(t, linuxPath, "Linux package not found")

	contents := getTarContents(t, linuxPath)
	t.Logf("Linux package contains %d entries", len(contents))

	// Linux package should contain linux binaries
	assert.True(t, containsPath(contents, linuxBinPath),
		"Linux package should contain %s folder", linuxBinPath)

	// Linux package should NOT contain windows binaries
	assert.False(t, containsPath(contents, windowsBinPath),
		"Linux package should NOT contain %s folder", windowsBinPath)
}

func TestWindowsPackageContents(t *testing.T) {
	windowsPath := filepath.Join(distDir, windowsTgz)
	require.FileExists(t, windowsPath, "Windows package not found")

	contents := getTarContents(t, windowsPath)
	t.Logf("Windows package contains %d entries", len(contents))

	// Windows package should contain windows binaries
	assert.True(t, containsPath(contents, windowsBinPath),
		"Windows package should contain %s folder", windowsBinPath)

	// Windows package should NOT contain linux binaries
	assert.False(t, containsPath(contents, linuxBinPath),
		"Windows package should NOT contain %s folder", linuxBinPath)
}

func TestMultiOSPackageContents(t *testing.T) {
	multiOSPath := filepath.Join(distDir, multiOSTgz)
	require.FileExists(t, multiOSPath, "Multi-OS package not found")

	contents := getTarContents(t, multiOSPath)
	t.Logf("Multi-OS package contains %d entries", len(contents))

	// Multi-OS package should contain both linux and windows binaries
	assert.True(t, containsPath(contents, linuxBinPath),
		"Multi-OS package should contain %s folder", linuxBinPath)

	assert.True(t, containsPath(contents, windowsBinPath),
		"Multi-OS package should contain %s folder", windowsBinPath)
}

func TestPackageMandatoryFiles(t *testing.T) {
	packages := map[string]string{
		"Multi-OS": multiOSTgz,
		"Linux":    linuxTgz,
		"Windows":  windowsTgz,
	}

	for pkgName, pkgFile := range packages {
		pkgPath := filepath.Join(distDir, pkgFile)
		t.Run(pkgName, func(t *testing.T) {
			require.FileExists(t, pkgPath, "%s package not found", pkgName)

			mandatoryPaths := []string{
				modularInputName + "/configs/agent_config.yaml",
				modularInputName + "/default/app.conf",
				modularInputName + "/default/inputs.conf",
				modularInputName + "/README/inputs.conf.spec",
				modularInputName + "/static/appIcon_2x.png",
				modularInputName + "/static/appIcon.png",
			}

			addLinuxExecutable := pkgName != "Windows"
			if addLinuxExecutable {
				mandatoryPaths = append(mandatoryPaths, modularInputName+"/linux_x86_64/bin/Splunk_TA_OTel_Collector")
			}

			addWindowsExecutable := pkgName != "Linux"
			if addWindowsExecutable {
				mandatoryPaths = append(mandatoryPaths, modularInputName+"/windows_x86_64/bin/Splunk_TA_OTel_Collector.exe")
			}

			contents := getTarContents(t, pkgPath)

			for _, mandatoryPath := range mandatoryPaths {
				found := false
				for _, entry := range contents {
					if entry == mandatoryPath {
						found = true
						break
					}
				}
				assert.True(t, found, "%s package should contain %s", pkgName, mandatoryPath)
			}

			for _, path := range contents {
				assert.Contains(t, mandatoryPaths, path,
					"%s package contains unexpected file: %s", pkgName, path)
			}
		})
	}
}
