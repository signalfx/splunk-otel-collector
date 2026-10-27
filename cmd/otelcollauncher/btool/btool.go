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

	"gopkg.in/ini.v1"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

const usage = "usage: otelcollauncher btool [--app=<app>] [--dir=<etc-dir>] <conf-file> list [stanza-prefix] [--debug]"

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
	options, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(streams.Err, "%v\n%s\n", err, usage)
		return cli.ExitUsage
	}

	etcDir := options.dir
	if etcDir == "" {
		home := os.Getenv("SPLUNK_HOME")
		if home == "" {
			fmt.Fprintln(streams.Err, "$SPLUNK_HOME is not defined; set it or use --dir=<etc-dir>")
			return cli.ExitFailure
		}
		etcDir = filepath.Join(home, "etc")
	}

	dirs, err := configDirs(etcDir, options.app)
	if err != nil {
		fmt.Fprintln(streams.Err, err)
		return cli.ExitFailure
	}
	merged, err := readMerged(options.confName, dirs)
	if err != nil {
		fmt.Fprintf(streams.Err, "failed to read %s: %v\n", options.confName, err)
		return cli.ExitFailure
	}
	if err := printSettings(streams.Out, merged, options.prefix, options.debug); err != nil {
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
}

func parseArgs(args []string) (commandOptions, error) {
	var options commandOptions
	var positional []string
	for _, arg := range args {
		switch {
		case arg == "--debug":
			options.debug = true
		case strings.HasPrefix(arg, "--app="):
			options.app = strings.TrimPrefix(arg, "--app=")
			if !validName(options.app) {
				return options, fmt.Errorf("invalid app name %q", options.app)
			}
		case strings.HasPrefix(arg, "--dir="):
			options.dir = strings.TrimPrefix(arg, "--dir=")
			if !filepath.IsAbs(options.dir) {
				return options, fmt.Errorf("--dir must be an absolute path: %q", options.dir)
			}
		case strings.HasPrefix(arg, "-"):
			return options, fmt.Errorf("unsupported option %q", arg)
		default:
			positional = append(positional, arg)
		}
	}
	if len(positional) < 2 || len(positional) > 3 || positional[1] != "list" {
		return options, errors.New("expected <conf-file> list [stanza-prefix]")
	}
	options.confName = positional[0]
	if !strings.HasSuffix(options.confName, ".conf") {
		options.confName += ".conf"
	}
	if !validName(options.confName) {
		return options, fmt.Errorf("invalid configuration file %q", positional[0])
	}
	if len(positional) == 3 {
		options.prefix = positional[2]
	}
	return options, nil
}

func validName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, `/\\`)
}

// configDirs returns paths in increasing precedence order. Global Splunk
// settings give alphabetically earlier apps priority over later apps.
func configDirs(etcDir, selectedApp string) ([]string, error) {
	appsDir := filepath.Join(etcDir, "apps")
	if selectedApp != "" {
		appDir := filepath.Join(appsDir, selectedApp)
		info, err := os.Stat(appDir)
		if errors.Is(err, os.ErrNotExist) || err == nil && !info.IsDir() {
			return nil, fmt.Errorf("application does not exist: %s", selectedApp)
		}
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", appDir, err)
		}
		return []string{filepath.Join(appDir, "default"), filepath.Join(appDir, "local")}, nil
	}

	entries, err := os.ReadDir(appsDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read %s: %w", appsDir, err)
	}
	var apps []string
	for _, entry := range entries {
		if entry.IsDir() {
			apps = append(apps, entry.Name())
		}
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

func readMerged(confName string, dirs []string) (map[string]*stanzaSettings, error) {
	merged := make(map[string]*stanzaSettings)
	defaults := make(map[string]valueSource)
	for _, dir := range dirs {
		path := filepath.Join(dir, confName)
		payload, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		parsed, err := ini.Load(payload)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, section := range parsed.Sections() {
			name := section.Name()
			if name == ini.DefaultSection {
				name = "default"
				if len(section.Keys()) == 0 {
					continue
				}
			}
			if confName == "inputs.conf" && strings.EqualFold(name, "default") {
				for _, key := range section.Keys() {
					defaults[key.Name()] = valueSource{value: key.Value(), path: path}
				}
				continue
			}
			settings := merged[name]
			if settings == nil {
				settings = &stanzaSettings{values: make(map[string]valueSource)}
				merged[name] = settings
			}
			settings.path = path
			for _, key := range section.Keys() {
				settings.values[key.Name()] = valueSource{value: key.Value(), path: path}
			}
		}
	}
	if confName == "inputs.conf" {
		for _, settings := range merged {
			for key, value := range defaults {
				if _, exists := settings.values[key]; !exists {
					settings.values[key] = value
				}
			}
		}
	}
	return merged, nil
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
	for _, name := range stanzas {
		settings := merged[name]
		if debug {
			if _, err := fmt.Fprintf(out, "%-*s [%s]\n", pathWidth, settings.path, name); err != nil {
				return err
			}
		} else if _, err := fmt.Fprintf(out, "[%s]\n", name); err != nil {
			return err
		}
		keys := make([]string, 0, len(settings.values))
		for key := range settings.values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			setting := settings.values[key]
			var err error
			if debug {
				_, err = fmt.Fprintf(out, "%-*s %s = %s\n", pathWidth, setting.path, key, setting.value)
			} else {
				_, err = fmt.Fprintf(out, "%s = %s\n", key, setting.value)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
