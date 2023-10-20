package subproc

import (
	"os"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
)

// DefaultPythonRuntimeConfig returns the runtime config that uses the bundled Python
// runtime.
func DefaultPythonRuntimeConfig(bundleDir string, pkgName string) *RuntimeConfig {
	// The PYTHONHOME envvar is set in agent core when config is processed.
	env := os.Environ()
	env = append(env, config.BundlePythonHomeEnvvar(bundleDir))

	return &RuntimeConfig{
		Binary: defaultPythonBinaryExecutable(bundleDir),
		Args:   defaultPythonBinaryArgs(pkgName),
		Env:    env,
	}
}
