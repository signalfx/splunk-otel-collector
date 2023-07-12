package subproc

import (
	"os"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
)

// DefaultPythonRuntimeConfig returns the runtime config that uses the bundled Python
// runtime.
func DefaultPythonRuntimeConfig(pkgName string) *RuntimeConfig {
	// The PYTHONHOME envvar is set in agent core when config is processed.
	env := os.Environ()
	env = append(env, config.BundlePythonHomeEnvvar())

	return &RuntimeConfig{
		Binary: defaultPythonBinaryExecutable(),
		Args:   defaultPythonBinaryArgs(pkgName),
		Env:    env,
	}
}
