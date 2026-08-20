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

// Command gen transforms the OpenTelemetry Collector Builder (ocb) output into
// the public baseline API. ocb emits a throwaway `package main` with an
// unexported components() func that assembles factories via MakeFactoryMap; this
// tool rewrites that into `package baseline`'s NewBaseline() constructor, which
// returns the raw factory slices so flavours can append to them before Build().
//
// It is invoked by `go generate` after ocb runs — see baseline/gen.go.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

// kind maps the otelcol.Factories field name (as referenced in ocb's
// components() body) to the unexported Baseline slice field the generated
// literal populates and the component-kind package used for its element type
// (e.g. Receivers -> receivers []receiver.Factory).
var kinds = []struct{ factoriesField, field, pkg string }{
	{"Extensions", "extensions", "extension"},
	{"Receivers", "receivers", "receiver"},
	{"Processors", "processors", "processor"},
	{"Exporters", "exporters", "exporter"},
	{"Connectors", "connectors", "connector"},
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: gen <ocb-components.go> <out.go>")
		os.Exit(2)
	}
	src, out := os.Args[1], os.Args[2]
	if err := run(src, out); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run(src, out string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, src, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", src, err)
	}

	// Collect the aliased factory imports (alias -> path). ocb aliases every
	// component import; the bare kind packages (receiver, exporter, …) and
	// otelcol come in unaliased and are handled separately.
	factoryImports := map[string]string{} // alias -> path
	for _, imp := range f.Imports {
		if imp.Name == nil {
			continue // unaliased: kind packages + otelcol, added by us as needed
		}
		path := strings.Trim(imp.Path.Value, `"`)
		factoryImports[imp.Name.Name] = path
	}

	// Find components() and extract, per kind, the ordered list of factory
	// expressions passed to <kind>.MakeFactoryMap(...).
	lists := map[string][]string{} // field -> ["ackextension.NewFactory()", …]
	used := map[string]bool{}      // aliases actually referenced
	fn := findFunc(f, "components")
	if fn == nil {
		return fmt.Errorf("%s: func components() not found", src)
	}
	for _, stmt := range fn.Body.List {
		as, ok := stmt.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != 2 || len(as.Rhs) != 1 {
			continue
		}
		field := factoriesField(as.Lhs[0]) // "Receivers" from factories.Receivers
		if field == "" {
			continue
		}
		call, ok := as.Rhs[0].(*ast.CallExpr)
		if !ok || !isMakeFactoryMap(call.Fun) {
			continue
		}
		for _, arg := range call.Args {
			expr := exprString(fset, arg)
			lists[field] = append(lists[field], expr)
			if alias := leadingIdent(arg); alias != "" {
				used[alias] = true
			}
		}
	}

	return writeFile(out, factoryImports, used, lists)
}

func writeFile(out string, factoryImports map[string]string, used map[string]bool, lists map[string][]string) error {
	var b bytes.Buffer
	b.WriteString(`// Copyright Splunk, Inc.
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

// Code generated from builder-config.yaml by "go generate"; DO NOT EDIT.

package baseline

import (
`)
	// Kind packages used by the slice-literal element types.
	for _, k := range kinds {
		fmt.Fprintf(&b, "\t\"go.opentelemetry.io/collector/%s\"\n", k.pkg)
	}
	b.WriteString("\n")
	// Aliased factory imports, sorted by alias, only those referenced.
	var aliases []string
	for a := range used {
		aliases = append(aliases, a)
	}
	sort.Strings(aliases)
	for _, a := range aliases {
		fmt.Fprintf(&b, "\t%s %q\n", a, factoryImports[a])
	}
	b.WriteString(")\n\n")

	b.WriteString("// NewBaseline returns the shared baseline component set: upstream collector\n")
	b.WriteString("// core and contrib components only, with no Splunk-specific components. It is\n")
	b.WriteString("// the common denominator every flavour layers onto and is not shipped on its own.\n")
	b.WriteString("func NewBaseline() *Baseline {\n")
	b.WriteString("\treturn &Baseline{\n")
	for _, k := range kinds {
		fmt.Fprintf(&b, "\t\t%s: []%s.Factory{\n", k.field, k.pkg)
		for _, e := range lists[k.factoriesField] {
			fmt.Fprintf(&b, "\t\t\t%s,\n", e)
		}
		b.WriteString("\t\t},\n")
	}
	b.WriteString("\t}\n}\n")

	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return fmt.Errorf("gofmt generated source: %w\n%s", err, b.String())
	}
	return os.WriteFile(out, formatted, 0o644)
}

func findFunc(f *ast.File, name string) *ast.FuncDecl {
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == name && fn.Body != nil {
			return fn
		}
	}
	return nil
}

// factoriesField returns "Receivers" for an lhs of the form `factories.Receivers`.
func factoriesField(lhs ast.Expr) string {
	sel, ok := lhs.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	if id, ok := sel.X.(*ast.Ident); !ok || id.Name != "factories" {
		return ""
	}
	return sel.Sel.Name
}

// isMakeFactoryMap reports whether fun is `<pkg>.MakeFactoryMap`, allowing for
// the generic call form `<pkg>.MakeFactoryMap[T](...)` that ocb emits (v0.159+),
// where the callee is an index expression wrapping the selector.
func isMakeFactoryMap(fun ast.Expr) bool {
	switch e := fun.(type) {
	case *ast.IndexExpr:
		return isMakeFactoryMap(e.X)
	case *ast.IndexListExpr:
		return isMakeFactoryMap(e.X)
	case *ast.SelectorExpr:
		return e.Sel.Name == "MakeFactoryMap"
	default:
		return false
	}
}

// leadingIdent returns the package alias of an expression like `foo.NewFactory()`.
func leadingIdent(arg ast.Expr) string {
	call, ok := arg.(*ast.CallExpr)
	if !ok {
		return ""
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return id.Name
}

func exprString(fset *token.FileSet, e ast.Expr) string {
	var b bytes.Buffer
	_ = format.Node(&b, fset, e)
	return b.String()
}
