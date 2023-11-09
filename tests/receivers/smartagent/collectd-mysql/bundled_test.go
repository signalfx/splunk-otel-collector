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
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	defer cluster.Teardown()
	cluster.Create()
	cluster.LoadLocalCollectorImageIfNecessary()

	namespace := manifests.Namespace{Name: "test-namespace"}
	serviceAccount := manifests.ServiceAccount{Name: "some.serviceacount", Namespace: namespace.Name}
	configMap := manifests.ConfigMap{
		Name: "collector.config", Namespace: namespace.Name,
		Data: `config: |
  exporters:
    otlp:
      endpoint: ${OTLP_ENDPOINT}
      tls:
        insecure: true
  service:
    pipelines:
      metrics:
        exporters:
          - otlp
`}

	sout, serr, err := cluster.Apply(manifests.RenderAll(t, namespace, serviceAccount, configMap))
	require.NoError(t, err, "stdout: %s, stderr: %s", sout, serr)

	mysqlUID := cluster.createMySQL("target.mysql", namespace.Name, serviceAccount.Name)

	clusterRole, clusterRoleBinding := clusterRoleAndBinding(namespace.Name, serviceAccount.Name)
	ds := cluster.daemonSet(namespace.Name, serviceAccount.Name, configMap.Name)
	sout, serr, err = cluster.Apply(manifests.RenderAll(t, clusterRole, clusterRoleBinding, ds))
	require.NoError(t, err, "stdout: %s, stderr: %s", sout, serr)

	pods := cluster.WaitForPods(ds.Name, namespace.Name, 5*time.Minute)
	require.Len(t, pods, 1)
	collectorName := pods[0].Name

	expectedMetrics := tc.ResourceMetrics("all.yaml")
	require.NoError(t, tc.OTLPReceiverSink.AssertAllMetricsReceived(t, *expectedMetrics, 30*time.Second))

	stdOut, stdErr, err := cluster.Kubectl("logs", "-n", namespace.Name, collectorName)
	require.NoError(t, err, stdErr.String())
	require.Contains(
		t, stdOut.String(),
		fmt.Sprintf(`Successfully discovered "smartagent/collectd/mysql" using "k8s_observer" endpoint "k8s_observer/%s/(3306)`, mysqlUID),
	)
}

type testCluster struct{ *kubeutils.KindCluster }

// uses *error for deferred require error content
// so that %v verbs don't result in rendered addresses
type printableError struct {
	err *error
}

func (p printableError) String() string {
	if p.err != nil {
		return fmt.Sprintf("%v", *p.err)
	}
	return "<nil>"
}

func (cluster testCluster) createMySQL(name, namespace, serviceAccount string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mysql, err := cluster.Clientset.CoreV1().Pods(namespace).Create(
		ctx, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image:           "mysql:latest",
						Name:            "mysql-server",
						Ports:           []corev1.ContainerPort{{ContainerPort: 3306}},
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{
							{Name: "MYSQL_DATABASE", Value: dbName},
							{Name: "MYSQL_USER", Value: user},
							{Name: "MYSQL_PASSWORD", Value: password},
							{Name: "MYSQL_ROOT_PASSWORD", Value: password},
						},
					},
				},
				ServiceAccountName: serviceAccount,
				HostNetwork:        true,
			},
		}, metav1.CreateOptions{},
	)
	require.NoError(cluster.Testcase, err)

	cluster.WaitForPods(mysql.Name, namespace, 5*time.Minute)

	require.Eventually(cluster.Testcase, func() bool {
		stdOut, _, err := cluster.Kubectl("logs", "-n", namespace, mysql.Name)
		require.NoError(cluster.Testcase, err)
		return strings.Contains(stdOut.String(), "port: 3306  MySQL Community Server")
	}, time.Minute, time.Second)

	stdOut, stdErr := new(bytes.Buffer), new(bytes.Buffer)
	pe := &printableError{err: &err}
	for _, u := range []string{"root", user} {
		require.Eventually(cluster.Testcase, func() bool {
			*stdOut, *stdErr, *pe.err = cluster.Kubectl(
				"exec", "-n", namespace, mysql.Name, "--",
				"mysql", "--protocol=tcp", "-h127.0.0.1", "-p3306",
				fmt.Sprintf("-u%s", u), "-ptestpass", "-e", "show status",
			)
			return *pe.err == nil
		}, 30*time.Second, 5*time.Second, "db check for %s: %q, %q, %q", user, stdOut, stdErr, pe)
	}

	for _, cmd := range [][]string{
		{"-uroot", "-ptestpass", "-e", "grant PROCESS on *.* TO 'testuser'@'%'; flush privileges;"},
		{"-utestuser", "-ptestpass", "-Dtestdb", "-e", "CREATE TABLE a_table (name VARCHAR(255), preference VARCHAR(255))"},
		{"-utestuser", "-ptestpass", "-Dtestdb", "-e", "ALTER TABLE a_table ADD COLUMN id INT AUTO_INCREMENT PRIMARY KEY"},
		{"-utestuser", "-ptestpass", "-Dtestdb", "-e", "INSERT INTO a_table (name, preference) VALUES ('some.name', 'some preference')"},
		{"-utestuser", "-ptestpass", "-Dtestdb", "-e", "INSERT INTO a_table (name, preference) VALUES ('another.name', 'another preference');"},
		{"-utestuser", "-ptestpass", "-Dtestdb", "-e", "UPDATE a_table SET preference = 'the real preference' WHERE name = 'some.name'"},
		{"-utestuser", "-ptestpass", "-Dtestdb", "-e", "SELECT * FROM a_table"},
		{"-utestuser", "-ptestpass", "-Dtestdb", "-e", "DELETE FROM a_table WHERE name = 'another.name'"},
	} {
		args := append([]string{
			"exec", "-n", namespace, mysql.Name, "--",
			"mysql", "--protocol=tcp", "-h127.0.0.1", "-p3306",
		}, cmd...)
		stdout, stderr, e := cluster.Kubectl(args...)
		require.NoError(cluster.Testcase, e, fmt.Sprintf("stdout: %q, stderr: %q", stdout, stderr))
	}

	return string(mysql.UID)
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

func (cluster testCluster) daemonSet(namespace, serviceAccount, configMap string) manifests.DaemonSet {
	splat := strings.Split(cluster.Testcase.OTLPEndpoint, ":")
	port := splat[len(splat)-1]
	var hostFromContainer string
	if runtime.GOOS == "darwin" {
		hostFromContainer = "host.docker.internal"
	} else {
		hostFromContainer = cluster.GetDefaultGatewayIP()
	}
	otlpEndpoint := fmt.Sprintf("%s:%s", hostFromContainer, port)

	return manifests.DaemonSet{
		Name:           "an.agent.daemonset",
		Namespace:      namespace,
		ServiceAccount: serviceAccount,
		Labels:         map[string]string{"label.key": "label.value"},
		Containers: []corev1.Container{
			{
				Image: testutils.GetCollectorImageOrSkipTest(cluster.Testcase),
				Command: []string{
					"/otelcol", "--config=/config/config.yaml", "--discovery",
					"--set", "splunk.discovery.receivers.smartagent/collectd/mysql.config.username='${MYSQL_USERNAME}'",
					"--set", "splunk.discovery.receivers.smartagent/collectd/mysql.config.password='${MYSQL_PASSWORD}'",
					"--set", "splunk.discovery.receivers.smartagent/collectd/mysql.config.databases=[{name: 'testdb'}]",
					"--set", "splunk.discovery.receivers.smartagent/collectd/mysql.config.innodbStats=true",
					"--set", `splunk.discovery.receivers.smartagent/collectd/mysql.config.extraMetrics=["*"]`,
				},
				Env: []corev1.EnvVar{
					{Name: "OTLP_ENDPOINT", Value: otlpEndpoint},
					{Name: "MYSQL_USERNAME", Value: user},
					{Name: "MYSQL_PASSWORD", Value: password},
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
}
