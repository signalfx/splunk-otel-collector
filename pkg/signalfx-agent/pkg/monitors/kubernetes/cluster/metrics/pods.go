package metrics

import (
	"regexp"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	k8sutil "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/utils"
	atypes "github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func datapointsForPod(pod *v1.Pod) []*datapoint.Datapoint {
	dimensions := map[string]string{
		"metric_source": "kubernetes",
		// Try and be consistent with other plugin dimensions, despite
		// verbosity
		"kubernetes_namespace": pod.Namespace,
		"kubernetes_pod_uid":   string(pod.UID),
		"kubernetes_pod_name":  pod.Name,
		"kubernetes_node":      pod.Spec.NodeName,
	}

	dps := []*datapoint.Datapoint{
		datapoint.New(
			"kubernetes.pod_phase",
			dimensions,
			datapoint.NewIntValue(phaseToInt(pod.Status.Phase)),
			datapoint.Gauge,
			time.Time{}),
	}

	containersInPodByName := make(map[string]map[string]string)

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.ContainerID == "" {
			continue
		}

		contDims := getAllContainerDimensions(cs.ContainerID, cs.Name, cs.Image, dimensions)
		containersInPodByName[cs.Name] = contDims

		dps = append(dps, datapointsForContainerStatus(cs, contDims)...)
	}

	for _, c := range pod.Spec.Containers {
		contDims := containersInPodByName[c.Name]

		dps = append(dps, datapointsForContainerSpec(c, contDims)...)
	}

	return dps
}

func dimensionForPod(pod *v1.Pod) *atypes.Dimension {
	props, tags := k8sutil.PropsAndTagsFromLabels(pod.Labels)

	props["pod_creation_timestamp"] = pod.CreationTimestamp.Format(time.RFC3339)

	for _, or := range pod.OwnerReferences {
		props[utils.LowercaseFirstChar(or.Kind)] = or.Name
		props[utils.LowercaseFirstChar(or.Kind)+"_uid"] = string(or.UID)

		// defer syncing replicaset and job workload properties
		// to handleAddPod, handleAddReplicaSet, handleAddJob
		if or.Kind == "ReplicaSet" || or.Kind == "Job" {
			continue
		}
		props["kubernetes_workload"] = or.Kind
		props["kubernetes_workload_name"] = or.Name
	}

	_ = getPropsFromTolerations(pod.Spec.Tolerations)

	return &atypes.Dimension{
		Name:              "kubernetes_pod_uid",
		Value:             string(pod.UID),
		Properties:        props,
		Tags:              tags,
		MergeIntoExisting: true,
	}
}

func dimensionForPodWorkload(pod *v1.Pod, workloadName string, workloadType string) *atypes.Dimension {
	return &atypes.Dimension{
		Name:  "kubernetes_pod_uid",
		Value: string(pod.UID),
		Properties: map[string]string{
			"kubernetes_workload":      workloadType,
			"kubernetes_workload_name": workloadName,
		},
		MergeIntoExisting: true,
	}
}

func dimensionForPodServices(pod *v1.Pod, serviceNames []string, isAdd bool) *atypes.Dimension {
	dim := &atypes.Dimension{
		Name:              "kubernetes_pod_uid",
		Value:             string(pod.UID),
		Tags:              map[string]bool{},
		MergeIntoExisting: true,
	}

	for _, srv := range serviceNames {
		dim.Tags["kubernetes_service_"+srv] = isAdd
	}
	return dim
}

func dimensionForPodDeployment(pod *v1.Pod, deploymentName string, deploymentUID types.UID) *atypes.Dimension {
	return &atypes.Dimension{
		Name:  "kubernetes_pod_uid",
		Value: string(pod.UID),
		Properties: map[string]string{
			"kubernetes_workload":      "Deployment",
			"kubernetes_workload_name": deploymentName,
			"deployment":               deploymentName,
			"deployment_uid":           string(deploymentUID),
		},
		MergeIntoExisting: true,
	}
}

func dimensionForPodCronJob(pod *v1.Pod, cronJobName string, cronJobUID types.UID) *atypes.Dimension {
	return &atypes.Dimension{
		Name:  "kubernetes_pod_uid",
		Value: string(pod.UID),
		Properties: map[string]string{
			"kubernetes_workload":      "CronJob",
			"kubernetes_workload_name": cronJobName,
			"cronJob":                  cronJobName,
			"cronJob_uid":              string(cronJobUID),
		},
		MergeIntoExisting: true,
	}
}

func getPropsFromTolerations(tolerations []v1.Toleration) map[string]string {
	unsupportedPattern := regexp.MustCompile("[^a-zA-Z0-9_-]")

	props := make(map[string]string)

	for _, t := range tolerations {
		keyValueCombo := "toleration"
		if len(t.Key) > 0 {
			keyValueCombo += ("_" + t.Key)
		}
		if len(t.Value) > 0 {
			keyValueCombo += ("_" + t.Value)
		}
		keyValueCombo = unsupportedPattern.ReplaceAllString(keyValueCombo, "_")

		if _, exists := props[keyValueCombo]; exists {
			props[keyValueCombo] += ("," + string(t.Effect))
		} else {
			props[keyValueCombo] = string(t.Effect)
		}
	}

	return props
}

func phaseToInt(phase v1.PodPhase) int64 {
	switch phase {
	case v1.PodPending:
		return 1
	case v1.PodRunning:
		return 2
	case v1.PodSucceeded:
		return 3
	case v1.PodFailed:
		return 4
	case v1.PodUnknown:
		return 5
	default:
		return 5
	}
}
