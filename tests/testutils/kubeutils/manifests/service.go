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
	"testing"

	corev1 "k8s.io/api/core/v1"
)

// Service is a "k8s.io/api/core/v1.Service" convenience type
type Service struct {
	Name      string
	Namespace string
	Type      corev1.ServiceType
	Selector  map[string]string
	Ports     []corev1.ServicePort
}

const serviceTemplate = `---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  type: {{ .Type }}
  selector:
  {{- range $key, $value := .Selector }}
    {{ $key }}: {{ $value }}
  {{- end }}
{{- if .Ports }}
  ports:
{{ .Ports | toYaml | indent 4 }}
{{- end }}
`

var sm = Manifest[Service](serviceTemplate)

func (s Service) Render(t testing.TB) string {
	return sm.Render(s, t)
}
