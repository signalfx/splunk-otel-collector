// Copyright  Splunk, Inc.
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

package utils

import (
	"bytes"
	"text/template"
)

// RenderSimpleTemplate processes a simple, self-contained template with a
// given context and returns the final result.
func RenderSimpleTemplate(tmpl string, context interface{}) (string, error) {
	template, err := template.New("nested").Parse(tmpl)
	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	// fill in any templates with the whole config struct passed into this method
	err = template.Option("missingkey=error").Execute(&out, context)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}
