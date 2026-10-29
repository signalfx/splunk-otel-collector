// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package btool implements the read-only btool list command for Splunk .conf files.
package btool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

const usage = "usage: otelcollauncher btool [--app=<app>] [--dir=<etc-dir>] [--expand-stanzas=<true|false>] <conf-file> list [stanza-prefix] [--debug]"

// builtinInputs are the inputs.conf stanza types that Splunk handles itself.
// Any other "<scheme>://" stanza is a modular input, and the bare "[<scheme>]"
// stanzas that configure them are not listed by btool.
var builtinInputs = map[string]bool{
	"SSL":           true,
	"batch":         true,
	"fifo":          true,
	"fschange":      true,
	"http":          true,
	"monitor":       true,
	"script":        true,
	"splunktcp":     true,
	"splunktcp-ssl": true,
	"tcp":           true,
	"tcp-ssl":       true,
	"udp":           true,
}

type Manager struct{}

func New() *Manager { return &Manager{} }

func (m *Manager) Verbs() []cli.VerbInfo {
	return []cli.VerbInfo{{
		Name:  "btool",
		Help:  "inspect merged Splunk .conf settings",
		Usage: usage + "\n\nReads $SPLUNK_HOME/etc unless --dir is supplied.",
	}}
}

func (m *Manager) Dispatch(_ context.Context, _ string, args []string, streams cli.IO) int {
	options, code, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(streams.Err, "%v\n%s\n", err, usage)
		return code
	}

	home := os.Getenv("SPLUNK_HOME")
	etcDir := options.dir
	if etcDir == "" {
		if home == "" {
			fmt.Fprintln(streams.Err, "$SPLUNK_HOME is not defined; set it or use --dir=<etc-dir>")
			return cli.ExitFailure
		}
		etcDir = filepath.Join(home, "etc")
	} else if home == "" {
		home = filepath.Dir(etcDir)
	}

	dirs, err := configDirs(etcDir, options.app, options.confName)
	if err != nil {
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}
	vars := expansionVars{lookupEnv: os.LookupEnv, etcDir: etcDir, home: home}
	merged, err := readMerged(options.confName, dirs, vars, options.expand)
	if err != nil {
		fmt.Fprintf(streams.Err, "failed to read %s: %v\n", options.confName, err)
		return cli.ExitFailure
	}
	visible := resolve(options.confName, merged)
	if err := printSettings(streams.Out, visible, options.prefix, options.debug); err != nil {
		fmt.Fprintf(streams.Err, "failed to write settings: %v\n", err)
		return cli.ExitFailure
	}
	return cli.ExitOK
}

type commandOptions struct {
	confName string
	prefix   string
	app      string
	dir      string
	debug    bool
	expand   bool
}

// parseArgs returns the exit code to use alongside any error, matching the
// codes of splunk btool: 2 for a malformed command, 1 for anything else.
func parseArgs(args []string) (commandOptions, int, error) {
	options := commandOptions{expand: true}
	var positional []string
	for _, arg := range args {
		switch {
		case arg == "--debug":
			options.debug = true
		case arg == "--raw", arg == "--no-log":
			// Accepted for compatibility; they have no effect on list output.
		case arg == "--expand-stanzas=true":
			options.expand = true
		case arg == "--expand-stanzas=false":
			options.expand = false
		case strings.HasPrefix(arg, "--app="):
			options.app = strings.TrimPrefix(arg, "--app=")
			if !validName(options.app) {
				return options, cli.ExitFailure, fmt.Errorf("invalid app name %q", options.app)
			}
		case strings.HasPrefix(arg, "--dir="):
			options.dir = strings.TrimPrefix(arg, "--dir=")
			if !filepath.IsAbs(options.dir) {
				return options, cli.ExitFailure, fmt.Errorf("--dir must be an absolute path: %q", options.dir)
			}
		case strings.HasPrefix(arg, "-"):
			return options, cli.ExitFailure, fmt.Errorf("unsupported option %q", arg)
		default:
			positional = append(positional, arg)
		}
	}
	if len(positional) < 2 || len(positional) > 3 || positional[1] != "list" {
		return options, cli.ExitUsage, errors.New("expected <conf-file> list [stanza-prefix]")
	}
	options.confName = positional[0] + ".conf"
	if !validName(options.confName) {
		return options, cli.ExitFailure, fmt.Errorf("invalid configuration file %q", positional[0])
	}
	if len(positional) == 3 {
		options.prefix = positional[2]
	}
	return options, cli.ExitOK, nil
}

func validName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, `/\\`)
}

// appDisabled reports whether the app's [install] state setting, merged from
// default/app.conf and local/app.conf, is present and anything but "enabled".
func appDisabled(appDir string) bool {
	state, found := "", false
	for _, layer := range []string{"default", "local"} {
		payload, err := os.ReadFile(filepath.Join(appDir, layer, "app.conf"))
		if err != nil {
			continue
		}
		for _, s := range parseConf(payload) {
			if s.stanza == "install" && s.key == "state" && !s.header {
				state, found = s.value, true
			}
		}
	}
	return found && state != "enabled"
}

// configDirs returns paths in increasing precedence order. Global Splunk
// settings give alphabetically earlier apps priority over later apps.
// Disabled apps are skipped, except for their app.conf which is how the
// disabled state itself is reported.
func configDirs(etcDir, selectedApp, confName string) ([]string, error) {
	appsDir := filepath.Join(etcDir, "apps")
	skipDisabled := confName != "app.conf"
	if selectedApp != "" {
		appDir := filepath.Join(appsDir, selectedApp)
		info, err := os.Stat(appDir)
		if errors.Is(err, os.ErrNotExist) || err == nil && !info.IsDir() {
			return nil, fmt.Errorf("application does not exist: %s", selectedApp)
		}
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", appDir, err)
		}
		if skipDisabled && appDisabled(appDir) {
			return nil, fmt.Errorf("application is disabled: %s", selectedApp)
		}
		return []string{filepath.Join(appDir, "default"), filepath.Join(appDir, "local")}, nil
	}

	entries, err := os.ReadDir(appsDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read %s: %w", appsDir, err)
	}
	var apps []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if skipDisabled && appDisabled(filepath.Join(appsDir, entry.Name())) {
			continue
		}
		apps = append(apps, entry.Name())
	}
	dirs := []string{filepath.Join(etcDir, "system", "default")}
	// os.ReadDir sorts by filename; reverse it because later layers win.
	sort.Sort(sort.Reverse(sort.StringSlice(apps)))
	for _, app := range apps {
		dirs = append(dirs, filepath.Join(appsDir, app, "default"))
	}
	for _, app := range apps {
		dirs = append(dirs, filepath.Join(appsDir, app, "local"))
	}
	dirs = append(dirs, filepath.Join(etcDir, "system", "local"))
	return dirs, nil
}

type valueSource struct {
	value string
	path  string
}

type stanzaSettings struct {
	values map[string]valueSource
	path   string
}

func readMerged(confName string, dirs []string, vars expansionVars, expand bool) (map[string]*stanzaSettings, error) {
	merged := make(map[string]*stanzaSettings)
	for _, dir := range dirs {
		path := filepath.Join(dir, confName)
		payload, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		for _, s := range parseConf(payload) {
			name := s.stanza
			if expand {
				name = expandStanzaName(name, vars)
			}
			if confName == "inputs.conf" {
				name = resolveRelativeInput(name, inputBaseDir(vars.etcDir, dir))
			}
			settings := merged[name]
			if settings == nil {
				settings = &stanzaSettings{values: make(map[string]valueSource)}
				merged[name] = settings
			}
			settings.path = path
			if !s.header {
				settings.values[s.key] = valueSource{value: s.value, path: path}
			}
		}
	}
	return merged, nil
}

// inputBaseDir is the directory that "<scheme>://./" input paths are relative
// to: the app owning the layer, or the layer directory itself outside apps.
func inputBaseDir(etcDir, layerDir string) string {
	appsDir := filepath.Join(etcDir, "apps") + string(filepath.Separator)
	if rel, ok := strings.CutPrefix(layerDir, appsDir); ok {
		app, _, _ := strings.Cut(rel, string(filepath.Separator))
		return filepath.Join(appsDir, app)
	}
	return layerDir
}

func resolveRelativeInput(name, baseDir string) string {
	scheme, rest, ok := strings.Cut(name, "://")
	if !ok {
		return name
	}
	if relative, ok := strings.CutPrefix(rest, "./"); ok {
		return scheme + "://" + baseDir + "/" + relative
	}
	return name
}

// resolve applies [default] inheritance and the inputs.conf specifics, and
// drops the stanzas that btool does not list.
func resolve(confName string, merged map[string]*stanzaSettings) map[string]*stanzaSettings {
	isInputs := confName == "inputs.conf"
	defaults := merged[defaultStanza]
	visible := make(map[string]*stanzaSettings, len(merged))
	for name, settings := range merged {
		if name == defaultStanza {
			if !isInputs {
				visible[name] = settings
			}
			continue
		}
		if isInputs && !strings.Contains(name, ":") && !builtinInputs[name] {
			continue
		}
		resolved := &stanzaSettings{path: settings.path, values: make(map[string]valueSource, len(settings.values))}
		for key, value := range settings.values {
			resolved.values[key] = value
		}
		visible[name] = resolved

		inherit := func(from *stanzaSettings, keys ...string) {
			if from == nil {
				return
			}
			if len(keys) == 0 {
				for key := range from.values {
					keys = append(keys, key)
				}
			}
			for _, key := range keys {
				if _, exists := resolved.values[key]; exists {
					continue
				}
				if value, ok := from.values[key]; ok {
					resolved.values[key] = value
				}
			}
		}
		if scheme, _, ok := strings.Cut(name, "://"); isInputs && ok && !builtinInputs[scheme] {
			// Modular inputs inherit from their bare scheme stanza, and from
			// [default] only host and index, which have no source file.
			inherit(merged[scheme])
			for _, key := range []string{"host", "index"} {
				if _, exists := resolved.values[key]; !exists && defaults != nil {
					if value, ok := defaults.values[key]; ok {
						resolved.values[key] = valueSource{value: value.value}
					}
				}
			}
			continue
		}
		if confName == "indexes.conf" && isVirtualIndexProvider(name) {
			for key := range resolved.values {
				if !strings.HasPrefix(key, "vix.") {
					delete(resolved.values, key)
				}
			}
			continue
		}
		inherit(defaults)
	}
	return visible
}

// isVirtualIndexProvider reports indexes.conf stanzas that configure a
// virtual index provider. btool lists only their vix.* settings and they do
// not inherit from [default].
func isVirtualIndexProvider(name string) bool {
	return strings.HasPrefix(name, "provider:") || strings.HasPrefix(name, "provider-family:")
}

func printSettings(out io.Writer, merged map[string]*stanzaSettings, prefix string, debug bool) error {
	stanzas := make([]string, 0, len(merged))
	for name := range merged {
		if strings.HasPrefix(name, prefix) {
			stanzas = append(stanzas, name)
		}
	}
	sort.Strings(stanzas)
	pathWidth := 0
	if debug {
		for _, name := range stanzas {
			settings := merged[name]
			pathWidth = max(pathWidth, len(settings.path))
			for _, setting := range settings.values {
				pathWidth = max(pathWidth, len(setting.path))
			}
		}
	}
	line := func(path, text string) error {
		var err error
		switch {
		case !debug || path == "":
			_, err = fmt.Fprintln(out, text)
		default:
			_, err = fmt.Fprintf(out, "%-*s %s\n", pathWidth, path, text)
		}
		return err
	}
	for _, name := range stanzas {
		settings := merged[name]
		if err := line(settings.path, "["+name+"]"); err != nil {
			return err
		}
		keys := make([]string, 0, len(settings.values))
		for key := range settings.values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			setting := settings.values[key]
			if err := line(setting.path, key+" = "+setting.value); err != nil {
				return err
			}
		}
	}
	return nil
}
