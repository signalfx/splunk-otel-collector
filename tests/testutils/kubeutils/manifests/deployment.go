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

// Deployment is a "k8s.io/api/apps/v1.Deployment" convenience type
type Deployment struct {
	Labels         map[string]string
	MatchLabels    map[string]string
	Namespace      string
	Name           string
	ServiceAccount string
	Containers     []corev1.Container
	Volumes        []corev1.Volume
	Replicas       int
}

const deploymentTemplate = `---
apiVersion: apps/v1
kind: Deployment
metadata:
{{- if .Name }}
  name: {{ .Name }}
{{- end -}}
{{- if .Namespace }}
  namespace: {{ .Namespace }}
{{- end }}
{{- if .Labels }}
  labels:
  {{- range $key, $value := .Labels }}
    {{ $key }}: {{ $value }}
  {{- end }}
{{- end }}
spec:
  replicas: {{ .Replicas }}
  selector:
    matchLabels:
{{- if .Name }}
      name: {{ .Name }}
{{- end }}
{{- if .MatchLabels }}
  {{- range $key, $value := .MatchLabels }}
      {{ $key }}: {{ $value }}
  {{- end }}
{{- end }}
  template:
    metadata:
      labels:
{{- if .Name }}
        name: {{ .Name }}
{{- end }}
{{- if .MatchLabels }}
  {{- range $key, $value := .MatchLabels }}
        {{ $key }}: {{ $value }}
  {{- end }}
{{- end }}
    spec:
{{- if .ServiceAccount }}
      serviceAccountName: {{ .ServiceAccount }}
{{- end }}
{{- if .Containers }}
      containers:
{{ .Containers | toYaml | indent  6 }}
{{- end }}
{{- if .Volumes }}
      volumes:
{{ .Volumes | toYaml | indent  6 }}
{{- end }}
`

var dm = Manifest[Deployment](deploymentTemplate)

func (d Deployment) Render(t testing.TB) string {
	return dm.Render(d, t)
}
