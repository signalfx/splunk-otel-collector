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
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
	"github.com/signalfx/splunk-otel-collector/tests/testutils/kubeutils"
	"github.com/signalfx/splunk-otel-collector/tests/testutils/kubeutils/manifests"
)

func TestDiscoveryReceiverWithK8sObserverProvidesEndpointLogs(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	tc := testutils.NewTestcase(t, testutils.OTLPReceiverSinkAllInterfaces)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cluster := kubeutils.NewKindCluster(tc)
	defer cluster.Teardown()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	namespace := manifests.Namespace{Name: "test-namespace"}
	serviceAccount := manifests.ServiceAccount{Name: "some.serviceacount", Namespace: namespace.Name}
	configMap := manifests.ConfigMap{
		Name: "collector.config", Namespace: namespace.Name,
		Data: configMapData(t, "k8s_observer_endpoints_config.yaml"),
	}

	sout, serr, err := cluster.Apply(manifests.RenderAll(t, namespace, serviceAccount, configMap))
	require.NoError(t, err, "stdout: %s, stderr: %s", sout, serr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_, err = cluster.Clientset.CoreV1().Nodes().Create(ctx, &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"some.annotation": "annotation.value",
			},
			Labels: map[string]string{
				"some.label": "label.value",
			},
			Name: "some.node",
		},
		Spec: corev1.NodeSpec{
			// ensure we aren't scheduling subsequent pod to this node
			Taints: []corev1.Taint{{Key: "not", Value: "schedulable", Effect: corev1.TaintEffectNoSchedule}},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	_ = createRedis(cluster, "some.pod", namespace.Name, serviceAccount.Name)

	clusterRole, clusterRoleBinding := clusterRoleAndBinding(namespace.Name, serviceAccount.Name)
	ds := daemonSet(cluster, namespace.Name, serviceAccount.Name, configMap.Name)

	sout, serr, err = cluster.Apply(manifests.RenderAll(t, clusterRole, clusterRoleBinding, ds))
	require.NoError(t, err, "stdout: %s, stderr: %s", sout, serr)

	cluster.WaitForPods(ds.Name, namespace.Name, 5*time.Minute)

	expectedResourceLogs := tc.ResourceLogs("k8s_observer_endpoints.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedResourceLogs, 30*time.Second))
}

func TestDiscoveryReceiverWithK8sObserverAndSmartAgentRedisReceiverProvideStatusLogs(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	tc := testutils.NewTestcase(t, testutils.OTLPReceiverSinkAllInterfaces)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cluster := kubeutils.NewKindCluster(tc)
	defer cluster.Teardown()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	namespace := manifests.Namespace{Name: "test-namespace"}
	serviceAccount := manifests.ServiceAccount{Name: "some.serviceacount", Namespace: namespace.Name}
	configMap := manifests.ConfigMap{
		Name: "collector.config", Namespace: namespace.Name,
		Data: configMapData(t, "k8s_observer_smart_agent_redis_config.yaml"),
	}

	sout, serr, err := cluster.Apply(manifests.RenderAll(t, namespace, serviceAccount, configMap))
	require.NoError(t, err, "stdout: %s, stderr: %s", sout, serr)

	redis := createRedis(cluster, "target.redis", namespace.Name, serviceAccount.Name)

	clusterRole, clusterRoleBinding := clusterRoleAndBinding(namespace.Name, serviceAccount.Name)
	ds := daemonSet(cluster, namespace.Name, serviceAccount.Name, configMap.Name)

	sout, serr, err = cluster.Apply(manifests.RenderAll(t, clusterRole, clusterRoleBinding, ds))
	require.NoError(t, err, "stdout: %s, stderr: %s", sout, serr)

	cluster.WaitForPods(redis, namespace.Name, 5*time.Minute)
	cluster.WaitForPods(ds.Name, namespace.Name, 5*time.Minute)

	expectedResourceLogs := tc.ResourceLogs("k8s_observer_smart_agent_redis_statuses.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedResourceLogs, 30*time.Second))
}

func createRedis(cluster *kubeutils.KindCluster, name, namespace, serviceAccount string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	redis, err := cluster.Clientset.CoreV1().Pods(namespace).Create(
		ctx, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"another.annotation": "another.annotation.value",
				},
				Labels: map[string]string{
					"another.label": "another.label.value",
				},
				Name: name,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image:           "redis",
						Name:            "redis",
						Ports:           []corev1.ContainerPort{{ContainerPort: 6379}},
						ImagePullPolicy: corev1.PullIfNotPresent,
					},
				},
				ServiceAccountName: serviceAccount,
				HostNetwork:        true,
			},
		}, metav1.CreateOptions{},
	)
	require.NoError(cluster.Testcase, err)
	return redis.Name
}

func configMapData(t testing.TB, configPath string) string {
	config, err := os.ReadFile(filepath.Join(".", "testdata", configPath))
	configStore := map[string]any{"config": string(config)}
	configYaml, err := yaml.Marshal(configStore)
	require.NoError(t, err)
	return string(configYaml)
}

func clusterRoleAndBinding(namespace, serviceAccount string) (manifests.ClusterRole, manifests.ClusterRoleBinding) {
	cr := manifests.ClusterRole{
		Name:      "cluster-role",
		Namespace: namespace,
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{
					"events",
					"namespaces",
					"namespaces/status",
					"nodes",
					"nodes/spec",
					"nodes/stats",
					"nodes/proxy",
					"pods",
					"pods/status",
					"persistentvolumeclaims",
					"persistentvolumes",
					"replicationcontrollers",
					"replicationcontrollers/status",
					"resourcequotas",
					"services",
				},
				Verbs: []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{
					"daemonsets",
					"deployments",
					"replicasets",
					"statefulsets",
				},
				Verbs: []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"extensions"},
				Resources: []string{
					"daemonsets",
					"deployments",
					"replicasets",
				},
				Verbs: []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{
					"jobs",
					"cronjobs",
				},
				Verbs: []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"autoscaling"},
				Resources: []string{
					"horizontalpodautoscalers",
				},
				Verbs: []string{"get", "list", "watch"},
			},
			{
				NonResourceURLs: []string{"/metrics"},
				Verbs:           []string{"get", "list", "watch"},
			},
		},
	}
	crb := manifests.ClusterRoleBinding{
		Namespace:          namespace,
		Name:               "cluster-role-binding",
		ClusterRoleName:    cr.Name,
		ServiceAccountName: serviceAccount,
	}

	return cr, crb
}

func daemonSet(cluster *kubeutils.KindCluster, namespace, serviceAccount, configMap string) manifests.DaemonSet {
	splat := strings.Split(cluster.Testcase.OTLPEndpoint, ":")
	port := splat[len(splat)-1]
	var hostFromContainer string
	if runtime.GOOS == "darwin" {
		hostFromContainer = "host.docker.internal"
	} else {
		hostFromContainer = cluster.GetDefaultGatewayIP()
	}
	otlpEndpoint := fmt.Sprintf("%s:%s", hostFromContainer, port)

	ds := manifests.DaemonSet{
		Name:           "an.agent.daemonset",
		Namespace:      namespace,
		ServiceAccount: serviceAccount,
		Labels:         map[string]string{"label.key": "label.value"},
		Containers: []corev1.Container{
			{
				Image:   testutils.GetCollectorImageOrSkipTest(cluster.Testcase),
				Command: []string{"/otelcol", "--config=/config/config.yaml"},
				Env: []corev1.EnvVar{
					{Name: "OTLP_ENDPOINT", Value: otlpEndpoint},
				},
				Name:            "otel-collector",
				ImagePullPolicy: corev1.PullIfNotPresent,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name: "config-map-volume", MountPath: "/config",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "config-map-volume",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMap,
						},
						Items: []corev1.KeyToPath{
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
	return ds
}
