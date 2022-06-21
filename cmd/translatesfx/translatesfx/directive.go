// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translatesfx

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type source int

const (
	directiveSourceUnknown source = iota
	directiveSourceFile
	directiveSourceEnv
	directiveSourceZK
	directiveSourceEtcd2
	directiveSourceConsul
	directiveSourceVault
)

func parseDirective(m map[any]any, wd string) (out directive, isDirective bool, err error) {
	fromRaw, ok := m["#from"]
	if !ok {
		return
	}
	from, ok := fromRaw.(string)
	if !ok {
		return out, isDirective, fmt.Errorf("error parsing from %v", m)
	}

	isDirective = true

	out.fromType, out.fromPath, err = parseFrom(from)
	if err != nil {
		return
	}

	out.flatten, err = parseFlatten(m)
	if err != nil {
		return
	}

	out.defaultV, err = parseDefault(m)
	if err != nil {
		return
	}

	out.optional, err = parseOptional(m)
	if err != nil {
		return
	}

	out.wd = wd

	return
}

func parseFrom(from string) (source, string, error) {
	idx := strings.Index(from, ":")
	if idx == -1 {
		return directiveSourceFile, from, nil
	}
	s := strToSource(from[:idx])
	if s == directiveSourceUnknown {
		return s, "", fmt.Errorf("#from fromType type unknown: %s", from)
	}
	return s, from[idx+1:], nil
}

func strToSource(s string) source {
	switch s {
	case "", "file":
		return directiveSourceFile
	case "env":
		return directiveSourceEnv
	case "zookeeper", "zk":
		return directiveSourceZK
	case "etcd2":
		return directiveSourceEtcd2
	case "consul":
		return directiveSourceConsul
	case "vault":
		return directiveSourceVault
	}
	return directiveSourceUnknown
}

func parseFlatten(m map[any]any) (bool, error) {
	return parseField(m, "flatten")
}

func parseOptional(m map[any]any) (bool, error) {
	return parseField(m, "optional")
}

func parseField(m map[any]any, field string) (bool, error) {
	v, ok := m[field]
	if !ok {
		return false, nil
	}
	out, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("unable to parse field %q: %v", field, m)
	}
	return out, nil
}

func parseDefault(m map[any]any) (string, error) {
	rawDefault, ok := m["default"]
	if !ok {
		return "", nil
	}
	out, ok := rawDefault.(string)
	if !ok {
		return "", fmt.Errorf("unable to parse default value: %v", m)
	}
	return out, nil
}

type directive struct {
	fromPath string
	defaultV string
	wd       string
	fromType source
	flatten  bool
	optional bool
}

func (d directive) render(forceExpand bool, vaultPaths *[]string) (any, error) {
	switch d.fromType {
	case directiveSourceFile:
		return d.handleFileType(forceExpand)
	case directiveSourceEnv:
		return d.expandEnv()
	case directiveSourceZK:
		return d.expandZK()
	case directiveSourceEtcd2:
		return d.expandEtcd2()
	case directiveSourceVault:
		return d.expandVault(vaultPaths)
	case directiveSourceUnknown:
		return nil, fmt.Errorf("#from fromType type unknown: %v", d.fromType)
	default:
		return nil, fmt.Errorf("#from fromType type not supported by translatesfx at this time: %v", d.fromType)
	}
}

func (d directive) expandEnv() (any, error) {
	return fmt.Sprintf("${%s}", d.fromPath), nil
}

func directiveSource(from string) string {
	idx := strings.Index(from, ":")
	if idx == -1 {
		return ""
	}
	return from[:idx]
}

func (d directive) handleFileType(forceExpand bool) (any, error) {
	// configsource doesn't handle flatten, glob, or default values at this time, so
	// we inline the value if any of those are specified in the #from directive
	if forceExpand || d.flatten || hasGlob(d.fromPath) || d.defaultV != "" {
		return d.expandFiles()
	}
	// otherwise we replace the #from directive with a configsource one
	return fmt.Sprintf("${include:%s}", d.fromPath), nil
}

func hasGlob(path string) bool {
	return strings.ContainsAny(path, "*?[]")
}

func (d directive) expandFiles() (any, error) {
	path := resolvePath(d.fromPath, d.wd)
	filepaths, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}

	if len(filepaths) == 0 {
		if d.defaultV != "" {
			return d.defaultV, nil
		}

		if !d.optional {
			return nil, fmt.Errorf("#from files not found and directive not marked optional: %v", d)
		}
	}

	var items []any
	for _, p := range filepaths {
		unmarshaled, err := unmarshal(p)
		if err != nil {
			return nil, err
		}
		items = append(items, unmarshaled)
	}
	return merge(items)
}

func (d directive) expandZK() (any, error) {
	return fmt.Sprintf("${zookeeper:%s}", d.fromPath), nil
}

func (d directive) expandEtcd2() (any, error) {
	return fmt.Sprintf("${etcd2:%s}", d.fromPath), nil
}

func (d directive) expandVault(vaultPaths *[]string) (any, error) {
	path, keys := parseVaultPath(d.fromPath)
	idx, found := indexOf(*vaultPaths, path)
	if !found {
		idx = len(*vaultPaths)
		*vaultPaths = append(*vaultPaths, path)
	}
	return fmt.Sprintf("${vault/%d:%s}", idx, keys), nil
}

func indexOf(a []string, s string) (int, bool) {
	for i, v := range a {
		if v == s {
			return i, true
		}
	}
	return 0, false
}

var vaultRegexp = regexp.MustCompile(`(.*)\[(.*)]`)

func parseVaultPath(p string) (path string, keys string) {
	found := vaultRegexp.FindAllStringSubmatch(p, -1)
	return found[0][1], found[0][2]
}

func resolvePath(path, wd string) string {
	if path[:1] == string(os.PathSeparator) {
		return path
	}
	return filepath.Join(wd, path)
}

func unmarshal(path string) (any, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var replacement any
	err = yaml.UnmarshalStrict(bytes, &replacement)
	if err != nil {
		return nil, err
	}
	return replacement, nil
}

func merge(items []any) (any, error) {
	switch len(items) {
	case 0:
		return nil, nil
	case 1:
		return items[0], nil
	}
	switch items[0].(type) {
	case []any:
		return mergeSlices(items)
	case map[any]any:
		return mergeMaps(items)
	}
	return nil, fmt.Errorf("unable to merge: %v", items)
}

func mergeSlices(items []any) (any, error) {
	var out []any
	for _, item := range items {
		l, ok := item.([]any)
		if !ok {
			return nil, fmt.Errorf("mergeSlices: type coersion failed for item %v in items %v", item, items)
		}
		out = append(out, l...)
	}
	return out, nil
}

func mergeMaps(items []any) (any, error) {
	out := map[any]any{}
	for _, item := range items {
		m, ok := item.(map[any]any)
		if !ok {
			return nil, fmt.Errorf("mergeMaps: type coersion failed for item %v in items %v", item, items)
		}
		for k, v := range m {
			out[k] = v
		}
	}
	return out, nil
}
