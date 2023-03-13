package hostid

import (
	"context"
	"os"

	"github.com/signalfx/signalfx-agent/pkg/core/common/kubernetes"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/cluster"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesNodeUID returns the UID of the current K8s node, if running
// on K8s and if the appropriate envvar (MY_NODE_NAME) has been injected in the
// agent pod by the downward API mechanism.  This requires being able to do a
// request to the K8s API server to get the UID from the node name, since UID
// is not available locally to the agent.
func KubernetesNodeUID(monitorConf []config.MonitorConfig) string {
	nodeName := os.Getenv("MY_NODE_NAME")
	if nodeName == "" || os.Getenv("SKIP_KUBERNETES_NODE_UID_DIM") != "" {
		return ""
	}

	var clusterConf cluster.Config

	// Try and extract K8s API config from the `kubernetes-cluster` monitor if
	// available.  Otherwise use the standard service account method.
	for i := range monitorConf {
		conf := monitorConf[i]
		if conf.Type != "kubernetes-cluster" {
			continue
		}

		err := config.DecodeExtraConfig(&conf, &clusterConf, false)
		if err != nil {
			logrus.Errorf("Could not decode Kubernetes API config from kubernetes-cluster monitor: %s", err)
		}
		break
	}

	// This will use default in cluster config if clusterConf.KubernetesAPI is
	// nil.
	client, err := kubernetes.MakeClient(clusterConf.KubernetesAPI)
	if err != nil {
		logrus.Errorf("Could not create Kubernetes client to get node UID: %s", err)
		os.Exit(10)
	}
	node, err := client.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("Could not fetch node with name %s from K8s API: %v", nodeName, err)
		os.Exit(11)
	}

	return string(node.UID)
}
