// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkhome

import "github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"

// OutputMapperFactory maps one recognized outputs.conf stanza kind to the
// exporter Component the provider emits. It is the confmap-layer analogue of
// splunk_outputs' SubExporterFactory: the client registers one factory per
// output kind and the provider resolves against it by Scheme.
//
// Scheme returns the output stanza kind to match (the Kind from
// stanza.ParseOutputName, e.g. "tcpout"). Kinds are matched case-sensitively,
// matching Splunk UF behavior.
//
// MapOutput owns only the .conf -> exporter-config translation. Registering the
// exporter's component.Factory with the collector is the client's job, done in
// its own component set; the two must agree on the component TYPE that MapOutput
// puts in Component.ID (e.g. "splunk_s2sout"). Build the name segment with
// StableName so IDs stay stable and collision-free across the whole tree.
type OutputMapperFactory interface {
	Scheme() string
	MapOutput(conf.Output) (Component, error)
}

// Option configures a provider factory built by NewFactory. Options are applied
// once, at factory construction, to a per-factory registry; there is no global
// mutable state, so two factories in the same process can carry different mapper
// sets.
type Option func(*registry)

// WithOutputMapper registers f for the output stanza kind it reports from
// Scheme(). If another mapper is already registered for the same kind
// (including a built-in), it is replaced. This is the hook a client flavor uses
// to wire an additional UF-ported exporter, such as the S2S exporter from the
// private data-runtimes repo:
//
//	splunkhome.NewFactory(
//		splunkhome.WithOutputMapper(dataruntimes.S2SOutputMapper{}),
//	)
func WithOutputMapper(f OutputMapperFactory) Option {
	return func(r *registry) {
		if f == nil {
			return
		}
		r.outputs[f.Scheme()] = f
	}
}

// registry holds the output-stanza mappers a provider instance resolves against.
// It is seeded with the built-in mappers and then extended by Options.
type registry struct {
	outputs map[string]OutputMapperFactory
}

// newRegistry returns a registry preloaded with the built-in mappers.
func newRegistry() *registry {
	return &registry{
		outputs: map[string]OutputMapperFactory{
			"hecout": mapperFunc{scheme: "hecout", fn: mapHECOutput},
		},
	}
}

// mapperFunc adapts a plain mapping func to OutputMapperFactory. Built-in
// mappers stay simple funcs; only registration goes through the interface, so
// built-in and client-registered mappers share one resolution path.
type mapperFunc struct {
	scheme string
	fn     func(conf.Output) (Component, error)
}

func (m mapperFunc) Scheme() string                             { return m.scheme }
func (m mapperFunc) MapOutput(o conf.Output) (Component, error) { return m.fn(o) }
