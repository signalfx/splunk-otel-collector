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

//go:build discovery_integration_envoy_k8s

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

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestEnvoyK8sObserver(t *testing.T) {
	f := otlpreceiver.NewFactory()
	port := testutils.GetAvailablePort(t)
	otlpReceiverConfig := f.CreateDefaultConfig().(*otlpreceiver.Config)
	otlpReceiverConfig.GRPC = configoptional.Some(configgrpc.ServerConfig{
		NetAddr: confignet.AddrConfig{
			Endpoint:  fmt.Sprintf("0.0.0.0:%d", port),
			Transport: "tcp",
		},
	})
	otlpReceiverConfig.HTTP = configoptional.None[otlpreceiver.HTTPConfig]()
	sink := &consumertest.MetricsSink{}
	receiver, err := f.CreateMetrics(context.Background(), receivertest.NewNopSettings(f.Type()), otlpReceiverConfig, sink)
	require.NoError(t, err)
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, receiver.Shutdown(context.Background()))
	})

	require.NoError(t, err)
	dockerHost := "172.18.0.1"
	if runtime.GOOS == "darwin" {
		dockerHost = "host.docker.internal"
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	require.NoError(t, err)
	client, err := kubernetes.NewForConfig(kubeConfig)

	decode := scheme.Codecs.UniversalDeserializer().Decode
	// Cluster role, role binding, and service account
	stream, err := os.ReadFile(filepath.Join("testdata", "k8s", "clusterRole.yaml"))
	require.NoError(t, err)
	clusterRole, _, err := decode(stream, nil, nil)
	require.NoError(t, err)
	cr, err := client.RbacV1().ClusterRoles().Create(context.Background(), clusterRole.(*rbacv1.ClusterRole), metav1.CreateOptions{})

	stream, err = os.ReadFile(filepath.Join("testdata", "k8s", "clusterRoleBinding.yaml"))
	require.NoError(t, err)
	clusterRoleBinding, _, err := decode(stream, nil, nil)
	require.NoError(t, err)
	crb, err := client.RbacV1().ClusterRoleBindings().Create(context.Background(), clusterRoleBinding.(*rbacv1.ClusterRoleBinding), metav1.CreateOptions{})
	require.NoError(t, err)

	stream, err = os.ReadFile(filepath.Join("testdata", "k8s", "serviceAccount.yaml"))
	require.NoError(t, err)
	serviceAccount, _, err := decode(stream, nil, nil)
	require.NoError(t, err)
	sa, err := client.CoreV1().ServiceAccounts("default").Create(context.Background(), serviceAccount.(*v1.ServiceAccount), metav1.CreateOptions{})
	require.NoError(t, err)
	// Configmap
	stream, err = os.ReadFile(filepath.Join("testdata", "k8s", "config.yaml"))
	require.NoError(t, err)
	cfgMap, _, err := decode(stream, nil, nil)
	require.NoError(t, err)

	c, err := client.CoreV1().ConfigMaps("default").Create(context.Background(), cfgMap.(*v1.ConfigMap), metav1.CreateOptions{})
	require.NoError(t, err)
	// Collector:
	stream, err = os.ReadFile(filepath.Join("testdata", "k8s", "collector.yaml"))
	require.NoError(t, err)
	streamStr := strings.Replace(string(stream), "$OTLP_ENDPOINT", fmt.Sprintf("%s:%d", dockerHost, port), 1)
	collectorDeployment, _, err := decode([]byte(streamStr), nil, nil)
	require.NoError(t, err)

	d, err := client.AppsV1().Deployments("default").Create(context.Background(), collectorDeployment.(*appsv1.Deployment), metav1.CreateOptions{})
	require.NoError(t, err)

	t.Cleanup(func() {
		if os.Getenv("SKIP_TEARDOWN") == "true" {
			return
		}
		require.NoError(t, client.CoreV1().ConfigMaps("default").Delete(context.Background(), c.Name, metav1.DeleteOptions{}))
		require.NoError(t, client.AppsV1().Deployments("default").Delete(context.Background(), d.Name, metav1.DeleteOptions{}))

		require.NoError(t, client.CoreV1().ServiceAccounts("default").Delete(context.Background(), sa.Name, metav1.DeleteOptions{}))
		require.NoError(t, client.RbacV1().ClusterRoleBindings().Delete(context.Background(), crb.Name, metav1.DeleteOptions{}))
		require.NoError(t, client.RbacV1().ClusterRoles().Delete(context.Background(), cr.Name, metav1.DeleteOptions{}))
	})

	expected, err := golden.ReadMetrics(filepath.Join("testdata", "expected_k8s.yaml"))
	require.NoError(t, err)

	index := 0
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		if len(sink.AllMetrics()) == 0 {
			assert.Fail(tt, "No metrics collected")
			return
		}
		var err error
		newIndex := len(sink.AllMetrics())
		for i := index; i < newIndex; i++ {
			err = pmetrictest.CompareMetrics(expected, sink.AllMetrics()[i],
				pmetrictest.IgnoreResourceAttributeValue("service.instance.id"),
				pmetrictest.IgnoreResourceAttributeValue("net.host.port"),
				pmetrictest.IgnoreResourceAttributeValue("net.host.name"),
				pmetrictest.IgnoreResourceAttributeValue("server.address"),
				pmetrictest.IgnoreResourceAttributeValue("container.name"),
				pmetrictest.IgnoreResourceAttributeValue("server.port"),
				pmetrictest.IgnoreResourceAttributeValue("service.name"),
				pmetrictest.IgnoreResourceAttributeValue("service_instance_id"),
				pmetrictest.IgnoreResourceAttributeValue("service_version"),
				pmetrictest.IgnoreResourceAttributeValue("discovery.endpoint.id"),
				pmetrictest.IgnoreMetricAttributeValue("service_version"),
				pmetrictest.IgnoreMetricAttributeValue("service_instance_id"),
				pmetrictest.IgnoreResourceAttributeValue("server.address"),
				pmetrictest.IgnoreResourceAttributeValue("k8s.pod.name"),
				pmetrictest.IgnoreResourceAttributeValue("k8s.pod.uid"),
				pmetrictest.IgnoreTimestamp(),
				pmetrictest.IgnoreStartTimestamp(),
				pmetrictest.IgnoreMetricDataPointsOrder(),
				pmetrictest.IgnoreScopeMetricsOrder(),
				pmetrictest.IgnoreScopeVersion(),
				pmetrictest.IgnoreResourceMetricsOrder(),
				pmetrictest.IgnoreMetricsOrder(),
				pmetrictest.IgnoreMetricValues(),
			)
			if err == nil {
				return
			}
		}
		index = newIndex
		assert.NoError(tt, err)
	}, 120*time.Second, 1*time.Second)
}
