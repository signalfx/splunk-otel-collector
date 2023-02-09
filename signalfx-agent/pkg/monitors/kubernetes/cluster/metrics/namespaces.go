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
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	v1 "k8s.io/api/core/v1"
)

func datapointsForNamespace(ns *v1.Namespace) []*datapoint.Datapoint {
	dims := map[string]string{
		"kubernetes_namespace": ns.Name,
	}

	return []*datapoint.Datapoint{
		sfxclient.Gauge("kubernetes.namespace_phase", dims, namespacePhaseValues[ns.Status.Phase]),
	}
}

var namespacePhaseValues = map[v1.NamespacePhase]int64{
	v1.NamespaceActive:      1,
	v1.NamespaceTerminating: 0,
	// If phase is blank for some reason, send as -1 for unknown
	v1.NamespacePhase(""): -1,
}
