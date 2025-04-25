// Copyright Splunk, Inc.
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

package packaging

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mapset "github.com/deckarep/golang-set/v2"
)

func TestPackaging(t *testing.T) {
	tempPath := t.TempDir()
	actualPath := filepath.Join(tempPath, "Sample_Addon.tgz")
	err := PackageAddon(filepath.Join("testdata", "Sample_Addon"), actualPath)
	require.NoError(t, err)
	sha256sum, err := calculateSHA256(actualPath)
	require.NoError(t, err)
	assert.Equal(t, "94627778c3ba3e20ac0a00733ad054f1c1ed4ded57b344058613345995dcb3c8", sha256sum)

	// check paths
	paths, err := getFilesFromTarGz(actualPath)
	require.NoError(t, err)
	expectedPaths := mapset.NewSet[string]("Sample_Addon/default/inputs.conf", "Sample_Addon/README/inputs.conf.spec", "Sample_Addon/linux_x86_64/bin/Sample_Addon")
	assert.EqualValues(t, expectedPaths, paths, "expected paths to match, missing: %v ; extra: %v", expectedPaths.Difference(paths), paths.Difference(expectedPaths))

}

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getFilesFromTarGz(tarGzPath string) (mapset.Set[string], error) {
	fileSet := mapset.NewSet[string]()

	file, err := os.Open(tarGzPath)
	if err != nil {
		return fileSet, fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fileSet, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fileSet, fmt.Errorf("error reading tar entry: %w", err)
		}

		if header.Typeflag != tar.TypeDir {
			normalizedPath := filepath.ToSlash(header.Name)
			fileSet.Add(normalizedPath)
		}
	}

	return fileSet, nil
}
