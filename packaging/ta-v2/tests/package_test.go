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
	distDir       = "../out/distribution"
	multiOSName   = "Splunk_TA_otel"
	linuxName     = "Splunk_TA_otel_linux_x86_64"
	windowsName   = "Splunk_TA_otel_windows_x86_64"
	multiOSTgz    = multiOSName + ".tgz"
	linuxTgz      = linuxName + ".tgz"
	windowsTgz    = windowsName + ".tgz"
	linuxBinDir   = "linux_x86_64"
	windowsBinDir = "windows_x86_64"
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

// getTarFileContent returns the content of a specific file in a tar.gz archive.
// Returns empty string if the file is not found.
func getTarFileContent(t *testing.T, tgzPath, targetPath string) string {
	file, err := os.Open(tgzPath)
	require.NoError(t, err, "Failed to open tgz file: %s", tgzPath)
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	require.NoError(t, err, "Failed to create gzip reader for: %s", tgzPath)
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err, "Failed to read tar entry")
		if header.Name == targetPath {
			content, readErr := io.ReadAll(tr)
			require.NoError(t, readErr, "Failed to read file content: %s", targetPath)
			return string(content)
		}
	}
	return ""
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

	packageName := linuxName
	// Linux package should contain linux binaries
	linuxBinPath := filepath.Join(packageName, linuxBinDir)
	assert.True(t, containsPath(contents, linuxBinPath),
		"Linux package should contain %s folder", linuxBinPath)

	// Linux package should NOT contain windows binaries
	windowsBinPath := filepath.Join(packageName, windowsBinDir)
	assert.False(t, containsPath(contents, windowsBinPath),
		"Linux package should NOT contain %s folder", windowsBinPath)
}

func TestWindowsPackageContents(t *testing.T) {
	windowsPath := filepath.Join(distDir, windowsTgz)
	require.FileExists(t, windowsPath, "Windows package not found")

	contents := getTarContents(t, windowsPath)
	t.Logf("Windows package contains %d entries", len(contents))

	packageName := windowsName

	// Windows package should contain windows binaries
	windowsBinPath := filepath.Join(packageName, windowsBinDir)
	assert.True(t, containsPath(contents, windowsBinPath),
		"Windows package should contain %s folder", windowsBinPath)

	// Windows package should NOT contain linux binaries
	linuxBinPath := filepath.Join(packageName, linuxBinDir)
	assert.False(t, containsPath(contents, linuxBinPath),
		"Windows package should NOT contain %s folder", linuxBinPath)
}

func TestMultiOSPackageContents(t *testing.T) {
	multiOSPath := filepath.Join(distDir, multiOSTgz)
	require.FileExists(t, multiOSPath, "Multi-OS package not found")

	contents := getTarContents(t, multiOSPath)
	t.Logf("Multi-OS package contains %d entries", len(contents))

	packageName := multiOSName

	// Multi-OS package should contain both linux and windows binaries
	linuxBinPath := filepath.Join(packageName, linuxBinDir)
	assert.True(t, containsPath(contents, linuxBinPath),
		"Multi-OS package should contain %s folder", linuxBinPath)

	windowsBinPath := filepath.Join(packageName, windowsBinDir)
	assert.True(t, containsPath(contents, windowsBinPath),
		"Multi-OS package should contain %s folder", windowsBinPath)
}

func TestPackageMandatoryFiles(t *testing.T) {
	packages := map[string]struct {
		tgz  string
		root string
	}{
		"Multi-OS": {multiOSTgz, multiOSName},
		"Linux":    {linuxTgz, linuxName},
		"Windows":  {windowsTgz, windowsName},
	}

	for pkgName, pkg := range packages {
		pkgPath := filepath.Join(distDir, pkg.tgz)
		root := pkg.root
		t.Run(pkgName, func(t *testing.T) {
			require.FileExists(t, pkgPath, "%s package not found", pkgName)

			mandatoryPaths := []string{
				filepath.Join(root, "configs", "agent_config.yaml"),
				filepath.Join(root, "configs", "gateway_config.yaml"),
				filepath.Join(root, "default", "app.conf"),
				filepath.Join(root, "default", "inputs.conf"),
				filepath.Join(root, "README", "inputs.conf.spec"),
				filepath.Join(root, "static", "appIcon_2x.png"),
				filepath.Join(root, "static", "appIcon.png"),
			}

			addLinuxExecutable := pkgName != "Windows"
			if addLinuxExecutable {
				mandatoryPaths = append(mandatoryPaths, filepath.Join(root, "linux_x86_64", "bin", "Splunk_TA_otel"))
			}

			addWindowsExecutable := pkgName != "Linux"
			if addWindowsExecutable {
				mandatoryPaths = append(mandatoryPaths, filepath.Join(root, "windows_x86_64", "bin", "Splunk_TA_otel.exe"))
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

			// The Makefile sets: version = <git-tag-without-v-prefix>
			appConfContent := getTarFileContent(t, pkgPath, filepath.Join(root, "default", "app.conf"))
			require.NotEmpty(t, appConfContent, "app.conf not found or empty in %s package", pkgName)
			var appConfVersion string
			for _, line := range strings.Split(appConfContent, "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "version") && strings.ContainsRune(trimmed, '=') {
					parts := strings.SplitN(trimmed, "=", 2)
					appConfVersion = strings.TrimSpace(parts[1])
					break
				}
			}
			require.NotEmpty(t, appConfVersion, "version not found in app.conf of %s package", pkgName)
			// Example of expected version formats: "1.2.3", "1.2.3-4-16-gabcdef0"
			require.Regexp(t, `^\d+\.\d+\.\d+(-[0-9]+-[a-z0-9]+)?$`, appConfVersion)

			// Check that the [package] id matches the package name.
			const sectionHeaderSentinel = "["
			var appConfPackageID string
			inPackageSection := false
			for _, line := range strings.Split(appConfContent, "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed == "[package]" {
					inPackageSection = true
					continue
				}
				if inPackageSection {
					trimmed := strings.ReplaceAll(trimmed, " ", "")
					if strings.HasPrefix(trimmed, sectionHeaderSentinel) {
						// We've reached the next section without finding an id, so stop looking for it
						break
					}
					parts := strings.SplitN(trimmed, "=", 2)
					if len(parts) == 2 && parts[0] == "id" {
						appConfPackageID = parts[1]
						break
					}
				}
			}
			assert.Equal(t, root, appConfPackageID,
				"%s app.conf [package] id should be %q", pkgName, root)
		})
	}
}
