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

// component is one emitted confmap entry: a fully-qualified component ID
// (type/name) and its serializable config map.
type emitted struct {
	id  string
	cfg map[string]any
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

// stableName produces a collision-free, deterministic component name for a
// stanza. slug() alone is NOT injective (distinct targets can reduce to the
// same readable string), which would collapse two stanzas onto one
// component.ID and silently merge their checkpoints and receiver hashes. We
// append a short hash of the raw stanza name so the name stays readable but is
// unique per stanza and stable across reloads.
func stableName(prefix, raw string) string {
	sum := sha256.Sum256([]byte(raw))
	h := hex.EncodeToString(sum[:])[:8]
	s := slug(raw)
	if s == "" {
		return prefix + "-" + h
	}
	return prefix + "-" + s + "-" + h
}

// mapInputs maps enabled input stanzas to typed wrapper receiver configs. The
// full merged props/transforms set is attached to every receiver, matching how
// splunk_inputs feeds the same set to each sub-receiver. Per-source matching is
// done receiver-internally, so parity with splunk_inputs is preserved.
func mapInputs(prefix string, inputs []conf.Input, props []conf.Prop, transforms []conf.Transform) (out []emitted, skipped []string, err error) {
	for _, in := range inputs {
		st := in.Configuration.Stanza
		if st.IsDisabled() {
			continue
		}
		name, perr := stanza.ParseName(st.Name)
		if perr != nil {
			return nil, nil, fmt.Errorf("parse stanza %q: %w", st.Name, perr)
		}

		var component *emitted
		switch name.Kind {
		case "monitor":
			component = emitMonitor(prefix, st, name, props, transforms)
		case "tcp":
			component = emitTCP(prefix, st, name, props, transforms)
		case "udp":
			component = emitUDP(prefix, st, name, props, transforms)
		case "script", "":
			component = emitScript(prefix, st, name, props, transforms)
		case "batch":
			component = emitBatch(prefix, st, name, props, transforms)
		case "wineventlog":
			component = emitWineventlog(prefix, st, name, props, transforms)
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

func emitMonitor(prefix string, st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *emitted {
	id := splunkmonitor.TypeStr + "/" + stableName(prefix, st.Name)
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
	return &emitted{id: id, cfg: cfg}
}

func emitTCP(prefix string, st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *emitted {
	id := "splunk_tcp/" + stableName(prefix, st.Name)
	// Parse the target address:port from the stanza name (e.g., "0.0.0.0:5514")
	addr, port := parseListenAddress(name.Target)
	cfg := map[string]any{
		"listen_address": addr,
		"port":           port,
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &emitted{id: id, cfg: cfg}
}

func emitUDP(prefix string, st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *emitted {
	id := "splunk_udp/" + stableName(prefix, st.Name)
	// Parse the target address:port from the stanza name (e.g., "0.0.0.0:5515")
	addr, port := parseListenAddress(name.Target)
	cfg := map[string]any{
		"listen_address": addr,
		"port":           port,
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &emitted{id: id, cfg: cfg}
}

func emitScript(prefix string, st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *emitted {
	id := "splunk_script/" + stableName(prefix, st.Name)
	cfg := map[string]any{
		"script_filename": name.Target,
	}
	if interval := getParam(st, "interval", ""); interval != "" {
		cfg["interval"] = interval
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &emitted{id: id, cfg: cfg}
}

func emitBatch(prefix string, st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *emitted {
	id := "splunk_batch/" + stableName(prefix, st.Name)
	cfg := map[string]any{
		"file_path": name.Target,
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &emitted{id: id, cfg: cfg}
}

func emitWineventlog(prefix string, st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *emitted {
	id := "splunk_wineventlog/" + stableName(prefix, st.Name)
	cfg := map[string]any{
		"event_log_name": name.Target,
	}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &emitted{id: id, cfg: cfg}
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

// outputMapper builds one exporter entry from a single output stanza.
type outputMapper func(prefix string, out conf.Output) (emitted, error)

// outputMappers registers the supported outputs.conf stanza kinds, keyed by the
// stanza Kind from stanza.ParseOutputName (e.g. "hecout", later "tcpout"). No
// stanza is hardcoded as required: mapOutputs emits exactly the exporters the
// .conf declares, and a stanza whose kind is not registered here fails with a
// specific error rather than being silently dropped or forcing a fixed output.
var outputMappers = map[string]outputMapper{
	"hecout": mapHECOutput,
}

// mapOutputs maps every output stanza in the merged outputs.conf to its
// exporter entry. It requires no particular stanza; it emits one exporter per
// recognized stanza and fails if any stanza kind has no registered mapper. A
// pipeline needs at least one exporter, so an outputs.conf with no output
// stanzas is also an error (conf.ErrNoOutputStanzas).
func mapOutputs(prefix string, merged conf.Map) ([]emitted, error) {
	groups, err := conf.OutputGroups(merged)
	if err != nil {
		return nil, err // ErrNoOutputStanzas when the tree has no output stanzas
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
		m, ok := outputMappers[parsed.Kind]
		if !ok {
			unsupported = append(unsupported, fmt.Sprintf("[%s] (kind %q)", name, parsed.Kind))
			continue
		}
		e, merr := m(prefix, g)
		if merr != nil {
			return nil, merr
		}
		out = append(out, e)
	}
	if len(unsupported) > 0 {
		return nil, fmt.Errorf("unsupported output stanza(s) %s; no exporter registered (supported kinds: %s)",
			strings.Join(unsupported, ", "), knownOutputKinds())
	}
	return out, nil
}

// knownOutputKinds returns the registered output stanza kinds, sorted, for
// error messages.
func knownOutputKinds() string {
	kinds := make([]string, 0, len(outputMappers))
	for k := range outputMappers {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return strings.Join(kinds, ", ")
}

// mapHECOutput maps a [hecout] stanza to a splunk_hecout exporter entry.
func mapHECOutput(prefix string, out conf.Output) (emitted, error) {
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
	return emitted{id: splunkhecout.TypeStr + "/" + prefix, cfg: cfg}, nil
}
