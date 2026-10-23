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
	"archive/zip"
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteZip(t *testing.T) {
	root := filepath.Join(t.TempDir(), "splunk-support-bundle-123")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "logs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "logs", "a.log"), []byte("hello"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "stdout.log"), []byte("world"), 0o600))

	dest := filepath.Join(t.TempDir(), "bundle.zip")
	require.NoError(t, writeZip(t.Context(), root, dest))
	require.FileExists(t, dest)

	zr, err := zip.OpenReader(dest)
	require.NoError(t, err)
	defer zr.Close()

	found := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		data, err := io.ReadAll(rc)
		rc.Close()
		require.NoError(t, err)
		found[f.Name] = string(data)
	}

	assert.Equal(t, "hello", found["splunk-support-bundle-123/logs/a.log"])
	assert.Equal(t, "world", found["splunk-support-bundle-123/stdout.log"])
}

func TestWriteZip_RefusesToOverwriteExisting(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o600))

	dest := filepath.Join(t.TempDir(), "bundle.zip")
	require.NoError(t, os.WriteFile(dest, []byte("unrelated existing content"), 0o600))

	err := writeZip(t.Context(), root, dest)
	require.Error(t, err)

	got, readErr := os.ReadFile(dest)
	require.NoError(t, readErr)
	assert.Equal(t, "unrelated existing content", string(got))
}

func TestWriteZip_RespectsCancellation(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o600))

	dest := filepath.Join(t.TempDir(), "bundle.zip")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := writeZip(ctx, root, dest)
	require.ErrorIs(t, err, context.Canceled)
	assert.NoFileExists(t, dest)
}

// TestWriteZip_RespectsCancellation_DuringLastEntry proves cancellation is
// noticed mid-copy rather than only between entries. The last entry is sized
// and filled with incompressible random data so compressing it takes
// measurably longer than the context timeout, giving a wide, non-flaky
// margin for the cancellation to land mid-copy instead of before or after it.
func TestWriteZip_RespectsCancellation_DuringLastEntry(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a-first.txt"), []byte("x"), 0o600))

	big := make([]byte, 32*1024*1024)
	_, err := cryptorand.Read(big)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "z-last.bin"), big, 0o600))

	dest := filepath.Join(t.TempDir(), "bundle.zip")
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Millisecond)
	defer cancel()

	err = writeZip(ctx, root, dest)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NoFileExists(t, dest)
}

func TestCreateArchive(t *testing.T) {
	tempDir := filepath.Join(t.TempDir(), fmt.Sprintf("splunk-support-bundle-%d", time.Now().UnixNano()))
	require.NoError(t, os.MkdirAll(tempDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "stdout.log"), []byte("x"), 0o600))

	logger, _ := newTestLogger()
	path, err := createArchive(t.Context(), logger, tempDir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(path) })

	assert.FileExists(t, path)
}
