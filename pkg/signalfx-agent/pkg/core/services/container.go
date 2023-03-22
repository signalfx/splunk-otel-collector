package services

import (
	"fmt"
	"strings"
)

// Container information
type Container struct {
	// The ID of the container exposing the endpoint
	ID string `yaml:"container_id"`
	// A list of container names of the container exposing the endpoint
	Names []string `yaml:"container_names"`
	// The image name of the container exposing the endpoint
	Image string `yaml:"container_image"`
	// The command used when running the container exposing the endpoint
	Command string `yaml:"container_command"`
	// The container state, will usually be "running" since otherwise the
	// container wouldn't have a port exposed to be discovered.
	State string `yaml:"container_state"`
	// A map that contains container label key/value pairs. You can use the
	// `Contains` and `Get` helper functions in discovery rules to make use of
	// this. See [Endpoint
	// Discovery](../auto-discovery.md#additional-functions). For containers
	// managed by Kubernetes, this will be set to the pod's labels, as
	// individual containers do not have labels in Kubernetes proper.
	Labels map[string]string `yaml:"container_labels"`
}

// PrimaryName is the first container name, with all slashes stripped from the
// beginning.
func (c *Container) PrimaryName() string {
	if len(c.Names) > 0 {
		return strings.TrimLeft(c.Names[0], "/")
	}
	return ""
}

func (c *Container) String() string {
	return fmt.Sprintf("%#v", c)
}
