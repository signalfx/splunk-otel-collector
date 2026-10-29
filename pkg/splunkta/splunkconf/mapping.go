// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

// emitted is one confmap entry: a fully-qualified component ID (type/name) and
// its serializable config map.
type emitted struct {
	cfg map[string]any
	id  string
}

// slug reduces a stanza identity to a readable, charset-safe component name
// segment. The collector's name regex allows more than this, but we reduce to
// [a-z0-9] plus single dashes so emitted YAML keys stay readable and never
// collide with the type/name separator.
func slug(raw string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(raw) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// stableName produces a collision-free component name segment for a stanza.
// slug alone is not injective: distinct targets can reduce to the same readable
// string, which would collapse two stanzas onto one component ID and silently
// merge their checkpoints and receiver hashes. A short hash of the raw stanza
// name keeps the name readable but unique, and stable across reloads, which the
// partial-reload hash key requires.
func stableName(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	h := hex.EncodeToString(sum[:])[:8]
	s := slug(raw)
	if s == "" {
		return h
	}
	return s + "-" + h
}

// newConf builds the tree handed to the mappers. It parses every enabled
// stanza's name once, so a mapper filters on an already-split kind and a
// malformed name is reported here instead of by each mapper in turn, and it
// resolves props.conf and transforms.conf per declaring app.
//
// Resolving them once across the whole configuration root instead would hand
// every receiver every app's props, which is not what a stanza resolves to:
// splunkd gives an app's inputs its own layers merged with system and nothing
// from a sibling app.
func newConf(confRoot string, inputs []conf.Input) (Conf, error) {
	byApp := map[string][]conf.Input{}
	for _, in := range inputs {
		if in.Configuration.Stanza.IsDisabled() {
			continue
		}
		byApp[in.AppDir] = append(byApp[in.AppDir], in)
	}

	// Fixed order so the emitted set, and the error naming an unmappable
	// stanza, do not depend on map iteration.
	appDirs := make([]string, 0, len(byApp))
	for appDir := range byApp {
		appDirs = append(appDirs, appDir)
	}
	sort.Strings(appDirs)

	systemDir := filepath.Join(confRoot, "system")
	var tree Conf
	for _, appDir := range appDirs {
		dirs := tabuilder.AppConfRootDirs(confRoot, appDir)
		if appDir == systemDir {
			dirs = tabuilder.SystemConfRootDirs(confRoot)
		}
		props, err := tabuilder.ReadProps(dirs)
		if err != nil {
			return Conf{}, fmt.Errorf("splunkconf: read props.conf for %s: %w", appDir, err)
		}
		transforms, err := tabuilder.ReadTransforms(dirs)
		if err != nil {
			return Conf{}, fmt.Errorf("splunkconf: read transforms.conf for %s: %w", appDir, err)
		}

		for _, in := range byApp[appDir] {
			name := in.Configuration.Stanza.Name
			parsed, err := stanza.ParseName(name)
			if err != nil {
				return Conf{}, fmt.Errorf("splunkconf: parse stanza %q: %w", name, err)
			}
			tree.Inputs = append(tree.Inputs, Input{
				Input:      in,
				Name:       parsed,
				Props:      props,
				Transforms: transforms,
			})
		}
	}
	return tree, nil
}

// mapInputs asks every registered mapper for the instances it owns. Each mapper
// sees the whole tree and claims by stanza name, so this keeps the two things a
// mapper must not decide: the component ID, which has to stay a pure function
// of stanza identity, and ownership, which has to be exactly one mapper per
// stanza.
func (p *provider) mapInputs(tree Conf) (out []emitted, skipped []string, err error) {
	claimedBy := map[string]string{}
	for _, m := range p.inputMappers {
		typeStr := m.Type().String()
		instances, mapErr := m.MapInputs(tree)
		if mapErr != nil {
			return nil, nil, fmt.Errorf("splunkconf: %s: %w", typeStr, mapErr)
		}
		claimed, claimErr := claim(claimedBy, typeStr, instances)
		if claimErr != nil {
			return nil, nil, fmt.Errorf("splunkconf: %w", claimErr)
		}
		out = append(out, claimed...)
	}

	for _, in := range tree.Inputs {
		if _, claimed := claimedBy[in.Configuration.Stanza.Name]; !claimed {
			skipped = append(skipped, in.Configuration.Stanza.Name)
		}
	}
	return out, skipped, nil
}

// mapProcessors asks every registered mapper for its instances, in registration
// order, which is the order they take in the pipeline. There is no claiming: a
// processor answers to no stanza, so nothing can collide and nothing is skipped.
func (p *provider) mapProcessors(tree Conf) ([]emitted, error) {
	var out []emitted
	for _, m := range p.processorMappers {
		typeStr := m.Type().String()
		instances, err := m.MapProcessors(tree)
		if err != nil {
			return nil, fmt.Errorf("splunkconf: %s: %w", typeStr, err)
		}
		for _, name := range sortedNames(instances) {
			marshaled, mErr := marshalConfig(instances[name])
			if mErr != nil {
				return nil, fmt.Errorf("splunkconf: %s: processor %q: %w", typeStr, name, mErr)
			}
			id := typeStr
			if name != "" {
				id = typeStr + "/" + name
			}
			out = append(out, emitted{id: id, cfg: marshaled})
		}
	}
	return out, nil
}

// sortedNames keeps the emitted pipeline independent of map iteration.
func sortedNames(m map[string]component.Config) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// claim records the instances one mapper returned and assigns each the component
// ID the provider owns, rejecting a stanza something else already claimed.
func claim(claimedBy map[string]string, typeStr string, instances map[string]component.Config) ([]emitted, error) {
	out := make([]emitted, 0, len(instances))
	for name, cfg := range instances {
		if other, dup := claimedBy[name]; dup {
			return nil, fmt.Errorf("stanza [%s] claimed by both %s and %s", name, other, typeStr)
		}
		claimedBy[name] = typeStr

		marshaled, err := marshalConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("%s: stanza [%s]: %w", typeStr, name, err)
		}
		out = append(out, emitted{id: typeStr + "/" + stableName(name), cfg: marshaled})
	}
	return out, nil
}

// marshalConfig turns a mapper's typed Config into the serialized form a
// confmap fragment carries, which the resolver then unmarshals back into that
// same type. The round trip is what lets a mapper work in its own Config
// instead of an untyped map it could misspell a key in unnoticed.
func marshalConfig(cfg component.Config) (map[string]any, error) {
	cm := confmap.New()
	if err := cm.Marshal(cfg); err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	return cm.ToStringMap(), nil
}

// parseOutputs builds the output stanzas handed to the mappers, parsing every
// name once for the reason newConf does it for inputs.
func parseOutputs(merged conf.Map) ([]Output, error) {
	groups, err := conf.OutputGroups(merged)
	if err != nil {
		return nil, err
	}
	outputs := make([]Output, 0, len(groups))
	for _, g := range groups {
		name := g.Configuration.Stanza.Name
		parsed, perr := stanza.ParseOutputName(name)
		if perr != nil {
			return nil, fmt.Errorf("parse output stanza [%s]: %w", name, perr)
		}
		outputs = append(outputs, Output{Output: g, Name: parsed})
	}
	return outputs, nil
}

// mapOutputs asks every registered mapper for the instances it owns, claiming by
// stanza name the way mapInputs does. An outputs.conf with no stanzas at all is
// conf.ErrNoOutputStanzas.
func (p *provider) mapOutputs(merged conf.Map) (out []emitted, skipped []string, err error) {
	outputs, err := parseOutputs(merged)
	if err != nil {
		return nil, nil, err
	}

	claimedBy := map[string]string{}
	for _, m := range p.outputMappers {
		typeStr := m.Type().String()
		instances, mapErr := m.MapOutputs(outputs)
		if mapErr != nil {
			return nil, nil, fmt.Errorf("%s: %w", typeStr, mapErr)
		}
		claimed, claimErr := claim(claimedBy, typeStr, instances)
		if claimErr != nil {
			return nil, nil, claimErr
		}
		out = append(out, claimed...)
	}

	for _, o := range outputs {
		if _, claimed := claimedBy[o.Configuration.Stanza.Name]; !claimed {
			skipped = append(skipped, o.Configuration.Stanza.Name)
		}
	}
	return out, skipped, nil
}
