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

// ClusterRoleBinding is a "k8s.io/api/rbac/v1.ClusterRoleBinding" convenience type
type ClusterRoleBinding struct {
	Labels             map[string]string
	RoleRef            *rbacv1.RoleRef
	Namespace          string
	Name               string
	ClusterRoleName    string
	ServiceAccountName string
	Subjects           []rbacv1.Subject
}

const clusterRoleBindingTemplate = `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
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
{{- if .RoleRef }}
roleRef:
{{ .RoleRef | toYaml | indent 2 }}
{{- else if .ClusterRoleName }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ .ClusterRoleName }}
{{- end }}
{{- if .Subjects }}
subjects:
{{ .Subjects | toYaml }}
{{- else if .ServiceAccountName }}
subjects:
- kind: ServiceAccount
  name: {{ .ServiceAccountName }}
{{- if .Namespace }}
  namespace: {{ .Namespace }}
{{- end }}
{{- end }}
`

var crbm = Manifest[ClusterRoleBinding](clusterRoleBindingTemplate)

func (crb ClusterRoleBinding) Render() (string, error) {
	return crbm.Render(crb)
}
