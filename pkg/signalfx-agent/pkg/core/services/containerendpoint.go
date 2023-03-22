package services

import (
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// ContainerEndpoint contains information for single network endpoint of a
// discovered containerized service.  A single real-world service could have
// multiple distinct instances if it exposes multiple ports or is discovered by
// more than one observer.
type ContainerEndpoint struct {
	EndpointCore `yaml:",inline"`
	// Used for services that are accessed through some kind of
	// NAT redirection as Docker does.  This could be either the public port
	// or the private one.
	AltPort       uint16        `yaml:"alternate_port"`
	Container     Container     `yaml:",inline"`
	Orchestration Orchestration `yaml:",inline"`
	// A map of labels on the container port
	PortLabels map[string]string `yaml:"port_labels"`
}

// PublicPort is the port that the endpoint is accessed on externally.  It may
// be different from the PrivatePort.
func (ce *ContainerEndpoint) PublicPort() uint16 {
	if ce.Orchestration.PortPref == PUBLIC {
		return ce.Port
	}
	return ce.AltPort
}

// PrivatePort is the port that the service is configured to listen on
func (ce *ContainerEndpoint) PrivatePort() uint16 {
	if ce.Orchestration.PortPref == PRIVATE {
		return ce.Port
	}
	return ce.AltPort
}

// CONTAINER_ENDPOINT_VAR(container_name): The first and primary name of the container as
// it is known to the container runtime (e.g. Docker).

// CONTAINER_ENDPOINT_VAR(public_port): The port exposed outside the container

// CONTAINER_ENDPOINT_VAR(private_port): The port that the service endpoint runs on
// inside the container

// DerivedFields returns aliased and computed variable fields for this endpoint
func (ce *ContainerEndpoint) DerivedFields() map[string]interface{} {
	return utils.MergeInterfaceMaps(
		ce.EndpointCore.DerivedFields(),
		utils.StringMapToInterfaceMap(ce.Dimensions()),
		map[string]interface{}{
			"public_port":  ce.PublicPort(),
			"private_port": ce.PrivatePort(),
		})
}

// CONTAINER_DIMENSION(container_name): The primary name of the running
// container -- Docker containers can have multiple names but this will be the
// first name, if any.

// CONTAINER_DIMENSION(container_image): The image name (including tags) of the
// running container

// CONTAINER_DIMENSION(container_id): The container id of the container running
// this endpoint.

// Dimensions returns the dimensions associated with this endpoint
func (ce *ContainerEndpoint) Dimensions() map[string]string {
	return utils.MergeStringMaps(
		ce.EndpointCore.Dimensions(),
		utils.RemoveEmptyMapValues(map[string]string{
			"container_name":  ce.Container.PrimaryName(),
			"container_image": ce.Container.Image,
			"container_id":    ce.Container.ID,
		}))
}
