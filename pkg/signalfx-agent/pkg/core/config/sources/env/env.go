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
