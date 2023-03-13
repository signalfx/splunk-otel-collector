package subproc

import (
	"os"
	"path/filepath"
)

// DefaultJavaRuntimeConfig returns the runtime config that uses the bundled Java
// runtime.
func DefaultJavaRuntimeConfig(jarPath string) *RuntimeConfig {
	// The JAVA_HOME envvar is set in agent core when config is processed.
	env := os.Environ()

	javaHome := os.Getenv("JAVA_HOME")

	return &RuntimeConfig{
		Binary: filepath.Join(javaHome, "bin/java"),
		Args: []string{
			// Enable assertions by default
			"-ea",
		},
		Env: env,
	}
}
