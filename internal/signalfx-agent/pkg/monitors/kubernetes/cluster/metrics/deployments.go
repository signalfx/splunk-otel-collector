// nolint: dupl
package metrics

import (
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/cluster/meta"
	k8sutil "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/utils"
	atypes "github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
)

func datapointsForDeployment(dep *appsv1.Deployment) []*datapoint.Datapoint {
	dimensions := map[string]string{
		"metric_source":        "kubernetes",
		"kubernetes_namespace": dep.Namespace,
		"kubernetes_uid":       string(dep.UID),
		"kubernetes_name":      dep.Name,
	}

	if dep.Spec.Replicas == nil { // || dep.Status.AvailableReplicas == nil {
		return nil
	}

	dps := makeReplicaDPs("deployment", dimensions,
		*dep.Spec.Replicas, dep.Status.AvailableReplicas)

	dps = append(dps, datapoint.New(
		meta.KubernetesDeploymentUpdated,
		dimensions,
		datapoint.NewIntValue(int64(dep.Status.UpdatedReplicas)),
		datapoint.Gauge,
		time.Time{}),
	)

	return dps
}

func dimensionForDeployment(dep *appsv1.Deployment) *atypes.Dimension {
	props, tags := k8sutil.PropsAndTagsFromLabels(dep.Labels)
	props["kubernetes_workload_name"] = dep.Name
	props["deployment"] = dep.Name
	props["kubernetes_workload"] = "Deployment"
	props["deployment_creation_timestamp"] = dep.GetCreationTimestamp().Format(time.RFC3339)

	for _, or := range dep.OwnerReferences {
		props[utils.LowercaseFirstChar(or.Kind)] = or.Name
		props[utils.LowercaseFirstChar(or.Kind)+"_uid"] = string(or.UID)
	}

	return &atypes.Dimension{
		Name:       "kubernetes_uid",
		Value:      string(dep.UID),
		Properties: props,
		Tags:       tags,
	}
}
