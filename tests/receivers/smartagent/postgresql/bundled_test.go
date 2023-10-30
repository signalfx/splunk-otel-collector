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
	"gopkg.in/yaml.v3"
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

	cluster := testCluster{kubeutils.NewKindCluster(tc)}
	defer cluster.Delete()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	namespace, serviceAccount := cluster.createNamespaceAndServiceAccount()
	postgresUID := cluster.createPostgres("target.postgres", namespace, serviceAccount)
	cluster.createClusterRoleAndRoleBinding(namespace, serviceAccount)
	configMap := cluster.createConfigMap(namespace)
	daemonSet := cluster.daemonSetManifest(namespace, serviceAccount, configMap)

	var collectorName string
	// wait for collector to run
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
			collectorName = cPod.Name
			return cPod.Status.Phase == corev1.PodRunning
		}
		return false
	}, 5*time.Minute, 1*time.Second)

	expectedMetrics := tc.ResourceMetrics("all.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedMetrics, 30*time.Second))

	stdOut, stdErr, err := cluster.Kubectl("logs", "-n", namespace, collectorName)
	require.NoError(t, err)
	require.Contains(
		t, stdOut.String(),
		fmt.Sprintf(`Successfully discovered "smartagent/postgresql" using "k8s_observer" endpoint "k8s_observer/%s/(5432)`, postgresUID),
		stdErr.String(),
	)
}

type testCluster struct{ *kubeutils.KindCluster }

func (cluster testCluster) createPostgres(name, namespace, serviceAccount string) string {
	dbsql, err := os.ReadFile(filepath.Join(".", "testdata", "server", "initdb.d", "db.sql"))
	require.NoError(cluster.Testcase, err)
	cmContent := map[string]any{"db.sql": string(dbsql)}

	initsh, err := os.ReadFile(filepath.Join(".", "testdata", "server", "initdb.d", "init.sh"))
	require.NoError(cluster.Testcase, err)
	cmContent["init.sh"] = string(initsh)

	requests, err := os.ReadFile(filepath.Join(".", "testdata", "client", "requests.sh"))
	require.NoError(cluster.Testcase, err)
	cmContent["requests.sh"] = string(requests)

	configMapContent, err := yaml.Marshal(cmContent)
	require.NoError(cluster.Testcase, err)

	cm := manifests.ConfigMap{
		Namespace: namespace,
		Name:      "postgres",
		Data:      string(configMapContent),
	}
	sout, serr, err := cluster.Apply(cm.Render(cluster.Testcase))
	cluster.Testcase.Logger.Debug("applying ConfigMap", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(cluster.Testcase, err)

	fileMode := int32(0777)
	postgresID := int64(70)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	postgres, err := cluster.Clientset.CoreV1().Pods(namespace).Create(
		ctx, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image:           "postgres:13-alpine",
						Name:            "postgres-server",
						Ports:           []corev1.ContainerPort{{ContainerPort: 5432}},
						ImagePullPolicy: corev1.PullIfNotPresent,
						SecurityContext: &corev1.SecurityContext{
							RunAsUser:  &postgresID,
							RunAsGroup: &postgresID,
						},
						Command: []string{
							"docker-entrypoint.sh",
							"-c", "shared_preload_libraries=pg_stat_statements",
							"-c", "wal_level=logical",
							"-c", "max_replication_slots=2",
						},
						Env: []corev1.EnvVar{
							{
								Name:  "POSTGRES_DB",
								Value: "test_db",
							},
							{
								Name:  "POSTGRES_USER",
								Value: "postgres",
							},
							{
								Name:  "POSTGRES_PASSWORD",
								Value: "postgres",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name: "initdb", MountPath: "/docker-entrypoint-initdb.d",
							},
						},
					},
					{
						Image:           "postgres:13-alpine",
						Name:            "postgres-client",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"/opt/requests/requests.sh"},
						Env: []corev1.EnvVar{
							{
								Name:  "POSTGRES_SERVER",
								Value: "localhost",
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name: "requests", MountPath: "/opt/requests",
							},
						},
					},
				},
				ServiceAccountName: serviceAccount,
				HostNetwork:        true,
				Volumes: []corev1.Volume{
					{
						Name: "initdb",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cm.Name,
								},
								Items: []corev1.KeyToPath{
									{
										Key:  "init.sh",
										Path: "init.sh",
										Mode: &fileMode,
									},
									{
										Key:  "db.sql",
										Path: "db.sql",
									},
								},
							},
						},
					},
					{
						Name: "requests",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cm.Name,
								},
								Items: []corev1.KeyToPath{
									{
										Key:  "requests.sh",
										Path: "requests.sh",
										Mode: &fileMode,
									},
								},
							},
						},
					},
				},
			},
		}, metav1.CreateOptions{},
	)
	require.NoError(cluster.Testcase, err)

	require.Eventually(cluster.Testcase, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rPod, err := cluster.Clientset.CoreV1().Pods(namespace).Get(ctx, postgres.Name, metav1.GetOptions{})
		require.NoError(cluster.Testcase, err)
		cluster.Testcase.Logger.Debug(fmt.Sprintf("postgres is: %s\n", rPod.Status.Phase))
		return rPod.Status.Phase == corev1.PodRunning
	}, 5*time.Minute, 1*time.Second)
	return string(postgres.UID)
}

func (cluster testCluster) createNamespaceAndServiceAccount() (string, string) {
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

func (cluster testCluster) createConfigMap(namespace string) string {
	config := map[string]any{
		"config": `exporters:
  otlp:
    endpoint: ${OTLP_ENDPOINT}
    tls:
      insecure: true
service:
  pipelines:
    metrics:
      exporters:
        - otlp
`,
	}

	data, err := yaml.Marshal(config)
	require.NoError(cluster.Testcase, err)

	cm := manifests.ConfigMap{
		Namespace: namespace,
		Name:      "collector.config",
		Data:      string(data),
	}
	sout, serr, err := cluster.Apply(cm.Render(cluster.Testcase))
	cluster.Testcase.Logger.Debug("applying ConfigMap", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(cluster.Testcase, err)
	return cm.Name
}

func (cluster testCluster) createClusterRoleAndRoleBinding(namespace, serviceAccount string) {
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
	sout, serr, err := cluster.Apply(cr.Render(cluster.Testcase))
	cluster.Testcase.Logger.Debug("applying ClusterRole", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(cluster.Testcase, err)

	crb := manifests.ClusterRoleBinding{
		Namespace:          namespace,
		Name:               "cluster-role-binding",
		ClusterRoleName:    cr.Name,
		ServiceAccountName: serviceAccount,
	}
	sout, serr, err = cluster.Apply(crb.Render(cluster.Testcase))
	cluster.Testcase.Logger.Debug("applying ClusterRoleBinding", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(cluster.Testcase, err)
}

func (cluster testCluster) daemonSetManifest(namespace, serviceAccount, configMap string) string {
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
					"/otelcol", "--config=/config/config.yaml", "--discovery",
					"--set", "splunk.discovery.receivers.smartagent/postgresql.config.params::username='${PG_USERNAME}'",
					"--set", "splunk.discovery.receivers.smartagent/postgresql.config.params::password='${PG_PASSWORD}'",
					"--set", "splunk.discovery.receivers.smartagent/postgresql.config.masterDBName=test_db",
					"--set", `splunk.discovery.receivers.smartagent/postgresql.config.extraMetrics=["*"]`,
				},
				Env: []corev1.EnvVar{
					{Name: "PG_USERNAME", Value: "test_user"},
					{Name: "PG_PASSWORD", Value: "test_password"},
					{Name: "OTLP_ENDPOINT", Value: otlpEndpoint},
					// Helpful for debugging
					// {Name: "SPLUNK_DISCOVERY_DURATION", Value: "20s"},
					// {Name: "SPLUNK_DISCOVERY_LOG_LEVEL", Value: "debug"},
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
	sout, serr, err := cluster.Apply(ds.Render(cluster.Testcase))
	cluster.Testcase.Logger.Debug("applying DaemonSet", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(cluster.Testcase, err)
	return ds.Name
}
