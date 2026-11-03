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
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyConfig(t *testing.T) {
	srcDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "config.d"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "agent_config.yaml"), []byte("receivers: {}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "config.d", "nested.yaml"), []byte("nested: true\n"), 0o600))

	tempDir := t.TempDir()
	logger, _ := newTestLogger()

	require.NoError(t, copyConfig(t.Context(), logger, srcDir, tempDir))

	assert.FileExists(t, filepath.Join(tempDir, configDirName, "agent_config.yaml"))
	assert.FileExists(t, filepath.Join(tempDir, configDirName, "config.d", "nested.yaml"))

	got, err := os.ReadFile(filepath.Join(tempDir, configDirName, "agent_config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "receivers: {}\n", string(got))
}

func TestCopyConfig_MissingDirectory(t *testing.T) {
	logger, _ := newTestLogger()
	err := copyConfig(t.Context(), logger, filepath.Join(t.TempDir(), "does-not-exist"), t.TempDir())
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestCopyConfig_MasksTokens(t *testing.T) {
	const original = "SPLUNK_ACCESS_TOKEN=access-secret\nSPLUNK_REALM=us0\nSPLUNK_HEC_TOKEN=\"hec-secret\"\nOTELCOL_OPTIONS=\"--discovery\"\n"
	const expected = "SPLUNK_ACCESS_TOKEN=********\nSPLUNK_REALM=us0\nSPLUNK_HEC_TOKEN=********\nOTELCOL_OPTIONS=\"--discovery\"\n"
	srcDir, tempDir := t.TempDir(), t.TempDir()
	src := filepath.Join(srcDir, "splunk-otel-collector.conf")
	require.NoError(t, os.WriteFile(src, []byte(original), 0o600))

	logger, _ := newTestLogger()
	require.NoError(t, copyConfig(t.Context(), logger, srcDir, tempDir))
	got, err := os.ReadFile(filepath.Join(tempDir, configDirName, "splunk-otel-collector.conf"))
	require.NoError(t, err)
	assert.Equal(t, expected, string(got))
	got, err = os.ReadFile(src)
	require.NoError(t, err)
	assert.Equal(t, original, string(got))
}

func TestCopyConfig_DoesNotOverwriteExistingFile(t *testing.T) {
	srcDir, tempDir := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "agent_config.yaml"), []byte("new config"), 0o600))
	destDir := filepath.Join(tempDir, configDirName)
	require.NoError(t, os.Mkdir(destDir, 0o755))
	dest := filepath.Join(destDir, "agent_config.yaml")
	require.NoError(t, os.WriteFile(dest, []byte("existing config"), 0o600))

	logger, _ := newTestLogger()
	require.ErrorIs(t, copyConfig(t.Context(), logger, srcDir, tempDir), os.ErrExist)
	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "existing config", string(got))
}

func TestCopyConfig_NotADirectory(t *testing.T) {
	file := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o600))

	logger, _ := newTestLogger()
	err := copyConfig(t.Context(), logger, file, t.TempDir())
	require.Error(t, err)
}

func TestCopyConfig_RejectsDestinationWithinSource(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		relativeConfig, relativeTemp bool
	}{
		{name: "absolute paths"},
		{name: "relative config", relativeConfig: true},
		{name: "relative temp", relativeTemp: true},
		{name: "relative paths", relativeConfig: true, relativeTemp: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			configDir, err := filepath.EvalSymlinks(t.TempDir())
			require.NoError(t, err)
			tempDir := filepath.Join(configDir, "staging")
			require.NoError(t, os.MkdirAll(tempDir, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(configDir, "agent_config.yaml"), []byte("x"), 0o600))
			t.Chdir(configDir)

			src, dst := configDir, tempDir
			if tc.relativeConfig {
				src = "."
			}
			if tc.relativeTemp {
				dst = "staging"
			}

			logger, _ := newTestLogger()
			err = copyConfig(t.Context(), logger, src, dst)
			require.ErrorContains(t, err, "overlaps bundle destination")
			assert.NoDirExists(t, filepath.Join(tempDir, configDirName))
		})
	}
}

func TestCopyConfig_RelativePaths(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "source")
	tempDir := filepath.Join(root, "bundle")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.MkdirAll(tempDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "agent_config.yaml"), []byte("receivers: {}\n"), 0o600))
	t.Chdir(root)

	logger, _ := newTestLogger()
	require.NoError(t, copyConfig(t.Context(), logger, "source", "bundle"))
	got, err := os.ReadFile(filepath.Join(tempDir, configDirName, "agent_config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "receivers: {}\n", string(got))
}

func TestCopyConfig_RejectsEqualSourceAndDestination(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "agent_config.yaml"), []byte("x"), 0o600))

	logger, _ := newTestLogger()
	// tempDir == configDir means dest (tempDir/config) is trivially nested
	// inside configDir too.
	err := copyConfig(t.Context(), logger, dir, dir)
	require.Error(t, err)
}

func TestCopyConfig_RespectsCancellation(t *testing.T) {
	srcDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "agent_config.yaml"), []byte("x"), 0o600))

	tempDir := t.TempDir()
	logger, _ := newTestLogger()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := copyConfig(ctx, logger, srcDir, tempDir)
	require.ErrorIs(t, err, context.Canceled)
	assert.NoDirExists(t, filepath.Join(tempDir, configDirName))
}

// TestCopyConfig_NonPermissionFailureNotMisreported ensures a copy failure
// unrelated to permissions (here, a destination path component blocked by a
// pre-existing regular file) is not described as a permission problem.
func TestCopyConfig_NonPermissionFailureNotMisreported(t *testing.T) {
	srcDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "subdir", "file.txt"), []byte("x"), 0o600))

	tempDir := t.TempDir()
	destConfigDir := filepath.Join(tempDir, configDirName)
	require.NoError(t, os.MkdirAll(destConfigDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(destConfigDir, "subdir"), []byte("blocking"), 0o600))

	logger, _ := newTestLogger()
	err := copyConfig(t.Context(), logger, srcDir, tempDir)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "permission denied")
}

func TestIsWithin(t *testing.T) {
	cases := []struct {
		parent, child string
		want          bool
	}{
		{parent: "/a", child: "/a", want: true},
		{parent: "/a", child: "/a/b", want: true},
		{parent: "/a", child: "/a/b/c", want: true},
		{parent: "/a", child: "/b", want: false},
		{parent: "/a", child: "/ab", want: false},
		{parent: "/a/b", child: "/a", want: false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, isWithin(c.parent, c.child), "isWithin(%q, %q)", c.parent, c.child)
	}
}
