//go:build darwin
// +build darwin

package subproc

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
)

// The Darwin specific process attribute that make the Python runner be in the
// same process group as the agent so they get shutdown together.
func procAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}

func defaultPythonBinaryExecutable() string {
	return filepath.Join(os.Getenv(constants.BundleDirEnvVar), "bin/python")
}

func defaultPythonBinaryArgs(pkgName string) []string {
	return []string{
		"-u",
		"-m",
		pkgName,
	}
}
