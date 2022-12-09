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
	corev1 "k8s.io/api/core/v1"
)

// DaemonSet is a "k8s.io/api/apps/v1.DaemonSet" convenience type
type DaemonSet struct {
	Namespace      string
	Name           string
	ServiceAccount string
	Labels         map[string]string
	Containers     []corev1.Container
	Volumes        []corev1.Volume
}

const daemonSetTemplate = `---
apiVersion: apps/v1
kind: DaemonSet
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
  selector:
    matchLabels:
{{- if .Name }}
      name: {{ .Name }}
{{- end }}
  template:
    metadata:
      labels:
{{- if .Name }}
        name: {{ .Name }}
{{- end }}
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
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

var dsm = Manifest[DaemonSet](daemonSetTemplate)

func (d DaemonSet) Render() (string, error) {
	return dsm.Render(d)
}
