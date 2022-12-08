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

type ConfigMap struct {
	Namespace string
	Name      string
	Data      string
}

const configMapTemplate = `---
apiVersion: v1
kind: ConfigMap
metadata:
{{- if .Name }}
  name: {{ .Name }}
{{- end -}}
{{- if .Namespace }}
  namespace: {{ .Namespace }}
{{- end }}
data:
{{- if .Data }}
{{ .Data | indent 2 -}}
{{ end }}
`

var cmm = Manifest[ConfigMap](configMapTemplate)

func (cm ConfigMap) Render() (string, error) {
	return cmm.Render(cm)
}
