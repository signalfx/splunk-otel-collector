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
	mysqlUID := cluster.createMySQL("target.mysql", namespace, serviceAccount)
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

	require.Eventually(cluster.Testcase, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rPod, err := cluster.Clientset.CoreV1().Pods(namespace).Get(ctx, mysql.Name, metav1.GetOptions{})
		require.NoError(cluster.Testcase, err)
		cluster.Testcase.Logger.Debug(fmt.Sprintf("mysql is: %s\n", rPod.Status.Phase))
		return rPod.Status.Phase == corev1.PodRunning
	}, 5*time.Minute, 1*time.Second)

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
	cmm, err := cm.Render()
	require.NoError(cluster.Testcase, err)

	sout, serr, err := cluster.Apply(cmm)
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
	crManifest, err := cr.Render()
	require.NoError(cluster.Testcase, err)
	sout, serr, err := cluster.Apply(crManifest)
	cluster.Testcase.Logger.Debug("applying ClusterRole", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(cluster.Testcase, err)

	crb := manifests.ClusterRoleBinding{
		Namespace:          namespace,
		Name:               "cluster-role-binding",
		ClusterRoleName:    cr.Name,
		ServiceAccountName: serviceAccount,
	}
	crbManifest, err := crb.Render()
	require.NoError(cluster.Testcase, err)

	sout, serr, err = cluster.Apply(crbManifest)
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
	dsm, err := ds.Render()
	require.NoError(cluster.Testcase, err)
	sout, serr, err := cluster.Apply(dsm)
	cluster.Testcase.Logger.Debug("applying DaemonSet", zap.String("stdout", sout.String()), zap.String("stderr", serr.String()))
	require.NoError(cluster.Testcase, err)
	return ds.Name
}
