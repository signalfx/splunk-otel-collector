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
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// createArchive tars and gzips tempDir to /tmp/<base of tempDir>.tar.gz,
// matching the fixed /tmp destination the original shell script always used
// regardless of where -t pointed.
func createArchive(ctx context.Context, logger *logger, tempDir string) (string, error) {
	logger.Info("Creating tarball...")

	dest := filepath.Join(os.TempDir(), filepath.Base(tempDir)+".tar.gz")
	if err := writeTarGz(ctx, tempDir, dest); err != nil {
		return "", fmt.Errorf("support bundle was not properly created: %w", err)
	}
	return dest, nil
}

func writeTarGz(ctx context.Context, root, dest string) error {
	entries, err := collectArchiveEntries(ctx, root)
	if err != nil {
		return err
	}

	f, err := createOutputFile(dest)
	if err != nil {
		return err
	}

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	writeErr := writeTarEntries(ctx, tw, entries)
	closeErr := errors.Join(tw.Close(), gz.Close())
	return finishOutputFile(f, dest, writeErr, closeErr)
}

func writeTarEntries(ctx context.Context, tw *tar.Writer, entries []archiveEntry) error {
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := addTarEntry(ctx, tw, e); err != nil {
			return err
		}
	}
	return nil
}

func addTarEntry(ctx context.Context, tw *tar.Writer, e archiveEntry) error {
	in, err := os.Open(e.path)
	if err != nil {
		return fmt.Errorf("open archive entry %s: %w", e.path, err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat archive entry %s: %w", e.path, err)
	}

	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("build tar header for %s: %w", e.path, err)
	}
	hdr.Name = e.name
	if writeErr := tw.WriteHeader(hdr); writeErr != nil {
		return fmt.Errorf("write tar header for %s: %w", e.path, writeErr)
	}

	return copyEntryContent(ctx, tw, in, e.path, info.Size())
}
