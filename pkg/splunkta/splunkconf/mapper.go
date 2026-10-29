// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
)

// Input is one enabled input stanza with its name already split into kind and
// target, and the props/transforms it resolves against. The provider parses
// every name once and reports a malformed one, so a mapper filters on Name.Kind
// instead of re-parsing and re-reporting.
//
// Props and Transforms are resolved for the app that declared this stanza: its
// own layers merged with system, and nothing from a sibling app, which is what
// splunkd gives it. They are per-input rather than one set on Conf so a mapper
// cannot attach a set that outranges the stanza it came from. Within that set no
// further filtering is done, because per-source matching happens inside the
// receiver.
type Input struct {
	conf.Input
	Name       stanza.Name
	Props      []conf.Prop
	Transforms []conf.Transform

	// SchemeDefault is the [<scheme>] stanza's params for a modular input, nil
	// for anything else. Those params are already folded into this input, but
	// the fold is lossy: afterwards an inherited setting looks the same as one
	// the instance declared. A receiver that has to tell them apart reads this,
	// which is what a use_single_instance scheme needs, since it schedules from
	// the scheme default's interval alone and ignores every instance's own.
	SchemeDefault conf.Params
}

// Conf is the resolved .conf tree a mapper builds its instances from.
type Conf struct {
	Inputs []Input
}

// InputMapper is implemented by a receiver that builds its own instances from a
// .conf tree, which is what makes the stanza-to-config translation the
// receiver's concern rather than this package's. A receiver living in another
// repo registers through the same interface as the built-in ones, so a
// closed-source input needs no hook here.
//
// MapInputs returns one config per instance the receiver owns, keyed by the raw
// stanza name it was built from, and ignores every stanza it does not own. The
// key is the raw name rather than a component ID because two invariants stay
// with the provider instead of being reimplemented per receiver: component IDs
// are a pure function of stanza identity, which partial reload's hash key
// requires, and a stanza no mapper claims is reported rather than silently
// dropped.
type InputMapper interface {
	MapInputs(Conf) (map[string]component.Config, error)
}

// MappingReceiverFactory is a receiver factory that maps its own .conf inputs.
// Registration takes the factory rather than a bare InputMapper so the type the
// instances are emitted under comes from the factory itself and cannot disagree
// with the one registered in the collector's component set.
type MappingReceiverFactory interface {
	receiver.Factory
	InputMapper
}

// Output is one output stanza with its name already split into kind and target.
// As with Input, the provider parses every name once, so a mapper filters on
// Name.Kind instead of re-parsing and re-reporting.
//
// It carries no props/transforms counterpart to Input's: those resolve per
// input, so an output stanza carries nothing beyond itself.
type Output struct {
	Name stanza.Name
	conf.Output
}

// OutputMapper is the output-side counterpart of InputMapper, so this package
// models no output kind itself. An exporter living in another repo registers
// through the same interface as the in-repo one: it supplies the factory and the
// mapping, while reading and merging outputs.conf stays here.
//
// MapOutputs returns one config per instance the exporter owns, keyed by the raw
// stanza name it was built from, and ignores every stanza it does not own. The
// key is the raw name for the same reasons it is on InputMapper.
type OutputMapper interface {
	MapOutputs([]Output) (map[string]component.Config, error)
}

// MappingExporterFactory is an exporter factory that maps its own .conf outputs.
// Registration takes the factory rather than a bare OutputMapper for the reason
// it does on MappingReceiverFactory.
type MappingExporterFactory interface {
	exporter.Factory
	OutputMapper
}

// ProcessorMapper is implemented by a processor that belongs in the emitted
// pipeline, for a receiver whose records are not ready to export as they are.
//
// Nothing in inputs.conf declares a processor, so unlike InputMapper it claims
// no stanza and an empty result is not an error. It is given the whole tree
// because props.conf and transforms.conf resolve per input, and keys by the
// instance's own name because there is no stanza to take one from.
type ProcessorMapper interface {
	MapProcessors(Conf) (map[string]component.Config, error)
}

// MappingProcessorFactory is a processor factory that maps itself into the
// emitted pipeline, taking the factory for the reason MappingReceiverFactory
// does.
type MappingProcessorFactory interface {
	processor.Factory
	ProcessorMapper
}

// MapProcessorsFunc is a processor's mapping as a plain function.
type MapProcessorsFunc func(Conf) (map[string]component.Config, error)

// NewMappingProcessor pairs a processor factory with its mapping.
func NewMappingProcessor(f processor.Factory, mapProcessors MapProcessorsFunc) MappingProcessorFactory {
	return mappingProcessor{Factory: f, mapProcessors: mapProcessors}
}

type mappingProcessor struct {
	processor.Factory
	mapProcessors MapProcessorsFunc
}

func (m mappingProcessor) MapProcessors(c Conf) (map[string]component.Config, error) {
	return m.mapProcessors(c)
}

// Attrs holds the settings every input stanza can carry, whatever its kind.
type Attrs struct {
	Extra      map[string]string
	Index      string
	Source     string
	Sourcetype string
	Host       string
}

// StanzaAttrs splits a stanza's params into the four resource attributes and
// everything left over, which a Config carries in its ",remain" Extra field so
// an unmodeled UF setting is neither dropped at emit nor rejected by strict
// unmarshal. consumed names are the ones the caller binds to its own typed
// fields. disabled is always left out: it gates emission, not the receiver.
func StanzaAttrs(st conf.Stanza, consumed ...string) Attrs {
	skip := map[string]bool{"disabled": true}
	for _, k := range consumed {
		skip[k] = true
	}

	var attrs Attrs
	for name, field := range map[string]*string{
		"index":      &attrs.Index,
		"source":     &attrs.Source,
		"sourcetype": &attrs.Sourcetype,
		"host":       &attrs.Host,
	} {
		if p := st.Params.Get(name); p != nil {
			*field = p.Value
		}
		skip[name] = true
	}

	for _, p := range st.Params {
		if skip[p.Name] {
			continue
		}
		if attrs.Extra == nil {
			attrs.Extra = map[string]string{}
		}
		attrs.Extra[p.Name] = p.Value
	}
	return attrs
}

// Option configures the provider.
type Option func(*provider)

// WithInputMappers registers the receivers that map input stanzas. A provider
// with none emits no receivers, so the built-in set is passed in by the caller
// that also registers those receivers with the collector, which keeps the two
// lists from drifting and gives an out-of-repo receiver the same path in.
func WithInputMappers(mappers ...MappingReceiverFactory) Option {
	return func(p *provider) {
		p.inputMappers = append(p.inputMappers, mappers...)
	}
}

// WithOutputMappers registers the exporters that map output stanzas, the
// output-side counterpart of WithInputMappers. A provider with none emits no
// exporters, so even the in-repo [hecout] mapping comes in this way, which is
// what gives an out-of-repo exporter the same path in.
func WithOutputMappers(mappers ...MappingExporterFactory) Option {
	return func(p *provider) {
		p.outputMappers = append(p.outputMappers, mappers...)
	}
}

// WithProcessorMappers registers the processors that place themselves in the
// emitted pipeline; a provider with none emits receivers and exporters alone.
//
// Registration order is pipeline order, since the .conf tree does not state it.
// Instances of one mapper are ordered by name, so a mapper needing a particular
// order between them should be registered as several.
func WithProcessorMappers(mappers ...MappingProcessorFactory) Option {
	return func(p *provider) {
		p.processorMappers = append(p.processorMappers, mappers...)
	}
}
