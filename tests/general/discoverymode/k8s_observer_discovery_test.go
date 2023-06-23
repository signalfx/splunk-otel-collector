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

func TestK8sObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	tc := testutils.NewTestcase(t, testutils.OTLPReceiverSinkAllInterfaces)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cluster := kubeutils.NewKindCluster(tc)
	defer cluster.Delete()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	namespace, serviceAccount := createNamespaceAndServiceAccount(cluster)

	configMap, configMapManifest := configToConfigMapManifest(t, "k8s-otlp-exporter-no-internal-prometheus.yaml", namespace)
	sout, serr, err := cluster.Apply(configMapManifest)
	tc.Logger.Debug("applying ConfigMap", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(t, err)

	redisName, redisUID := createRedis(cluster, "target.redis", namespace, serviceAccount)

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
		rPod, err := cluster.Clientset.CoreV1().Pods(namespace).Get(ctx, redisName, metav1.GetOptions{})
		require.NoError(t, err)
		tc.Logger.Debug(fmt.Sprintf("redis is: %s\n", rPod.Status.Phase))
		return rPod.Status.Phase == corev1.PodRunning
	}, 5*time.Minute, 1*time.Second)

	var collectorPodName string
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
			collectorPodName = cPod.Name
			require.NoError(t, err)
			return cPod.Status.Phase == corev1.PodRunning
		}
		return false
	}, 5*time.Minute, 1*time.Second)

	expectedMetrics := tc.ResourceMetrics("k8s-observer-smart-agent-redis.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedMetrics, 30*time.Second))

	stdout, stderr, err := cluster.Kubectl(
		"exec", "-n", namespace, collectorPodName, "--", "bash", "-c",
		`SPLUNK_DEBUG_CONFIG_SERVER=false \
SPLUNK_DISCOVERY_EXTENSIONS_host_observer_ENABLED=false \
SPLUNK_DISCOVERY_EXTENSIONS_docker_observer_ENABLED=false \
SPLUNK_DISCOVERY_RECEIVERS_smartagent_CONFIG_extraDimensions_x3a__x3a_three_x2e_key='three.value.from.env.var.property' \
/otelcol --config=/config/config.yaml --config-dir=/config.d --discovery --dry-run`)
	require.NoError(t, err)
	require.Equal(t, `exporters:
  otlp:
    endpoint: ${OTLP_ENDPOINT}
    tls:
      insecure: true
extensions:
  k8s_observer:
    auth_type: serviceAccount
processors:
  filter:
    metrics:
      include:
        match_type: strict
        metric_names:
        - gauge.connected_clients
receivers:
  receiver_creator/discovery:
    receivers:
      smartagent:
        config:
          extraDimensions:
            three.key: three.value.from.env.var.property
          type: collectd/redis
        resource_attributes:
          one.key: one.value
          two.key: two.value
        rule: type == "port" && pod.name == "${TARGET_POD_NAME}"
    watch_observers:
    - k8s_observer
service:
  extensions:
  - k8s_observer
  pipelines:
    metrics:
      exporters:
      - otlp
      processors:
      - filter
      receivers:
      - receiver_creator/discovery
  telemetry:
    metrics:
      address: ""
      level: none
`, stdout.String())
	require.Contains(
		t, stderr.String(),
		fmt.Sprintf(`Discovering for next 10s...
Successfully discovered "smartagent" using "k8s_observer" endpoint "k8s_observer/%s/(6379)".
Discovery complete.
`, redisUID),
	)
}

func createRedis(cluster *kubeutils.KindCluster, name, namespace, serviceAccount string) (string, string) {
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
	return redis.Name, string(redis.UID)
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

func configToConfigMapManifest(t testing.TB, cfg, namespace string) (name, manifest string) {
	config, err := os.ReadFile(filepath.Join(".", "testdata", cfg))
	require.NoError(t, err)
	configStore := map[string]any{"config": string(config)}

	k8sObserver, err := os.ReadFile(filepath.Join(".", "testdata", "k8s-observer-config.d", "extensions", "k8s-observer.discovery.yaml"))
	configStore["extensions"] = string(k8sObserver)

	saReceiver, err := os.ReadFile(filepath.Join(".", "testdata", "k8s-observer-config.d", "receivers", "smart-agent-redis.discovery.yaml"))
	configStore["receivers"] = string(saReceiver)

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
				Image: testutils.GetCollectorImageOrSkipTest(cluster.Testcase),
				Command: []string{
					"/otelcol", "--config=/config/config.yaml", "--config-dir=/config.d", "--discovery",
					// TODO update w/ resource_attributes when supported
					"--set", "splunk.discovery.receivers.smartagent.config.extraDimensions::three.key='three.value.from.cmdline.property'",
					"--set", `splunk.discovery.extensions.host_observer.enabled=false`,
					"--set", `splunk.discovery.extensions.docker_observer.enabled=false`,
				},
				Env: []corev1.EnvVar{
					{Name: "OTLP_ENDPOINT", Value: otlpEndpoint},
					{Name: "TARGET_POD_NAME", Value: "target.redis"},
				},
				Name:            "otel-collector",
				ImagePullPolicy: corev1.PullIfNotPresent,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name: "config-map-volume", MountPath: "/config",
					},
					{
						Name: "config-d-volume", MountPath: "/config.d",
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
			{
				Name: "config-d-volume",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMap,
						},
						Items: []corev1.KeyToPath{
							{
								Key:  "extensions",
								Path: "extensions/k8s-observer.discovery.yaml",
							},
							{
								Key:  "receivers",
								Path: "receivers/smart-agent-redis.discovery.yaml",
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
