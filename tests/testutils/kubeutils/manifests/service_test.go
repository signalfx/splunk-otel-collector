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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestService(t *testing.T) {
	s := Service{
		Name:      "some.service",
		Namespace: "some.namespace",
		Type:      corev1.ServiceTypeLoadBalancer,
		Selector: map[string]string{
			"selector.key": "selector.value",
		},
		Ports: []corev1.ServicePort{
			{
				Name:       "port-one",
				Protocol:   corev1.ProtocolTCP,
				Port:       12345,
				TargetPort: intstr.FromInt32(23456),
			},
			{
				Name:       "port-one",
				Protocol:   corev1.ProtocolTCP,
				Port:       12345,
				TargetPort: intstr.FromInt32(23456),
			},
		},
	}

	require.Equal(t,
		`---
apiVersion: v1
kind: Service
metadata:
  name: some.service
  namespace: some.namespace
spec:
  type: LoadBalancer
  selector:
    selector.key: selector.value
  ports:
    - name: port-one
      port: 12345
      protocol: TCP
      targetPort: 23456
    - name: port-one
      port: 12345
      protocol: TCP
      targetPort: 23456
`, s.Render(t))
}
