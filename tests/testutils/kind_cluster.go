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

package testutils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/cmd/kind"
)

type KindCluster struct {
	Testcase   *Testcase
	Clientset  *kubernetes.Clientset
	Name       string
	Kubeconfig string
}

func (k KindCluster) Delete() {
	defer func() { require.NoError(k.Testcase, os.Remove(k.Kubeconfig)) }()
	runKindCmd(k.Testcase, []string{"delete", "cluster", "--name", k.Name})
}

func CreateKindCluster(t *Testcase) KindCluster {
	clusterName := fmt.Sprintf("cluster-%s", t.ID)
	f, err := os.CreateTemp("", fmt.Sprintf("kubeconfig-%s--", clusterName))
	require.NoError(t, err)
	kubeconfig := f.Name()
	runKindCmd(t, []string{
		"create", "cluster",
		"--name", clusterName,
		"--kubeconfig", kubeconfig,
	})

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	require.NoError(t, err)
	clientset, err := kubernetes.NewForConfig(config)
	require.NoError(t, err)

	return KindCluster{
		Name:       clusterName,
		Kubeconfig: kubeconfig,
		Clientset:  clientset,
		Testcase:   t,
	}
}

func runKindCmd(t testing.TB, args []string) {
	kindCmd := kind.NewCommand(cmd.NewLogger(), cmd.StandardIOStreams())
	kindCmd.SetArgs(args)
	err := kindCmd.Execute()
	require.NoError(t, err, "failed running kind command %v", args)
}
