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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Permission bits on the created file are intentionally not asserted here:
// Go reports a writable Windows file as mode 0666 regardless of what
// permission bits were requested at creation (see
// https://go.dev/src/os/types_windows.go), since Windows access control is
// ACL-based rather than POSIX permission bits. That assertion lives in
// archive_unix_test.go instead.
func TestCreateOutputFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.bin")

	f, err := createOutputFile(dest)
	require.NoError(t, err)
	defer f.Close()

	assert.FileExists(t, dest)
}

func TestCreateOutputFile_RefusesExisting(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.bin")
	require.NoError(t, os.WriteFile(dest, []byte("existing"), 0o600))

	_, err := createOutputFile(dest)
	require.Error(t, err)

	got, readErr := os.ReadFile(dest)
	require.NoError(t, readErr)
	assert.Equal(t, "existing", string(got), "existing file must not be truncated")
}

func TestFinishOutputFile_PropagatesErrorsAndCleansUp(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.bin")
	f, err := createOutputFile(dest)
	require.NoError(t, err)

	err = finishOutputFile(f, dest, errors.New("writer close failed"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "writer close failed")
	assert.NoFileExists(t, dest, "partially written output must be removed on failure")
}

func TestFinishOutputFile_Success(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.bin")
	f, err := createOutputFile(dest)
	require.NoError(t, err)

	require.NoError(t, finishOutputFile(f, dest))
	assert.FileExists(t, dest)
}

func TestCollectArchiveEntries_RespectsCancellation(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o600))

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := collectArchiveEntries(ctx, root)
	require.ErrorIs(t, err, context.Canceled)
}

func TestContextReader(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	r := contextReader{ctx: ctx, r: strings.NewReader("hello world")}

	buf := make([]byte, 5)
	n, err := r.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", string(buf))

	cancel()
	_, err = r.Read(buf)
	require.ErrorIs(t, err, context.Canceled)
}
