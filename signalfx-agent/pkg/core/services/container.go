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
