// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build testutils

package kubeutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestOTLPSinkDeploymentManifest(t *testing.T) {
	t.Cleanup(func() func() {
		ciev := "SPLUNK_OTEL_COLLECTOR_IMAGE"
		prev, ok := os.LookupEnv(ciev)
		os.Setenv(ciev, "some.otelcol.image:with.tag")
		return func() {
			if ok {
				os.Setenv(ciev, prev)
			} else {
				os.Unsetenv(ciev)
			}
		}
	}(),
	)

	deployment := OTLPSinkDeployment{tc: testutils.NewTestcase(t)}
	sb, err := newSBRest(t.Logf)
	require.NoError(t, err)
	deployment.sb = sb
	deployment.apiEndpoint = "some.api.endpoint"
	deployment.otlpEndpoint = "some.otlp.endpoint"

	expectedManifests := `---
apiVersion: v1
kind: Namespace
metadata:
  name: testing
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otlp-sink
  namespace: testing
  labels:
    app: otlp-sink
spec:
  replicas: 1
  selector:
    matchLabels:
      name: otlp-sink
      app: otlp-sink
  template:
    metadata:
      labels:
        name: otlp-sink
        app: otlp-sink
    spec:
      containers:
      - command:
        - /otelcol
        - --config=/conf/relay.yaml
        env:
        - name: SPLUNK_MEMORY_TOTAL_MIB
          value: "128"
        image: some.otelcol.image:with.tag
        imagePullPolicy: IfNotPresent
        livenessProbe:
          httpGet:
            path: /
            port: 13133
          terminationGracePeriodSeconds: 10
        name: otel-collector
        ports:
        - containerPort: 6060
          name: http-forwarder
          protocol: TCP
        - containerPort: 4317
          name: otlp
          protocol: TCP
        - containerPort: 4318
          name: otlp-http
          protocol: TCP
        - containerPort: 7276
          name: sapm
          protocol: TCP
        - containerPort: 9943
          name: signalfx
          protocol: TCP
        - containerPort: 8088
          name: splunk-hec
          protocol: TCP
        resources:
          limits:
            cpu: 200m
            memory: 128Mi
        volumeMounts:
        - mountPath: /conf
          name: otlp-sink-configmap
      volumes:
      - configMap:
          items:
          - key: relay
            path: relay.yaml
          name: otlp-sink
        name: otlp-sink-configmap
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: otlp-sink
  namespace: testing
data:
  relay: |
    extensions:
      health_check:
      http_forwarder:
        egress:
          endpoint: http://some.api.endpoint
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318
      sapm:
        endpoint: 0.0.0.0:7276
      signalfx:
        access_token_passthrough: true
        endpoint: 0.0.0.0:9943
      splunk_hec:
        endpoint: 0.0.0.0:8088
    exporters:
      otlp:
        endpoint: some.otlp.endpoint
        tls:
          insecure: true
    service:
      extensions:
        - health_check
        - http_forwarder
      pipelines:
        logs:
          exporters:
          - otlp
          receivers:
          - otlp
          - splunk_hec
          - signalfx
        metrics:
          exporters:
          - otlp
          receivers:
          - otlp
          - signalfx
          - splunk_hec
        traces:
          exporters:
          - otlp
          receivers:
          - otlp
          - sapm
---
apiVersion: v1
kind: Service
metadata:
  name: otlp-sink
  namespace: testing
spec:
  type: ClusterIP
  selector:
    app: otlp-sink
  ports:
    - name: http-forwarder
      port: 26060
      protocol: TCP
      targetPort: http-forwarder
    - name: otlp
      port: 24317
      protocol: TCP
      targetPort: otlp
    - name: otlp-http
      port: 24318
      protocol: TCP
      targetPort: otlp-http
    - name: sapm
      port: 27276
      protocol: TCP
      targetPort: sapm
    - name: signalfx
      port: 29943
      protocol: TCP
      targetPort: signalfx
    - name: splunk-hec
      port: 28088
      protocol: TCP
      targetPort: splunk-hec
`
	require.Equal(t, expectedManifests, deployment.manifests())
}
