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

//go:build testutils

package manifests

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestDeployment(t *testing.T) {
	dep := Deployment{
		Name:      "some.deployment",
		Namespace: "some.namespace",
		Labels: map[string]string{
			"label.one": "value.one",
			"label.two": "value.two",
		},
		MatchLabels: map[string]string{
			"match.label.one": "match.value.one",
			"match.label.two": "match.value.two",
		},
		Replicas: 123,
		Containers: []v1.Container{
			{
				Name:  "a.container",
				Image: "an.image",
				Env: []v1.EnvVar{
					{Name: "an.env.var", Value: "a.value"},
				},
			},
			{
				Name:  "another.container",
				Image: "another.image",
				VolumeMounts: []v1.VolumeMount{
					{
						Name: "config-map-volume", MountPath: "/config/map",
					},
				},
			},
		},
		ServiceAccount: "some.service.account",
		Volumes: []v1.Volume{
			{
				Name: "config-map-volume",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "some-config-map",
						},
						Items: []v1.KeyToPath{
							{
								Key:  "config",
								Path: "config.yaml",
							},
						},
					},
				},
			},
		},
	}

	require.Equal(t,
		`---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: some.deployment
  namespace: some.namespace
  labels:
    label.one: value.one
    label.two: value.two
spec:
  replicas: 123
  selector:
    matchLabels:
      name: some.deployment
      match.label.one: match.value.one
      match.label.two: match.value.two
  template:
    metadata:
      labels:
        name: some.deployment
        match.label.one: match.value.one
        match.label.two: match.value.two
    spec:
      serviceAccountName: some.service.account
      containers:
      - env:
        - name: an.env.var
          value: a.value
        image: an.image
        name: a.container
        resources: {}
      - image: another.image
        name: another.container
        resources: {}
        volumeMounts:
        - mountPath: /config/map
          name: config-map-volume
      volumes:
      - configMap:
          items:
          - key: config
            path: config.yaml
          name: some-config-map
        name: config-map-volume
`, dep.Render(t))
}
