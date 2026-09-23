// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkhome

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkhome/splunkhecout"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/splunkhome/splunkmonitor"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
)

// Component is one emitted confmap entry: a fully-qualified component ID
// (type/name) and its serializable config map. It is the value an OutputMapper
// returns, so it is part of the public registration API (see options.go).
type Component struct {
	ID  string
	Cfg map[string]any
}

// slug turns a stanza identity into a deterministic, charset-safe component
// name segment. It is a pure function of the input string, so the same stanza
// always yields the same component.ID across reloads. That stability is what
// lets the ID double as the storage/fishbucket checkpoint key and keeps the
// partial-reload receiver hash keyed consistently.
//
// The collector name regex is `^[^\pZ\pC\pS]+$` (allows `/` and `.`, rejects
// spaces, control, symbols). We are stricter on purpose: reduce to
// [a-z0-9] plus single `-` separators so the emitted YAML keys stay readable
// and never collide with the type/name `/` separator.
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

// StableName produces a collision-free, deterministic component name segment
// for a stanza. slug() alone is NOT injective (distinct targets can reduce to
// the same readable string), which would collapse two stanzas onto one
// component.ID and silently merge their checkpoints and receiver hashes. We
// append a short hash of the raw stanza name so the name stays readable but is
// unique per stanza and stable across reloads.
//
// It is exported so a client-registered OutputMapper builds its component ID
// (e.g. "splunk_s2sout/"+StableName(name)) with the same naming rule as the
// built-in mappers, keeping checkpoint/hash keys stable across the whole tree.
func StableName(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	h := hex.EncodeToString(sum[:])[:8]
	s := slug(raw)
	if s == "" {
		return h
	}
	return s + "-" + h
}

// mapInputs maps enabled input stanzas to typed wrapper receiver configs. The
// full merged props/transforms set is attached to every receiver, matching how
// splunk_inputs feeds the same set to each sub-receiver. Per-source matching is
// done receiver-internally, so parity with splunk_inputs is preserved.
func mapInputs(inputs []conf.Input, props []conf.Prop, transforms []conf.Transform) (out []Component, skipped []string, err error) {
	for _, in := range inputs {
		st := in.Configuration.Stanza
		if st.IsDisabled() {
			continue
		}
		name, perr := stanza.ParseName(st.Name)
		if perr != nil {
			return nil, nil, fmt.Errorf("parse stanza %q: %w", st.Name, perr)
		}

		var component *Component
		switch name.Kind {
		case "monitor":
			component = emitMonitor(st, name, props, transforms)
		case "tcp":
			component = emitTCP(st, name, props, transforms)
		case "udp":
			component = emitUDP(st, name, props, transforms)
		case "script", "":
			component = emitScript(st, name, props, transforms)
		case "batch":
			component = emitBatch(st, name, props, transforms)
		case "wineventlog":
			component = emitWineventlog(st, name, props, transforms)
		default:
			skipped = append(skipped, st.Name)
			continue
		}

		if component != nil {
			out = append(out, *component)
		}
	}
	return out, skipped, nil
}

// attachPropsTransforms attaches the shared props/transforms set to a receiver
// cfg. Every receiver gets the full merged set, matching splunk_inputs.
func attachPropsTransforms(cfg map[string]any, props []conf.Prop, transforms []conf.Transform) {
	if len(props) > 0 {
		cfg["props"] = props
	}
	if len(transforms) > 0 {
		cfg["transforms"] = transforms
	}
}

func emitMonitor(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *Component {
	id := splunkmonitor.TypeStr + "/" + StableName(st.Name)
	cfg := map[string]any{
		"include": []any{name.Target},
	}
	if p := st.Params.Get("blacklist"); p != nil {
		cfg["exclude"] = []any{p.Value}
	}
	for confKey, mapKey := range map[string]string{
		"index":         "index",
		"source":        "source",
		"sourcetype":    "sourcetype",
		"host":          "host",
		"CHARSET":       "charset",
		"EVENT_BREAKER": "event_breaker",
	} {
		if p := st.Params.Get(confKey); p != nil {
			cfg[mapKey] = p.Value
		}
	}
	if p := st.Params.Get("TRUNCATE"); p != nil {
		if n, convErr := strconv.Atoi(p.Value); convErr == nil {
			cfg["truncate"] = n
		}
	}
	addExtras(cfg, st, "index", "source", "sourcetype", "host",
		"CHARSET", "EVENT_BREAKER", "TRUNCATE", "blacklist")
	attachPropsTransforms(cfg, props, transforms)
	return &Component{ID: id, Cfg: cfg}
}

func emitTCP(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *Component {
	id := "splunk_tcp/" + StableName(st.Name)
	// Parse the target address:port from the stanza name (e.g., "0.0.0.0:5514")
	addr, port := parseListenAddress(name.Target)
	cfg := map[string]any{
		"listen_address": addr,
		"port":           port,
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &Component{ID: id, Cfg: cfg}
}

func emitUDP(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *Component {
	id := "splunk_udp/" + StableName(st.Name)
	// Parse the target address:port from the stanza name (e.g., "0.0.0.0:5515")
	addr, port := parseListenAddress(name.Target)
	cfg := map[string]any{
		"listen_address": addr,
		"port":           port,
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &Component{ID: id, Cfg: cfg}
}

func emitScript(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *Component {
	id := "splunk_script/" + StableName(st.Name)
	cfg := map[string]any{
		"script_filename": name.Target,
	}
	if interval := getParam(st, "interval", ""); interval != "" {
		cfg["interval"] = interval
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &Component{ID: id, Cfg: cfg}
}

func emitBatch(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *Component {
	id := "splunk_batch/" + StableName(st.Name)
	cfg := map[string]any{
		"file_path": name.Target,
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &Component{ID: id, Cfg: cfg}
}

func emitWineventlog(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *Component {
	id := "splunk_wineventlog/" + StableName(st.Name)
	cfg := map[string]any{
		"event_log_name": name.Target,
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &Component{ID: id, Cfg: cfg}
}

func getParam(st conf.Stanza, name, defaultVal string) string {
	if p := st.Params.Get(name); p != nil {
		return p.Value
	}
	return defaultVal
}

func parsePort(s string) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err == nil && n > 0 && n <= 65535 {
		return n
	}
	return 0
}

func addResourceAttrs(cfg map[string]any, st conf.Stanza) {
	for confKey, mapKey := range map[string]string{
		"index":      "index",
		"source":     "source",
		"sourcetype": "sourcetype",
		"host":       "host",
	} {
		if p := st.Params.Get(confKey); p != nil {
			cfg[mapKey] = p.Value
		}
	}
	addExtras(cfg, st, "index", "source", "sourcetype", "host")
}

// addExtras copies every stanza param not already consumed into cfg under its
// original key. Modeled params (passed in consumed) bind to typed wrapper
// fields; everything else must still round-trip so the wrapper's ",remain"
// Extra field can carry it into conf.Input. Without this, an unmodeled UF param
// is dropped at emit and, worse, strict confmap unmarshal would reject it
// outright. "disabled" is always skipped (it gates emission, not the receiver).
func addExtras(cfg map[string]any, st conf.Stanza, consumed ...string) {
	skip := map[string]bool{"disabled": true}
	for _, k := range consumed {
		skip[k] = true
	}
	for _, p := range st.Params {
		if skip[p.Name] {
			continue
		}
		if _, exists := cfg[p.Name]; exists {
			continue
		}
		cfg[p.Name] = p.Value
	}
}

func parseListenAddress(target string) (string, int) {
	// Parse "address:port" or just "port"
	if strings.Contains(target, ":") {
		parts := strings.SplitN(target, ":", 2)
		addr := parts[0]
		if addr == "" {
			addr = "0.0.0.0"
		}
		port := parsePort(parts[1])
		return addr, port
	}
	// Just a port number
	return "0.0.0.0", parsePort(target)
}

// mapOutputs maps every output stanza in the merged outputs.conf to its
// exporter entry, using the mapper registered for each stanza kind. It requires
// no particular stanza; it emits one exporter per recognized stanza and fails
// if any stanza kind has no registered mapper. A pipeline needs at least one
// exporter, so an outputs.conf with no output stanzas is also an error
// (conf.ErrNoOutputStanzas). mappers is the per-provider registry
// (built-ins + client-registered), so a client-supplied kind like "tcpout"
// resolves here exactly like the built-in "hecout".
func mapOutputs(mappers map[string]OutputMapperFactory, merged conf.Map) ([]Component, error) {
	groups, err := conf.OutputGroups(merged)
	if err != nil {
		return nil, err // ErrNoOutputStanzas when the tree has no output stanzas
	}
	var (
		out         []Component
		unsupported []string
	)
	for _, g := range groups {
		name := g.Configuration.Stanza.Name
		parsed, perr := stanza.ParseOutputName(name)
		if perr != nil {
			return nil, fmt.Errorf("parse output stanza [%s]: %w", name, perr)
		}
		m, ok := mappers[parsed.Kind]
		if !ok {
			unsupported = append(unsupported, fmt.Sprintf("[%s] (kind %q)", name, parsed.Kind))
			continue
		}
		e, merr := m.MapOutput(g)
		if merr != nil {
			return nil, merr
		}
		out = append(out, e)
	}
	if len(unsupported) > 0 {
		return nil, fmt.Errorf("unsupported output stanza(s) %s; no exporter registered (supported kinds: %s)",
			strings.Join(unsupported, ", "), knownOutputKinds(mappers))
	}
	return out, nil
}

// knownOutputKinds returns the registered output stanza kinds, sorted, for
// error messages.
func knownOutputKinds(mappers map[string]OutputMapperFactory) string {
	kinds := make([]string, 0, len(mappers))
	for k := range mappers {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return strings.Join(kinds, ", ")
}

// mapHECOutput maps a [hecout] stanza to a splunk_hecout exporter entry.
func mapHECOutput(out conf.Output) (Component, error) {
	get := func(k string) string {
		if p := out.Configuration.Stanza.Params.Get(k); p != nil {
			return p.Value
		}
		return ""
	}
	cfg := map[string]any{
		"endpoint": get("uri"),
		"token":    get("httpEventCollectorToken"),
		"tls": map[string]any{
			"insecure_skip_verify": true,
		},
	}
	return Component{ID: splunkhecout.TypeStr + "/" + StableName(out.Configuration.Stanza.Name), Cfg: cfg}, nil
}
