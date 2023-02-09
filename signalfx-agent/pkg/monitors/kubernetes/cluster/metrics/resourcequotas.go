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
	"strings"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	v1 "k8s.io/api/core/v1"
)

func datapointsForResourceQuota(rq *v1.ResourceQuota) []*datapoint.Datapoint {
	dps := []*datapoint.Datapoint{}

	for _, t := range []struct {
		typ string
		rl  v1.ResourceList
	}{
		{
			"hard",
			rq.Status.Hard,
		},
		{
			"used",
			rq.Status.Used,
		},
	} {
		for k, v := range t.rl {
			dims := map[string]string{
				"resource":             string(k),
				"quota_name":           rq.Name,
				"kubernetes_namespace": rq.Namespace,
			}

			val := v.Value()
			if strings.HasSuffix(string(k), ".cpu") {
				val = v.MilliValue()
			}

			dps = append(dps, sfxclient.Gauge("kubernetes.resource_quota_"+t.typ, dims, val))
		}
	}
	return dps
}
