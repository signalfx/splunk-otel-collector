package collectd

import (
	"os"
	"path/filepath"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
)

// MakePythonPluginPath takes file path components below the BundleDir for
// Python plugins and returns an os appropriate file path.  The environment
// variable SIGNALFX_BUNDLE_DIR is used as the root of the path
func MakePythonPluginPath(components ...string) string {
	components = append([]string{os.Getenv(constants.BundleDirEnvVar), "collectd-python"}, components...)
	return filepath.Join(components...)
}

// DefaultTypesDBPath returns the default types.db path based on the bundle dir
// envvar.
func DefaultTypesDBPath() string {
	return filepath.Join(os.Getenv(constants.BundleDirEnvVar), "types.db")
}
