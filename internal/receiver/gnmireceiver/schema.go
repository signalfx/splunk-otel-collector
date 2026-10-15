// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gnmireceiver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

// maxLeafrefDepth bounds leafref-chasing so a schema cycle cannot hang startup.
const maxLeafrefDepth = 10

// valueKind is the wire-value shape a YANG leaf resolves to. It confines
// goyang's yang.TypeKind to this file so the rest of the package stays free
// of the goyang dependency.
type valueKind int

const (
	valueKindUnknown valueKind = iota
	valueKindInt
	valueKindFloat
	valueKindString
)

// resolvedMetric is what a single YANG leaf tells us about how its values
// should become a metric.
type resolvedMetric struct {
	MetricConfig
	kind valueKind
}

// yangSchema is an immutable, precomputed gNMI-path -> resolvedMetric index
// built once at startup from a set of YANG modules. The zero value has no
// entries; a nil *yangSchema means "no schema configured".
//
// The index is a single flat tree, keyed only by path: it is not aware of
// gNMI origin (unlike SubscriptionConfig.Origin, which subscriptionFor does
// match against). All loaded yang_modules are expected to describe one
// schema tree; see the "yang_modules and origin" README note.
type yangSchema struct {
	byPath map[string]resolvedMetric
}

func (s *yangSchema) lookup(elems []string) (resolvedMetric, bool) {
	if s == nil {
		return resolvedMetric{}, false
	}
	rm, ok := s.byPath[strings.Join(stripModulePrefixes(elems), "/")]
	return rm, ok
}

func stripModulePrefixes(elems []string) []string {
	out := make([]string, len(elems))
	for i, e := range elems {
		if idx := strings.IndexByte(e, ':'); idx >= 0 {
			e = e[idx+1:]
		}
		out[i] = e
	}
	return out
}

func loadYangSchema(paths []string) (*yangSchema, error) {
	if len(paths) == 0 {
		return &yangSchema{byPath: map[string]resolvedMetric{}}, nil
	}

	ms := yang.NewModules()

	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("yang_modules: %w", err)
		}
		if !info.IsDir() {
			ms.AddPath(filepath.Dir(p))
			files = append(files, p)
			continue
		}
		dirs, err := yang.PathsWithModules(p)
		if err != nil {
			return nil, fmt.Errorf("yang_modules: %w", err)
		}
		ms.AddPath(dirs...)
		for _, dir := range dirs {
			entries, err := os.ReadDir(dir)
			if err != nil {
				return nil, fmt.Errorf("yang_modules: %w", err)
			}
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".yang") {
					files = append(files, filepath.Join(dir, e.Name()))
				}
			}
		}
	}

	for _, f := range files {
		if err := ms.Read(f); err != nil {
			return nil, fmt.Errorf("yang_modules: reading %q: %w", f, err)
		}
	}

	if errs := ms.Process(); len(errs) != 0 {
		return nil, yangProcessError(errs)
	}

	entryIndex := map[string]*yang.Entry{}
	pathEntries := map[string][]*yang.Entry{}
	seen := map[*yang.Module]bool{}
	for _, mod := range ms.Modules {
		if seen[mod] {
			continue
		}
		seen[mod] = true
		root := yang.ToEntry(mod)
		if root == nil {
			continue
		}
		for _, child := range root.Dir {
			indexEntries(child, nil, entryIndex, pathEntries)
		}
	}

	byPath := map[string]resolvedMetric{}
	for path, entries := range pathEntries {
		rm, ok := resolveEntry(entries[0], entryIndex, 0)
		if !ok {
			continue
		}
		if err := checkNoPathConflict(path, entries[1:], entryIndex, rm); err != nil {
			return nil, err
		}
		byPath[path] = rm
	}

	return &yangSchema{byPath: byPath}, nil
}

func checkNoPathConflict(path string, rest []*yang.Entry, index map[string]*yang.Entry, first resolvedMetric) error {
	for _, e := range rest {
		rm, ok := resolveEntry(e, index, 0)
		if !ok || sameResolution(first, rm) {
			continue
		}
		return fmt.Errorf(
			"yang_modules: leaf path %q is defined by more than one loaded module with different results "+
				"(type=%q/unit=%q vs type=%q/unit=%q); rename the conflicting leaf or split yang_modules so only one definition loads",
			path, first.Type, first.Unit, rm.Type, rm.Unit)
	}
	return nil
}

func sameResolution(a, b resolvedMetric) bool {
	if a.Type != b.Type || a.Unit != b.Unit || a.kind != b.kind || len(a.EnumValues) != len(b.EnumValues) {
		return false
	}
	for i := range a.EnumValues {
		if a.EnumValues[i] != b.EnumValues[i] {
			return false
		}
	}
	return true
}

func yangProcessError(errs []error) error {
	const maxShown = 5
	msgs := make([]string, 0, maxShown+1)
	for i, err := range errs {
		if i >= maxShown {
			msgs = append(msgs, fmt.Sprintf("... and %d more error(s)", len(errs)-maxShown))
			break
		}
		msgs = append(msgs, err.Error())
	}
	return fmt.Errorf(
		"failed to process YANG modules (every imported/included module must be reachable via yang_modules): %s",
		strings.Join(msgs, "; "))
}

func indexEntries(e *yang.Entry, names []string, index map[string]*yang.Entry, pathEntries map[string][]*yang.Entry) {
	if e.IsChoice() || e.IsCase() {
		for _, child := range e.Dir {
			indexEntries(child, names, index, pathEntries)
		}
		return
	}

	names = append(names[:len(names):len(names)], e.Name)

	if e.IsLeaf() || e.IsLeafList() {
		path := strings.Join(names, "/")
		index[path] = e
		pathEntries[path] = append(pathEntries[path], e)
		return
	}

	for _, child := range e.Dir {
		indexEntries(child, names, index, pathEntries)
	}
}

func resolveEntry(e *yang.Entry, index map[string]*yang.Entry, depth int) (resolvedMetric, bool) {
	t := e.Type
	if t == nil || depth > maxLeafrefDepth {
		return resolvedMetric{}, false
	}

	unit := e.Units
	if unit == "" {
		unit = t.Units
	}

	switch t.Kind {
	case yang.Yint8, yang.Yint16, yang.Yint32, yang.Yint64,
		yang.Yuint8, yang.Yuint16, yang.Ybool:
		return resolvedMetric{
			MetricConfig: MetricConfig{Type: metricTypeGauge, Unit: unit},
			kind:         valueKindInt,
		}, true

	case yang.Yuint32, yang.Yuint64:
		metricType := metricTypeGauge
		if isCounterType(t) {
			metricType = metricTypeSum
		}
		return resolvedMetric{
			MetricConfig: MetricConfig{Type: metricType, Unit: unit},
			kind:         valueKindInt,
		}, true

	case yang.Ydecimal64:
		return resolvedMetric{
			MetricConfig: MetricConfig{Type: metricTypeGauge, Unit: unit},
			kind:         valueKindFloat,
		}, true

	case yang.Yenum:
		if t.Enum == nil {
			return resolvedMetric{}, false
		}
		return newEnumResolvedMetric(unit, t.Enum.Names()), true

	case yang.Yidentityref:
		if t.IdentityBase == nil || len(t.IdentityBase.Values) == 0 {
			return resolvedMetric{}, false
		}
		values := make([]string, len(t.IdentityBase.Values))
		for i, v := range t.IdentityBase.Values {
			values[i] = v.Name
		}
		return newEnumResolvedMetric(unit, values), true

	case yang.Ystring:
		return resolvedMetric{
			MetricConfig: MetricConfig{Type: metricTypeGauge, Unit: unit},
			kind:         valueKindString,
		}, true

	case yang.Yleafref:
		target := resolveLeafrefTarget(t.Path, pathOf(e), index)
		if target == nil {
			return resolvedMetric{}, false
		}
		return resolveEntry(target, index, depth+1)

	default:
		// Yunion, Ybinary, Yempty, Ybits, Ynone and anything unrecognized:
		// values can be ambiguous or non-scalar, so fall back to config.
		return resolvedMetric{}, false
	}
}

func newEnumResolvedMetric(unit string, values []string) resolvedMetric {
	rm := resolvedMetric{
		MetricConfig: MetricConfig{Type: metricTypeGauge, Unit: unit, EnumValues: values},
		kind:         valueKindString,
	}
	cacheNormalizedEnumValues(&rm.MetricConfig)
	return rm
}

func isCounterType(t *yang.YangType) bool {
	for cur := t; cur != nil; cur = baseYangType(cur) {
		if isCounterTypeName(cur.Name) {
			return true
		}
	}
	return false
}

func baseYangType(t *yang.YangType) *yang.YangType {
	if t.Base == nil {
		return nil
	}
	return t.Base.YangType
}

func isCounterTypeName(name string) bool {
	return name == "counter32" || name == "counter64"
}

func pathOf(e *yang.Entry) []string {
	var names []string
	for cur := e; cur != nil && cur.Parent != nil; cur = cur.Parent {
		if cur.IsChoice() || cur.IsCase() {
			continue
		}
		names = append(names, cur.Name)
	}
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
	return names
}

func resolveLeafrefTarget(path string, fromPath []string, index map[string]*yang.Entry) *yang.Entry {
	if path == "" {
		return nil
	}

	var base []string
	steps := strings.Split(path, "/")
	if strings.HasPrefix(path, "/") {
		steps = steps[1:]
	} else {
		if len(fromPath) == 0 {
			return nil
		}
		base = append(base, fromPath...)
	}

	for _, step := range steps {
		if step == "" {
			continue
		}
		if step == ".." {
			if len(base) == 0 {
				return nil
			}
			base = base[:len(base)-1]
			continue
		}
		if idx := strings.IndexByte(step, '['); idx >= 0 {
			step = step[:idx]
		}
		if idx := strings.IndexByte(step, ':'); idx >= 0 {
			step = step[idx+1:]
		}
		if step == "" {
			continue
		}
		base = append(base, step)
	}

	return index[strings.Join(base, "/")]
}
