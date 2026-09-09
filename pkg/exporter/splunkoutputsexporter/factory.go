// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkoutputsexporter

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
)

// NewFactory returns an exporter factory for splunk_outputs.
//
// By default, the factory supports tarunner's built-in [httpout] stanza.
// Additional schemes, or overrides for built-in schemes, can be registered with
// WithSubExporter.
func NewFactory(opts ...Option) exporter.Factory {
	options := newFactoryOptions(opts...)
	return exporter.NewFactory(
		component.MustNewType("splunk_outputs"),
		createDefaultConfig,
		exporter.WithLogs(options.createLogsFunc, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return Config{}
}
