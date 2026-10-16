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
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCopyConfig_SymlinkRoot is Unix-only because creating a directory
// symlink on Windows requires elevated privilege or Developer Mode, which
// isn't guaranteed on CI runners; the underlying fix (resolving the root via
// filepath.EvalSymlinks) is itself platform-agnostic.
func TestCopyConfig_SymlinkRoot(t *testing.T) {
	realDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(realDir, "agent_config.yaml"), []byte("receivers: {}\n"), 0o600))

	link := filepath.Join(t.TempDir(), "config-link")
	require.NoError(t, os.Symlink(realDir, link))

	tempDir := t.TempDir()
	logger, _ := newTestLogger()

	require.NoError(t, copyConfig(t.Context(), logger, link, tempDir))
	assert.FileExists(t, filepath.Join(tempDir, configDirName, "agent_config.yaml"))
}

func TestCopyConfig_SymlinkEntries(t *testing.T) {
	const production = "receivers: {}\n"
	const common = "exporters: {}\n"
	src := t.TempDir()
	shared := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(src, "profiles"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "profiles", "production.yaml"), []byte(production), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(shared, "common.yaml"), []byte(common), 0o600))
	for name, target := range map[string]string{
		"agent_config.yaml":  filepath.Join("profiles", "production.yaml"),
		"active.yaml":        "agent_config.yaml",
		"selected":           "profiles",
		"shared":             shared,
		"shared-again":       shared,
		"shared_config.yaml": filepath.Join(shared, "common.yaml"),
	} {
		require.NoError(t, os.Symlink(target, filepath.Join(src, name)))
	}

	tempDir := t.TempDir()
	logger, _ := newTestLogger()
	require.NoError(t, copyConfig(t.Context(), logger, src, tempDir))

	expected := map[string]string{
		"profiles/production.yaml": production,
		"agent_config.yaml":        production,
		"active.yaml":              production,
		"selected/production.yaml": production,
		"shared/common.yaml":       common,
		"shared-again/common.yaml": common,
		"shared_config.yaml":       common,
	}
	for name, content := range expected {
		path := filepath.Join(tempDir, configDirName, filepath.FromSlash(name))
		info, err := os.Lstat(path)
		require.NoError(t, err)
		assert.True(t, info.Mode().IsRegular(), "%s must be a regular file", name)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), name)
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, content, string(got), name)
	}
	for _, name := range []string{"selected", "shared", "shared-again"} {
		info, err := os.Lstat(filepath.Join(tempDir, configDirName, name))
		require.NoError(t, err)
		assert.True(t, info.IsDir(), "%s must be a directory", name)
	}

	archivePath := filepath.Join(t.TempDir(), "bundle.tar.gz")
	require.NoError(t, writeTarGz(t.Context(), tempDir, archivePath))
	f, err := os.Open(archivePath)
	require.NoError(t, err)
	defer f.Close()
	gz, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gz.Close()
	tr := tar.NewReader(gz)
	found := make(map[string]string)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		content, err := io.ReadAll(tr)
		require.NoError(t, err)
		rel, err := filepath.Rel(filepath.Join(filepath.Base(tempDir), configDirName), hdr.Name)
		require.NoError(t, err)
		found[filepath.ToSlash(rel)] = string(content)
	}
	assert.Equal(t, expected, found)
}

func TestCopyConfig_RejectsSymlinkCycles(t *testing.T) {
	for _, tc := range []struct {
		name, link, target, message string
	}{
		{name: "root ancestor", link: "loop", target: ".", message: "symlink cycle"},
		{name: "nested ancestor", link: "nested/back", target: "..", message: "symlink cycle"},
		{name: "same directory", link: "nested/loop", target: ".", message: "symlink cycle"},
		{name: "self-referencing link", link: "loop", target: "loop", message: "resolve configuration path"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := t.TempDir()
			link := filepath.Join(src, filepath.FromSlash(tc.link))
			require.NoError(t, os.MkdirAll(filepath.Dir(link), 0o755))
			require.NoError(t, os.Symlink(tc.target, link))

			tempDir := t.TempDir()
			logger, _ := newTestLogger()
			err := copyConfig(t.Context(), logger, src, tempDir)
			require.ErrorContains(t, err, tc.message)
			_, err = os.Lstat(filepath.Join(tempDir, configDirName, filepath.FromSlash(tc.link)))
			require.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestCopyConfig_RejectsBrokenSymlink(t *testing.T) {
	src := t.TempDir()
	require.NoError(t, os.Symlink("missing.yaml", filepath.Join(src, "agent_config.yaml")))

	logger, _ := newTestLogger()
	err := copyConfig(t.Context(), logger, src, t.TempDir())
	require.ErrorIs(t, err, os.ErrNotExist)
	assert.ErrorContains(t, err, "agent_config.yaml")
}

func TestCopyConfig_RejectsSymlinksOverlappingDestination(t *testing.T) {
	for _, tc := range []struct {
		name, target string
		file         bool
	}{
		{name: "working directory", target: "bundle"},
		{name: "configuration destination", target: "bundle/config"},
		{name: "common ancestor", target: "."},
		{name: "output file", target: "bundle/config/existing.yaml", file: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			src, tempDir := filepath.Join(root, "source"), filepath.Join(root, "bundle")
			require.NoError(t, os.Mkdir(src, 0o755))
			require.NoError(t, os.Mkdir(tempDir, 0o755))
			target := filepath.Join(root, filepath.FromSlash(tc.target))
			if tc.file {
				require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
				require.NoError(t, os.WriteFile(target, []byte("keep existing output"), 0o600))
			}
			require.NoError(t, os.Symlink(target, filepath.Join(src, "linked")))

			logger, _ := newTestLogger()
			err := copyConfig(t.Context(), logger, src, tempDir)
			require.ErrorContains(t, err, "overlaps bundle destination")
			_, err = os.Lstat(filepath.Join(tempDir, configDirName, "linked"))
			require.ErrorIs(t, err, os.ErrNotExist)
			if tc.file {
				got, err := os.ReadFile(target)
				require.NoError(t, err)
				assert.Equal(t, "keep existing output", string(got))
			}
		})
	}
}
