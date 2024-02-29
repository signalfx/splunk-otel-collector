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

func TestK8sObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	tc := testutils.NewTestcase(t, testutils.OTLPReceiverSinkAllInterfaces)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cluster := kubeutils.NewKindCluster(tc)
	defer cluster.Teardown()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	namespace := manifests.Namespace{Name: "test-namespace"}
	serviceAccount := manifests.ServiceAccount{Name: "some.serviceaccount", Namespace: "test-namespace"}
	configMap := manifests.ConfigMap{
		Name: "collector.config", Namespace: namespace.Name,
		Data: configMapData(t, "k8s-otlp-exporter-no-internal-prometheus.yaml"),
	}
	sout, serr, err := cluster.Apply(manifests.RenderAll(t, namespace, serviceAccount, configMap))
	require.NoError(t, err, "stdout: %s, stderr: %s", sout, serr)

	redisName, redisUID := createRedis(cluster, "target.redis", namespace.Name, serviceAccount.Name)

	clusterRole, clusterRoleBinding := clusterRoleAndBinding(namespace.Name, serviceAccount.Name)
	sout, serr, err = cluster.Apply(manifests.RenderAll(t, clusterRole, clusterRoleBinding))
	require.NoError(t, err, "stdout: %s, stderr: %s", sout, serr)

	ds := daemonSet(cluster, namespace.Name, serviceAccount.Name, configMap.Name)
	sout, serr, err = cluster.Apply(ds.Render(t))
	require.NoError(t, err, "stdout: %s, stderr: %s", sout, serr)

	cluster.WaitForPods(redisName, namespace.Name, 5*time.Minute)

	pods := cluster.WaitForPods(ds.Name, namespace.Name, 5*time.Minute)
	require.Len(t, pods, 1)
	collectorPodName := pods[0].Name

	expectedMetrics := tc.ResourceMetrics("k8s-observer-smart-agent-redis.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedMetrics, 30*time.Second))

	stdout, stderr, err := cluster.Kubectl(
		"exec", "-n", namespace.Name, collectorPodName, "--", "bash", "-c",
		`SPLUNK_DEBUG_CONFIG_SERVER=false \
SPLUNK_DISCOVERY_EXTENSIONS_host_observer_ENABLED=false \
SPLUNK_DISCOVERY_EXTENSIONS_docker_observer_ENABLED=false \
SPLUNK_DISCOVERY_RECEIVERS_redis_ENABLED=false \
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
    resource:
      splunk_autodiscovery: "true"
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

func configMapData(t testing.TB, cfg string) string {
	config, err := os.ReadFile(filepath.Join(".", "testdata", cfg))
	require.NoError(t, err)
	configStore := map[string]any{"config": string(config)}

	k8sObserver, err := os.ReadFile(filepath.Join(".", "testdata", "k8s_observer-config.d", "extensions", "k8s_observer.discovery.yaml"))
	configStore["extensions"] = string(k8sObserver)

	saReceiver, err := os.ReadFile(filepath.Join(".", "testdata", "k8s_observer-config.d", "receivers", "smart_agent_redis.discovery.yaml"))
	configStore["receivers"] = string(saReceiver)

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
								Path: "extensions/k8s_observer.discovery.yaml",
							},
							{
								Key:  "receivers",
								Path: "receivers/smart_agent_redis.discovery.yaml",
							},
						},
					},
				},
			},
		},
	}
	return ds
}
