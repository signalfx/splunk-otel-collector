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
//
// It is a deliberate subset of native btool. Only "list" is supported: layer,
// add, delete, check, check-schemes, validate, the *-debug commands, and the
// --user, --peername, --peer-apps, --slave-apps, --searchpool, --json,
// --debug-print, --debug-logfile and --log-warnings options are rejected.
//
// Known differences from native btool within list:
//   - Only system/{default,local} and apps/*/{default,local} are read. The
//     conf.conf path mappings are not consulted, so peer-app locations are
//     not supported, and the "learned" app is ordered like any other app
//     rather than at the lowest precedence of its tier.
//   - Stanza names are expanded ($VARIABLE) before layering and prefix
//     filtering, while native does both on the raw name and expands only for
//     display. A stanza spelled with a variable and one spelled literally
//     therefore merge by layer order here, whereas native lets the literal
//     stanza win.
//   - Files are decoded as UTF-8 or BOM-prefixed UTF-16 (UTF-32 is not
//     supported), and a stanza header cannot be continued across lines with a
//     trailing backslash. The set of whitespace characters trimmed also differs
//     slightly for exotic Unicode spaces.
//   - A configuration file that cannot be read for a reason other than not
//     existing fails the command; native logs a warning and skips it. Errors
//     are written to stderr rather than stdout.
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

const usage = "usage: otelcollauncher btool [--app=<app>] [--dir=<etc-dir>] [--expand-stanzas=<true|false>] <conf-file> list [stanza-prefix]"

type Manager struct{}

func New() *Manager { return &Manager{} }

func (m *Manager) Verbs() []cli.VerbInfo {
	return []cli.VerbInfo{{
		Name:  "btool",
		Help:  "inspect merged Splunk .conf settings",
		Usage: usage + "\n\nReads --dir, else $SPLUNK_ETC, else $SPLUNK_HOME/etc.",
	}}
}

func (m *Manager) Dispatch(_ context.Context, _ string, args []string, streams cli.IO) int {
	options, code, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(streams.Err, "%v\n%s\n", err, usage)
		return code
	}

	home := os.Getenv("SPLUNK_HOME")
	envEtcDir := os.Getenv("SPLUNK_ETC")
	etcDir := options.dir
	if etcDir == "" {
		etcDir = etcFromEnv(envEtcDir)
	}
	if etcDir == "" {
		if home == "" {
			fmt.Fprintln(streams.Err, "neither $SPLUNK_HOME nor $SPLUNK_ETC is defined; set one or use --dir=<etc-dir>")
			return cli.ExitFailure
		}
		etcDir = filepath.Join(home, "etc")
	}

	// --dir selects files without changing environment references in stanza names.
	expansionEtcDir := envEtcDir
	if expansionEtcDir == "" {
		if home != "" {
			expansionEtcDir = filepath.Join(home, "etc")
		} else {
			expansionEtcDir = etcDir
		}
	}
	if home == "" {
		home = filepath.Dir(etcDir)
	}

	dirs, err := configDirs(etcDir, options.app, options.confName)
	if err != nil {
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}
	vars := expansionVars{lookupEnv: os.LookupEnv, etcDir: expansionEtcDir, home: home}
	merged, err := readMerged(options.confName, dirs, vars, options.expand)
	if err != nil {
		fmt.Fprintf(streams.Err, "failed to read %s: %v\n", options.confName, err)
		return cli.ExitFailure
	}
	visible := resolve(merged)
	if err := printSettings(streams.Out, visible, options.prefix); err != nil {
		fmt.Fprintf(streams.Err, "failed to write settings: %v\n", err)
		return cli.ExitFailure
	}
	return cli.ExitOK
}

// etcFromEnv returns the SPLUNK_ETC value if it names an existing absolute
// directory, and "" otherwise so that the caller falls back to $SPLUNK_HOME/etc.
func etcFromEnv(value string) string {
	if value == "" || !filepath.IsAbs(value) {
		return ""
	}
	if info, err := os.Stat(value); err != nil || !info.IsDir() {
		return ""
	}
	return filepath.Clean(value)
}

type commandOptions struct {
	confName string
	prefix   string
	app      string
	dir      string
	expand   bool
}

// parseArgs returns the exit code to use alongside any error, matching the
// codes of splunk btool: 2 for a malformed command, 1 for anything else.
func parseArgs(args []string) (commandOptions, int, error) {
	options := commandOptions{expand: true}
	var positional []string
	for _, arg := range args {
		switch {
		case arg == "--no-log":
			// Accepted for compatibility; it has no effect on list output.
		case strings.HasPrefix(arg, "--expand-stanzas="):
			options.expand = parseBool(strings.TrimPrefix(arg, "--expand-stanzas="))
		case strings.HasPrefix(arg, "--app="):
			options.app = strings.TrimPrefix(arg, "--app=")
			if !validName(options.app) {
				return options, cli.ExitFailure, fmt.Errorf("invalid app name %q", options.app)
			}
		case strings.HasPrefix(arg, "--dir="):
			dir := strings.TrimPrefix(arg, "--dir=")
			abs, err := filepath.Abs(dir)
			if dir == "" || err != nil {
				return options, cli.ExitFailure, fmt.Errorf("invalid --dir %q", dir)
			}
			options.dir = abs
		case strings.HasPrefix(arg, "-"):
			return options, cli.ExitFailure, fmt.Errorf("unsupported option %q", arg)
		default:
			positional = append(positional, arg)
		}
	}
	// Native accepts a fourth positional (a key) that list ignores.
	if len(positional) < 2 || len(positional) > 4 || positional[1] != "list" {
		return options, cli.ExitUsage, errors.New("expected <conf-file> list [stanza-prefix]")
	}
	options.confName = positional[0] + ".conf"
	if !validName(options.confName) {
		return options, cli.ExitFailure, fmt.Errorf("invalid configuration file %q", positional[0])
	}
	if len(positional) >= 3 {
		options.prefix = positional[2]
	}
	return options, cli.ExitOK, nil
}

// parseBool accepts the spellings of native btool; empty or unrecognized
// values mean true.
func parseBool(value string) bool {
	switch strings.ToLower(value) {
	case "f", "false", "no", "0":
		return false
	}
	return true
}

func validName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, `/\\`)
}

// appDisabled reports whether the app's [install] state setting, merged from
// default/app.conf and local/app.conf, is present and anything but "enabled".
func appDisabled(appDir string) (bool, error) {
	state, found := "", false
	for _, layer := range []string{"default", "local"} {
		path := filepath.Join(appDir, layer, "app.conf")
		payload, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		settings, err := parseConf(payload)
		if err != nil {
			return false, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, s := range settings {
			if s.stanza == "install" && s.key == "state" && !s.header {
				state, found = s.value, true
			}
		}
	}
	return found && state != "enabled", nil
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
		if skipDisabled {
			disabled, err := appDisabled(appDir)
			if err != nil {
				return nil, err
			}
			if disabled {
				return nil, fmt.Errorf("application is disabled: %s", selectedApp)
			}
		}
		return []string{filepath.Join(appDir, "default"), filepath.Join(appDir, "local")}, nil
	}

	entries, err := os.ReadDir(appsDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read %s: %w", appsDir, err)
	}
	var apps []string
	for _, entry := range entries {
		// Like native btool, follow symlinks to app directories.
		if !entry.IsDir() && !isDirSymlink(filepath.Join(appsDir, entry.Name()), entry) {
			continue
		}
		if skipDisabled {
			disabled, err := appDisabled(filepath.Join(appsDir, entry.Name()))
			if err != nil {
				return nil, err
			}
			if disabled {
				continue
			}
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

func isDirSymlink(path string, entry os.DirEntry) bool {
	if entry.Type()&os.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

type stanzaSettings struct {
	values map[string]string
}

func readMerged(confName string, dirs []string, vars expansionVars, expand bool) (map[string]*stanzaSettings, error) {
	merged := make(map[string]*stanzaSettings)
	for _, dir := range dirs {
		path := filepath.Join(dir, confName)
		payload, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		// Unlike native btool, which warns and skips an unreadable layer, fail:
		// skipping would silently produce a plausible but wrong effective config.
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		parsed, err := parseConf(payload)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, s := range parsed {
			name := s.stanza
			if expand {
				name = expandStanzaName(name, vars)
			}
			settings := merged[name]
			if settings == nil {
				settings = &stanzaSettings{values: make(map[string]string)}
				merged[name] = settings
			}
			if !s.header {
				settings.values[s.key] = s.value
			}
		}
	}
	return merged, nil
}

// resolve applies [default] inheritance.
func resolve(merged map[string]*stanzaSettings) map[string]*stanzaSettings {
	defaults := merged[defaultStanza]
	visible := make(map[string]*stanzaSettings, len(merged))
	for name, settings := range merged {
		if name == defaultStanza {
			visible[name] = settings
			continue
		}
		resolved := &stanzaSettings{values: make(map[string]string, len(settings.values))}
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
		inherit(defaults)
	}
	return visible
}

func printSettings(out io.Writer, merged map[string]*stanzaSettings, prefix string) error {
	stanzas := make([]string, 0, len(merged))
	for name := range merged {
		if strings.HasPrefix(name, prefix) {
			stanzas = append(stanzas, name)
		}
	}
	sort.Strings(stanzas)
	for _, name := range stanzas {
		if _, err := fmt.Fprintf(out, "[%s]\n", name); err != nil {
			return err
		}
		settings := merged[name]
		keys := make([]string, 0, len(settings.values))
		for key := range settings.values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if _, err := fmt.Fprintf(out, "%s = %s\n", key, settings.values[key]); err != nil {
				return err
			}
		}
	}
	return nil
}
