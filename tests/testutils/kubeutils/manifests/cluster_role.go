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
	rbacv1 "k8s.io/api/rbac/v1"
)

// ClusterRole is a "k8s.io/api/rbac/v1.ClusterRole" convenience type
type ClusterRole struct {
	Namespace string
	Name      string
	Labels    map[string]string
	Rules     []rbacv1.PolicyRule
}

const clusterRoleTemplate = `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
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
{{- if .Rules }}
rules:
{{ .Rules | toYaml }}
{{- end }}
`

var crm = Manifest[ClusterRole](clusterRoleTemplate)

func (cr ClusterRole) Render() (string, error) {
	return crm.Render(cr)
}
