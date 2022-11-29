// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
	"github.com/signalfx/splunk-otel-collector/tests/testutils/kubeutils"
	"github.com/signalfx/splunk-otel-collector/tests/testutils/kubeutils/manifests"
)

func TestDiscoveryReceiverWithK8sObserverProvidesEndpointLogs(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cluster := kubeutils.NewKindCluster(tc)
	defer cluster.Delete()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ns, err := cluster.Clientset.CoreV1().Namespaces().Create(
		ctx, &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)
	require.NoError(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serviceAccount, err := cluster.Clientset.CoreV1().ServiceAccounts(ns.Name).Create(
		ctx, &apiv1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "some.serviceaccount"},
		},
		metav1.CreateOptions{})
	require.NoError(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_, err = cluster.Clientset.CoreV1().Nodes().Create(ctx, &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"some.annotation": "annotation.value",
			},
			Labels: map[string]string{
				"some.label": "label.value",
			},
			Name: "some.node",
		},
		Spec: apiv1.NodeSpec{
			// ensure we aren't scheduling subsequent pod to this node
			Taints: []apiv1.Taint{{Key: "not", Value: "schedulable", Effect: apiv1.TaintEffectNoSchedule}},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = cluster.Clientset.CoreV1().Pods(ns.Name).Create(
		ctx, &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"another.annotation": "another.annotation.value",
				},
				Labels: map[string]string{
					"another.label": "another.label.value",
				},
				Name: "some.pod",
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  "redis",
						Image: "redis",
					},
				},
				ServiceAccountName: serviceAccount.Name,
			},
		}, metav1.CreateOptions{},
	)
	require.NoError(tc, err)

	_, shutdown := tc.SplunkOtelCollectorProcess(
		"k8s_observer_endpoints_config.yaml",
		func(collector testutils.Collector) testutils.Collector {
			return collector.WithEnv(map[string]string{
				"KUBECONFIG": cluster.Kubeconfig,
			})
		},
	)
	defer shutdown()

	expectedResourceLogs := tc.ResourceLogs("k8s_observer_endpoints.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedResourceLogs, 30*time.Second))
}

func TestDiscoveryReceiverWithK8sObserverAndSmartAgentRedisReceiverProvideStatusLogs(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cluster := kubeutils.NewKindCluster(tc)
	defer cluster.Delete()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	namespace, err := cluster.Clientset.CoreV1().Namespaces().Create(
		ctx, &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)
	require.NoError(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serviceAccount, err := cluster.Clientset.CoreV1().ServiceAccounts(namespace.Name).Create(
		ctx, &apiv1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "some.serviceaccount"},
		},
		metav1.CreateOptions{})
	require.NoError(t, err)

	config, err := os.ReadFile(filepath.Join(".", "testdata", "k8s_observer_smart_agent_redis_config.yaml"))
	require.NoError(t, err)
	cm := manifests.ConfigMap{
		Name: "collector.config", Namespace: namespace.Name,
		Data: string(config),
	}
	cmm, err := cm.Render()
	require.NoError(t, err)
	sout, serr, err := cluster.Apply(cmm)
	fmt.Printf("stdout: %s\n", sout.String())
	fmt.Printf("stderr: %s\n", serr.String())
	require.NoError(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	redis, err := cluster.Clientset.CoreV1().Pods(namespace.Name).Create(
		ctx, &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"another.annotation": "another.annotation.value",
				},
				Labels: map[string]string{
					"another.label": "another.label.value",
				},
				Name: "target.redis",
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Image: "redis",
						Name:  "redis",
						// currently we're creating the port but are unable to reach it
						// until we add extraPortMappings to the cluster.
						Ports:           []apiv1.ContainerPort{{ContainerPort: 6379}},
						ImagePullPolicy: apiv1.PullIfNotPresent,
					},
				},
				ServiceAccountName: serviceAccount.Name,
				HostNetwork:        true,
			},
		}, metav1.CreateOptions{},
	)
	require.NoError(t, err)

	crManifests := `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-role
  namespace: test-namespace
  labels:
rules:
- apiGroups:
  - ""
  resources:
  - events
  - namespaces
  - namespaces/status
  - nodes
  - nodes/spec
  - nodes/stats
  - nodes/proxy
  - pods
  - pods/status
  - persistentvolumeclaims
  - persistentvolumes
  - replicationcontrollers
  - replicationcontrollers/status
  - resourcequotas
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - apps
  resources:
  - daemonsets
  - deployments
  - replicasets
  - statefulsets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - daemonsets
  - deployments
  - replicasets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  - cronjobs
  verbs:
  - get
  - list
  - watch
- apiGroups:
    - autoscaling
  resources:
    - horizontalpodautoscalers
  verbs:
    - get
    - list
    - watch
- nonResourceURLs:
  - /metrics
  verbs:
  - get
  - list
  - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cluster-role-binding
  namespace: test-namespace
  labels:
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-role
subjects:
- kind: ServiceAccount
  name: some.serviceaccount
  namespace: test-namespace
`
	cluster.Apply(crManifests)
	require.NoError(t, err)

	splat := strings.Split(tc.OTLPEndpoint, ":")
	port := splat[len(splat)-1]
	otlpEndpoint := fmt.Sprintf(
		"%s:%s", cluster.GetDefaultGatewayIP(), port,
	)

	ds := manifests.DaemonSet{
		Name:           "an.agent.daemonset",
		Namespace:      namespace.Name,
		ServiceAccount: serviceAccount.Name,
		Labels:         map[string]string{"label.key": "label.value"},
		Image:          testutils.GetCollectorImageOrSkipTest(t),
		ConfigMap:      cm.Name,
		OTLPEndpoint:   otlpEndpoint,
	}
	dsm, err := ds.Render()
	require.NoError(t, err)

	sout, serr, e := cluster.Apply(dsm)
	fmt.Printf("stdout: %s\n", sout.String())
	fmt.Printf("stderr: %s\n", serr.String())
	require.NoError(t, e)

	require.Eventually(t, func() bool {
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rPod, err := cluster.Clientset.CoreV1().Pods(redis.Namespace).Get(ctx, redis.Name, metav1.GetOptions{})
		require.NoError(t, err)
		fmt.Printf("redis is: %s\n", rPod.Status.Phase)
		return rPod.Status.Phase == apiv1.PodRunning
	}, 5*time.Minute, 1*time.Second)

	require.Eventually(t, func() bool {
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		dsPods, err := cluster.Clientset.CoreV1().Pods(namespace.Name).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("name = %s", ds.Name),
		})
		require.NoError(t, err)
		if len(dsPods.Items) > 0 {
			collectorPod := dsPods.Items[0]
			fmt.Printf("collectorPod is %v\n", collectorPod.Status.Phase)
			cPod, err := cluster.Clientset.CoreV1().Pods(collectorPod.Namespace).Get(ctx, collectorPod.Name, metav1.GetOptions{})
			require.NoError(t, err)
			return cPod.Status.Phase == apiv1.PodRunning
		}
		return false
	}, 5*time.Minute, 1*time.Second)

	expectedResourceLogs := tc.ResourceLogs("k8s_observer_smart_agent_redis_statuses.yaml")
	// give time for redis to start
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedResourceLogs, 30*time.Second))
}
