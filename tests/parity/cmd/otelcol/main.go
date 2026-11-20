// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Program otelcol is the collector the parity tests run as their candidate. It
// is test-only: it wires the splunkconf provider to the UF-native receivers and
// exporter under internal/, which are not part of the shipped Splunk OTel
// Collector, plus the few upstream components the hand-written case configs
// use.
package main

import (
	"log"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/service/telemetry/otelconftelemetry"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf"
	"github.com/signalfx/splunk-otel-collector/tests/parity/cmd/otelcol/internal/exporter/splunkhecout"
	"github.com/signalfx/splunk-otel-collector/tests/parity/cmd/otelcol/internal/receiver/splunkbatch"
	"github.com/signalfx/splunk-otel-collector/tests/parity/cmd/otelcol/internal/receiver/splunkmonitor"
	"github.com/signalfx/splunk-otel-collector/tests/parity/cmd/otelcol/internal/receiver/splunkscript"
	"github.com/signalfx/splunk-otel-collector/tests/parity/cmd/otelcol/internal/receiver/splunktcp"
	"github.com/signalfx/splunk-otel-collector/tests/parity/cmd/otelcol/internal/receiver/splunkudp"
	"github.com/signalfx/splunk-otel-collector/tests/parity/cmd/otelcol/internal/receiver/splunkwineventlog"
)

func main() {
	if err := otelcol.NewCommand(settings()).Execute(); err != nil {
		log.Fatal(err)
	}
}

func settings() otelcol.CollectorSettings {
	return otelcol.CollectorSettings{
		BuildInfo: component.BuildInfo{
			Command:     "otelcol",
			Description: "Splunk OTel Collector parity test candidate",
			Version:     "parity",
		},
		Factories: factories,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				ProviderFactories: []confmap.ProviderFactory{
					envprovider.NewFactory(),
					fileprovider.NewFactory(),
					splunkConfProviderFactory(),
				},
				DefaultScheme: "env",
			},
		},
	}
}

// inputMappers are both the receivers the collector runs and the mappers the
// splunkconf provider emits them with, so the two cannot drift apart.
func inputMappers() []splunkconf.MappingReceiverFactory {
	return []splunkconf.MappingReceiverFactory{
		splunkmonitor.NewFactory(),
		splunktcp.NewFactory(),
		splunkudp.NewFactory(),
		splunkscript.NewFactory(),
		splunkbatch.NewFactory(),
		splunkwineventlog.NewFactory(),
	}
}

func splunkConfProviderFactory() confmap.ProviderFactory {
	return splunkconf.NewFactory(
		splunkconf.WithInputMappers(inputMappers()...),
		splunkconf.WithOutputMappers(splunkhecout.NewFactory()),
	)
}

func factories() (otelcol.Factories, error) {
	receivers := []receiver.Factory{filelogreceiver.NewFactory()}
	for _, f := range inputMappers() {
		receivers = append(receivers, f)
	}
	r, err := otelcol.MakeFactoryMap(receivers...)
	if err != nil {
		return otelcol.Factories{}, err
	}
	p, err := otelcol.MakeFactoryMap(resourceprocessor.NewFactory())
	if err != nil {
		return otelcol.Factories{}, err
	}
	e, err := otelcol.MakeFactoryMap[exporter.Factory](
		splunkhecexporter.NewFactory(),
		splunkhecout.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}
	return otelcol.Factories{
		Receivers:  r,
		Processors: p,
		Exporters:  e,
		Telemetry:  otelconftelemetry.NewFactory(),
	}, nil
}
