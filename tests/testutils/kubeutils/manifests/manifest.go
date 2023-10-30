// Copyright Splunk, Inc.
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

package manifests

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

type renderable interface {
	Render(tb testing.TB) string
}

type M[T any] struct {
	template string
}

// Manifest is a generic function that returns a Render-able type with the text/template
// using an instance of the type argument
func Manifest[T any](template string) M[T] {
	return M[T]{template: template}
}

// Render takes an instance of the type argument and renders the Manifest's
// template w/ its fields.
func (m M[T]) Render(t T, tb testing.TB) string {
	out := &bytes.Buffer{}
	tpl, err := template.New("").Funcs(funcMap()).Parse(m.template)
	require.NoError(tb, err)
	err = tpl.Execute(out, t)
	require.NoError(tb, err)
	return out.String()
}

func RenderAll(tb testing.TB, manifests ...renderable) string {
	var rendered []string
	for _, m := range manifests {
		rendered = append(rendered, m.Render(tb))
	}
	return strings.Join(rendered, "")
}

// funcMap provides all sprig functions with an additional k8s yaml helper.
// This enables manifest templates to be able to render k8s core/api types:
// {{ .SomeK8sType | toYaml }}
func funcMap() map[string]any {
	fm := sprig.TxtFuncMap()
	fm["toYaml"] = func(k8sType any) (string, error) {
		rendered, err := yaml.Marshal(k8sType)
		if err == nil {
			return strings.TrimSuffix(string(rendered), "\n"), nil
		}
		return "", err
	}
	return fm
}
