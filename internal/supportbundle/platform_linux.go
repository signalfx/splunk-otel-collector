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

//go:build linux

package supportbundle

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const serviceName = "splunk-otel-collector"

func defaultConfigDir() string {
	return "/etc/otel/collector"
}

func defaultTempRoot() string {
	return os.TempDir()
}

// collectServiceStatus and collectServiceLogs are the only two steps that
// still shell out: there is no reasonable native alternative to systemd's
// D-Bus/journal protocols without pulling in a new dependency, so this
// mirrors exactly what the original script ran.

func collectServiceStatus(ctx context.Context, logger *logger, tempDir string) {
	logger.Info("Getting status...")
	dest := filepath.Join(tempDir, logsDirName, "splunk-otel-collector.txt")
	runCommandToFile(ctx, logger, dest, "systemctl", "status", serviceName)
}

func collectServiceLogs(ctx context.Context, logger *logger, tempDir string) {
	logger.Info("Getting logs...")
	dest := filepath.Join(tempDir, logsDirName, "splunk-otel-collector.log")
	runCommandToFile(ctx, logger, dest, "journalctl", "-u", serviceName)
}

// collectExtras has nothing additional to gather on Linux
func collectExtras(context.Context, *logger, string) {}

func runCommandToFile(ctx context.Context, logger *logger, dest, name string, args ...string) {
	if _, err := exec.LookPath(name); err != nil {
		logger.Warnf("%s could not be found. Please install it to capture a full support bundle.", name)
	}

	f, err := createOutputFile(dest)
	if err != nil {
		logger.Warnf("failed to create %s: %v", dest, err)
		return
	}
	var writeErr error
	defer func() {
		if err := finishOutputFile(f, dest, writeErr); err != nil {
			logger.Warnf("failed to write %s: %v", dest, err)
		}
	}()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = f
	cmd.Stderr = f
	if err := cmd.Run(); err != nil {
		// Keep useful diagnostics when the command exits nonzero.
		_, writeErr = fmt.Fprintf(f, "\n%s exited with error: %v\n", name, err)
	}
}
