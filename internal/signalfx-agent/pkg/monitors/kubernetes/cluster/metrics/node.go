package metrics

import (
	"fmt"
	"regexp"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/cluster/meta"
	k8sutil "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/utils"
	atypes "github.com/signalfx/signalfx-agent/pkg/monitors/types"
	v1 "k8s.io/api/core/v1"
)

var resourceNameToMetric = map[v1.ResourceName]string{
	v1.ResourceCPU:              meta.KubernetesNodeAllocatableCPU,
	v1.ResourceMemory:           meta.KubernetesNodeAllocatableMemory,
	v1.ResourceStorage:          meta.KubernetesNodeAllocatableStorage,
	v1.ResourceEphemeralStorage: meta.KubernetesNodeAllocatableEphemeralStorage,
}

func datapointsForNode(
	node *v1.Node,
	nodeConditionTypesToReport []string,
) []*datapoint.Datapoint {
	dims := map[string]string{
		"kubernetes_node":     node.Name,
		"kubernetes_node_uid": string(node.UID),
	}

	datapoints := make([]*datapoint.Datapoint, 0)
	for _, nodeConditionTypeValue := range nodeConditionTypesToReport {
		nodeConditionMetric := fmt.Sprintf("kubernetes.node_%s", strcase.ToSnake(nodeConditionTypeValue))
		v1NodeConditionTypeValue := v1.NodeConditionType(nodeConditionTypeValue)
		datapoints = append(
			datapoints,
			sfxclient.Gauge(
				nodeConditionMetric, dims, nodeConditionValue(node, v1NodeConditionTypeValue),
			),
		)
	}

	for _, res := range [4]v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory, v1.ResourceStorage, v1.ResourceEphemeralStorage} {
		quant, ok := node.Status.Allocatable[res]
		if ok {
			metric, ok := resourceNameToMetric[res]
			if !ok {
				panic("programmer forgot an entry")
			}

			datapoints = append(datapoints, sfxclient.GaugeF(metric, dims, float64(quant.MilliValue())/1000.0))
		}
	}
	return datapoints
}

func dimensionsForNode(node *v1.Node, updatesForNodeDimension bool) []*atypes.Dimension {
	var out []*atypes.Dimension
	props, tags := k8sutil.PropsAndTagsFromLabels(node.Labels)
	_ = getPropsFromTaints(node.Spec.Taints)

	props["node_creation_timestamp"] = node.GetCreationTimestamp().Format(time.RFC3339)

	if updatesForNodeDimension {
		propsCopy := make(map[string]string)
		for k, v := range props {
			propsCopy[k] = v
		}
		tagsCopy := make(map[string]bool)
		for k, v := range tags {
			tagsCopy[k] = v
		}
		out = append(out, &atypes.Dimension{
			Name:       "kubernetes_node",
			Value:      node.Name,
			Properties: propsCopy,
			Tags:       tagsCopy,
		})
	}

	props["kubernetes_node"] = node.Name
	out = append(out, &atypes.Dimension{
		Name:       "kubernetes_node_uid",
		Value:      string(node.UID),
		Properties: props,
		Tags:       tags,
	})

	return out
}

func getPropsFromTaints(taints []v1.Taint) map[string]string {
	unsupportedPattern := regexp.MustCompile("[^a-zA-Z0-9_-]")

	props := make(map[string]string)

	for _, t := range taints {
		keyValueCombo := "taint"
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

var nodeConditionValues = map[v1.ConditionStatus]int64{
	v1.ConditionTrue:    1,
	v1.ConditionFalse:   0,
	v1.ConditionUnknown: -1,
}

func nodeConditionValue(node *v1.Node, condType v1.NodeConditionType) int64 {
	status := v1.ConditionUnknown
	for _, c := range node.Status.Conditions {
		if c.Type == condType {
			status = c.Status
			break
		}
	}
	return nodeConditionValues[status]
}
