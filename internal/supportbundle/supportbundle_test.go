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

package supportbundle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() (*logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return newLogger(&buf), &buf
}

func TestRun(t *testing.T) {
	configDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "agent_config.yaml"), []byte("receivers: {}\n"), 0o600))

	tempDir := filepath.Join(t.TempDir(), fmt.Sprintf("splunk-support-bundle-test-%d", time.Now().UnixNano()))
	var stdout bytes.Buffer

	archivePath, err := Run(t.Context(), Options{
		ConfigDir: configDir,
		TempDir:   tempDir,
		Stdout:    &stdout,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(archivePath) })

	assert.FileExists(t, archivePath)
	assert.FileExists(t, filepath.Join(tempDir, "config", "agent_config.yaml"))
	assert.FileExists(t, filepath.Join(tempDir, "stdout.log"))
	assert.Contains(t, stdout.String(), "Support bundle available at:")
}

func TestRun_MissingConfigDir(t *testing.T) {
	tempDir := filepath.Join(t.TempDir(), "splunk-support-bundle-test")
	var stdout bytes.Buffer

	_, err := Run(t.Context(), Options{
		ConfigDir: filepath.Join(t.TempDir(), "does-not-exist"),
		TempDir:   tempDir,
		Stdout:    &stdout,
	})
	assert.Error(t, err)
}

func TestCreateTempDir(t *testing.T) {
	for _, name := range []string{"bundle", filepath.Join("parent", "bundle"), "bundle" + string(filepath.Separator)} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir() + string(filepath.Separator) + name

			require.NoError(t, createTempDir(dir))
			assert.DirExists(t, dir)
			assert.DirExists(t, filepath.Join(dir, logsDirName))
			assert.DirExists(t, filepath.Join(dir, metricsDirName))
			assert.DirExists(t, filepath.Join(dir, zpagesDirName))
		})
	}
}

func TestCreateTempDir_AlreadyExists(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bundle")
	require.NoError(t, createTempDir(dir))

	err := createTempDir(dir)
	assert.ErrorIs(t, err, os.ErrExist)
}

func TestCreateTempDir_ExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bundle")
	require.NoError(t, os.WriteFile(path, []byte("keep existing file"), 0o600))

	require.ErrorIs(t, createTempDir(path), os.ErrExist)
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "keep existing file", string(got))
}

func TestCreateTempDir_Concurrent(t *testing.T) {
	const attempts = 16
	dir := filepath.Join(t.TempDir(), "bundle")
	start := make(chan struct{})
	results := make(chan error, attempts)
	for range attempts {
		go func() {
			<-start
			results <- createTempDir(dir)
		}()
	}
	close(start)

	successes := 0
	for range attempts {
		if err := <-results; err == nil {
			successes++
		} else {
			require.ErrorIs(t, err, os.ErrExist)
		}
	}
	assert.Equal(t, 1, successes)
}

func TestDefaultTempDir(t *testing.T) {
	dir := defaultTempDir()
	assert.Contains(t, dir, "splunk-support-bundle-")
	assert.Equal(t, filepath.Clean(defaultTempRoot()), filepath.Dir(dir))
}
