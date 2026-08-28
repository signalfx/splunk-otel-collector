// Copyright Splunk, Inc.
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

// Package baseline is the shared, upstream-only component set that every Splunk
// OTel Collector flavor layers on top of.
//
// The component set lives in components.go. It is hand-maintained for now;
// a follow-up adds an OpenTelemetry Collector Builder manifest and generator
// that produce that file, after which it should not be edited by hand.
//
// A flavor composes its component set by starting from NewBaseline, adding its
// own factories through the Add* methods, and calling Build:
//
//	b := baseline.NewBaseline()
//	b.AddExporters(appdexporter.NewFactory()) // private
//	factories, err := b.Build()
//
// The Add* methods are append-only: a flavor can layer new components on top of
// the baseline but cannot replace or drop the shared set. Adding a component whose
// type collides with a baseline component (an override attempt) is rejected by
// Build.
package baseline

import (
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/service/telemetry/otelconftelemetry"
	"go.uber.org/multierr"
)

// Baseline is the composable, pre-assembly form of a component set: the raw
// factory slices before they are collapsed into otelcol.Factories. The slices
// are unexported and mutated only through the Add* methods, so a flavor can
// layer its own factories on top of the baseline — applying its own feature
// gates inline (add only when a gate is enabled) without any hook machinery —
// but cannot replace or drop the shared set.
type Baseline struct {
	receivers  []receiver.Factory
	processors []processor.Factory
	exporters  []exporter.Factory
	extensions []extension.Factory
	connectors []connector.Factory
}

// AddReceivers appends flavor receiver factories on top of the baseline.
func (b *Baseline) AddReceivers(factories ...receiver.Factory) {
	b.receivers = append(b.receivers, factories...)
}

// AddProcessors appends flavor processor factories on top of the baseline.
func (b *Baseline) AddProcessors(factories ...processor.Factory) {
	b.processors = append(b.processors, factories...)
}

// AddExporters appends flavor exporter factories on top of the baseline.
func (b *Baseline) AddExporters(factories ...exporter.Factory) {
	b.exporters = append(b.exporters, factories...)
}

// AddExtensions appends flavor extension factories on top of the baseline.
func (b *Baseline) AddExtensions(factories ...extension.Factory) {
	b.extensions = append(b.extensions, factories...)
}

// AddConnectors appends flavor connector factories on top of the baseline.
func (b *Baseline) AddConnectors(factories ...connector.Factory) {
	b.connectors = append(b.connectors, factories...)
}

// Build assembles the baseline into otelcol.Factories. Duplicate component types
// (e.g. a contribution colliding with a baseline factory) surface as an error
// from otelcol.MakeFactoryMap rather than silently overwriting.
func (b *Baseline) Build() (otelcol.Factories, error) {
	var errs []error

	extensions, err := otelcol.MakeFactoryMap(b.extensions...)
	if err != nil {
		errs = append(errs, err)
	}
	receivers, err := otelcol.MakeFactoryMap(b.receivers...)
	if err != nil {
		errs = append(errs, err)
	}
	exporters, err := otelcol.MakeFactoryMap(b.exporters...)
	if err != nil {
		errs = append(errs, err)
	}
	processors, err := otelcol.MakeFactoryMap(b.processors...)
	if err != nil {
		errs = append(errs, err)
	}
	connectors, err := otelcol.MakeFactoryMap(b.connectors...)
	if err != nil {
		errs = append(errs, err)
	}

	return otelcol.Factories{
		Extensions: extensions,
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
		Connectors: connectors,
		Telemetry:  otelconftelemetry.NewFactory(),
	}, multierr.Combine(errs...)
}
