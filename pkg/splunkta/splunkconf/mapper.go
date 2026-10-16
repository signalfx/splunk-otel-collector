// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf

import (
	"go.opentelemetry.io/collector/component"
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
		p.mappers = append(p.mappers, mappers...)
	}
}
