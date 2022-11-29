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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDaemonSet(t *testing.T) {
	ds := DaemonSet{
		Name:           "some.daemon.set",
		Namespace:      "some.namespace",
		Image:          "some.image",
		ConfigMap:      "some.config.map",
		OTLPEndpoint:   "some.otlp.endpoint",
		ServiceAccount: "some.service.account",
	}

	manifest, err := ds.Render()
	require.NoError(t, err)
	require.Equal(t,
		`---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: some.daemon.set
  namespace: some.namespace
spec:
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  selector:
    matchLabels:
      name: some.daemon.set
  template:
    metadata:
      labels:
        name: some.daemon.set
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      serviceAccountName: some.service.account
      containers:
      - name: otel-collector
        command:
        - /otelcol
        - --config=/config/config.yaml
        image: some.image
        imagePullPolicy: IfNotPresent
        env:
          - name: OTLP_ENDPOINT
            value: some.otlp.endpoint
        resources:
          limits:
            cpu: 200m
            memory: 500Mi
        volumeMounts:
          - mountPath: /config
            name: config-map-volume
      terminationGracePeriodSeconds: 5
      volumes:
        - name: config-map-volume
          configMap:
            name: some.config.map
            items:
              - key: data
                path: config.yaml
`, manifest)
}

func TestEmptyDaemonSet(t *testing.T) {
	ds := DaemonSet{}
	manifest, err := ds.Render()
	require.NoError(t, err)
	require.Equal(t,
		`---
apiVersion: apps/v1
kind: DaemonSet
metadata:
spec:
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  selector:
    matchLabels:
  template:
    metadata:
      labels:
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: otel-collector
        command:
        - /otelcol
        - --config=/config/config.yaml
        imagePullPolicy: IfNotPresent
        env:
        resources:
          limits:
            cpu: 200m
            memory: 500Mi
        volumeMounts:
      terminationGracePeriodSeconds: 5
      volumes:
`, manifest)
}
