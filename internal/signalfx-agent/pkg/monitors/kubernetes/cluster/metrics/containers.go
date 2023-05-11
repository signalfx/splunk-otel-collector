package metrics

import (
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	atypes "github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/k8sutil"
	v1 "k8s.io/api/core/v1"
)

func datapointsForContainerStatus(cs v1.ContainerStatus, contDims map[string]string) []*datapoint.Datapoint {
	dps := []*datapoint.Datapoint{
		datapoint.New(
			"kubernetes.container_restart_count",
			contDims,
			datapoint.NewIntValue(int64(cs.RestartCount)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			"kubernetes.container_ready",
			contDims,
			datapoint.NewIntValue(int64(utils.BoolToInt(cs.Ready))),
			datapoint.Gauge,
			time.Time{}),
	}

	return dps
}

func datapointsForContainerSpec(c v1.Container, contDims map[string]string) []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint

	if val, ok := c.Resources.Requests[v1.ResourceCPU]; ok {
		dps = append(dps,
			datapoint.New(
				"kubernetes.container_cpu_request",
				contDims,
				datapoint.NewFloatValue(float64(val.MilliValue())/1000.0),
				datapoint.Gauge,
				time.Time{}))
	}

	if val, ok := c.Resources.Limits[v1.ResourceCPU]; ok {
		dps = append(dps,
			datapoint.New(
				"kubernetes.container_cpu_limit",
				contDims,
				datapoint.NewFloatValue(float64(val.MilliValue())/1000.0),
				datapoint.Gauge,
				time.Time{}))
	}

	if val, ok := c.Resources.Requests[v1.ResourceMemory]; ok {
		dps = append(dps,
			datapoint.New(
				"kubernetes.container_memory_request",
				contDims,
				datapoint.NewIntValue(val.Value()),
				datapoint.Gauge,
				time.Time{}))
	}

	if val, ok := c.Resources.Limits[v1.ResourceMemory]; ok {
		dps = append(dps,
			datapoint.New(
				"kubernetes.container_memory_limit",
				contDims,
				datapoint.NewIntValue(val.Value()),
				datapoint.Gauge,
				time.Time{}))
	}

	if val, ok := c.Resources.Requests[v1.ResourceEphemeralStorage]; ok {
		dps = append(dps,
			datapoint.New(
				"kubernetes.container_ephemeral_storage_request",
				contDims,
				datapoint.NewIntValue(val.Value()),
				datapoint.Gauge,
				time.Time{}))
	}

	if val, ok := c.Resources.Limits[v1.ResourceEphemeralStorage]; ok {
		dps = append(dps,
			datapoint.New(
				"kubernetes.container_ephemeral_storage_limit",
				contDims,
				datapoint.NewIntValue(val.Value()),
				datapoint.Gauge,
				time.Time{}))
	}

	return dps
}

func dimensionsForPodContainers(pod *v1.Pod) []*atypes.Dimension {
	var out []*atypes.Dimension
	for _, cs := range pod.Status.ContainerStatuses {
		// Do not send container id dimensions updates if container id
		// returned is empty
		if cs.ContainerID == "" {
			continue
		}
		out = append(out, dimensionForContainer(cs))
	}
	return out
}

func dimensionForContainer(cs v1.ContainerStatus) *atypes.Dimension {
	containerProps := make(map[string]string)

	if cs.State.Running != nil {
		containerProps["container_status"] = "running"
		containerProps["container_status_reason"] = ""
	}

	if cs.State.Terminated != nil {
		containerProps["container_status"] = "terminated"
		containerProps["container_status_reason"] = cs.State.Terminated.Reason
	}

	if cs.State.Waiting != nil {
		containerProps["container_status"] = "waiting"
		containerProps["container_status_reason"] = cs.State.Waiting.Reason
	}

	return &atypes.Dimension{
		Name:              "container_id",
		Value:             k8sutil.StripContainerID(cs.ContainerID),
		Properties:        containerProps,
		MergeIntoExisting: true,
	}
}

func getAllContainerDimensions(id string, name string, image string, dims map[string]string) map[string]string {
	out := utils.CloneStringMap(dims)

	out["container_id"] = k8sutil.StripContainerID(id)
	out["container_spec_name"] = name
	out["container_image"] = image

	return out
}
