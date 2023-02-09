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

package kubernetes

import (
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// PodEndpoint is an endpoint that represents an entire pod and not a specific
// port in a container in the pod.
type PodEndpoint struct {
	services.EndpointCore `yaml:",inline"`
	Orchestration         services.Orchestration `yaml:",inline"`
}

// NewPodEndpoint makes a new PodEndpoint and sets the Target properly.
func NewPodEndpoint(ec *services.EndpointCore, orch *services.Orchestration) *PodEndpoint {
	ec.Target = services.TargetTypePod
	return &PodEndpoint{
		EndpointCore:  *ec,
		Orchestration: *orch,
	}
}

// DerivedFields returns aliased and computed variable fields for this endpoint
func (pe *PodEndpoint) DerivedFields() map[string]interface{} {
	return utils.MergeInterfaceMaps(
		pe.EndpointCore.DerivedFields(),
		utils.StringMapToInterfaceMap(pe.Dimensions()))
}
