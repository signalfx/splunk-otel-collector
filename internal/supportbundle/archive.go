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
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
)

type archiveEntry struct {
	path string
	name string
}

// collectArchiveEntries walks root and returns its regular files, ready to
// hand to a tar or zip writer.
func collectArchiveEntries(ctx context.Context, root string) ([]archiveEntry, error) {
	base := filepath.Base(root)
	var entries []archiveEntry

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if !d.Type().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		entries = append(entries, archiveEntry{
			path: path,
			name: filepath.ToSlash(filepath.Join(base, rel)),
		})
		return nil
	})
	return entries, err
}

// contextReader interrupts a copy promptly
type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr contextReader) Read(p []byte) (int, error) {
	if err := cr.ctx.Err(); err != nil {
		return 0, err
	}
	return cr.r.Read(p)
}

// copyEntryContent copies exactly size bytes from src to dst, interrupting
// the copy if ctx is canceled
func copyEntryContent(ctx context.Context, dst io.Writer, src io.Reader, path string, size int64) error {
	n, err := io.CopyN(dst, contextReader{ctx: ctx, r: src}, size)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("copy %s: %w", path, err)
	}
	if n != size {
		return fmt.Errorf("%s changed size while reading: expected %d bytes, copied %d", path, size, n)
	}
	return ctx.Err()
}
