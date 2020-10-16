// Copyright 2020 Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Program otelcol is the OpenTelemetry Collector that collects stats
// and traces and exports to a configured backend.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/version"
)

const ballastEnvVarName = "SPLUNK_BALLAST_SIZE_MIB"

func main() {
	factories, err := components.Get()
	if err != nil {
		log.Fatalf("failed to build default components: %v", err)
	}

	info := component.ApplicationStartInfo{
		ExeName:  "otelcol",
		LongName: "OpenTelemetry Collector",
		Version:  version.Version,
		GitHash:  version.GitHash,
	}

	useBallastSizeFromEnvVar()

	if err := run(service.Parameters{ApplicationStartInfo: info, Factories: factories}); err != nil {
		log.Fatal(err)
	}
}

func useBallastSizeFromEnvVar() {

	// Check if the ballast is specified via the env var.
	ballastSize := os.Getenv(ballastEnvVarName)
	if ballastSize != "" {
		// Check if it is a numeric value.
		_, err := strconv.Atoi(ballastSize)
		if err != nil {
			log.Fatalf("Expected a number in %s env variable but got %s", ballastEnvVarName, ballastSize)
		}

		// Inject the command line flag that controls the ballast size.
		os.Args = append(os.Args, "--mem-ballast-size-mib="+ballastSize)
	}
}

func runInteractive(params service.Parameters) error {
	app, err := service.New(params)
	if err != nil {
		return fmt.Errorf("failed to construct the application: %w", err)
	}

	err = app.Run()
	if err != nil {
		return fmt.Errorf("application run finished with error: %w", err)
	}

	return nil
}
