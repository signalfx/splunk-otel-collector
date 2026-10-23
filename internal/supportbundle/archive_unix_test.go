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

//go:build !windows

package supportbundle

import (
	"archive/tar"
	"compress/gzip"
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteTarGz(t *testing.T) {
	root := filepath.Join(t.TempDir(), "splunk-support-bundle-123")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "logs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "logs", "a.log"), []byte("hello"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "stdout.log"), []byte("world"), 0o600))

	dest := filepath.Join(t.TempDir(), "bundle.tar.gz")
	require.NoError(t, writeTarGz(t.Context(), root, dest))
	require.FileExists(t, dest)

	f, err := os.Open(dest)
	require.NoError(t, err)
	defer f.Close()

	gz, err := gzip.NewReader(f)
	require.NoError(t, err)
	tr := tar.NewReader(gz)

	found := map[string]string{}
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		data, err := io.ReadAll(tr)
		require.NoError(t, err)
		found[hdr.Name] = string(data)
	}

	assert.Equal(t, "hello", found["splunk-support-bundle-123/logs/a.log"])
	assert.Equal(t, "world", found["splunk-support-bundle-123/stdout.log"])
}

func TestWriteTarGz_Permissions(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "secret.txt"), []byte("x"), 0o600))

	dest := filepath.Join(t.TempDir(), "bundle.tar.gz")
	require.NoError(t, writeTarGz(t.Context(), root, dest))

	info, err := os.Stat(dest)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "archive must not be group/world readable regardless of umask")
}

func TestWriteTarGz_RefusesToOverwriteExisting(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o600))

	dest := filepath.Join(t.TempDir(), "bundle.tar.gz")
	require.NoError(t, os.WriteFile(dest, []byte("unrelated existing content"), 0o600))

	err := writeTarGz(t.Context(), root, dest)
	require.Error(t, err)

	got, readErr := os.ReadFile(dest)
	require.NoError(t, readErr)
	assert.Equal(t, "unrelated existing content", string(got))
}

func TestWriteTarGz_RefusesSymlinkDestination(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "unrelated.txt")
	require.NoError(t, os.WriteFile(target, []byte("do not touch"), 0o600))

	link := filepath.Join(dir, "dest.tar.gz")
	require.NoError(t, os.Symlink(target, link))

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o600))

	err := writeTarGz(t.Context(), root, link)
	require.Error(t, err)

	got, readErr := os.ReadFile(target)
	require.NoError(t, readErr)
	assert.Equal(t, "do not touch", string(got))
}

func TestWriteTarGz_RespectsCancellation(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o600))

	dest := filepath.Join(t.TempDir(), "bundle.tar.gz")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := writeTarGz(ctx, root, dest)
	require.ErrorIs(t, err, context.Canceled)
	assert.NoFileExists(t, dest)
}

// TestCreateOutputFile_Permissions covers createOutputFile in isolation. On
// Windows, Go reports a writable file as mode 0666 regardless of the
// requested permissions (Windows access control is ACL-based, not POSIX
// permission bits), so this assertion only holds on Unix.
func TestCreateOutputFile_Permissions(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.bin")

	f, err := createOutputFile(dest)
	require.NoError(t, err)
	defer f.Close()

	info, err := os.Stat(dest)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

// TestWriteTarGz_RespectsCancellation_DuringLastEntry proves cancellation is
// noticed mid-copy rather than only between entries. The last entry is sized
// and filled with incompressible random data so gzip+tar take measurably
// longer than the context timeout, giving a wide, non-flaky margin for the
// cancellation to land mid-copy instead of before or after it.
func TestWriteTarGz_RespectsCancellation_DuringLastEntry(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a-first.txt"), []byte("x"), 0o600))

	big := make([]byte, 32*1024*1024)
	_, err := cryptorand.Read(big)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "z-last.bin"), big, 0o600))

	dest := filepath.Join(t.TempDir(), "bundle.tar.gz")
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Millisecond)
	defer cancel()

	err = writeTarGz(ctx, root, dest)
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

	assert.Equal(t, filepath.Join("/tmp", filepath.Base(tempDir)+".tar.gz"), path)
	assert.FileExists(t, path)
}
