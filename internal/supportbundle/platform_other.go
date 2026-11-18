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

//go:build !linux && !windows

// This file exists so the package (and its test suite) build and run on
// platforms the collector is not packaged for, such as macOS during local
// development and in CI. The Splunk OpenTelemetry Collector only ships
// splunk-support-bundle on Linux and Windows.

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
	return "/etc/otel/collector"
}

func defaultTempRoot() string {
	return os.TempDir()
}

func collectServiceStatus(_ context.Context, logger *logger, tempDir string) {
	logger.Info("Getting status...")
	writeUnsupported(logger, filepath.Join(tempDir, logsDirName, "splunk-otel-collector.txt"))
}

func collectServiceLogs(_ context.Context, logger *logger, tempDir string) {
	logger.Info("Getting logs...")
	writeUnsupported(logger, filepath.Join(tempDir, logsDirName, "splunk-otel-collector.log"))
}

func collectExtras(context.Context, *logger, string) {}

func writeUnsupported(logger *logger, dest string) {
	msg := fmt.Sprintf("Service status/log collection for %s is not supported on %s.\n", serviceName, runtime.GOOS)
	writeOrWarn(logger, dest, msg)
}
