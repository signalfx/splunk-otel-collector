// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package kubeutils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kubectlcmd "k8s.io/kubectl/pkg/cmd"
	kubectlcmdutil "k8s.io/kubectl/pkg/cmd/util"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/cmd/kind"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const kindConfigTemplate = `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    image: kindest/node:v1.26.0
    {{- if .ExposedPorts }}
    extraPortMappings:
    {{- range $hostPort, $containerPort := .ExposedPorts }}
      - containerPort: {{ $containerPort }}
        hostPort: {{ $hostPort }}
        listenAddress: "0.0.0.0"
        protocol: tcp
    {{- end }}
    {{- end }}
`

type KindCluster struct {
	Testcase          *testutils.Testcase
	Clientset         *kubernetes.Clientset
	ExposedPorts      map[uint16]uint16
	Name              string
	Kubeconfig        string
	Config            string
	hostFromContainer string
}

func NewKindCluster(t *testutils.Testcase) *KindCluster {
	return &KindCluster{
		Name:         fmt.Sprintf("cluster-%s", t.ID),
		Testcase:     t,
		ExposedPorts: map[uint16]uint16{},
	}
}

func (k *KindCluster) Create() {
	f, err := os.CreateTemp("", fmt.Sprintf("kubeconfig-%s--", k.Name))
	require.NoError(k.Testcase, err)
	k.Kubeconfig = f.Name()

	f, err = os.CreateTemp("", fmt.Sprintf("kindonfig-%s--", k.Name))
	require.NoError(k.Testcase, err)
	_, err = f.WriteString(k.renderConfig())
	require.NoError(k.Testcase, err)
	k.Config = f.Name()

	k.runKindCmd([]string{
		"create", "cluster",
		"--name", k.Name,
		"--kubeconfig", k.Kubeconfig,
		"--config", k.Config,
	})

	restConfig, err := clientcmd.BuildConfigFromFlags("", k.Kubeconfig)
	require.NoError(k.Testcase, err)
	k.Clientset, err = kubernetes.NewForConfig(restConfig)
	require.NoError(k.Testcase, err)

	for _, pod := range []string{
		"kube-apiserver", "kube-controller-manager", "kube-scheduler", "etcd-cluster", "coredns", "kindnet", "kube-proxy",
	} {
		k.WaitForPods(fmt.Sprintf("%s-.*", pod), "kube-system", 2*time.Minute)
	}
	k.WaitForPods("local-path-provisioner-.*", "local-path-storage", 2*time.Minute)
}

func (k *KindCluster) Teardown() {
	defer func() { require.NoError(k.Testcase, os.Remove(k.Kubeconfig)) }()
	defer func() { require.NoError(k.Testcase, os.Remove(k.Config)) }()
	k.runKindCmd([]string{"delete", "cluster", "--name", k.Name})
}

func (k KindCluster) LoadDockerImage(image string) {
	k.runKindCmd([]string{"--name", k.Name, "load", "docker-image", image})
}

func (k KindCluster) LoadLocalCollectorImageIfNecessary() {
	collectorImage := testutils.GetCollectorImageOrSkipTest(k.Testcase)
	// currently doing a simple repo name check, since remotes are
	// unavailable to load. If not robust enough, adopt something from
	// "github.com/docker/distribution/reference" or similar
	if !strings.Contains(collectorImage, "/") {
		k.LoadDockerImage(collectorImage)
	}
}

func (k KindCluster) GetDefaultGatewayIP() string {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	client.NegotiateAPIVersion(context.Background())
	require.NoError(k.Testcase, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	network, err := client.NetworkInspect(ctx, "kind", types.NetworkInspectOptions{})
	require.NoError(k.Testcase, err)
	for _, ipam := range network.IPAM.Config {
		return ipam.Gateway
	}
	k.Testcase.Fatal("no kind network gateway detected. Host IP is inaccessible.")
	return ""
}

func (k *KindCluster) HostFromContainer() string {
	if k.hostFromContainer != "" {
		return k.hostFromContainer
	}
	if runtime.GOOS == "darwin" {
		k.hostFromContainer = "host.docker.internal"
	} else {
		k.hostFromContainer = k.GetDefaultGatewayIP()
	}
	return k.hostFromContainer
}

func (k KindCluster) OTLPEndointFromContainer() string {
	hostFromContainer := k.HostFromContainer()
	splat := strings.Split(k.Testcase.OTLPEndpoint, ":")
	port := splat[len(splat)-1]
	return fmt.Sprintf("%s:%s", hostFromContainer, port)
}

func (k KindCluster) Kubectl(args ...string) (stdOut, stdErr bytes.Buffer, err error) {
	return k.runKubectl(nil, args...)
}

func (k KindCluster) tmpManifestFile(manifests string) string {
	sha := sha256.Sum256([]byte(manifests))
	f, err := os.CreateTemp("", fmt.Sprintf("manifests-%x", sha[:8]))
	require.NoError(k.Testcase, err)
	n, err := f.Write([]byte(manifests))
	require.NoError(k.Testcase, err)
	require.Equal(k.Testcase, len(manifests), n)
	require.NoError(k.Testcase, f.Sync())
	require.NoError(k.Testcase, f.Close())
	return f.Name()
}

func (k KindCluster) Apply(manifests string) (stdOut, stdErr bytes.Buffer, err error) {
	return k.runKubectl(bytes.NewReader([]byte(manifests)), "apply", "-f", k.tmpManifestFile(manifests))
}

func (k KindCluster) Delete(manifests string) (stdOut, stdErr bytes.Buffer, err error) {
	return k.runKubectl(bytes.NewReader([]byte(manifests)), "delete", "-f", k.tmpManifestFile(manifests))
}

func (k KindCluster) WaitForPods(podNameRegex, namespaceName string, duration time.Duration) {
	namePattern, e := regexp.Compile(podNameRegex)
	require.NoError(k.Testcase, e)
	require.Eventually(k.Testcase, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pods, err := k.Clientset.CoreV1().Pods(namespaceName).List(ctx, metav1.ListOptions{})
		require.NoError(k.Testcase, err)
		cancel()

		matches := map[string]struct{}{}

		for i := range pods.Items {
			pod := pods.Items[i]
			if namePattern.MatchString(pod.Name) {
				matches[pod.Name] = struct{}{}
			}
		}

		if len(matches) == 0 {
			return false
		}

		for name := range matches {
			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			pod, err := k.Clientset.CoreV1().Pods(namespaceName).Get(ctx, name, metav1.GetOptions{})
			require.NoError(k.Testcase, err)
			cancel()
			k.Testcase.Logger.Debug(fmt.Sprintf("%s is: %s", pod.Name, pod.Status.Phase))
			if pod.Status.Phase == corev1.PodRunning {
				delete(matches, name)
			}
		}

		return len(matches) == 0
	}, duration, 1*time.Second)
}

func (k KindCluster) runKubectl(stdin io.Reader, args ...string) (stdOut, stdErr bytes.Buffer, err error) {
	stdOut = bytes.Buffer{}
	stdErr = bytes.Buffer{}
	fullArgs := append([]string{"--kubeconfig", k.Kubeconfig}, args...)
	kubectl := kubectlcmd.NewDefaultKubectlCommandWithArgs(
		kubectlcmd.KubectlOptions{
			Arguments: append([]string{"<ignored-placeholder>"}, args...),
			IOStreams: genericclioptions.IOStreams{In: stdin, Out: &stdOut, ErrOut: &stdErr},
			// don't use default or persist (pin local kubeconfig)
			ConfigFlags: genericclioptions.NewConfigFlags(false),
		},
	)
	kubectl.SetArgs(fullArgs)

	kubectlcmdutil.BehaviorOnFatal(func(msg string, code int) {
		// panic here to prevent swallowing what would have been a fatal error
		panic(fmt.Errorf("os.Exit(%d): %q", code, msg))
	})

	defer func() {
		// recovered from fatal kubectl command
		if e := recover(); e != nil {
			ee, ok := e.(error)
			if !ok {
				ee = fmt.Errorf("%v", e)
			}
			err = multierr.Combine(err, ee)
		}
	}()

	if e := kubectl.Execute(); e != nil {
		err = multierr.Combine(err, e)
	}

	return
}

func (k KindCluster) runKindCmd(args []string) {
	kindCmd := kind.NewCommand(kindcmd.NewLogger(), kindcmd.StandardIOStreams())
	kindCmd.SetArgs(args)
	err := kindCmd.Execute()
	require.NoError(k.Testcase, err, "failed running kind command %v", args)
}

func (k KindCluster) renderConfig() string {
	out := &bytes.Buffer{}
	tpl, err := template.New("").Parse(kindConfigTemplate)
	require.NoError(k.Testcase, err)
	err = tpl.Execute(out, k)
	require.NoError(k.Testcase, err)
	return out.String()
}
