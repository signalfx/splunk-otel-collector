// Copyright Splunk, Inc.
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

package manifests

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

var sprigFuncMap = sprig.TxtFuncMap()

type M[T any] struct {
	template string
}

func Manifest[T any](template string) M[T] {
	return M[T]{template: template}
}

func (m M[T]) Render(t T) (string, error) {
	out := &bytes.Buffer{}
	tpl, err := template.New("").Funcs(sprigFuncMap).Parse(m.template)
	if err != nil {
		return "", err
	}
	if err = tpl.Execute(out, t); err != nil {
		return "", err
	}
	return out.String(), nil
}
