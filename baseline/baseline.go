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
// The component set lives in components.go. Its component-to-module metadata
// is generated from that source and baseline/go.mod by go generate.
//
// A flavor composes its component set by starting from NewBaseline, adding its
// own factories with their owning Go module paths, and calling Build:
//
//	b := baseline.NewBaseline()
//	b.AddExportersWithModulePath("example.com/appd/exporter", appdexporter.NewFactory())
//	factories, err := b.Build()
//
// The Add*WithModulePath methods are append-only: a flavor can layer new
// components on top of the baseline but cannot replace or drop the shared set.
// Adding a component whose type collides with a baseline component (an override
// attempt) is rejected by Build.
package baseline

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"go.opentelemetry.io/collector/component"
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
// are unexported and mutated only through the Add*WithModulePath methods, so a
// flavor can layer its own factories on top of the baseline — applying its own
// feature gates inline (add only when a gate is enabled) without any hook
// machinery — but cannot replace or drop the shared set.
type Baseline struct {
	receivers        []receiver.Factory
	receiverModules  []moduleMetadata
	processors       []processor.Factory
	processorModules []moduleMetadata
	exporters        []exporter.Factory
	exporterModules  []moduleMetadata
	extensions       []extension.Factory
	extensionModules []moduleMetadata
	connectors       []connector.Factory
	connectorModules []moduleMetadata
	moduleVersions   map[string]string
}

// Option configures a Baseline.
type Option func(*Baseline)

// WithModuleVersion supplies the release version for a component module whose
// version is unavailable from Go build information. This is normally needed
// only for components built from the final binary's main module, which Go
// reports as "(devel)" when built from a source checkout.
func WithModuleVersion(modulePath, version string) Option {
	return func(b *Baseline) {
		if b.moduleVersions == nil {
			b.moduleVersions = map[string]string{}
		}
		b.moduleVersions[modulePath] = version
	}
}

// AddReceiversWithModulePath appends flavor receiver factories and records
// their owning Go module path. Build resolves the selected module version from
// the final binary.
func (b *Baseline) AddReceiversWithModulePath(modulePath string, factories ...receiver.Factory) {
	b.receivers = append(b.receivers, factories...)
	b.receiverModules = appendRepeated(b.receiverModules, modulePath, len(factories))
}

// AddProcessorsWithModulePath appends flavor processor factories and records
// their owning Go module path. Build resolves the selected module version from
// the final binary.
func (b *Baseline) AddProcessorsWithModulePath(modulePath string, factories ...processor.Factory) {
	b.processors = append(b.processors, factories...)
	b.processorModules = appendRepeated(b.processorModules, modulePath, len(factories))
}

// AddExportersWithModulePath appends flavor exporter factories and records
// their owning Go module path. Build resolves the selected module version from
// the final binary.
func (b *Baseline) AddExportersWithModulePath(modulePath string, factories ...exporter.Factory) {
	b.exporters = append(b.exporters, factories...)
	b.exporterModules = appendRepeated(b.exporterModules, modulePath, len(factories))
}

// AddExtensionsWithModulePath appends flavor extension factories and records
// their owning Go module path. Build resolves the selected module version from
// the final binary.
func (b *Baseline) AddExtensionsWithModulePath(modulePath string, factories ...extension.Factory) {
	b.extensions = append(b.extensions, factories...)
	b.extensionModules = appendRepeated(b.extensionModules, modulePath, len(factories))
}

// AddConnectorsWithModulePath appends flavor connector factories and records
// their owning Go module path. Build resolves the selected module version from
// the final binary.
func (b *Baseline) AddConnectorsWithModulePath(modulePath string, factories ...connector.Factory) {
	b.connectors = append(b.connectors, factories...)
	b.connectorModules = appendRepeated(b.connectorModules, modulePath, len(factories))
}

// Build assembles the baseline into otelcol.Factories. Duplicate component types
// (e.g. a contribution colliding with a baseline factory) surface as an error
// from otelcol.MakeFactoryMap rather than silently overwriting.
func (b *Baseline) Build() (otelcol.Factories, error) {
	var errs []error
	moduleResolver, err := newBuildModuleResolver(b.moduleVersions)
	if err != nil {
		return otelcol.Factories{}, fmt.Errorf("resolve baseline module metadata: %w", err)
	}

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
	extensionModules, err := makeModuleMap(b.extensions, b.extensionModules, moduleResolver)
	if err != nil {
		errs = append(errs, fmt.Errorf("extensions: %w", err))
	}
	receiverModules, err := makeModuleMap(b.receivers, b.receiverModules, moduleResolver)
	if err != nil {
		errs = append(errs, fmt.Errorf("receivers: %w", err))
	}
	exporterModules, err := makeModuleMap(b.exporters, b.exporterModules, moduleResolver)
	if err != nil {
		errs = append(errs, fmt.Errorf("exporters: %w", err))
	}
	processorModules, err := makeModuleMap(b.processors, b.processorModules, moduleResolver)
	if err != nil {
		errs = append(errs, fmt.Errorf("processors: %w", err))
	}
	connectorModules, err := makeModuleMap(b.connectors, b.connectorModules, moduleResolver)
	if err != nil {
		errs = append(errs, fmt.Errorf("connectors: %w", err))
	}

	return otelcol.Factories{
		Extensions:       extensions,
		ExtensionModules: extensionModules,
		Receivers:        receivers,
		ReceiverModules:  receiverModules,
		Processors:       processors,
		ProcessorModules: processorModules,
		Exporters:        exporters,
		ExporterModules:  exporterModules,
		Connectors:       connectors,
		ConnectorModules: connectorModules,
		Telemetry:        otelconftelemetry.NewFactory(),
	}, multierr.Combine(errs...)
}

type moduleMetadata struct {
	path     string
	fallback string
}

func appendRepeated(modules []moduleMetadata, modulePath string, count int) []moduleMetadata {
	for range count {
		modules = append(modules, moduleMetadata{path: modulePath})
	}
	return modules
}

type buildModuleResolver struct {
	modules           map[string]debug.Module
	moduleVersions    map[string]string
	allowTestFallback bool
}

func newBuildModuleResolver(moduleVersions map[string]string) (*buildModuleResolver, error) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("go build information is unavailable")
	}
	resolver := moduleResolverFromBuildInfo(buildInfo)
	resolver.moduleVersions = moduleVersions
	return resolver, nil
}

func moduleResolverFromBuildInfo(buildInfo *debug.BuildInfo) *buildModuleResolver {
	modules := make(map[string]debug.Module, len(buildInfo.Deps)+1)
	if buildInfo.Main.Path != "" {
		modules[buildInfo.Main.Path] = buildInfo.Main
	}
	for _, dependency := range buildInfo.Deps {
		modules[dependency.Path] = *dependency
	}
	return &buildModuleResolver{
		modules:           modules,
		allowTestFallback: strings.HasSuffix(buildInfo.Path, ".test") && len(buildInfo.Deps) == 0,
	}
}

func (r *buildModuleResolver) resolve(modulePath string) (string, error) {
	if modulePath == "" {
		return "", errors.New("module path is empty")
	}
	module, ok := r.modules[modulePath]
	if ok && module.Version != "" && module.Version != "(devel)" {
		return modulePath + " " + module.Version, nil
	}
	if version := r.moduleVersions[modulePath]; version != "" && version != "(devel)" {
		return modulePath + " " + version, nil
	}
	if !ok {
		return "", fmt.Errorf("module %q is not present in the final binary and has no configured version", modulePath)
	}
	return "", fmt.Errorf("module %q has no release version in the final binary", modulePath)
}

func (r *buildModuleResolver) resolveWithTestFallback(modulePath, fallback string) (string, error) {
	if modulePath == "" {
		return r.resolve(modulePath)
	}
	_, present := r.modules[modulePath]
	_, configured := r.moduleVersions[modulePath]
	if !r.allowTestFallback || present || configured {
		return r.resolve(modulePath)
	}
	if fallback == "" {
		return "", fmt.Errorf("module %q has no generated test fallback", modulePath)
	}
	return fallback, nil
}

type deprecatedAliasProvider interface {
	DeprecatedAlias() component.Type
}

func makeModuleMap[T component.Factory](
	factories []T,
	modules []moduleMetadata,
	resolver *buildModuleResolver,
) (map[component.Type]string, error) {
	if len(factories) != len(modules) {
		return nil, fmt.Errorf("factory and module counts differ: %d factories, %d modules", len(factories), len(modules))
	}

	result := make(map[component.Type]string, len(factories))
	var errs []error
	for i, factory := range factories {
		module, err := resolver.resolveWithTestFallback(modules[i].path, modules[i].fallback)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", factory.Type(), err))
		}
		result[factory.Type()] = module

		if aliasProvider, ok := any(factory).(deprecatedAliasProvider); ok {
			alias := aliasProvider.DeprecatedAlias()
			if alias.String() != "" {
				result[alias] = module
			}
		}
	}
	return result, multierr.Combine(errs...)
}
