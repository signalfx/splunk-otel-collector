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
	"fmt"
	"os"
	"path/filepath"
)

// createArchive zips tempDir into the current working directory, matching
// where the original PowerShell script (via ${PWD}) saved its .zip.
func createArchive(ctx context.Context, logger *logger, tempDir string) (string, error) {
	logger.Info("Creating support bundle...")

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("support bundle was not properly created: %w", err)
	}
	dest := filepath.Join(cwd, filepath.Base(tempDir)+".zip")
	if err := writeZip(ctx, tempDir, dest); err != nil {
		return "", fmt.Errorf("support bundle was not properly created: %w", err)
	}
	return dest, nil
}

func writeZip(ctx context.Context, root, dest string) error {
	entries, err := collectArchiveEntries(ctx, root)
	if err != nil {
		return err
	}

	f, err := createOutputFile(dest)
	if err != nil {
		return err
	}

	zw := zip.NewWriter(f)

	writeErr := writeZipEntries(ctx, zw, entries)
	return finishOutputFile(f, dest, writeErr, zw.Close())
}

func writeZipEntries(ctx context.Context, zw *zip.Writer, entries []archiveEntry) error {
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := addZipEntry(ctx, zw, e); err != nil {
			return err
		}
	}
	return nil
}

func addZipEntry(ctx context.Context, zw *zip.Writer, e archiveEntry) error {
	in, err := os.Open(e.path)
	if err != nil {
		return fmt.Errorf("open archive entry %s: %w", e.path, err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat archive entry %s: %w", e.path, err)
	}

	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("build zip header for %s: %w", e.path, err)
	}
	hdr.Name = e.name
	hdr.Method = zip.Deflate

	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return fmt.Errorf("write zip header for %s: %w", e.path, err)
	}

	return copyEntryContent(ctx, w, in, e.path, info.Size())
}
