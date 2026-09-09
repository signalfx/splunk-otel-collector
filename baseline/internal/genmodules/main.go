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

// Command genmodules generates the ordered module paths paired with the
// factories in baseline/components.go.

//go:generate go run -mod=readonly . -components ../../components.go -output ../../modules_generated.go

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var componentKinds = []struct {
	field    string
	variable string
}{
	{field: "extensions", variable: "baselineExtensionModules"},
	{field: "receivers", variable: "baselineReceiverModules"},
	{field: "processors", variable: "baselineProcessorModules"},
	{field: "exporters", variable: "baselineExporterModules"},
	{field: "connectors", variable: "baselineConnectorModules"},
}

type module struct {
	Replace *module
	Path    string
	Version string
}

type moduleFile struct {
	Require []module
	Replace []moduleReplacement
}

type moduleReplacement struct {
	Old module
	New module
}

func main() {
	components := flag.String("components", "components.go", "path to the baseline component source")
	output := flag.String("output", "modules_generated.go", "path to the generated module metadata")
	flag.Parse()

	if err := run(*components, *output); err != nil {
		fmt.Fprintln(os.Stderr, "genmodules:", err)
		os.Exit(1)
	}
}

func run(componentsPath, outputPath string) error {
	componentImports, err := parseComponentImports(componentsPath)
	if err != nil {
		return err
	}

	modules, err := listModules(filepath.Dir(componentsPath))
	if err != nil {
		return err
	}

	componentModules := make(map[string][]module, len(componentKinds))
	for _, kind := range componentKinds {
		for _, importPath := range componentImports[kind.field] {
			componentModule, err := moduleForImport(importPath, modules)
			if err != nil {
				return fmt.Errorf("%s factory import %q: %w", kind.field, importPath, err)
			}
			componentModules[kind.field] = append(componentModules[kind.field], componentModule)
		}
	}

	return writeGenerated(outputPath, componentModules)
}

func parseComponentImports(filename string) (map[string][]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}

	imports := make(map[string]string, len(file.Imports))
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("parse import path %s: %w", spec.Path.Value, err)
		}
		alias := path.Base(importPath)
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		if alias == "_" || alias == "." {
			continue
		}
		imports[alias] = importPath
	}

	baselineLiteral := findBaselineLiteral(file)
	if baselineLiteral == nil {
		return nil, errors.New(filename + ": NewBaseline Baseline literal not found")
	}

	componentImports := make(map[string][]string, len(componentKinds))
	seen := make(map[string]bool, len(componentKinds))
	for _, element := range baselineLiteral.Elts {
		field, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		name, ok := field.Key.(*ast.Ident)
		if !ok || !isComponentKind(name.Name) {
			continue
		}
		list, ok := field.Value.(*ast.CompositeLit)
		if !ok {
			return nil, fmt.Errorf("%s: baseline field %s is not a slice literal", filename, name.Name)
		}
		seen[name.Name] = true
		for _, factoryExpr := range list.Elts {
			importAlias, err := factoryImportAlias(factoryExpr)
			if err != nil {
				return nil, fmt.Errorf("%s: baseline field %s: %w", filename, name.Name, err)
			}
			importPath, ok := imports[importAlias]
			if !ok {
				return nil, fmt.Errorf("%s: baseline field %s: no import for factory qualifier %q", filename, name.Name, importAlias)
			}
			componentImports[name.Name] = append(componentImports[name.Name], importPath)
		}
	}

	for _, kind := range componentKinds {
		if !seen[kind.field] {
			return nil, fmt.Errorf("%s: baseline field %s not found", filename, kind.field)
		}
	}
	return componentImports, nil
}

func findBaselineLiteral(file *ast.File) *ast.CompositeLit {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "NewBaseline" || function.Body == nil {
			continue
		}

		var result *ast.CompositeLit
		ast.Inspect(function.Body, func(node ast.Node) bool {
			literal, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			identifier, ok := literal.Type.(*ast.Ident)
			if ok && identifier.Name == "Baseline" {
				result = literal
				return false
			}
			return true
		})
		return result
	}
	return nil
}

func isComponentKind(field string) bool {
	for _, kind := range componentKinds {
		if kind.field == field {
			return true
		}
	}
	return false
}

func factoryImportAlias(expression ast.Expr) (string, error) {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return "", errors.New("factory expression is not a function call")
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", errors.New("factory constructor is not package-qualified")
	}
	identifier, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", errors.New("factory constructor qualifier is not an identifier")
	}
	return identifier.Name, nil
}

func listModules(directory string) ([]module, error) {
	command := exec.Command("go", "mod", "edit", "-json")
	command.Dir = directory
	command.Env = withoutEnvironmentVariable(os.Environ(), "GOWORK")
	command.Env = append(command.Env, "GOWORK=off")
	output, err := command.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return nil, fmt.Errorf("read go.mod: %w: %s", err, strings.TrimSpace(string(exitError.Stderr)))
		}
		return nil, fmt.Errorf("read go.mod: %w", err)
	}

	var parsed moduleFile
	if err := json.Unmarshal(output, &parsed); err != nil {
		return nil, fmt.Errorf("decode go.mod: %w", err)
	}
	for i := range parsed.Require {
		required := &parsed.Require[i]
		for j := range parsed.Replace {
			replacement := &parsed.Replace[j]
			if replacement.Old.Path == required.Path &&
				(replacement.Old.Version == "" || replacement.Old.Version == required.Version) {
				required.Replace = &replacement.New
				break
			}
		}
	}
	return parsed.Require, nil
}

func withoutEnvironmentVariable(environment []string, name string) []string {
	prefix := name + "="
	filtered := make([]string, 0, len(environment))
	for _, entry := range environment {
		if !strings.HasPrefix(entry, prefix) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func moduleForImport(importPath string, modules []module) (module, error) {
	var selected *module
	for i := range modules {
		candidate := &modules[i]
		if importPath != candidate.Path && !strings.HasPrefix(importPath, candidate.Path+"/") {
			continue
		}
		if selected == nil || len(candidate.Path) > len(selected.Path) {
			selected = candidate
		}
	}
	if selected == nil {
		return module{}, errors.New("owning module not found")
	}
	if selected.Replace != nil {
		return module{}, fmt.Errorf("owning module %s is replaced", selected.Path)
	}
	if selected.Version == "" {
		return module{}, fmt.Errorf("owning module %s has no version", selected.Path)
	}
	return *selected, nil
}

func writeGenerated(filename string, componentModules map[string][]module) error {
	var source bytes.Buffer
	source.WriteString(`// Copyright Splunk, Inc.
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

// Code generated from components.go and go.mod by "go generate"; DO NOT EDIT.
// Production versions are resolved from the final binary's Go build info so
// metadata follows Minimal Version Selection in consuming flavors. The pinned
// references are fallbacks only for dependency-less non-main test binaries.

package baseline

	`)
	for _, kind := range componentKinds {
		fmt.Fprintf(&source, "var %s = []moduleMetadata{\n", kind.variable)
		for _, componentModule := range componentModules[kind.field] {
			fmt.Fprintf(
				&source,
				"\t{path: %q, fallback: %q},\n",
				componentModule.Path,
				componentModule.Path+" "+componentModule.Version,
			)
		}
		source.WriteString("}\n\n")
	}

	formatted, err := format.Source(source.Bytes())
	if err != nil {
		return fmt.Errorf("format generated source: %w", err)
	}
	if err := os.WriteFile(filename, formatted, 0o644); err != nil { //nolint:gosec // Generated Go source should be repository-readable.
		return fmt.Errorf("write %s: %w", filename, err)
	}
	return nil
}
