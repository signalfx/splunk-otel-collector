// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkbatch"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkmonitor"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkscript"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunktcp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkudp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkwineventlog"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

// component is one emitted confmap entry: a fully-qualified component ID
// (type/name) and its serializable config map.
type component struct {
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

// mapInputs maps enabled input stanzas to wrapper receiver configs. The full
// merged props/transforms set is attached to every receiver, matching how
// splunk_inputs feeds the same set to each sub-receiver; per-source matching is
// done receiver-internally, so parity is preserved.
//
// Kinds are matched against the spellings inputs.conf.spec defines, exactly
// rather than case-folded, which is how splunk_outputs resolves its own stanza
// kinds. WinEventLog is the only input kind whose canonical spelling is not
// lowercase.
func mapInputs(inputs []conf.Input, props []conf.Prop, transforms []conf.Transform) (out []component, skipped []string, err error) {
	for _, in := range inputs {
		st := in.Configuration.Stanza
		if st.IsDisabled() {
			continue
		}
		name, perr := stanza.ParseName(st.Name)
		if perr != nil {
			return nil, nil, fmt.Errorf("parse stanza %q: %w", st.Name, perr)
		}

		var c *component
		switch name.Kind {
		case "monitor":
			c = emitMonitor(st, name, props, transforms)
		case "tcp":
			c = emitTCP(st, name, props, transforms)
		case "udp":
			c = emitUDP(st, name, props, transforms)
		case "script":
			c = emitScript(st, name, in.AppDir, props, transforms)
		case "batch":
			c = emitBatch(st, name, props, transforms)
		case "wineventlog", "WinEventLog":
			c = emitWineventlog(st, name, props, transforms)
		default:
			// Includes the unprefixed form ([my_script]), which resolves under
			// bin/<os>_<arch>/ rather than as a plain path. Routing it through
			// splunk_script would silently change where the executable is found,
			// so it is reported as unsupported instead.
			skipped = append(skipped, st.Name)
			continue
		}
		out = append(out, *c)
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

func emitMonitor(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *component {
	cfg := map[string]any{"path": name.Target}
	if p := st.Params.Get("CHARSET"); p != nil {
		cfg["charset"] = p.Value
	}
	if p := st.Params.Get("EVENT_BREAKER"); p != nil {
		cfg["event_breaker"] = p.Value
	}
	if p := st.Params.Get("TRUNCATE"); p != nil {
		if n, convErr := strconv.Atoi(p.Value); convErr == nil {
			cfg["truncate"] = n
		}
	}
	// whitelist and blacklist are deliberately not modeled: the monitor receiver
	// reads both straight off the stanza params, so they round-trip verbatim
	// through Extra instead of being re-encoded.
	addResourceAttrs(cfg, st, "CHARSET", "EVENT_BREAKER", "TRUNCATE")
	attachPropsTransforms(cfg, props, transforms)
	return &component{id: splunkmonitor.TypeStr + "/" + stableName(st.Name), cfg: cfg}
}

func emitTCP(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *component {
	addr, port := parseListenAddress(name.Target)
	cfg := map[string]any{"listen_address": addr, "port": port}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &component{id: splunktcp.TypeStr + "/" + stableName(st.Name), cfg: cfg}
}

func emitUDP(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *component {
	addr, port := parseListenAddress(name.Target)
	cfg := map[string]any{"listen_address": addr, "port": port}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &component{id: splunkudp.TypeStr + "/" + stableName(st.Name), cfg: cfg}
}

// emitScript carries appDir so a relative script target resolves against the app
// that declared the stanza. script.GetPath both resolves and sandboxes against
// that directory, so without it the executable would resolve against the
// collector's working directory.
func emitScript(st conf.Stanza, name stanza.Name, appDir string, props []conf.Prop, transforms []conf.Transform) *component {
	cfg := map[string]any{"script_filename": name.Target}
	if appDir != "" {
		cfg["app_dir"] = appDir
	}
	if p := st.Params.Get("interval"); p != nil {
		cfg["interval"] = p.Value
	}
	addResourceAttrs(cfg, st, "interval")
	attachPropsTransforms(cfg, props, transforms)
	return &component{id: splunkscript.TypeStr + "/" + stableName(st.Name), cfg: cfg}
}

func emitBatch(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *component {
	cfg := map[string]any{"file_path": name.Target}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &component{id: splunkbatch.TypeStr + "/" + stableName(st.Name), cfg: cfg}
}

func emitWineventlog(st conf.Stanza, name stanza.Name, props []conf.Prop, transforms []conf.Transform) *component {
	cfg := map[string]any{"event_log_name": name.Target}
	addResourceAttrs(cfg, st)
	attachPropsTransforms(cfg, props, transforms)
	return &component{id: splunkwineventlog.TypeStr + "/" + stableName(st.Name), cfg: cfg}
}

// addResourceAttrs copies the four resource attributes onto cfg under the
// wrapper's own key names, then every remaining stanza param under its original
// key so the wrapper's ",remain" Extra field can carry it through to conf.Input.
// Without the passthrough an unmodeled UF param would be dropped at emit, and
// strict confmap unmarshal would reject it outright. consumed names are the ones
// the caller already bound to a typed field; disabled gates emission, not the
// receiver, so it is always skipped.
func addResourceAttrs(cfg map[string]any, st conf.Stanza, consumed ...string) {
	skip := map[string]bool{"disabled": true}
	for _, k := range consumed {
		skip[k] = true
	}
	for _, attr := range []string{"index", "source", "sourcetype", "host"} {
		if p := st.Params.Get(attr); p != nil {
			cfg[attr] = p.Value
		}
		skip[attr] = true
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
	if addr, portStr, found := strings.Cut(target, ":"); found {
		if addr == "" {
			addr = "0.0.0.0"
		}
		return addr, parsePort(portStr)
	}
	return "0.0.0.0", parsePort(target)
}

func parsePort(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 || n > 65535 {
		return 0
	}
	return n
}

// mapOutputs maps every output stanza to its exporter entry. A pipeline needs at
// least one exporter, so an outputs.conf with no output stanzas is an error
// (conf.ErrNoOutputStanzas), as is a stanza kind with no mapping.
func mapOutputs(merged conf.Map) ([]component, error) {
	groups, err := conf.OutputGroups(merged)
	if err != nil {
		return nil, err
	}
	var (
		out         []component
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
		out = append(out, component{
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
