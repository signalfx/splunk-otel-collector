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

type DaemonSet struct {
	Namespace      string
	Name           string
	ServiceAccount string
	Labels         map[string]string
	Image          string
	ConfigMap      string
	OTLPEndpoint   string
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
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
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
      containers:
      - name: otel-collector
        command:
        - /otelcol
        - --config=/config/config.yaml
{{- if .Image }}
        image: {{ .Image }}
{{- end }}
        imagePullPolicy: IfNotPresent
        env:
{{- if .OTLPEndpoint }}
          - name: OTLP_ENDPOINT
            value: {{ .OTLPEndpoint }}
{{- end }}
        resources:
          limits:
            cpu: 200m
            memory: 500Mi
        volumeMounts:
{{- if .ConfigMap }}
          - mountPath: /config
            name: config-map-volume
{{- end }}
      terminationGracePeriodSeconds: 5
      volumes:
{{- if .ConfigMap }}
        - name: config-map-volume
          configMap:
            name: {{ .ConfigMap }}
            items:
              - key: data
                path: config.yaml
{{- end }}
`

var dsm = Manifest[DaemonSet](daemonSetTemplate)

func (d DaemonSet) Render() (string, error) {
	return dsm.Render(d)
}
