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
	"fmt"
	"log"
	"os"

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
	// TODO: Use same format as the collector
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	collectorSettings, err := settings.New(os.Args[1:])
	if err != nil {
		// Exit if --help flag was supplied and usage help was displayed.
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		log.Fatalf(`invalid settings detected: %v. Use "--help" to show valid usage`, err)
	}

	info := component.BuildInfo{
		Command: "otelcol",
		Version: version.Version,
	}

	configServer := configconverter.NewConfigServer()

	confMapConverterFactories := collectorSettings.ConfMapConverterFactories()
	dryRun := configconverter.NewDryRun(collectorSettings.IsDryRun(), confMapConverterFactories)
	confMapConverterFactories = append(confMapConverterFactories,
		configconverter.ConverterFactoryFromConverter(dryRun),
		configconverter.ConverterFactoryFromConverter(configServer))

	configSourceProvider := configsource.New(zap.NewNop(), []configsource.Hook{configServer, dryRun})

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

	os.Args = append(os.Args[:1], collectorSettings.ColCoreArgs()...)
	if err = run(serviceSettings); err != nil {
		log.Fatal(err)
	}
}

func runInteractive(settings otelcol.CollectorSettings) error {
	cmd := otelcol.NewCommandMustSetProvider(settings)
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("application run finished with error: %w", err)
	}

	return nil
}
