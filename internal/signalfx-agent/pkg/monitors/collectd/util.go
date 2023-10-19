package collectd

import (
	"path/filepath"
)

// MakePythonPluginPath takes file path components below the BundleDir for
// Python plugins and returns an os appropriate file path.
func MakePythonPluginPath(bundleDir string, components ...string) string {
	components = append([]string{bundleDir, "collectd-python"}, components...)
	return filepath.Join(components...)
}

// DefaultTypesDBPath returns the default types.db path based on the bundle dir
// envvar.
func DefaultTypesDBPath(bundleDir string) string {
	return filepath.Join(bundleDir, "types.db")
}
