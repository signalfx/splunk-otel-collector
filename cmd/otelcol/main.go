// Copyright Splunk, Inc.
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
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/shirou/gopsutil/v4/process"
	flag "github.com/spf13/pflag"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/components"
	"github.com/signalfx/splunk-otel-collector/internal/configconverter"
	"github.com/signalfx/splunk-otel-collector/internal/confmapprovider/configsource"
	"github.com/signalfx/splunk-otel-collector/internal/settings"
	"github.com/signalfx/splunk-otel-collector/internal/version"
)

func main() {
	runFromCmdLine(os.Args)
}

func runFromCmdLine(args []string) {
	// TODO: Use same format as the collector
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Handle the cases of running as a TA
	isModularInput, isQueryMode := isModularInputMode(args)
	if isModularInput && isQueryMode {
		// Query modes (scheme/validate) are empty no-ops for now.
		os.Exit(0)
	}

	collectorSettings, err := settings.New(args[1:])
	if err != nil {
		// Exit if --help flag was supplied and usage help was displayed.
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatalf(`invalid settings detected: %v. Use "--help" to show valid usage`, err)
	}

	info := component.BuildInfo{
		Command: "otelcol",
		Version: version.Version,
	}

	confMapConverterFactories := collectorSettings.ConfMapConverterFactories()
	dryRun := configconverter.NewDryRun(collectorSettings.IsDryRun(), confMapConverterFactories)
	expvarConverter := configconverter.GetExpvarConverter()
	confMapConverterFactories = append(confMapConverterFactories,
		configconverter.ConverterFactoryFromConverter(dryRun),
		configconverter.ConverterFactoryFromConverter(expvarConverter))

	configSourceProvider := configsource.New(zap.NewNop(), []configsource.Hook{expvarConverter, dryRun})

	var providerFactories []confmap.ProviderFactory
	for _, pf := range collectorSettings.ConfMapProviderFactories() {
		providerFactories = append(providerFactories, configSourceProvider.Wrap(pf))
	}

	serviceSettings := otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: components.Get,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs:               collectorSettings.ResolverURIs(),
				ProviderFactories:  providerFactories,
				ConverterFactories: confMapConverterFactories,
			},
		},
	}

	allArgs := args[:1]
	allArgs = append(allArgs, collectorSettings.ColCoreArgs()...)
	os.Args = allArgs
	if err = run(serviceSettings); err != nil {
		log.Fatal(err)
	}
}

var otelcolCmdTestCtx context.Context // Use to control termination during tests.

func runInteractive(settings otelcol.CollectorSettings) error {
	cmd := otelcol.NewCommand(settings)
	if otelcolCmdTestCtx != nil {
		cmd.SetContext(otelcolCmdTestCtx)
	}
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("application run finished with error: %w", err)
	}

	return nil
}

func isModularInputMode(args []string) (isModularInput, isQueryMode bool) {
	// SPLUNK_HOME must be defined if this is running as a modular input.
	_, hasSplunkHome := os.LookupEnv("SPLUNK_HOME")
	if !hasSplunkHome {
		return false, false
	}

	// Check if the parent process is splunkd
	if !isParentProcessSplunkd() {
		return false, false
	}

	// This is running as a modular input
	if len(args) == 2 && (args[1] == "--scheme" || args[1] == "--validate-arguments") {
		return true, true
	}

	// TODO: process the XML input for actual modular input operation

	return true, false
}

// isParentProcessSplunkd checks if the parent process name is splunkd (Linux) or splunkd.exe (Windows)
func isParentProcessSplunkd() bool {
	// Get parent process ID
	ppid := os.Getppid()
	parentProc, err := process.NewProcess(int32(ppid)) //nolint:gosec // disable G115
	if err != nil {
		log.Printf("ERROR unable to get parent process: %v\n", err)
		return false
	}

	// Get parent process name
	parentName, err := parentProc.Name()
	if err != nil {
		log.Printf("ERROR unable to get parent process name: %v\n", err)
		return false
	}

	// Check if parent process is splunkd (Linux) or splunkd.exe (Windows)
	return parentName == "splunkd" || parentName == "splunkd.exe"
}
