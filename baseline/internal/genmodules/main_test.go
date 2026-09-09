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

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedMetadataIsCurrent(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "modules_generated.go")
	if err := run("../../components.go", generated); err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile("../../modules_generated.go")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(generated)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("modules_generated.go is stale; run `make generate-baseline-modules` from the repository root")
	}
}

func TestParseComponentImports(t *testing.T) {
	source := `package baseline

import componentfactory "example.com/component/factory"

func NewBaseline() *Baseline {
	return &Baseline{
		extensions: []any{componentfactory.NewFactory()},
		receivers: []any{componentfactory.NewFactory()},
		processors: []any{componentfactory.NewFactory()},
		exporters: []any{componentfactory.NewFactory()},
		connectors: []any{componentfactory.NewFactory()},
	}
}
`
	filename := filepath.Join(t.TempDir(), "components.go")
	if err := os.WriteFile(filename, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	imports, err := parseComponentImports(filename)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range componentKinds {
		if len(imports[kind.field]) != 1 || imports[kind.field][0] != "example.com/component/factory" {
			t.Fatalf("unexpected %s imports: %v", kind.field, imports[kind.field])
		}
	}
}

func TestModuleForImportUsesLongestModulePath(t *testing.T) {
	modules := []module{
		{Path: "example.com/component", Version: "v1.0.0"},
		{Path: "example.com/component/factory", Version: "v2.0.0"},
	}

	componentModule, err := moduleForImport("example.com/component/factory/subpackage", modules)
	if err != nil {
		t.Fatal(err)
	}
	if componentModule.Path != "example.com/component/factory" || componentModule.Version != "v2.0.0" {
		t.Fatalf("unexpected module: %+v", componentModule)
	}
}

func TestModuleForImportRejectsReplacement(t *testing.T) {
	modules := []module{{
		Path:    "example.com/component/factory",
		Version: "v2.0.0",
		Replace: &module{Path: "../factory"},
	}}

	if _, err := moduleForImport("example.com/component/factory", modules); err == nil {
		t.Fatal("expected replaced module to be rejected")
	}
}
