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
