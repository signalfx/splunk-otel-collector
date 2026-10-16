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
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	logsDirName    = "logs"
	metricsDirName = "metrics"
	zpagesDirName  = "zpages"
	configDirName  = "config"
)

type Options struct {
	// Stdout receives the same progress output written to
	// <TempDir>/stdout.log. Defaults to os.Stdout.
	Stdout io.Writer
	// ConfigDir is the Splunk OpenTelemetry Collector configuration
	// directory to include in the bundle. Defaults to the platform-specific
	// collector configuration directory.
	ConfigDir string
	// TempDir is the unique working directory used to assemble the bundle
	// contents before archiving. It must not already exist. Defaults to a
	// timestamped directory under the platform-specific temp root.
	TempDir string
}

// Run collects a support bundle according to opts and returns the path to the
// resulting archive (.tar.gz on Linux/Unix, .zip on Windows).
func Run(ctx context.Context, opts Options) (archivePath string, err error) {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	tempDir := opts.TempDir
	if tempDir == "" {
		tempDir = defaultTempDir()
	}
	if createErr := createTempDir(tempDir); createErr != nil {
		return "", createErr
	}

	logFile, err := createOutputFile(filepath.Join(tempDir, "stdout.log"))
	if err != nil {
		return "", fmt.Errorf("create stdout log: %w", err)
	}
	// Keep the progress log even when collection or closing the file fails.
	defer func() {
		if closeErr := logFile.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close stdout log: %w", closeErr))
		}
	}()
	logger := newLogger(stdout, logFile)

	configDir := opts.ConfigDir
	if configDir == "" {
		configDir = defaultConfigDir()
	}
	if copyErr := copyConfig(ctx, logger, configDir, tempDir); copyErr != nil {
		return "", copyErr
	}

	collectServiceStatus(ctx, logger, tempDir)
	collectServiceLogs(ctx, logger, tempDir)
	collectMetrics(ctx, logger, tempDir)
	collectZpages(ctx, logger, tempDir)
	collectHostInfo(ctx, logger, tempDir)
	collectExtras(ctx, logger, tempDir)

	archivePath, err = createArchive(ctx, logger, tempDir)
	if err != nil {
		return "", err
	}
	logger.Infof("Support bundle available at: %s", archivePath)
	logger.Info("Please attach this to your support case")
	return archivePath, nil
}

type logger struct {
	w  io.Writer
	mu sync.Mutex
}

func newLogger(writers ...io.Writer) *logger {
	return &logger{w: io.MultiWriter(writers...)}
}

func (l *logger) write(prefix, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.w, "%s: %s\n", prefix, fmt.Sprintf(format, args...))
}

func (l *logger) Info(msg string) {
	l.write("INFO", "%s", msg)
}

func (l *logger) Infof(format string, args ...any) {
	l.write("INFO", format, args...)
}

func (l *logger) Warn(msg string) {
	l.write("WARN", "%s", msg)
}

func (l *logger) Warnf(format string, args ...any) {
	l.write("WARN", format, args...)
}

func defaultTempDir() string {
	return filepath.Join(defaultTempRoot(), fmt.Sprintf("splunk-support-bundle-%d", time.Now().Unix()))
}

func createTempDir(dir string) error {
	dir = filepath.Clean(dir)
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return fmt.Errorf("create temporary directory parent %q: %w", filepath.Dir(dir), err)
	}
	if err := os.Mkdir(dir, 0o700); err != nil {
		return fmt.Errorf("create temporary directory %q: %w", dir, err)
	}

	for _, sub := range []string{logsDirName, metricsDirName, zpagesDirName} {
		if err := os.Mkdir(filepath.Join(dir, sub), 0o755); err != nil {
			return fmt.Errorf("create temporary directory %q: %w", filepath.Join(dir, sub), err)
		}
	}
	return nil
}
