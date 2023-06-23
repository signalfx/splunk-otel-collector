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
	"go.uber.org/zap"
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
	defer cluster.Delete()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	namespace, serviceAccount := createNamespaceAndServiceAccount(cluster)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_, err := cluster.Clientset.CoreV1().Nodes().Create(ctx, &corev1.Node{
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

	_ = createRedis(cluster, "some.pod", namespace, serviceAccount)

	configMap, configMapManifest := configToConfigMapManifest(t, "k8s_observer_endpoints_config.yaml", namespace)
	sout, serr, err := cluster.Apply(configMapManifest)
	tc.Logger.Debug("applying ConfigMap", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(t, err)

	crManifest, crbManifest := clusterRoleAndBindingManifests(t, namespace, serviceAccount)
	sout, serr, err = cluster.Apply(crManifest)
	tc.Logger.Debug("applying ClusterRole", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(t, err)

	sout, serr, err = cluster.Apply(crbManifest)
	tc.Logger.Debug("applying ClusterRoleBinding", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(t, err)

	daemonSet, dsManifest := daemonSetManifest(cluster, namespace, serviceAccount, configMap)
	sout, serr, err = cluster.Apply(dsManifest)
	tc.Logger.Debug("applying DaemonSet", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		dsPods, err := cluster.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("name = %s", daemonSet),
		})
		require.NoError(t, err)
		if len(dsPods.Items) > 0 {
			collectorPod := dsPods.Items[0]
			tc.Logger.Debug(fmt.Sprintf("collector is: %s\n", collectorPod.Status.Phase))
			cPod, err := cluster.Clientset.CoreV1().Pods(collectorPod.Namespace).Get(ctx, collectorPod.Name, metav1.GetOptions{})
			require.NoError(t, err)
			return cPod.Status.Phase == corev1.PodRunning
		}
		return false
	}, 5*time.Minute, 1*time.Second)

	expectedResourceLogs := tc.ResourceLogs("k8s_observer_endpoints.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllLogsReceived(t, *expectedResourceLogs, 30*time.Second))
}

func TestDiscoveryReceiverWithK8sObserverAndSmartAgentRedisReceiverProvideStatusLogs(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	tc := testutils.NewTestcase(t, testutils.OTLPReceiverSinkAllInterfaces)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cluster := kubeutils.NewKindCluster(tc)
	defer cluster.Delete()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	namespace, serviceAccount := createNamespaceAndServiceAccount(cluster)

	configMap, configMapManifest := configToConfigMapManifest(t, "k8s_observer_smart_agent_redis_config.yaml", namespace)
	sout, serr, err := cluster.Apply(configMapManifest)
	tc.Logger.Debug("applying ConfigMap", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(t, err)

	redis := createRedis(cluster, "target.redis", namespace, serviceAccount)

	crManifest, crbManifest := clusterRoleAndBindingManifests(t, namespace, serviceAccount)
	sout, serr, err = cluster.Apply(crManifest)
	tc.Logger.Debug("applying ClusterRole", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(t, err)

	sout, serr, err = cluster.Apply(crbManifest)
	tc.Logger.Debug("applying ClusterRoleBinding", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(t, err)

	daemonSet, dsManifest := daemonSetManifest(cluster, namespace, serviceAccount, configMap)
	sout, serr, err = cluster.Apply(dsManifest)
	tc.Logger.Debug("applying DaemonSet", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rPod, err := cluster.Clientset.CoreV1().Pods(namespace).Get(ctx, redis, metav1.GetOptions{})
		require.NoError(t, err)
		tc.Logger.Debug(fmt.Sprintf("redis is: %s\n", rPod.Status.Phase))
		return rPod.Status.Phase == corev1.PodRunning
	}, 5*time.Minute, 1*time.Second)

	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		dsPods, err := cluster.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("name = %s", daemonSet),
		})
		require.NoError(t, err)
		if len(dsPods.Items) > 0 {
			collectorPod := dsPods.Items[0]
			tc.Logger.Debug(fmt.Sprintf("collector is: %s\n", collectorPod.Status.Phase))
			cPod, err := cluster.Clientset.CoreV1().Pods(collectorPod.Namespace).Get(ctx, collectorPod.Name, metav1.GetOptions{})
			require.NoError(t, err)
			return cPod.Status.Phase == corev1.PodRunning
		}
		return false
	}, 5*time.Minute, 1*time.Second)

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

func createNamespaceAndServiceAccount(cluster *kubeutils.KindCluster) (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ns, err := cluster.Clientset.CoreV1().Namespaces().Create(
		ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}},
		metav1.CreateOptions{},
	)
	require.NoError(cluster.Testcase, err)
	namespace := ns.Name

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	serviceAccount, err := cluster.Clientset.CoreV1().ServiceAccounts(namespace).Create(
		ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "some.serviceaccount"},
		},
		metav1.CreateOptions{})
	require.NoError(cluster.Testcase, err)
	return namespace, serviceAccount.Name
}

func configToConfigMapManifest(t testing.TB, configPath, namespace string) (name, manifest string) {
	config, err := os.ReadFile(filepath.Join(".", "testdata", configPath))
	configStore := map[string]any{"config": string(config)}
	configYaml, err := yaml.Marshal(configStore)
	require.NoError(t, err)
	cm := manifests.ConfigMap{
		Name: "collector.config", Namespace: namespace,
		Data: string(configYaml),
	}
	cmm, err := cm.Render()
	require.NoError(t, err)
	return cm.Name, cmm
}

func clusterRoleAndBindingManifests(t testing.TB, namespace, serviceAccount string) (string, string) {
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
	crManifest, err := cr.Render()
	require.NoError(t, err)

	crb := manifests.ClusterRoleBinding{
		Namespace:          namespace,
		Name:               "cluster-role-binding",
		ClusterRoleName:    cr.Name,
		ServiceAccountName: serviceAccount,
	}
	crbManifest, err := crb.Render()
	require.NoError(t, err)

	return crManifest, crbManifest
}

func daemonSetManifest(cluster *kubeutils.KindCluster, namespace, serviceAccount, configMap string) (name, manifest string) {
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
	dsm, err := ds.Render()
	require.NoError(cluster.Testcase, err)
	return ds.Name, dsm
}
