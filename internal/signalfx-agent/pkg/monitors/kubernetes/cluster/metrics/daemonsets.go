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

func datapointsForDaemonSet(ds *appsv1.DaemonSet) []*datapoint.Datapoint {
	dimensions := map[string]string{
		"metric_source":        "kubernetes",
		"kubernetes_namespace": ds.Namespace,
		"kubernetes_uid":       string(ds.UID),
		"kubernetes_name":      ds.Name,
	}

	return []*datapoint.Datapoint{
		datapoint.New(
			"kubernetes.daemon_set.current_scheduled",
			dimensions,
			datapoint.NewIntValue(int64(ds.Status.CurrentNumberScheduled)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			"kubernetes.daemon_set.desired_scheduled",
			dimensions,
			datapoint.NewIntValue(int64(ds.Status.DesiredNumberScheduled)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			"kubernetes.daemon_set.misscheduled",
			dimensions,
			datapoint.NewIntValue(int64(ds.Status.NumberMisscheduled)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			"kubernetes.daemon_set.ready",
			dimensions,
			datapoint.NewIntValue(int64(ds.Status.NumberReady)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			meta.KubernetesDaemonSetUpdated,
			dimensions,
			datapoint.NewIntValue(int64(ds.Status.UpdatedNumberScheduled)),
			datapoint.Gauge,
			time.Time{}),
	}
}

func dimensionForDaemonSet(ds *appsv1.DaemonSet) *atypes.Dimension {
	props, tags := k8sutil.PropsAndTagsFromLabels(ds.Labels)
	props["kubernetes_workload"] = "DaemonSet"
	props["kubernetes_workload_name"] = ds.Name
	props["daemonset_creation_timestamp"] = ds.GetCreationTimestamp().Format(time.RFC3339)

	for _, or := range ds.OwnerReferences {
		props[utils.LowercaseFirstChar(or.Kind)] = or.Name
		props[utils.LowercaseFirstChar(or.Kind)+"_uid"] = string(or.UID)
	}

	return &atypes.Dimension{
		Name:       "kubernetes_uid",
		Value:      string(ds.UID),
		Properties: props,
		Tags:       tags,
	}
}
