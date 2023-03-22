package docker

import (
	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
)

// FindHostMappedPort returns the port number of the docker port binding to the
// underlying host, or 0 if none exists.  It also returns the mapped ip that the
// port is bound to on the underlying host, or "" if none exists.
func FindHostMappedPort(cont *dtypes.ContainerJSON, exposedPort nat.Port) (int, string) {
	bindings := cont.NetworkSettings.Ports[exposedPort]

	for _, binding := range bindings {
		if port, err := nat.ParsePort(binding.HostPort); err == nil {
			return port, binding.HostIP
		}
	}
	return 0, ""
}
