//go:build windows
// +build windows

package subproc

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
)

// The Windows specific process attributes
func procAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		//Pdeathsig: syscall.SIGTERM,
	}
}

func defaultPythonBinaryExecutable() string {
	return filepath.Join(os.Getenv(constants.BundleDirEnvVar), "python", "python.exe")
}

func defaultPythonBinaryArgs(pkgName string) []string {
	return []string{
		"-u",
		"-m",
		pkgName,
	}
}
