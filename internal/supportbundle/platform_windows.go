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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const serviceName = "splunk-otel-collector"

func defaultConfigDir() string {
	base := os.Getenv("PROGRAMDATA")
	if base == "" {
		base = `C:\ProgramData`
	}
	return filepath.Join(base, "Splunk", "OpenTelemetry Collector")
}

func defaultTempRoot() string {
	base := os.Getenv("PROGRAMFILES")
	if base == "" {
		base = `C:\Program Files`
	}
	return filepath.Join(base, "Splunk", "OpenTelemetry Collector")
}

func collectServiceStatus(_ context.Context, logger *logger, tempDir string) {
	logger.Info("Getting status...")
	dest := filepath.Join(tempDir, logsDirName, serviceName+".txt")
	// Still need to implement
	writeUnsupported(logger, dest)
}

func collectServiceLogs(_ context.Context, logger *logger, tempDir string) {
	logger.Info("Getting logs...")
	dest := filepath.Join(tempDir, logsDirName, serviceName+".log")
	// Still need to implement
	writeUnsupported(logger, dest)
}

// TO DO: collects MSI/registry information here.
func collectExtras(context.Context, *logger, string) {}

func writeUnsupported(logger *logger, dest string) {
	msg := fmt.Sprintf("Service status/log collection for %s is not supported on %s.\n", serviceName, runtime.GOOS)
	writeOrWarn(logger, dest, msg)
}
