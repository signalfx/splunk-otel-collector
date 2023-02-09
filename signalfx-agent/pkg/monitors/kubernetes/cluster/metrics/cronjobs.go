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

package metrics

import (
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	k8sutil "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes/utils"
	atypes "github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
)

func datapointsForCronJob(cj *batchv1beta1.CronJob) []*datapoint.Datapoint {
	dimensions := map[string]string{
		"metric_source":        "kubernetes",
		"kubernetes_namespace": cj.Namespace,
		"kubernetes_uid":       string(cj.UID),
		"kubernetes_name":      cj.Name,
	}

	return []*datapoint.Datapoint{
		datapoint.New(
			"kubernetes.cronjob.active",
			dimensions,
			datapoint.NewIntValue(int64(len(cj.Status.Active))),
			datapoint.Gauge,
			time.Time{}),
	}
}

func dimensionForCronJob(cj *batchv1beta1.CronJob) *atypes.Dimension {
	props, tags := k8sutil.PropsAndTagsFromLabels(cj.Labels)

	props["cronjob_creation_timestamp"] = cj.GetCreationTimestamp().Format(time.RFC3339)
	props["kubernetes_workload"] = "CronJob"
	props["kubernetes_workload_name"] = cj.Name
	props["schedule"] = cj.Spec.Schedule
	props["concurrency_policy"] = string(cj.Spec.ConcurrencyPolicy)

	for _, or := range cj.OwnerReferences {
		props[utils.LowercaseFirstChar(or.Kind)] = or.Name
		props[utils.LowercaseFirstChar(or.Kind)+"_uid"] = string(or.UID)
	}

	return &atypes.Dimension{
		Name:       "kubernetes_uid",
		Value:      string(cj.UID),
		Properties: props,
		Tags:       tags,
	}
}
