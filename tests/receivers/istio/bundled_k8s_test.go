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

//go:build discovery_integration_istio_k8s

package tests

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/signalfx/splunk-otel-collector/tests/internal/discoverytest"
)

const (
	istioVersion = "1.24.2"
)

func downloadIstio(t *testing.T, version string) (string, string) {
	var url string
	if runtime.GOOS == "darwin" {
		url = fmt.Sprintf("https://github.com/istio/istio/releases/download/%s/istio-%s-osx.tar.gz", version, version)
	} else if runtime.GOOS == "linux" {
		url = fmt.Sprintf("https://github.com/istio/istio/releases/download/%s/istio-%s-linux-amd64.tar.gz", version, version)
	} else {
		t.Fatalf("unsupported operating system: %s", runtime.GOOS)
	}

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	gz, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)
	defer gz.Close()

	tr := tar.NewReader(gz)
	var istioDir string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		target := filepath.Join(".", hdr.Name)
		if hdr.FileInfo().IsDir() && istioDir == "" {
			istioDir = target
		}
		if hdr.FileInfo().IsDir() {
			require.NoError(t, os.MkdirAll(target, hdr.FileInfo().Mode()))
		} else {
			f, err := os.Create(target)
			require.NoError(t, err)
			defer f.Close()

			_, err = io.Copy(f, tr)
			require.NoError(t, err)
		}
	}
	require.NotEmpty(t, istioDir, "istioctl path not found")

	absIstioDir, err := filepath.Abs(istioDir)
	require.NoError(t, err, "failed to get absolute path for istioDir")

	istioctlPath := filepath.Join(absIstioDir, "bin", "istioctl")
	require.FileExists(t, istioctlPath, "istioctl binary not found")
	require.NoError(t, os.Chmod(istioctlPath, 0o755), "failed to set executable permission for istioctl")

	t.Cleanup(func() {
		os.RemoveAll(absIstioDir)
	})

	return absIstioDir, istioctlPath
}

func TestIstioEntities(t *testing.T) {
	kubeCfg := os.Getenv("KUBECONFIG")
	if kubeCfg == "" {
		t.Fatal("KUBECONFIG environment variable not set")
	}

	skipTearDown := false
	if os.Getenv("SKIP_TEARDOWN") == "true" {
		skipTearDown = true
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeCfg)
	require.NoError(t, err)

	clientset, err := kubernetes.NewForConfig(config)
	require.NoError(t, err)

	runCommand(t, "kubectl label ns default istio-injection=enabled")

	istioDir, istioctlPath := downloadIstio(t, istioVersion)
	runCommand(t, fmt.Sprintf("%s install -y", istioctlPath))

	t.Cleanup(func() {
		if skipTearDown {
			t.Log("Skipping teardown as SKIP_TEARDOWN is set to true")
			return
		}
		runCommand(t, fmt.Sprintf("%s uninstall --purge -y", istioctlPath))
	})

	// Patch ingress gateway to work in kind cluster
	patchResource(t, clientset, "istio-system", "istio-ingressgateway", "deployments", `{"spec":{"template":{"spec":{"containers":[{"name":"istio-proxy","ports":[{"containerPort":8080,"hostPort":80},{"containerPort":8443,"hostPort":443}]}]}}}}`)
	patchResource(t, clientset, "istio-system", "istio-ingressgateway", "services", `{"spec": {"type": "ClusterIP"}}`)

	demoYaml := filepath.Join(istioDir, "samples", "bookinfo", "platform", "kube", "bookinfo.yaml")
	runCommand(t, fmt.Sprintf("kubectl apply -f %s", demoYaml))
	t.Cleanup(func() {
		if skipTearDown {
			return
		}
		runCommand(t, fmt.Sprintf("kubectl delete -f %s", demoYaml))
	})

	waitForPodsReady(t, clientset, "")

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	require.NoError(t, err)

	expectedLogs := []map[string]string{}
	for _, pod := range pods.Items {
		message := "istio prometheus receiver is working for istio-proxy!"
		if pod.Labels["app"] == "istiod" {
			message = "istio prometheus receiver is working for istiod!"
		}
		if pod.Labels["app"] == "istiod" || pod.Labels["istio"] == "ingressgateway" || hasIstioProxyContainer(pod) {
			t.Logf("Matching pod: %s", pod.Name)
			podIP := pod.Status.PodIP
			promPort := "15090"
			if port, ok := pod.Annotations["prometheus.io/port"]; ok {
				promPort = port
			}
			expectedLogs = append(expectedLogs, map[string]string{
				"discovery.receiver.name": "istio",
				"discovery.receiver.type": "prometheus",
				"k8s.pod.name":            pod.Name,
				"k8s.namespace.name":      pod.Namespace,
				"discovery.message":       message,
				"discovery.observer.type": "k8s_observer",
				"discovery.status":        "successful",
				"endpoint":                podIP,
				"server.address":          podIP,
				"server.port":             promPort,
				"service.instance.id":     fmt.Sprintf("%s:%s", podIP, promPort),
				"type":                    "pod",
				"url.scheme":              "http",
			})
		}
	}
	discoverytest.RunWithK8s(t, expectedLogs, []string{"splunk.discovery.receivers.prometheus/istio.enabled=true"})
}

func hasIstioProxyContainer(pod corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == "istio-proxy" {
			return true
		}
	}
	return false
}

func runCommand(t *testing.T, command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run(), "failed to run command: %s", command)
}

func patchResource(t *testing.T, clientset *kubernetes.Clientset, namespace, name, resourceType, patch string) {
	var err error
	switch resourceType {
	case "deployments":
		_, err = clientset.AppsV1().Deployments(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
		require.NoError(t, err)
		waitForDeploymentRollout(t, clientset, namespace, name)
	case "services":
		_, err = clientset.CoreV1().Services(namespace).Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	}
	require.NoError(t, err)
}

func waitForDeploymentRollout(t *testing.T, clientset *kubernetes.Clientset, namespace, name string) {
	err := wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas &&
			deployment.Status.Replicas == *deployment.Spec.Replicas &&
			deployment.Status.AvailableReplicas == *deployment.Spec.Replicas &&
			deployment.Status.ObservedGeneration >= deployment.Generation {
			return true, nil
		}
		return false, nil
	})
	require.NoError(t, err, "Deployment %s in namespace %s did not roll out successfully", name, namespace)
}

func waitForPodsReady(t *testing.T, clientset *kubernetes.Clientset, namespace string) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	require.NoError(t, err)

	for _, pod := range pods.Items {
		if pod.DeletionTimestamp != nil {
			continue
		}

		err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
			p, err := clientset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			for _, cond := range p.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					return true, nil
				}
			}
			return false, nil
		})
		if k8serrors.IsNotFound(err) {
			continue
		}
		require.NoError(t, err, "Pod %s in namespace %s is not ready", pod.Name, pod.Namespace)
	}
}
