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
