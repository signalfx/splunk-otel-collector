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

//go:build integration

package kubeutils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestKindCluster(t *testing.T) {
	tc := testutils.NewTestcase(t)
	cluster := NewKindCluster(tc)
	ports := []int{int(testutils.GetAvailablePort(tc)), int(testutils.GetAvailablePort(tc))}
	sort.Ints(ports)
	portOne, portTwo := uint16(ports[0]), uint16(ports[1])
	cluster.ExposedPorts[portOne] = 12345
	cluster.ExposedPorts[portTwo] = 23456

	defer func() {
		cluster.Delete()

		// confirm node unavailable
		_, err := cluster.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, syscall.ECONNREFUSED))

		_, err = os.Stat(cluster.Kubeconfig)
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	}()

	cluster.Create()
	assert.Equal(t, fmt.Sprintf("cluster-%s", tc.ID), cluster.Name)
	require.NotNil(t, cluster.Clientset)

	config, err := os.ReadFile(cluster.Config)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 12345
        hostPort: %d
        listenAddress: "0.0.0.0"
        protocol: tcp
      - containerPort: 23456
        hostPort: %d
        listenAddress: "0.0.0.0"
        protocol: tcp
`, portOne, portTwo), string(config))

	nodes, err := cluster.Clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, nodes)
	assert.Len(t, nodes.Items, 1)
	assert.Equal(t, fmt.Sprintf("%s-control-plane", cluster.Name), nodes.Items[0].Name)

	namespace := &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}}
	ns, err := cluster.Clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	require.NoError(t, err)
	require.Equal(t, "test-namespace", ns.Name)

	require.NotEmpty(t, cluster.GetDefaultGatewayIP())
}
