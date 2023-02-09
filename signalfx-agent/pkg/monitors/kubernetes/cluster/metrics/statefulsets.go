// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// nolint: dupl
package metrics

import (
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	k8sutil "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/utils"
	atypes "github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
)

func datapointsForStatefulSet(ss *appsv1.StatefulSet) []*datapoint.Datapoint {
	dimensions := map[string]string{
		"metric_source":        "kubernetes",
		"kubernetes_namespace": ss.Namespace,
		"kubernetes_uid":       string(ss.UID),
		"kubernetes_name":      ss.Name,
	}

	if ss.Spec.Replicas == nil {
		return nil
	}

	return []*datapoint.Datapoint{
		datapoint.New(
			"kubernetes.stateful_set.desired",
			dimensions,
			datapoint.NewIntValue(int64(*ss.Spec.Replicas)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			"kubernetes.stateful_set.ready",
			dimensions,
			datapoint.NewIntValue(int64(ss.Status.ReadyReplicas)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			"kubernetes.stateful_set.current",
			dimensions,
			datapoint.NewIntValue(int64(ss.Status.CurrentReplicas)),
			datapoint.Gauge,
			time.Time{}),
		datapoint.New(
			"kubernetes.stateful_set.updated",
			dimensions,
			datapoint.NewIntValue(int64(ss.Status.UpdatedReplicas)),
			datapoint.Gauge,
			time.Time{}),
	}
}

func dimensionForStatefulSet(ss *appsv1.StatefulSet) *atypes.Dimension {
	props, tags := k8sutil.PropsAndTagsFromLabels(ss.Labels)
	props["kubernetes_workload"] = "StatefulSet"
	props["kubernetes_workload_name"] = ss.Name
	props["current_revision"] = ss.Status.CurrentRevision
	props["update_revision"] = ss.Status.UpdateRevision
	props["statefulset_creation_timestamp"] = ss.GetCreationTimestamp().Format(time.RFC3339)

	for _, or := range ss.OwnerReferences {
		props[utils.LowercaseFirstChar(or.Kind)] = or.Name
		props[utils.LowercaseFirstChar(or.Kind)+"_uid"] = string(or.UID)
	}

	return &atypes.Dimension{
		Name:       "kubernetes_uid",
		Value:      string(ss.UID),
		Properties: props,
		Tags:       tags,
	}
}
