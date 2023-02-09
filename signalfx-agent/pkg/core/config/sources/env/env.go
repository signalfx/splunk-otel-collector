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

package env

import (
	"os"

	"github.com/signalfx/signalfx-agent/pkg/core/config/types"
)

// Config for the file-based config source
type Config struct {
}

type envConfigSource struct{}

// New creates a new envvar remote config source from the target config
func (c *Config) New() (types.ConfigSource, error) {
	return New(), nil
}

// Validate the config
func (c *Config) Validate() error {
	return nil
}

var _ types.ConfigSourceConfig = &Config{}

// New makes a new fileConfigSource with the given config
func New() types.ConfigSource {
	return &envConfigSource{}
}

func (ecs *envConfigSource) Name() string {
	return "env"
}

func (ecs *envConfigSource) Get(path string) (map[string][]byte, uint64, error) {
	if value, ok := os.LookupEnv(path); ok {
		return map[string][]byte{path: []byte(value)}, 1, nil
	}

	return nil, 1, nil
}

// WaitForChange does nothing with envvars.  Technically they can change within
// the lifetime of the process but those changes are not picked up currently.
func (ecs *envConfigSource) WaitForChange(path string, version uint64, stop <-chan struct{}) error {
	<-stop
	return nil
}
