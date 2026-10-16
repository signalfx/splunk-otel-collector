// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package splunkmodinputexporter

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
)

// This file implements factory for Splunk modular input exporter.

const (
	typeStr        = "splunk_modinput"
	stabilityLevel = component.StabilityLevelDevelopment
)

// NewFactory creates a factory for Splunk modular input exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		exporter.WithLogs(newLogsExporter, stabilityLevel),
		exporter.WithMetrics(newMetricsExporter, stabilityLevel),
		exporter.WithTraces(newTracesExporter, stabilityLevel),
	)
}

// CreateDefaultConfig creates the default configuration for Splunk modular input exporter.
func createDefaultConfig() component.Config {
	return &Config{}
}
