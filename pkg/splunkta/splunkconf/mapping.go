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

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
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
	for _, m := range p.mappers {
		typeStr := m.Type().String()
		instances, mapErr := m.MapInputs(tree)
		if mapErr != nil {
			return nil, nil, fmt.Errorf("splunkconf: %s: %w", typeStr, mapErr)
		}
		for name, cfg := range instances {
			if other, dup := claimedBy[name]; dup {
				return nil, nil, fmt.Errorf("splunkconf: stanza [%s] claimed by both %s and %s", name, other, typeStr)
			}
			claimedBy[name] = typeStr

			marshaled, mErr := marshalConfig(cfg)
			if mErr != nil {
				return nil, nil, fmt.Errorf("splunkconf: %s: stanza [%s]: %w", typeStr, name, mErr)
			}
			out = append(out, emitted{id: typeStr + "/" + stableName(name), cfg: marshaled})
		}
	}

	for _, in := range tree.Inputs {
		if _, claimed := claimedBy[in.Configuration.Stanza.Name]; !claimed {
			skipped = append(skipped, in.Configuration.Stanza.Name)
		}
	}
	return out, skipped, nil
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

// mapOutputs maps every output stanza to its exporter entry. A pipeline needs at
// least one exporter, so an outputs.conf with no output stanzas is an error
// (conf.ErrNoOutputStanzas), as is a stanza kind with no mapping.
func mapOutputs(merged conf.Map) ([]emitted, error) {
	groups, err := conf.OutputGroups(merged)
	if err != nil {
		return nil, err
	}
	var (
		out         []emitted
		unsupported []string
	)
	for _, g := range groups {
		name := g.Configuration.Stanza.Name
		parsed, perr := stanza.ParseOutputName(name)
		if perr != nil {
			return nil, fmt.Errorf("parse output stanza [%s]: %w", name, perr)
		}
		if parsed.Kind != "hecout" {
			unsupported = append(unsupported, fmt.Sprintf("[%s] (kind %q)", name, parsed.Kind))
			continue
		}
		// splunk_hec is the contrib exporter, emitted directly: its config keys
		// are what tabuilder.HECOutputConfig produces, so a wrapper type would
		// only add an identity mapping. [tcpout] -> S2S does need its own type,
		// and arrives with the output-mapper registration API.
		out = append(out, emitted{
			id:  splunkhecexporter.NewFactory().Type().String() + "/" + stableName(name),
			cfg: tabuilder.HECOutputConfig(&g),
		})
	}
	if len(unsupported) > 0 {
		return nil, fmt.Errorf("unsupported output stanza(s) %s; only [hecout] is mapped so far",
			strings.Join(unsupported, ", "))
	}
	return out, nil
}
