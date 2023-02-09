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

func datapointsForReplicaSet(rs *appsv1.ReplicaSet) []*datapoint.Datapoint {
	dimensions := map[string]string{
		"metric_source":        "kubernetes",
		"kubernetes_namespace": rs.Namespace,
		"kubernetes_uid":       string(rs.UID),
		"kubernetes_name":      rs.Name,
	}

	if rs.Spec.Replicas == nil { //|| rs.Status.AvailableReplicas == nil {
		return nil
	}
	return makeReplicaDPs("replica_set", dimensions,
		*rs.Spec.Replicas, rs.Status.AvailableReplicas)
}

func dimensionForReplicaSet(rs *appsv1.ReplicaSet) *atypes.Dimension {
	props, tags := k8sutil.PropsAndTagsFromLabels(rs.Labels)
	props["kubernetes_workload_name"] = rs.Name
	props["kubernetes_workload"] = "ReplicaSet"
	props["replicaset_creation_timestamp"] = rs.GetCreationTimestamp().Format(time.RFC3339)

	for _, or := range rs.OwnerReferences {
		props[utils.LowercaseFirstChar(or.Kind)] = or.Name
		props[utils.LowercaseFirstChar(or.Kind)+"_uid"] = string(or.UID)
	}

	return &atypes.Dimension{
		Name:       "kubernetes_uid",
		Value:      string(rs.UID),
		Properties: props,
		Tags:       tags,
	}
}
