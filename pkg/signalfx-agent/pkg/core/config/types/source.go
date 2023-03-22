package types

import "github.com/signalfx/signalfx-agent/pkg/core/config/validation"

// ConfigSourceConfig is a config type for a ConfigSource.  It can make new
// ConfigSource instances.
type ConfigSourceConfig interface {
	validation.Validatable
	New() (ConfigSource, error)
}

// ConfigSource represents a data store for which we can get and watch paths.
type ConfigSource interface {
	// Name should return the name used as the scheme of the URL that is
	// provided as the value of the "#from" key.
	Name() string
	// Get should return the content and version of the given path.  Path can
	// be globbed, but a backend may choose to return an error if globbing is
	// not supported.  Get should return an instance of ErrNotFound from this
	// package if the path does not exist in the source so that the agent can
	// distinguish that from other errors to allow for optional paths.
	Get(path string) (content map[string][]byte, version uint64, err error)
	// WaitForChange should accept a path and version and only return if either
	// the provided stop channel is closed, or if the path's content changes.
	// It should make every effort to not produce false positives.
	WaitForChange(path string, version uint64, stop <-chan struct{}) error
}

// ErrNotFound should be returned by config sources if a path does not exist,
// so that we can distinguish this from other errors in case the path is
// optional.
type ErrNotFound struct {
	msg string
}

// NewNotFoundError returns an instance of ErrNotFound with the given msg.
func NewNotFoundError(msg string) ErrNotFound {
	return ErrNotFound{
		msg: msg,
	}
}

func (e ErrNotFound) Error() string {
	return e.msg
}
