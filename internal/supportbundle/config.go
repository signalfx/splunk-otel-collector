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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func copyConfig(ctx context.Context, logger *logger, configDir, tempDir string) error {
	logger.Info("Getting configuration...")

	resolvedConfigDir, err := resolvePath(configDir)
	if err != nil {
		return fmt.Errorf("resolve directory (%s): %w", configDir, err)
	}
	info, err := os.Stat(resolvedConfigDir)
	if err != nil {
		return fmt.Errorf("could not access directory (%s): %w", configDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("configuration path (%s) is not a directory", configDir)
	}

	resolvedTempDir, err := resolvePath(tempDir)
	if err != nil {
		return fmt.Errorf("resolve directory (%s): %w", tempDir, err)
	}

	dest := filepath.Join(resolvedTempDir, configDirName)
	if copyErr := copyDir(ctx, resolvedConfigDir, dest); copyErr != nil {
		return fmt.Errorf("copy directory (%s): %w", configDir, copyErr)
	}
	return nil
}

func resolvePath(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return filepath.Abs(resolved)
}

func isWithin(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func copyDir(ctx context.Context, src, dst string) error {
	ancestors := make(map[string]bool)
	var copyEntry func(string, string) error
	copyEntry = func(source, target string) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		resolved, err := resolvePath(source)
		if err != nil {
			return fmt.Errorf("resolve configuration path %s: %w", source, err)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return err
		}
		if isWithin(dst, resolved) || (info.IsDir() && isWithin(resolved, dst)) {
			return fmt.Errorf("configuration path %s overlaps bundle destination %s", source, dst)
		}

		if info.Mode().IsRegular() {
			return copyFile(ctx, resolved, target)
		}
		if !info.IsDir() {
			return nil
		}
		if ancestors[resolved] {
			return fmt.Errorf("configuration directory symlink cycle at %s", source)
		}
		ancestors[resolved] = true
		defer delete(ancestors, resolved)

		if mkdirErr := os.MkdirAll(target, 0o755); mkdirErr != nil {
			return mkdirErr
		}
		entries, err := os.ReadDir(resolved)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyEntry(filepath.Join(resolved, entry.Name()), filepath.Join(target, entry.Name())); err != nil {
				return err
			}
		}
		return ctx.Err()
	}
	return copyEntry(src, dst)
}

func copyFile(ctx context.Context, src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}

	var reader io.Reader = in
	size := info.Size()
	if filepath.Base(dst) == "splunk-otel-collector.conf" {
		var contents strings.Builder
		if copyErr := copyEntryContent(ctx, &contents, in, src, size); copyErr != nil {
			return copyErr
		}
		lines := strings.Split(contents.String(), "\n")
		for i, line := range lines {
			if key, _, ok := strings.Cut(line, "="); ok && strings.Contains(strings.ToUpper(key), "TOKEN") {
				lines[i] = key + "=********"
			}
		}
		masked := strings.Join(lines, "\n")
		reader = strings.NewReader(masked)
		size = int64(len(masked))
	}

	out, err := createOutputFile(dst)
	if err != nil {
		return err
	}

	copyErr := copyEntryContent(ctx, out, reader, src, size)
	return finishOutputFile(out, dst, copyErr)
}
