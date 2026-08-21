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

// Package components defines Registry and Registries, a mutable set of component factories
// that Build assembles into an otelcol.Factories value. It has no dependency beyond the
// collector core, so any Go program can import it to assemble its own set of components by
// calling NewRegistries, then Register or Deregister factories of its own before calling
// Build. Programs that want this distribution's default set of components instead should
// import pkg/components/defaults, which registers them via RegisterDefaults; that package
// pulls in this distribution's full dependency graph, including Splunk-internal packages
// resolved only through this module's local replace directives, so it cannot be imported
// from outside this module.
package components

import (
	"sort"
	"sync"

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

// Registry is a mutable, concurrency-safe set of component factories of a single kind,
// keyed by their canonical component.Type.
type Registry[F component.Factory] struct {
	mu        sync.RWMutex
	factories map[component.Type]F
}

func newRegistry[F component.Factory]() *Registry[F] {
	return &Registry[F]{
		factories: make(map[component.Type]F),
	}
}

// Register adds f to the registry, replacing any existing factory registered under the same
// component type.
func (r *Registry[F]) Register(f F) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[f.Type()] = f
}

// Deregister removes the factory registered under name, if any. It returns an error only if
// name is not a syntactically valid component type.
func (r *Registry[F]) Deregister(name string) error {
	t, err := component.NewType(name)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.factories, t)
	return nil
}

// Names returns the canonical type names currently registered, sorted alphabetically.
func (r *Registry[F]) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.factories))
	for t := range r.factories {
		names = append(names, t.String())
	}
	sort.Strings(names)
	return names
}

// Factories returns a snapshot of the currently registered factories.
func (r *Registry[F]) Factories() []F {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]F, 0, len(r.factories))
	for _, f := range r.factories {
		out = append(out, f)
	}
	return out
}

// Registries is a caller-owned set of Extensions, Receivers, Processors, Exporters, and
// Connectors registries that Build assembles into an otelcol.Factories. Each Registries
// value is independent: registering or deregistering a factory on one has no effect on any
// other, so multiple distributions can be assembled concurrently without interfering with
// each other.
type Registries struct {
	Extensions *Registry[extension.Factory]
	Receivers  *Registry[receiver.Factory]
	Processors *Registry[processor.Factory]
	Exporters  *Registry[exporter.Factory]
	Connectors *Registry[connector.Factory]
}

// NewRegistries returns an empty Registries. Call RegisterDefaults to populate it with the
// components that ship with this collector distribution.
func NewRegistries() *Registries {
	return &Registries{
		Extensions: newRegistry[extension.Factory](),
		Receivers:  newRegistry[receiver.Factory](),
		Processors: newRegistry[processor.Factory](),
		Exporters:  newRegistry[exporter.Factory](),
		Connectors: newRegistry[connector.Factory](),
	}
}

// Build assembles an otelcol.Factories from the current contents of r's registries.
func (r *Registries) Build() (otelcol.Factories, error) {
	var errs []error

	extensions, err := otelcol.MakeFactoryMap(r.Extensions.Factories()...)
	if err != nil {
		errs = append(errs, err)
	}
	receivers, err := otelcol.MakeFactoryMap(r.Receivers.Factories()...)
	if err != nil {
		errs = append(errs, err)
	}
	processors, err := otelcol.MakeFactoryMap(r.Processors.Factories()...)
	if err != nil {
		errs = append(errs, err)
	}
	exporters, err := otelcol.MakeFactoryMap(r.Exporters.Factories()...)
	if err != nil {
		errs = append(errs, err)
	}
	connectors, err := otelcol.MakeFactoryMap(r.Connectors.Factories()...)
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
