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
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	quota "github.com/openshift/api/quota/v1"
	"github.com/signalfx/golib/v3/datapoint"
)

// If additional metrics are added probably need to add additional resources here.
var resources = []v1.ResourceName{
	v1.ResourceCPU,
	v1.ResourceMemory,
	v1.ResourcePods,
	v1.ResourceServices,
	v1.ResourcePersistentVolumeClaims,
	v1.ResourceServicesNodePorts,
	v1.ResourceServicesLoadBalancers,
}

func datapointsForClusterQuotas(quota *quota.ClusterResourceQuota) []*datapoint.Datapoint {
	dimensions := map[string]string{
		"metric_source":  "openshift",
		"kubernetes_uid": string(quota.UID),
		"quota_name":     quota.Name,
	}

	dps := buildDatapoints("openshift.clusterquota", dimensions, quota.Status.Total.Hard, quota.Status.Total.Used)

	for _, ns := range quota.Status.Namespaces {
		namespaceDims := map[string]string{
			"kubernetes_namespace": ns.Namespace,
		}
		for dim := range dimensions {
			namespaceDims[dim] = dimensions[dim]
		}
		dps = append(dps,
			buildDatapoints("openshift.appliedclusterquota", namespaceDims, ns.Status.Hard, ns.Status.Used)...)
	}

	return dps
}

func buildDatapoints(metricPrefix string, dimensions map[string]string,
	limit v1.ResourceList, used v1.ResourceList) []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint
	for _, resource := range resources {
		if quantity, ok := limit[resource]; ok {
			dps = append(dps,
				datapoint.New(
					fmt.Sprintf("%s.%s.limit", metricPrefix, resource),
					dimensions,
					datapoint.NewIntValue(quantity.Value()),
					datapoint.Gauge,
					time.Time{}))
		}

		if quantity, ok := used[resource]; ok {
			dps = append(dps,
				datapoint.New(
					fmt.Sprintf("%s.%s.used", metricPrefix, resource),
					dimensions,
					datapoint.NewIntValue(quantity.Value()),
					datapoint.Gauge,
					time.Time{}))
		}
	}
	return dps
}
