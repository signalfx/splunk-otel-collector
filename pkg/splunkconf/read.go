// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkconf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf/conf"
)

// ConfRootDirs returns the Splunk btool conf search path for a configuration
// root: the directory holding system/ and apps/*/. Callers that know an install
// layout pass $SPLUNK_HOME/etc (or $SPLUNK_ETC when the tree is relocated);
// ConfRootDirs itself makes no assumption about where the root lives.
func ConfRootDirs(confRoot string) []string {
	apps, _ := discoverApps(confRoot)

	dirs := []string{filepath.Join(confRoot, "system", "default")}
	for _, app := range apps {
		dirs = append(dirs, filepath.Join(app, "default"))
	}
	for _, app := range apps {
		dirs = append(dirs, filepath.Join(app, "local"))
	}
	return append(dirs, filepath.Join(confRoot, "system", "local"))
}

// AppConfRootDirs returns the conf search path a single app resolves against: the
// app's own layers merged with system, and no other app's. appDir is absolute,
// as conf.Input.AppDir reports it. Pass the configuration root rather than an
// install, as ConfRootDirs does.
func AppConfRootDirs(confRoot, appDir string) []string {
	return []string{
		filepath.Join(confRoot, "system", "default"),
		filepath.Join(appDir, "default"),
		filepath.Join(appDir, "local"),
		filepath.Join(confRoot, "system", "local"),
	}
}

// SystemConfRootDirs returns the conf search path a system-level stanza resolves
// against, which is the system layers alone.
func SystemConfRootDirs(confRoot string) []string {
	return []string{
		filepath.Join(confRoot, "system", "default"),
		filepath.Join(confRoot, "system", "local"),
	}
}

// ReadInputs discovers and merges inputs.conf files from the given search
// directories. Returns nil (no error) when no inputs.conf files are found.
func ReadInputs(dirs []string) ([]conf.Input, map[string]conf.Input, error) {
	var layers [][]conf.Input
	for _, path := range confFilePaths(dirs, "inputs.conf") {
		b, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		appDir := filepath.Dir(filepath.Dir(path)) // strip /default or /local
		inputs, err := conf.ReadInput(b, appDir)
		if err != nil {
			return nil, nil, fmt.Errorf("parse %s: %w", path, err)
		}
		layers = append(layers, inputs)
	}
	// Globals are folded after the layers merge, not per file: stanza
	// specificity outranks layer precedence, so each stanza has to be resolved
	// across layers before a global can fill in what it leaves unset.
	out, schemeDefaults := conf.ApplyGlobals(conf.MergeInputs(layers))
	return out, schemeDefaults, nil
}

// ReadProps discovers and merges props.conf files from the given search
// directories. Returns nil (no error) when absent.
func ReadProps(dirs []string) ([]conf.Prop, error) {
	payloads, err := readConfFiles(confFilePaths(dirs, "props.conf"))
	if err != nil {
		return nil, err
	}
	var layers [][]conf.Prop
	for _, b := range payloads {
		props, err := conf.ReadProps(b)
		if err != nil {
			return nil, err
		}
		layers = append(layers, props)
	}
	return conf.MergeProps(layers), nil
}

// ReadTransforms discovers and merges transforms.conf files from the given
// search directories. Returns nil (no error) when absent.
func ReadTransforms(dirs []string) ([]conf.Transform, error) {
	payloads, err := readConfFiles(confFilePaths(dirs, "transforms.conf"))
	if err != nil {
		return nil, err
	}
	var layers [][]conf.Transform
	for _, b := range payloads {
		transforms, err := conf.ReadTransforms(b)
		if err != nil {
			return nil, err
		}
		layers = append(layers, transforms)
	}
	return conf.MergeTransforms(layers), nil
}

// ReadOutputs merges outputs.conf across the given search directories. Use
// conf.OutputGroups or conf.HECOut to extract a specific stanza.
func ReadOutputs(dirs []string) (conf.Map, error) {
	payloads, err := readConfFiles(confFilePaths(dirs, "outputs.conf"))
	if err != nil {
		return nil, err
	}
	return conf.ParseAndMergeConf(payloads)
}

// discoverApps returns the app directories that app.conf leaves in the bundle,
// which is what splunkd reads config from.
func discoverApps(confRoot string) ([]string, error) {
	dirs, err := appDirs(confRoot)
	if err != nil {
		return nil, err
	}
	var enabled []string
	for _, appDir := range dirs {
		if appEnabled(appDir) {
			enabled = append(enabled, appDir)
		}
	}
	return enabled, nil
}

// appDirs returns every app directory under confRoot/apps, enabled or not.
// splunkd admits every app in the bundle regardless of name, so this applies no
// name pattern; it skips only dotfile directories.
func appDirs(confRoot string) ([]string, error) {
	appsDir := filepath.Join(confRoot, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("splunkconf: failed to scan %s: %w", appsDir, err)
	}
	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		dirs = append(dirs, filepath.Join(appsDir, entry.Name()))
	}
	return dirs, nil
}

// appEnabled reports whether app.conf leaves the app in the bundle. splunkd
// allowlists the exact literal "enabled" rather than looking for "disabled", so
// any other value, "ENABLED" included, excludes the app; an absent key or file
// leaves it enabled.
//
// Reading fails open: splunkd resolves the value through a conf search that
// yields nothing for an unreadable or malformed app.conf, so one bad file
// excludes no app and does not fail discovery.
func appEnabled(appDir string) bool {
	dirs := []string{filepath.Join(appDir, "default"), filepath.Join(appDir, "local")}
	payloads, err := readConfFiles(confFilePaths(dirs, "app.conf"))
	if err != nil {
		return true
	}
	merged, err := conf.ParseAndMergeConf(payloads)
	if err != nil {
		return true
	}
	state, set := merged["install"]["state"]
	return !set || state == "enabled"
}

func confFilePaths(dirs []string, filename string) []string {
	paths := make([]string, len(dirs))
	for i, dir := range dirs {
		paths[i] = filepath.Join(dir, filename)
	}
	return paths
}

func readConfFiles(paths []string) ([][]byte, error) {
	var payloads [][]byte
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, b)
	}
	return payloads, nil
}
