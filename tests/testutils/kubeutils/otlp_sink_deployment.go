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

package kubeutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

type OTLPSinkDeployment struct {
	tc      *testutils.Testcase
	cluster *KindCluster
	sb      *sbRest
}

func NewOTLPSinkDeployment(cluster *KindCluster) *OTLPSinkDeployment {
	dep := &OTLPSinkDeployment{
		tc:      cluster.Testcase,
		cluster: cluster,
	}
	var err error
	dep.sb, err = newSBRest(cluster.Testcase.Logf)
	require.NoError(dep.tc, err)
	dep.apply()
	cluster.WaitForPods("otlp-sink-.*", "testing", 2*time.Minute)
	return dep
}

func (dep OTLPSinkDeployment) Teardown() {
	dep.sb.shutdown()
	so, se, err := dep.cluster.Delete(dep.manifests())
	require.NoError(dep.tc, err, "stdout: %s, stderr: %s", so.String(), se.String())
}

func (dep OTLPSinkDeployment) apply() {
	so, se, err := dep.cluster.Apply(dep.manifests())
	require.NoError(dep.tc, err, "stdout: %s, stderr: %s", so.String(), se.String())
}

func (dep OTLPSinkDeployment) manifests() string {
	require.NotNil(dep.tc, dep.sb)
	_, port, err := net.SplitHostPort(dep.sb.addr)
	require.NoError(dep.tc, err)

	require.NotNil(dep.tc, dep.cluster)
	info := struct {
		APIUrl       string
		OTLPEndpoint string
	}{
		APIUrl:       fmt.Sprintf("%s:%s", dep.cluster.HostFromContainer(), port),
		OTLPEndpoint: dep.cluster.OTLPEndointFromContainer(),
	}

	tmpl := `---
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
    component: otlp-sink
spec:
  replicas: 1
  selector:
    matchLabels:
      app: otlp-sink
      component: otlp-sink
  template:
    metadata:
      labels:
        app: otlp-sink
        component: otlp-sink
    spec:
      containers:
      - name: otel-collector
        command:
        - /otelcol
        - --config=/conf/relay.yaml
        image: otelcol:latest
        imagePullPolicy: IfNotPresent
        env:
          - name: SPLUNK_MEMORY_TOTAL_MIB
            value: "128"
        ports:
        - name: http-forwarder
          containerPort: 6060
          protocol: TCP
        - name: otlp
          containerPort: 4317
          protocol: TCP
        - name: otlp-http
          containerPort: 4318
          protocol: TCP
        - name: signalfx
          containerPort: 9943
          protocol: TCP
        - name: sapm
          containerPort: 7276
          protocol: TCP
        - name: splunk-hec
          containerPort: 8088
          protocol: TCP
        readinessProbe:
          httpGet:
            path: /
            port: 13133
        livenessProbe:
          httpGet:
            path: /
            port: 13133
        resources:
          limits:
            cpu: 200m
            memory: 128Mi
        volumeMounts:
        - mountPath: /conf
          name: otlp-sink-configmap
      terminationGracePeriodSeconds: 600
      volumes:
      - name: otlp-sink-configmap
        configMap:
          name: otlp-sink
          items:
            - key: relay
              path: relay.yaml
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
          endpoint: http://{{ .APIUrl }}
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
        endpoint: {{ .OTLPEndpoint }}
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
  ports:
  - name: http-forwarder
    port: 26060
    targetPort: http-forwarder
    protocol: TCP
  - name: otlp
    port: 24317
    targetPort: otlp
    protocol: TCP
  - name: otlp-http
    port: 24318
    targetPort: otlp-http
    protocol: TCP
  - name: signalfx
    port: 29943
    targetPort: signalfx
    protocol: TCP
  - name: splunk-hec
    port: 28088
    targetPort: splunk-hec
    protocol: TCP
  - name: sapm
    port: 27276
    targetPort: sapm
    protocol: TCP
  selector:
    app: otlp-sink
    component: otlp-sink
`
	manifestTmpl, err := template.New("").Parse(tmpl)
	require.NoError(dep.tc, err)

	var out bytes.Buffer
	require.NoError(dep.tc, manifestTmpl.Execute(&out, info))
	return out.String()
}

var _ http.Handler = (*sbRest)(nil)

type sbRest struct {
	server *http.Server
	logf   func(format string, args ...any)
	addr   string
}

func newSBRest(logf func(format string, args ...any)) (*sbRest, error) {
	sbr := &sbRest{logf: logf}
	listener, err := net.Listen("tcp", "0.0.0.0:0") // nolint:gosec // not a production server
	if err != nil {
		return nil, err
	}
	sbr.addr = listener.Addr().String()
	sbr.server = &http.Server{Handler: sbr} // nolint:gosec // not a production server
	go sbr.server.Serve(listener)
	return sbr, nil
}

func (sbr *sbRest) shutdown() {
	sbr.server.Shutdown(context.Background())
}

func (sbr *sbRest) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	msg := &strings.Builder{}
	fmt.Fprintf(msg, "url: %s\n", request.URL)
	fmt.Fprintf(msg, "method: %v\n", request.Method)
	for k, v := range request.Header {
		fmt.Fprintf(msg, "header: %v: %v\n", k, v)
	}
	c, err := io.ReadAll(request.Body)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	body := map[string]any{}
	if err = json.Unmarshal(c, &body); err == nil {
		if c, err = json.MarshalIndent(body, "", " "); err != nil {
			fmt.Printf("jsonMarshalIndent: %v\n", err)
		}
	} else {
		fmt.Printf("jsonUnmarshal: %v\n", err)
	}
	fmt.Fprintf(msg, "body: %v\n\n", string(c))
	sbr.logf("received api request: %s", msg.String())
	writer.WriteHeader(200)
}
